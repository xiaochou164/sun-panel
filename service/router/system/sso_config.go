package system

import (
	"sun-panel/api/api_v1"
	"sun-panel/api/api_v1/middleware"

	"github.com/gin-gonic/gin"
)

// InitSsoConfigRouter admin SSO configs
func InitSsoConfigRouter(router *gin.RouterGroup) {
	api := api_v1.ApiGroupApp.ApiSystem.SsoConfigApi

	// TODO: verify admin role using middleware, keeping it within LoginInterceptor for now as it's the standard
	r := router.Group("/system/ssoConfig", middleware.LoginInterceptor)
	r.POST("/getList", api.GetList)
	r.POST("/save", api.Save)
}
