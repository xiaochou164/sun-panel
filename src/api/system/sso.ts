import { get, post } from '@/utils/request'

export interface SsoProvider {
  provider: string
  name: string
}

export interface UserBinding {
  provider: string
  providerUid: string
  createdAt: string
}

export function getProviders() {
  return get<SsoProvider[]>({ url: '/system/sso/providers' })
}

export function getUserBindings() {
  return post<UserBinding[]>({ url: '/system/sso/getUserBindings' })
}

export function unbindSso(provider: string) {
  return post({ url: '/system/sso/unbind', data: { provider } })
}

export interface SsoConfig {
  provider: string
  enabled: number
  name: string
  clientId: string
  clientSecret: string
  issuerUrl: string
  samlMetadata: string
  ext: string
}

export function getSsoConfigList() {
  return post<SsoConfig[]>({ url: '/system/ssoConfig/getList' })
}

export function saveSsoConfig(data: Partial<SsoConfig>) {
  return post({ url: '/system/ssoConfig/save', data })
}
