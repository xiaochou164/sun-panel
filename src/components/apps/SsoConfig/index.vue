<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NButton, NCard, NForm, NFormItem, NInput, NSpace, NSpin, NSwitch, useMessage } from 'naive-ui'
import { getSsoConfigList, saveSsoConfig } from '@/api/system/sso'
import type { SsoConfig } from '@/api/system/sso'
import { SvgIcon } from '@/components/common/'

const ms = useMessage()
const configs = ref<SsoConfig[]>([])
const loading = ref(false)

const predefinedProviders = ['github', 'google', 'oidc']

const loadConfigs = async () => {
  loading.value = true
  try {
    const res = await getSsoConfigList()
    if (res.code === 0 && res.data) {
      const dbConfigs = res.data
      const merged = predefinedProviders.map((p) => {
        const found = dbConfigs.find(c => c.provider === p)
        if (found)
          return found
        return {
          provider: p,
          enabled: 0,
          name: p === 'github' ? 'GitHub' : p === 'google' ? 'Google' : 'OpenID Connect',
          clientId: '',
          clientSecret: '',
          issuerUrl: '',
          samlMetadata: '',
          ext: '',
        }
      })
      configs.value = merged
    }
  }
  catch (e) {
    ms.error('获取配置失败')
  }
  finally {
    loading.value = false
  }
}

onMounted(() => {
  loadConfigs()
})

const handleSave = async (cfg: SsoConfig) => {
  try {
    const res = await saveSsoConfig(cfg)
    if (res.code === 0) {
      ms.success('保存成功')
      loadConfigs()
    }
    else {
      ms.error(res.msg || '保存失败')
    }
  }
  catch (e) {
    ms.error('保存失败')
  }
}
</script>

<template>
  <div class="h-full bg-slate-200 dark:bg-zinc-900 p-2 overflow-auto">
    <div class="text-lg font-bold mb-4 text-slate-700 dark:text-slate-300">
      SSO 单点登录配置
    </div>
    <NSpin :show="loading">
      <NSpace vertical>
        <NCard v-for="cfg in configs" :key="cfg.provider" size="small" style="border-radius:10px;">
          <template #header>
            <div class="flex items-center gap-2">
              <SvgIcon :icon="cfg.provider === 'github' ? 'mdi:github' : cfg.provider === 'google' ? 'mdi:google' : 'mdi:openid'" />
              <span>{{ cfg.name || cfg.provider }}</span>
            </div>
          </template>
          <template #header-extra>
            <NSwitch v-model:value="cfg.enabled" :checked-value="1" :unchecked-value="0" @update:value="handleSave(cfg)" />
          </template>

          <NForm :model="cfg" label-placement="left" label-width="120" class="mt-4">
            <NFormItem label="显示名称">
              <NInput v-model:value="cfg.name" placeholder="例如: GitHub" />
            </NFormItem>
            <NFormItem label="客户端 ID">
              <NInput v-model:value="cfg.clientId" />
            </NFormItem>
            <NFormItem label="客户端密钥">
              <NInput v-model:value="cfg.clientSecret" type="password" show-password-on="click" />
            </NFormItem>
            <NFormItem v-if="cfg.provider === 'oidc'" label="Issuer URL">
              <NInput v-model:value="cfg.issuerUrl" placeholder="https://example.com" />
            </NFormItem>
            <NFormItem>
              <NButton type="primary" @click="handleSave(cfg)">
                保存
              </NButton>
            </NFormItem>
          </NForm>
        </NCard>
      </NSpace>
    </NSpin>
  </div>
</template>
