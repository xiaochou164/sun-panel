<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NAlert, NButton, NCard, NForm, NFormItem, NInput, NPopconfirm, NSpace, NSpin, NSwitch, useMessage } from 'naive-ui'
import { createApiKey, deleteApiKey, getOpenApiConfig, setOpenApiEnabled } from '@/api/system/openApiConfig'
import type { ApiKeyInfo } from '@/api/system/openApiConfig'

const ms = useMessage()
const loading = ref(false)
const saving = ref(false)
const enabled = ref(false)
const keys = ref<ApiKeyInfo[]>([])
const name = ref('')
const description = ref('')
const newKey = ref('')

async function loadConfig() {
  loading.value = true
  try {
    const res = await getOpenApiConfig()
    if (res.code === 0 && res.data) {
      enabled.value = res.data.enabled
      keys.value = res.data.keys || []
    }
  }
  catch {
    ms.error('获取 OpenAPI 配置失败')
  }
  finally {
    loading.value = false
  }
}

async function toggleEnabled(value: boolean) {
  saving.value = true
  try {
    const res = await setOpenApiEnabled(value)
    if (res.code === 0) {
      enabled.value = value
      ms.success(value ? 'OpenAPI 已启用' : 'OpenAPI 已停用')
    }
    else {
      ms.error(res.msg || '保存失败')
    }
  }
  catch {
    ms.error('保存失败')
  }
  finally {
    saving.value = false
  }
}

async function handleCreate() {
  saving.value = true
  try {
    const res = await createApiKey(name.value || 'default', description.value)
    if (res.code === 0 && res.data?.key) {
      newKey.value = res.data.key
      name.value = ''
      description.value = ''
      await loadConfig()
      ms.success('API Key 创建成功，请立即复制保存')
    }
    else {
      ms.error(res.msg || '创建失败')
    }
  }
  catch {
    ms.error('创建失败')
  }
  finally {
    saving.value = false
  }
}

async function handleDelete(key: string) {
  try {
    const res = await deleteApiKey(key)
    if (res.code === 0) {
      ms.success('API Key 已删除')
      await loadConfig()
    }
    else {
      ms.error(res.msg || '删除失败')
    }
  }
  catch {
    ms.error('删除失败')
  }
}

async function copyKey() {
  await navigator.clipboard.writeText(newKey.value)
  ms.success('已复制 API Key')
}

onMounted(loadConfig)
</script>

<template>
  <div class="h-full bg-slate-200 dark:bg-zinc-900 p-2 overflow-auto">
    <div class="text-lg font-bold mb-4 text-slate-700 dark:text-slate-300">
      OpenAPI / MCP
    </div>
    <NSpin :show="loading">
      <NSpace vertical>
        <NCard size="small" style="border-radius:10px;">
          <div class="flex items-center justify-between">
            <div>
              <div class="font-bold">开放接口服务</div>
              <div class="text-sm text-slate-500 mt-1">启用后可通过 API Key 访问 OpenAPI 和 MCP。</div>
            </div>
            <NSwitch :value="enabled" :loading="saving" @update:value="toggleEnabled" />
          </div>
        </NCard>

        <NAlert v-if="newKey" type="warning" title="新 API Key 仅显示一次">
          <div class="break-all font-mono mb-2">{{ newKey }}</div>
          <NButton size="small" type="primary" @click="copyKey">复制 API Key</NButton>
        </NAlert>

        <NCard size="small" title="创建 API Key" style="border-radius:10px;">
          <NForm label-placement="top">
            <NFormItem label="名称">
              <NInput v-model:value="name" placeholder="例如：Hermes" />
            </NFormItem>
            <NFormItem label="描述">
              <NInput v-model:value="description" placeholder="用途说明（可选）" />
            </NFormItem>
            <NButton type="primary" :loading="saving" @click="handleCreate">创建 Key</NButton>
          </NForm>
        </NCard>

        <NCard size="small" title="已有 API Keys" style="border-radius:10px;">
          <div v-if="keys.length === 0" class="text-slate-500">暂无 API Key</div>
          <div v-for="item in keys" :key="item.keyMasked" class="border-b last:border-b-0 py-3">
            <div class="flex items-center justify-between gap-2">
              <div class="min-w-0">
                <div class="font-bold">{{ item.name }}</div>
                <div class="font-mono text-xs text-slate-500">{{ item.keyMasked }}</div>
                <div class="text-xs text-slate-500">创建：{{ item.createdAt }}<span v-if="item.lastUsedAt"> · 最近使用：{{ item.lastUsedAt }}</span></div>
                <div v-if="item.description" class="text-xs text-slate-500">{{ item.description }}</div>
              </div>
              <NPopconfirm positive-text="删除" negative-text="取消" @positive-click="handleDelete(item.name)">
                <template #trigger>
                  <NButton size="small" type="error" secondary>删除</NButton>
                </template>
                删除后无法恢复，确认继续？
              </NPopconfirm>
            </div>
          </div>
        </NCard>

        <NCard size="small" title="连接地址" style="border-radius:10px;">
          <div class="text-sm leading-7 font-mono break-all">
            OpenAPI: /api/openness/openapi/overview<br>
            MCP: /mcp<br>
            Header: Authorization: Bearer &lt;API Key&gt;
          </div>
        </NCard>
      </NSpace>
    </NSpin>
  </div>
</template>
