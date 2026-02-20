package system

import (
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/models"

	"github.com/gin-gonic/gin"
)

type SsoConfigApi struct{}

// GetList 获取所有配置
func (a *SsoConfigApi) GetList(c *gin.Context) {
	var configs []models.SsoConfig
	models.Db.Find(&configs)
	apiReturn.SuccessData(c, configs)
}

// Save 保存配置
func (a *SsoConfigApi) Save(c *gin.Context) {
	var req models.SsoConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	if req.Provider == "" {
		apiReturn.ErrorParamFomat(c, "provider is required")
		return
	}

	var existing models.SsoConfig
	if err := models.Db.Where("provider = ?", req.Provider).First(&existing).Error; err != nil {
		// create
		models.Db.Create(&req)
	} else {
		// update
		models.Db.Model(&existing).Select("enabled", "name", "client_id", "client_secret", "issuer_url", "saml_metadata", "ext").Updates(req)
	}

	apiReturn.Success(c)
}
