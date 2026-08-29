package openness

import (
	"sun-panel/api/api_v1"
	"sun-panel/api/api_v1/middleware"

	"github.com/gin-gonic/gin"
)

func InitOpenness(router *gin.RouterGroup) {
	api := api_v1.ApiGroupApp.ApiOpen.Openness
	openApi := api_v1.ApiGroupApp.ApiOpen.OpenApi
	{
		router.GET("loginConfig", api.LoginConfig)
		router.GET("getDisclaimer", api.GetDisclaimer)
		router.GET("getAboutDescription", api.GetAboutDescription)
	}

	// OpenAPI 资源端点（需 API Key 鉴权）
	openApiGroup := router.Group("openapi", middleware.OpenApiInterceptor)
	{
		openApiGroup.GET("groups", openApi.GetGroups)
		openApiGroup.GET("icons", openApi.GetIcons)
		openApiGroup.GET("files", openApi.GetFiles)
		openApiGroup.GET("systemInfo", openApi.GetSystemInfo)
		openApiGroup.GET("overview", openApi.GetOverview)
	}
}
