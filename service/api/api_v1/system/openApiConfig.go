package system

import (
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/lib/cmn/systemSetting"

	"github.com/gin-gonic/gin"
)

// OpenApiConfigApi OpenAPI 配置管理（仅管理员）
type OpenApiConfigApi struct {
}

// GetConfig 获取 OpenAPI 配置（key 脱敏）
func (a *OpenApiConfigApi) GetConfig(c *gin.Context) {
	keys, err := systemSetting.GetApiKeys()
	if err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	// 脱敏：只显示前 8 + 后 4 位
	sanitized := make([]map[string]interface{}, 0, len(keys))
	for _, k := range keys {
		sanitized = append(sanitized, map[string]interface{}{
			"name":        k.Name,
			"keyMasked":   maskKey(k.Key),
			"createdAt":   k.CreatedAt,
			"lastUsedAt":  k.LastUsedAt,
			"description": k.Description,
		})
	}
	apiReturn.SuccessData(c, gin.H{
		"enabled": systemSetting.IsOpenApiEnabled(),
		"keys":    sanitized,
	})
}

// SetEnabled 启用/停用 OpenAPI
func (a *OpenApiConfigApi) SetEnabled(c *gin.Context) {
	req := struct {
		Enabled bool `json:"enabled"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	if err := systemSetting.SetOpenApiEnabled(req.Enabled); err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	apiReturn.Success(c)
}

// CreateKey 创建 API Key
func (a *OpenApiConfigApi) CreateKey(c *gin.Context) {
	req := struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	if req.Name == "" {
		req.Name = "default"
	}
	key, err := systemSetting.AddApiKey(req.Name, req.Description)
	if err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	apiReturn.SuccessData(c, gin.H{
		"key": key, // 明文只此一次
	})
}

// DeleteKey 删除 API Key
func (a *OpenApiConfigApi) DeleteKey(c *gin.Context) {
	req := struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	if req.Key == "" && req.Name == "" {
		apiReturn.ErrorParamFomat(c, "key or name is required")
		return
	}
	var err error
	if req.Key != "" {
		err = systemSetting.DeleteApiKey(req.Key)
	} else {
		err = systemSetting.DeleteApiKeyByName(req.Name)
	}
	if err != nil {
		apiReturn.Error(c, err.Error())
		return
	}
	apiReturn.Success(c)
}

// maskKey 脱敏 key 显示
func maskKey(key string) string {
	if len(key) <= 12 {
		return "****"
	}
	return key[:8] + "****" + key[len(key)-4:]
}
