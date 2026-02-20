package models

// SsoConfig SSO配置表
type SsoConfig struct {
	BaseModel
	Provider     string `gorm:"uniqueIndex;type:varchar(50)" json:"provider" validate:"required"` // e.g. "github", "google", "oidc", "saml"
	Enabled      int    `gorm:"type:tinyint(1);default:0" json:"enabled"`                         // 1.启用 0.停用
	Name         string `gorm:"type:varchar(50)" json:"name"`                                     // Display name
	ClientId     string `gorm:"type:varchar(255)" json:"clientId"`
	ClientSecret string `gorm:"type:varchar(255)" json:"clientSecret"`
	IssuerUrl    string `gorm:"type:varchar(255)" json:"issuerUrl"` // For OIDC
	SamlMetadata string `gorm:"type:text" json:"samlMetadata"`      // For SAML metadata URL or XML
	Ext          string `gorm:"type:text" json:"ext"`               // Extended config JSON
}

// GetEnabledProviders 获取所有启用的SSO提供商配置
func (m *SsoConfig) GetEnabledProviders() ([]SsoConfig, error) {
	var configs []SsoConfig
	err := Db.Where("enabled = ?", 1).Find(&configs).Error
	return configs, err
}

// GetByProvider 根据提供商名称获取配置
func (m *SsoConfig) GetByProvider(provider string) (*SsoConfig, error) {
	config := SsoConfig{}
	err := Db.Where("provider = ?", provider).First(&config).Error
	return &config, err
}
