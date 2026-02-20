package system

import (
	"sun-panel/api/api_v1"
	"sun-panel/api/api_v1/middleware"

	"github.com/gin-gonic/gin"
)

func InitSsoRouter(router *gin.RouterGroup) {
	api := api_v1.ApiGroupApp.ApiSystem.SsoApi

	// 公开(免登录)获取SSO登录选项及回调逻辑
	routerPublic := router.Group("/system/sso")
	routerPublic.GET("/providers", api.GetProviders)
	routerPublic.GET("/login/:provider", api.Login)
	routerPublic.GET("/callback/:provider", api.Callback)

	// 需要登录后操作
	r := router.Group("/system/sso", middleware.LoginInterceptor)
	r.POST("/getUserBindings", api.GetUserBindings)
	r.POST("/unbind", api.Unbind)
}
