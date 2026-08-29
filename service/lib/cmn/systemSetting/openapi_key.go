package systemSetting

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"sun-panel/models"
)

// 获取所有 API Key（明文，内部使用）
func GetApiKeys() ([]ApiKey, error) {
	keys := []ApiKey{}
	mSetting := models.SystemSetting{}
	val, err := mSetting.Get(OPENAPI_KEYS)
	if err != nil {
		// 不存在则返回空
		return keys, nil
	}
	if val == "" {
		return keys, nil
	}
	if err := json.Unmarshal([]byte(val), &keys); err != nil {
		return keys, nil
	}
	return keys, nil
}

// 是否启用 OpenAPI
func IsOpenApiEnabled() bool {
	mSetting := models.SystemSetting{}
	val, err := mSetting.Get(OPENAPI_ENABLED)
	if err != nil || val == "" {
		return false
	}
	return val == "true"
}

// 生成一个 API Key（前缀 sp_ + 32位hex）
func GenerateApiKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "sp_" + hex.EncodeToString(b), nil
}

// 校验 API Key 是否有效，返回 key 信息
func ValidateApiKey(rawKey string) (*ApiKey, error) {
	if rawKey == "" {
		return nil, errors.New("empty api key")
	}
	keys, err := GetApiKeys()
	if err != nil {
		return nil, err
	}
	for i := range keys {
		if keys[i].Key == rawKey {
			// 更新最后使用时间（不阻塞主流程，忽略错误）
			keys[i].LastUsedAt = time.Now().Format("2006-01-02 15:04:05")
			_ = SaveApiKeys(keys)
			return &keys[i], nil
		}
	}
	return nil, errors.New("invalid api key")
}

// 添加 API Key
func AddApiKey(name string, description string) (string, error) {
	key, err := GenerateApiKey()
	if err != nil {
		return "", err
	}
	keys, _ := GetApiKeys()
	keys = append(keys, ApiKey{
		Name:        name,
		Key:         key,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
		LastUsedAt:  "",
		Description: description,
	})
	if err := SaveApiKeys(keys); err != nil {
		return "", err
	}
	return key, nil
}

// 删除 API Key（按完整 key 匹配，供 API 调用）
func DeleteApiKey(rawKey string) error {
	return deleteApiKey(func(k ApiKey) bool { return k.Key == rawKey })
}

// DeleteApiKeyByName 按名称删除 API Key，供后台页面使用（页面不接触完整 key）
func DeleteApiKeyByName(name string) error {
	if name == "" {
		return errors.New("api key name is required")
	}
	return deleteApiKey(func(k ApiKey) bool { return k.Name == name })
}

func deleteApiKey(match func(ApiKey) bool) error {
	keys, err := GetApiKeys()
	if err != nil {
		return err
	}
	newKeys := []ApiKey{}
	found := false
	for _, k := range keys {
		if match(k) && !found {
			found = true
			continue
		}
		newKeys = append(newKeys, k)
	}
	if !found {
		return errors.New("api key not found")
	}
	return SaveApiKeys(newKeys)
}

// 设置是否启用
func SetOpenApiEnabled(enabled bool) error {
	mSetting := models.SystemSetting{}
	if enabled {
		return mSetting.Set(OPENAPI_ENABLED, "true")
	}
	return mSetting.Set(OPENAPI_ENABLED, "false")
}

// SaveApiKeys 保存 API Key 列表
func SaveApiKeys(keys []ApiKey) error {
	b, err := json.Marshal(keys)
	if err != nil {
		return err
	}
	mSetting := models.SystemSetting{}
	return mSetting.Set(OPENAPI_KEYS, string(b))
}
