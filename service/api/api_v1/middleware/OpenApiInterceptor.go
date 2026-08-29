package middleware

import (
	"strings"

	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/lib/cmn/systemSetting"

	"github.com/gin-gonic/gin"
)

// OpenApiInterceptor OpenAPI 鉴权中间件
// 支持 Authorization: Bearer <key> 或 X-API-Key: <key>
func OpenApiInterceptor(c *gin.Context) {
	// 检查是否启用
	if !systemSetting.IsOpenApiEnabled() {
		apiReturn.Error(c, "OpenAPI is disabled")
		c.Abort()
		return
	}

	apiKey := ""
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Bearer token
		if strings.HasPrefix(authHeader, "Bearer ") {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			apiKey = authHeader
		}
	}
	if apiKey == "" {
		apiKey = c.GetHeader("X-API-Key")
	}

	if apiKey == "" {
		apiReturn.ErrorByCode(c, 1000)
		c.Abort()
		return
	}

	keyInfo, err := systemSetting.ValidateApiKey(apiKey)
	if err != nil || keyInfo == nil {
		apiReturn.Error(c, "Invalid API key")
		c.Abort()
		return
	}

	c.Set("apiKeyInfo", keyInfo)
	c.Next()
}
