package system

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/lib/cmn"
	"sun-panel/models"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type SsoApi struct{}

// GetProviders 获取启用的SSO提供商列表(用于登录页)
func (a *SsoApi) GetProviders(c *gin.Context) {
	mSsoConfig := models.SsoConfig{}
	configs, err := mSsoConfig.GetEnabledProviders()
	if err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	var safeConfigs []map[string]interface{}
	for _, config := range configs {
		safeConfigs = append(safeConfigs, map[string]interface{}{
			"provider": config.Provider,
			"name":     config.Name,
		})
	}
	apiReturn.SuccessData(c, safeConfigs)
}

func getOauth2Config(c *gin.Context, provider string) (*oauth2.Config, *models.SsoConfig, error) {
	config := &models.SsoConfig{}
	err := models.Db.Where("provider = ? AND enabled = 1", provider).First(config).Error
	if err != nil {
		return nil, nil, fmt.Errorf("provider not found or disabled")
	}

	scheme := "http://"
	if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https://"
	}

	host := c.Request.Host
	// In local dev, map 3002 to 1002 for the callback as well
	host = strings.Replace(host, ":3002", ":1002", 1)

	redirectUrl := scheme + host + "/api/system/sso/callback/" + provider

	switch provider {
	case "github":
		return &oauth2.Config{
			ClientID:     config.ClientId,
			ClientSecret: config.ClientSecret,
			Endpoint:     github.Endpoint,
			RedirectURL:  redirectUrl,
			Scopes:       []string{"read:user", "user:email"},
		}, config, nil
	case "google":
		return &oauth2.Config{
			ClientID:     config.ClientId,
			ClientSecret: config.ClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  redirectUrl,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		}, config, nil
	default:
		// Generic OIDC fallback
		providerOIDC, err := oidc.NewProvider(context.Background(), config.IssuerUrl)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to init OIDC provider: %v", err)
		}
		return &oauth2.Config{
			ClientID:     config.ClientId,
			ClientSecret: config.ClientSecret,
			Endpoint:     providerOIDC.Endpoint(),
			RedirectURL:  redirectUrl,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		}, config, nil
	}
}

// Login 重定向到SSO提供商登录页
func (a *SsoApi) Login(c *gin.Context) {
	provider := c.Param("provider")
	userToken := c.Query("token")

	oauthCfg, _, err := getOauth2Config(c, provider)
	if err != nil {
		redirectFrontend(c, "", err.Error())
		return
	}

	state := cmn.BuildRandCode(32, cmn.RAND_CODE_MODE2)
	if userToken != "" {
		state = state + "||" + userToken
	}

	global.VerifyCodeCachePool.SetDefault(state, "1")

	url := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}

func redirectFrontend(c *gin.Context, token string, errMsg string) {
	scheme := "http://"
	if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https://"
	}

	host := c.Request.Host
	// In local dev, Vite runs on 1002 while backend API runs on 3002.
	// We remap the redirect host for convenience during development.
	host = strings.Replace(host, ":3002", ":1002", 1)

	url := scheme + host + "/#/login"

	params := []string{}
	if token != "" {
		params = append(params, "ssoToken="+token)
	}
	if errMsg != "" {
		params = append(params, "ssoError="+errMsg)
	}

	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	c.Redirect(http.StatusFound, url)
}

// Callback SSO回调处理
func (a *SsoApi) Callback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.String(http.StatusBadRequest, "Missing code or state")
		return
	}

	if _, found := global.VerifyCodeCachePool.Get(state); !found {
		c.String(http.StatusBadRequest, "Invalid or expired state")
		return
	}
	global.VerifyCodeCachePool.Delete(state)

	var userToken string
	stateParts := strings.Split(state, "||")
	if len(stateParts) > 1 {
		userToken = stateParts[1]
	}

	oauthCfg, ssoConfig, err := getOauth2Config(c, provider)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	token, err := oauthCfg.Exchange(context.Background(), code)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to exchange token: "+err.Error())
		return
	}

	var providerUid string
	var email string
	var name string

	if provider == "github" {
		client := oauthCfg.Client(context.Background(), token)
		userResp, err := client.Get("https://api.github.com/user")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to get user info: "+err.Error())
			return
		}
		defer userResp.Body.Close()
		var user map[string]interface{}
		json.NewDecoder(userResp.Body).Decode(&user)
		if id, ok := user["id"].(float64); ok {
			providerUid = fmt.Sprintf("%.0f", id)
		}
		if e, ok := user["email"].(string); ok {
			email = e
		}
		if n, ok := user["name"].(string); ok {
			name = n
		}
	} else {
		providerOIDC, err := oidc.NewProvider(context.Background(), ssoConfig.IssuerUrl)
		if provider == "google" {
			providerOIDC, _ = oidc.NewProvider(context.Background(), "https://accounts.google.com")
		}
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to init OIDC provider: "+err.Error())
			return
		}
		verifier := providerOIDC.Verifier(&oidc.Config{ClientID: ssoConfig.ClientId})

		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			userInfo, err := providerOIDC.UserInfo(context.Background(), oauth2.StaticTokenSource(token))
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to get user info: "+err.Error())
				return
			}
			providerUid = userInfo.Subject
			var claims struct {
				Email string `json:"email"`
				Name  string `json:"name"`
			}
			userInfo.Claims(&claims)
			email = claims.Email
			name = claims.Name
		} else {
			idToken, err := verifier.Verify(context.Background(), rawIDToken)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to verify ID token: "+err.Error())
				return
			}
			var claims struct {
				Email string `json:"email"`
				Name  string `json:"name"`
			}
			idToken.Claims(&claims)
			providerUid = idToken.Subject
			email = claims.Email
			name = claims.Name
		}
	}

	if providerUid == "" {
		c.String(http.StatusInternalServerError, "Could not identify user from provider")
		return
	}

	mUserAuth := models.UserAuth{}
	authRec, _ := mUserAuth.GetByProviderAndUid(provider, providerUid)
	mUser := models.User{}

	if userToken != "" {
		dbToken, ok := global.CUserToken.Get(userToken)
		if !ok || dbToken == "" {
			redirectFrontend(c, "", "Session expired, please login again before binding")
			return
		}

		userInfo, err := mUser.GetUserInfoByToken(dbToken)
		if err != nil {
			redirectFrontend(c, "", "Invalid binding user")
			return
		}
		if authRec.ID != 0 {
			if authRec.UserId == userInfo.ID {
				redirectFrontend(c, "", "Already bound")
				return
			} else {
				redirectFrontend(c, "", "This account is already bound to another user")
				return
			}
		}

		newAuth := models.UserAuth{
			UserId:      userInfo.ID,
			Provider:    provider,
			ProviderUid: providerUid,
		}
		models.Db.Create(&newAuth)
		redirectFrontend(c, "", "Bind success")
		return
	}

	var loginUser models.User
	if authRec.ID != 0 {
		loginUser, err = mUser.GetUserInfoByUid(authRec.UserId)
		if err != nil || loginUser.Status != 1 {
			redirectFrontend(c, "", "User account is disabled or does not exist")
			return
		}
	} else {
		username := email
		if username == "" {
			username = providerUid + "@" + provider
		}
		loginUser = models.User{
			Username: username,
			Password: cmn.PasswordEncryption(cmn.BuildRandCode(12, cmn.RAND_CODE_MODE2)),
			Name:     name,
			Mail:     email,
			Status:   1,
			Role:     2,
		}
		if name == "" {
			loginUser.Name = loginUser.Username
		}
		for i := 0; i < 10; i++ {
			if _, err := mUser.CheckUsernameExist(loginUser.Username); err == nil {
				break
			}
			loginUser.Username = username + fmt.Sprintf("%d", i)
		}

		err = models.Db.Create(&loginUser).Error
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to create user: "+err.Error())
			return
		}

		newAuth := models.UserAuth{
			UserId:      loginUser.ID,
			Provider:    provider,
			ProviderUid: providerUid,
		}
		models.Db.Create(&newAuth)
	}

	bToken := loginUser.Token
	if bToken == "" {
		bToken = cmn.BuildRandCode(32, cmn.RAND_CODE_MODE2)
		models.Db.Model(&loginUser).Update("token", bToken)
	}
	cToken := uuid.NewString() + "-" + cmn.Md5(cmn.Md5("userId"+strconv.Itoa(int(loginUser.ID))))
	global.CUserToken.SetDefault(cToken, bToken)

	redirectFrontend(c, cToken, "")
}

// GetUserBindings 获取当前用户绑定的SSO账号
func (a *SsoApi) GetUserBindings(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	mUserAuth := models.UserAuth{}
	auths, err := mUserAuth.GetListByUserId(userInfo.ID)
	if err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	var bindList []map[string]interface{}
	for _, auth := range auths {
		bindList = append(bindList, map[string]interface{}{
			"provider":    auth.Provider,
			"providerUid": auth.ProviderUid,
			"createdAt":   auth.CreatedAt,
		})
	}
	apiReturn.SuccessData(c, bindList)
}

// Unbind 解绑SSO账号
func (a *SsoApi) Unbind(c *gin.Context) {
	type UnbindReq struct {
		Provider string `json:"provider" validate:"required"`
	}
	req := UnbindReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	userInfo, _ := base.GetCurrentUserInfo(c)
	err := models.Db.Where("user_id = ? AND provider = ?", userInfo.ID, req.Provider).Delete(&models.UserAuth{}).Error
	if err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	apiReturn.Success(c)
}
