<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { NButton, NCard, NDivider, NForm, NFormItem, NInput, NSelect, useDialog, useMessage } from 'naive-ui'
import { onMounted, ref } from 'vue'
import { useAppStore, useAuthStore, usePanelState, useUserStore } from '@/store'
import { languageOptions } from '@/utils/defaultData'
import type { Language, Theme } from '@/store/modules/app/helper'
import { logout } from '@/api'
import { RoundCardModal, SvgIcon } from '@/components/common/'
import { updateInfo, updatePassword } from '@/api/system/user'
import { getProviders, getUserBindings, unbindSso } from '@/api/system/sso'
import type { SsoProvider, UserBinding } from '@/api/system/sso'
import { updateLocalUserInfo } from '@/utils/cmn'
import { t } from '@/locales'

const userStore = useUserStore()
const authStore = useAuthStore()
const appStore = useAppStore()
const panelState = usePanelState()
const ms = useMessage()
const dialog = useDialog()

const languageValue = ref(appStore.language)
const themeValue = ref(appStore.theme)
const nickName = ref(authStore.userInfo?.name || '')
const isEditNickNameStatus = ref(false)
const formRef = ref<FormInst | null>(null)
const themeOptions: { label: string; key: string; value: Theme }[] = [
  { label: t('apps.userInfo.themeStyle.dark'), key: 'dark', value: 'dark' },
  { label: t('apps.userInfo.themeStyle.light'), key: 'light', value: 'light' },
  { label: t('apps.userInfo.themeStyle.auto'), key: 'Auto', value: 'auto' },
]
const updatePasswordModalState = ref({
  show: false,
  loading: false,
  form: {
    password: '',
    oldPassword: '',
    confirmPassword: '',
  },
})

const updatePasswordModalFormRules: FormRules = {
  oldPassword: {
    required: true,
    trigger: 'blur',
    min: 6,
    max: 20,
    message: t('adminSettingUsers.formRules.passwordLimit'),
  },
  password: {
    required: true,
    trigger: 'blur',
    min: 6,
    max: 20,
    message: t('adminSettingUsers.formRules.passwordLimit'),
  },
  confirmPassword: {
    required: true,
    trigger: 'blur',
    min: 6,
    max: 20,
    message: t('adminSettingUsers.formRules.passwordLimit'),
  },
}

const ssoProviders = ref<SsoProvider[]>([])
const userBindings = ref<UserBinding[]>([])
const ssoBindModalShow = ref(false)

const loadBindings = async () => {
  try {
    const bindRes = await getUserBindings()
    if (bindRes.code === 0 && bindRes.data)
      userBindings.value = bindRes.data
  }
  catch (e) {
  }
}

const loadProviders = async () => {
  try {
    const provRes = await getProviders()
    if (provRes.code === 0 && provRes.data)
      ssoProviders.value = provRes.data
  }
  catch (e) {
  }
}

onMounted(() => {
  loadBindings()
  loadProviders()
})

const handleUnbind = (provider: string) => {
  dialog.warning({
    title: t('common.warning'),
    content: '确定要解绑此账号吗？',
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      const res = await unbindSso(provider)
      if (res.code === 0) {
        ms.success(t('common.success'))
        loadBindings()
      }
      else {
        ms.error(res.msg || t('common.error'))
      }
    },
  })
}

const handleBind = (provider: string) => {
  window.location.href = `/api/system/sso/login/${provider}?token=${authStore.token}`
}

async function logoutApi() {
  await logout()
  userStore.resetUserInfo()
  authStore.removeToken()
  panelState.removeState()
  appStore.removeToken()
  ms.success(t('settingUserInfo.logoutSuccess'))
  // router.push({ path: '/login' })
  location.reload()// 强制刷新一下页面
}

function handleSaveInfo() {
  updateInfo(nickName.value).then(({ code, msg }) => {
    if (code === 0) {
      updateLocalUserInfo()
      isEditNickNameStatus.value = false
    }
    else {
      ms.error(`${t('common.editFail')}:${msg}`)
    }
  })
}

function handleUpdatePassword(e: MouseEvent) {
  e.preventDefault()
  formRef.value?.validate((errors) => {
    if (errors) {
      console.log(errors)
      return
    }

    if (updatePasswordModalState.value.form.password !== updatePasswordModalState.value.form.confirmPassword) {
      ms.error(t('settingUserInfo.confirmPasswordInconsistentMsg'))
      return
    }
    updatePasswordModalState.value.loading = true
    updatePassword(updatePasswordModalState.value.form.oldPassword, updatePasswordModalState.value.form.password).then(({ code }) => {
      if (code === 0) {
        // 成功
        updatePasswordModalState.value.show = false
        ms.success(t('common.success'))
      }
    }).finally(() => {
      updatePasswordModalState.value.loading = false
    }).catch(() => {
      ms.error(t('common.serverError'))
    })
  })
}

function handleLogout() {
  dialog.warning({
    title: t('common.warning'),
    content: t('settingUserInfo.confirmLogoutText'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => {
      logoutApi()
    },
  })
}

function handleChangeLanuage(value: Language) {
  languageValue.value = value
  appStore.setLanguage(value)
  location.reload()
}

function handleChangeTheme(value: Theme) {
  themeValue.value = value
  appStore.setTheme(value)
  // location.reload()
}
</script>

<template>
  <div class="bg-slate-200 dark:bg-zinc-900 p-2 h-full">
    <NCard style="border-radius:10px" size="small">
      <div>
        <div class="text-slate-500 font-bold">
          {{ $t('common.username') }}
        </div>
        {{ authStore.userInfo?.username }}
      </div>

      <div class="mt-[10px]">
        <div class="text-slate-500 font-bold">
          {{ $t('common.nikeName') }}
        </div>

        <div v-if="!isEditNickNameStatus">
          {{ authStore.userInfo?.name }}

          <NButton size="small" text type="info" @click="isEditNickNameStatus = !isEditNickNameStatus">
            {{ $t('common.edit') }}
          </NButton>
        </div>

        <div v-else class="flex items-center">
          <div class="max-w-[150px]">
            <NInput v-model:value="nickName" type="text" :placeholder="$t('common.inputPlaceholder')" />
          </div>
          <NButton size="small" quaternary type="info" @click="handleSaveInfo">
            {{ $t('common.save') }}
          </NButton>
        </div>
      </div>

      <div class="mt-[10px]">
        <div class="text-slate-500 font-bold">
          {{ $t('common.language') }}
        </div>
        <div class="max-w-[200px]">
          <NSelect v-model:value="languageValue" :options="languageOptions" @update-value="handleChangeLanuage" />
        </div>
      </div>

      <div class="mt-[10px]">
        <div class="text-slate-500 font-bold">
          {{ $t('apps.userInfo.theme') }}
        </div>
        <div class="max-w-[200px]">
          <NSelect v-model:value="themeValue" :options="themeOptions" @update-value="handleChangeTheme" />
        </div>
      </div>

      <NDivider style="margin: 10px 0;" dashed />

      <div class="mt-[10px]">
        <div class="text-slate-500 font-bold mb-2">
          社交账号 / 单点登录 (SSO)
        </div>
        <div class="flex flex-col gap-2">
          <div v-for="bind in userBindings" :key="bind.provider" class="flex items-center justify-between bg-slate-100 dark:bg-zinc-800 p-2 rounded">
            <div class="flex items-center gap-2">
              <SvgIcon :icon="bind.provider === 'github' ? 'mdi:github' : bind.provider === 'google' ? 'mdi:google' : 'mdi:link'" />
              <span>{{ bind.provider }}</span>
            </div>
            <NButton size="tiny" type="error" ghost @click="handleUnbind(bind.provider)">
              解绑
            </NButton>
          </div>
          <NButton size="small" type="primary" dashed @click="ssoBindModalShow = true">
            <template #icon>
              <SvgIcon icon="mdi:link-variant" />
            </template>
            绑定账号
          </NButton>
        </div>
      </div>

      <NDivider style="margin: 10px 0;" dashed />
      <div>
        <NButton size="small" text type="info" @click="updatePasswordModalState.show = !updatePasswordModalState.show">
          {{ $t('settingUserInfo.updatePassword') }}
        </NButton>
      </div>
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <NButton size="small" text type="error" @click="handleLogout">
        <template #icon>
          <SvgIcon icon="tabler:logout" />
        </template>
        {{ $t('settingUserInfo.logout') }}
      </NButton>
    </NCard>

    <RoundCardModal v-model:show="updatePasswordModalState.show" size="small" preset="card" style="width: 400px" :title="$t('settingUserInfo.updatePassword')">
      <NForm ref="formRef" :model="updatePasswordModalState.form" :rules="updatePasswordModalFormRules">
        <NFormItem path="oldPassword" :label="$t('settingUserInfo.oldPassword')">
          <NInput v-model:value="updatePasswordModalState.form.oldPassword" :maxlength="20" type="password" :placeholder="$t('settingUserInfo.oldPassword')" />
        </NFormItem>

        <NFormItem path="password" :label="$t('settingUserInfo.newPassword')">
          <NInput v-model:value="updatePasswordModalState.form.password" :maxlength="20" type="password" :placeholder="$t('settingUserInfo.newPassword')" />
        </NFormItem>

        <NFormItem path="confirmPassword" :label="$t('settingUserInfo.confirmPassword')">
          <NInput v-model:value="updatePasswordModalState.form.confirmPassword" :maxlength="20" type="password" :placeholder="$t('settingUserInfo.confirmPassword')" />
        </NFormItem>
      </NForm>

      <template #footer>
        <div class="float-right">
          <NButton type="success" size="small" :loading="updatePasswordModalState.loading" @click="handleUpdatePassword">
            {{ $t('common.save') }}
          </NButton>
        </div>
      </template>
    </RoundCardModal>

    <RoundCardModal v-model:show="ssoBindModalShow" size="small" preset="card" style="width: 400px" title="绑定账号">
      <div class="flex flex-col gap-2">
        <NButton
          v-for="provider in ssoProviders.filter(p => !userBindings.some(b => b.provider === p.provider))"
          :key="provider.provider"
          @click="handleBind(provider.provider)"
        >
          <template #icon>
            <SvgIcon :icon="provider.provider === 'github' ? 'mdi:github' : provider.provider === 'google' ? 'mdi:google' : 'mdi:login'" />
          </template>
          绑定 {{ provider.name }}
        </NButton>
        <div v-if="ssoProviders.filter(p => !userBindings.some(b => b.provider === p.provider)).length === 0" class="text-center text-slate-400">
          没有可绑定的账号提供商
        </div>
      </div>
    </RoundCardModal>
  </div>
</template>
