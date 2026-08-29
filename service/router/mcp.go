package router

import (
	"sun-panel/api/api_v1/middleware"
	"sun-panel/global"
	"sun-panel/mcp"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
)

// InitMcpRouter 挂载 MCP Streamable HTTP 端点到 /mcp
// 鉴权使用 OpenAPI 中间件（Authorization: Bearer <api_key>）
func InitMcpRouter(router *gin.Engine) {
	// 创建 MCP server
	mcpServer := mcp.NewServer()
	httpServer := server.NewStreamableHTTPServer(mcpServer, server.WithEndpointPath("/mcp"))

	// 全局中间件：OpenAPI 鉴权
	mcpRouter := router.Group("/mcp", middleware.OpenApiInterceptor)

	// Streamable HTTP: POST /mcp (session), GET /mcp (SSE), DELETE /mcp (close)
	mcpRouter.Any("", func(c *gin.Context) {
		httpServer.ServeHTTP(c.Writer, c.Request)
	})
	mcpRouter.Any("/", func(c *gin.Context) {
		httpServer.ServeHTTP(c.Writer, c.Request)
	})

	global.Logger.Info("MCP Streamable HTTP server mounted at /mcp")
}
