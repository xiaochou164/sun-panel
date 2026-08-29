package mcp

import (
	"context"
	"encoding/json"

	"sun-panel/global"
	"sun-panel/models"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewServer 创建 Sun-Panel MCP server，暴露面板只读工具
func NewServer() *server.MCPServer {
	s := server.NewMCPServer(
		"sun-panel",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(false),
		server.WithLogging(),
	)

	// 注册只读工具
	s.AddTool(mcp.NewTool(
		"get_groups",
		mcp.WithDescription("获取面板图标分组列表（管理员数据）"),
	), handleGetGroups)

	s.AddTool(mcp.NewTool(
		"get_icons",
		mcp.WithDescription("获取图标列表，可按分组ID过滤"),
		mcp.WithString("groupId", mcp.Description("分组ID（可选）")),
	), handleGetIcons)

	s.AddTool(mcp.NewTool(
		"get_overview",
		mcp.WithDescription("获取面板概览（全部分组和图标）"),
	), handleGetOverview)

	s.AddTool(mcp.NewTool(
		"get_system_info",
		mcp.WithDescription("获取系统信息配置"),
	), handleGetSystemInfo)

	s.AddTool(mcp.NewTool(
		"get_files",
		mcp.WithDescription("获取文件列表"),
		mcp.WithNumber("limit", mcp.Description("条数，默认20，最大100")),
		mcp.WithNumber("offset", mcp.Description("偏移量，默认0")),
	), handleGetFiles)

	return s
}

// getAdminUserId 获取管理员用户ID
func getAdminUserId() (uint, bool) {
	userInfo := models.User{}
	if err := global.Db.Where("role=?", 1).Order("id asc").First(&userInfo).Error; err != nil {
		return 0, false
	}
	return userInfo.ID, true
}

func handleGetGroups(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userId, ok := getAdminUserId()
	if !ok {
		return mcp.NewToolResultError("no admin user found"), nil
	}
	groups := []models.ItemIconGroup{}
	if err := global.Db.Order("sort ,created_at").Where("user_id=?", userId).Find(&groups).Error; err != nil {
		return mcp.NewToolResultError("database error: " + err.Error()), nil
	}
	return mcp.NewToolResultJSON(groups)
}

func handleGetIcons(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userId, ok := getAdminUserId()
	if !ok {
		return mcp.NewToolResultError("no admin user found"), nil
	}
	args := req.GetArguments()
	groupId, _ := args["groupId"].(string)

	icons := []models.ItemIcon{}
	query := global.Db.Order("sort ,created_at").Where("user_id=?", userId)
	if groupId != "" {
		query = query.Where("item_icon_group_id=?", groupId)
	}
	if err := query.Find(&icons).Error; err != nil {
		return mcp.NewToolResultError("database error: " + err.Error()), nil
	}

	// 解析 icon 对象输出
	type iconOut struct {
		ID            uint                   `json:"id"`
		Title         string                 `json:"title"`
		Url           string                 `json:"url"`
		LanUrl        string                 `json:"lanUrl"`
		Description   string                 `json:"description"`
		OpenMethod    int                    `json:"openMethod"`
		Sort          int                    `json:"sort"`
		ItemIconGroupId int                  `json:"itemIconGroupId"`
		Icon          map[string]interface{} `json:"icon"`
	}
	result := make([]iconOut, 0, len(icons))
	for _, icon := range icons {
		iconMap := map[string]interface{}{}
		if icon.IconJson != "" {
			_ = json.Unmarshal([]byte(icon.IconJson), &iconMap)
		}
		result = append(result, iconOut{
			ID:              icon.ID,
			Title:           icon.Title,
			Url:             icon.Url,
			LanUrl:          icon.LanUrl,
			Description:     icon.Description,
			OpenMethod:      icon.OpenMethod,
			Sort:            icon.Sort,
			ItemIconGroupId: icon.ItemIconGroupId,
			Icon:            iconMap,
		})
	}
	return mcp.NewToolResultJSON(result)
}

func handleGetOverview(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userId, ok := getAdminUserId()
	if !ok {
		return mcp.NewToolResultError("no admin user found"), nil
	}
	groups := []models.ItemIconGroup{}
	if err := global.Db.Order("sort ,created_at").Where("user_id=?", userId).Find(&groups).Error; err != nil {
		return mcp.NewToolResultError("database error: " + err.Error()), nil
	}
	icons := []models.ItemIcon{}
	if err := global.Db.Order("sort ,created_at").Where("user_id=?", userId).Find(&icons).Error; err != nil {
		return mcp.NewToolResultError("database error: " + err.Error()), nil
	}
	type iconOut struct {
		ID              uint                   `json:"id"`
		Title           string                 `json:"title"`
		Url             string                 `json:"url"`
		LanUrl          string                 `json:"lanUrl"`
		Description     string                 `json:"description"`
		OpenMethod      int                    `json:"openMethod"`
		Sort            int                    `json:"sort"`
		ItemIconGroupId int                    `json:"itemIconGroupId"`
		Icon            map[string]interface{} `json:"icon"`
	}
	result := make([]iconOut, 0, len(icons))
	for _, icon := range icons {
		iconMap := map[string]interface{}{}
		if icon.IconJson != "" {
			_ = json.Unmarshal([]byte(icon.IconJson), &iconMap)
		}
		result = append(result, iconOut{
			ID:              icon.ID,
			Title:           icon.Title,
			Url:             icon.Url,
			LanUrl:          icon.LanUrl,
			Description:     icon.Description,
			OpenMethod:      icon.OpenMethod,
			Sort:            icon.Sort,
			ItemIconGroupId: icon.ItemIconGroupId,
			Icon:            iconMap,
		})
	}
	return mcp.NewToolResultJSON(map[string]interface{}{
		"groups": groups,
		"icons":  result,
	})
}

func handleGetSystemInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	info := map[string]interface{}{}
	if err := global.SystemSetting.GetValueByInterface("system_application", &info); err != nil {
		// 忽略
	}
	return mcp.NewToolResultJSON(info)
}

func handleGetFiles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userId, ok := getAdminUserId()
	if !ok {
		return mcp.NewToolResultError("no admin user found"), nil
	}
	args := req.GetArguments()
	limit := 20
	offset := 0
	if v, ok := args["limit"].(float64); ok && v > 0 && v <= 100 {
		limit = int(v)
	}
	if v, ok := args["offset"].(float64); ok && v >= 0 {
		offset = int(v)
	}
	files := []models.File{}
	var count int64
	if err := global.Db.Model(&models.File{}).Where("user_id=?", userId).Count(&count).Error; err != nil {
		return mcp.NewToolResultError("database error: " + err.Error()), nil
	}
	if err := global.Db.Where("user_id=?", userId).Order("created_at desc").Limit(limit).Offset(offset).Find(&files).Error; err != nil {
		return mcp.NewToolResultError("database error: " + err.Error()), nil
	}
	return mcp.NewToolResultJSON(map[string]interface{}{
		"list":  files,
		"count": count,
	})
}
