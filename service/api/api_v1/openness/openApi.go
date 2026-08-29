package openness

import (
	"encoding/json"
	"strconv"

	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/global"
	"sun-panel/models"

	"github.com/gin-gonic/gin"
)

// OpenApi OpenAPI 只读资源端点
type OpenApi struct {
}

// getAdminUserId 获取管理员用户ID（OpenAPI 数据固定以管理员视角）
func getAdminUserId(c *gin.Context) (uint, bool) {
	userInfo := models.User{}
	if err := global.Db.Where("role=?", 1).Order("id asc").First(&userInfo).Error; err != nil {
		apiReturn.Error(c, "no admin user found")
		return 0, false
	}
	return userInfo.ID, true
}

// iconWithParsed 图标+解析后的 icon 对象
type iconWithParsed struct {
	models.ItemIcon
	Icon map[string]interface{} `json:"icon"`
}

// parseIconJson 解析 IconJson 到 icon 对象
func parseIconJson(iconJson string) map[string]interface{} {
	m := map[string]interface{}{}
	if iconJson != "" {
		_ = json.Unmarshal([]byte(iconJson), &m)
	}
	return m
}

// GetGroups 获取图标分组列表
func (a *OpenApi) GetGroups(c *gin.Context) {
	userId, ok := getAdminUserId(c)
	if !ok {
		return
	}
	groups := []models.ItemIconGroup{}
	if err := global.Db.Order("sort ,created_at").Where("user_id=?", userId).Find(&groups).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	apiReturn.SuccessListData(c, groups, int64(len(groups)))
}

// GetIcons 获取图标列表，支持按分组过滤 ?groupId=<id>
func (a *OpenApi) GetIcons(c *gin.Context) {
	userId, ok := getAdminUserId(c)
	if !ok {
		return
	}
	groupId := c.Query("groupId")

	icons := []models.ItemIcon{}
	query := global.Db.Order("sort ,created_at").Where("user_id=?", userId)
	if groupId != "" {
		query = query.Where("item_icon_group_id=?", groupId)
	}
	if err := query.Find(&icons).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	// 解析 icon 对象输出
	result := make([]iconWithParsed, 0, len(icons))
	for _, icon := range icons {
		result = append(result, iconWithParsed{
			ItemIcon: icon,
			Icon:     parseIconJson(icon.IconJson),
		})
	}
	apiReturn.SuccessListData(c, result, int64(len(result)))
}

// GetSystemInfo 获取系统信息
func (a *OpenApi) GetSystemInfo(c *gin.Context) {
	info := gin.H{}
	if err := global.SystemSetting.GetValueByInterface("system_application", &info); err != nil {
		// 忽略，返回默认
	}
	apiReturn.SuccessData(c, info)
}

// GetFiles 获取文件列表 ?limit=20&offset=0
func (a *OpenApi) GetFiles(c *gin.Context) {
	userId, ok := getAdminUserId(c)
	if !ok {
		return
	}
	limit := 20
	offset := 0
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	files := []models.File{}
	var count int64
	if err := global.Db.Model(&models.File{}).Where("user_id=?", userId).Count(&count).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	if err := global.Db.Where("user_id=?", userId).Order("created_at desc").Limit(limit).Offset(offset).Find(&files).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	apiReturn.SuccessListData(c, files, count)
}

// GetOverview 获取面板概览（分组+图标聚合）
func (a *OpenApi) GetOverview(c *gin.Context) {
	userId, ok := getAdminUserId(c)
	if !ok {
		return
	}
	groups := []models.ItemIconGroup{}
	if err := global.Db.Order("sort ,created_at").Where("user_id=?", userId).Find(&groups).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	icons := []models.ItemIcon{}
	if err := global.Db.Order("sort ,created_at").Where("user_id=?", userId).Find(&icons).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	result := make([]iconWithParsed, 0, len(icons))
	for _, icon := range icons {
		result = append(result, iconWithParsed{
			ItemIcon: icon,
			Icon:     parseIconJson(icon.IconJson),
		})
	}
	apiReturn.SuccessData(c, gin.H{
		"groups": groups,
		"icons":  result,
	})
}
