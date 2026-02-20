package models

import (
	"time"
)

// UserAuth 用户授权/SSO绑定表
type UserAuth struct {
	BaseModel
	UserId      uint      `gorm:"index;type:int(11)" json:"userId"`
	Provider    string    `gorm:"index;type:varchar(50)" json:"provider"` // e.g. "google", "github", "saml"
	ProviderUid string    `gorm:"index;type:varchar(255)" json:"providerUid"`
	CreatedAt   time.Time `json:"createdAt"`
}

// GetByProviderAndUid 根据提供商和提供商用户ID查询记录
func (m *UserAuth) GetByProviderAndUid(provider, providerUid string) (*UserAuth, error) {
	auth := UserAuth{}
	err := Db.Where("provider = ? AND provider_uid = ?", provider, providerUid).First(&auth).Error
	return &auth, err
}

// GetListByUserId 获取用户的绑定列表
func (m *UserAuth) GetListByUserId(userId uint) ([]UserAuth, error) {
	var auths []UserAuth
	err := Db.Where("user_id = ?", userId).Find(&auths).Error
	return auths, err
}
