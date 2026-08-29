package systemSetting

// OpenAPI 配置存储
const (
	OPENAPI_ENABLED = "openapi_enabled" // 是否开启 OpenAPI/MCP  "true"/"false"
	OPENAPI_KEYS    = "openapi_keys"    // API Key 列表 JSON 数组 [{name,key,createdAt,lastUsedAt}]
)

// ApiKey 结构
type ApiKey struct {
	Name        string `json:"name"`        // 备注名
	Key         string `json:"key"`         // API Key 明文（仅生成时展示一次）
	CreatedAt   string `json:"createdAt"`   // 创建时间
	LastUsedAt  string `json:"lastUsedAt"`  // 最后使用时间
	Description string `json:"description"` // 描述
}
