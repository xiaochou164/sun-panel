import { post } from '@/utils/request'

export interface ApiKeyInfo {
  name: string
  keyMasked: string
  createdAt: string
  lastUsedAt: string
  description: string
}

export interface OpenApiConfig {
  enabled: boolean
  keys: ApiKeyInfo[]
}

export function getOpenApiConfig() {
  return post<OpenApiConfig>({ url: '/system/openApiConfig/getConfig' })
}

export function setOpenApiEnabled(enabled: boolean) {
  return post({ url: '/system/openApiConfig/setEnabled', data: { enabled } })
}

export function createApiKey(name: string, description: string) {
  return post<{ key: string }>({ url: '/system/openApiConfig/createKey', data: { name, description } })
}

export function deleteApiKey(name: string) {
  return post({ url: '/system/openApiConfig/deleteKey', data: { name } })
}
