package system

import (
	"sun-panel/api/api_v1"
	"sun-panel/api/api_v1/middleware"

	"github.com/gin-gonic/gin"
)

// InitOpenApiConfigRouter OpenAPI 配置管理路由（管理员）
func InitOpenApiConfigRouter(router *gin.RouterGroup) {
	api := api_v1.ApiGroupApp.ApiSystem.OpenApiConfigApi
	r := router.Group("/system/openApiConfig", middleware.LoginInterceptor, middleware.AdminInterceptor)
	{
		r.POST("/getConfig", api.GetConfig)
		r.POST("/setEnabled", api.SetEnabled)
		r.POST("/createKey", api.CreateKey)
		r.POST("/deleteKey", api.DeleteKey)
	}
}
