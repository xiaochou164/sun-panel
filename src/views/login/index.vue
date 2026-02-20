<script setup lang="ts">
import { NButton, NCard, NForm, NFormItem, NGradientText, NInput, NSelect, useMessage } from 'naive-ui'
import { onMounted, ref } from 'vue'
import { login } from '@/api'
import { useAppStore, useAuthStore } from '@/store'
import { SvgIcon } from '@/components/common'
import { router } from '@/router'
import { t } from '@/locales'
import { languageOptions } from '@/utils/defaultData'
import type { Language } from '@/store/modules/app/helper'
import { getProviders } from '@/api/system/sso'
import type { SsoProvider } from '@/api/system/sso'
import { post } from '@/utils/request'

// const userStore = useUserStore()
const authStore = useAuthStore()
const appStore = useAppStore()
const ms = useMessage()
const loading = ref(false)
const languageValue = ref<Language>(appStore.language)

const ssoProviders = ref<SsoProvider[]>([])

onMounted(async () => {
  // Check for SSO callback token in URL hash
  const hash = window.location.hash
  if (hash.includes('?')) {
    const searchParams = new URLSearchParams(hash.split('?')[1])
    const ssoToken = searchParams.get('ssoToken')
    const ssoError = searchParams.get('ssoError')

    if (ssoError) {
      if (ssoError === 'Bind success') {
        ms.success('绑定成功')
        router.push({ path: '/' })
      }
      else {
        ms.error(decodeURIComponent(ssoError))
      }
    }

    if (ssoToken) {
      loading.value = true
      authStore.setToken(ssoToken)

      // Fetch user info through standard auth flow or simulate login success
      try {
        const userInfoRes = await post<any>({ url: '/user/getInfo' })
        if (userInfoRes.code === 0) {
          authStore.setUserInfo(userInfoRes.data)
          ms.success(`Hi ${userInfoRes.data.name || ''},${t('login.welcomeMessage')}`)
          router.push({ path: '/' })
        }
        else {
          ms.error('Failed to retrieve user info')
          authStore.removeToken()
        }
      }
      catch (e) {
        authStore.removeToken()
      }
      finally {
        loading.value = false
      }
    }
  }

  // Load SSO providers
  try {
    const res = await getProviders()
    if (res.code === 0 && res.data)
      ssoProviders.value = res.data
  }
  catch (e) {
    console.error(e)
  }
})

const goSsoLogin = (provider: string) => {
  window.location.href = `/api/system/sso/login/${provider}`
}

const form = ref<Login.LoginReqest>({
  username: '',
  password: '',
})

const loginPost = async () => {
  loading.value = true
  try {
    const res = await login<Login.LoginResponse>(form.value)
    if (res.code === 0) {
      authStore.setToken(res.data.token)
      authStore.setUserInfo(res.data)

      setTimeout(() => {
        ms.success(`Hi ${res.data.name},${t('login.welcomeMessage')}`)
        loading.value = false
        router.push({ path: '/' })
      }, 500)
    }
    else {
      loading.value = false
    }
  }
  catch (error) {
    loading.value = false
    console.log(error)
  }
}

function handleSubmit() {
  loginPost()
}

function handleChangeLanuage(value: Language) {
// ... continuing code
  languageValue.value = value
  appStore.setLanguage(value)
}
</script>

<template>
  <div class="login-container">
    <NCard class="login-card" style="border-radius: 20px;">
      <div class="mb-5 flex items-center justify-end">
        <div class="mr-2">
          <SvgIcon icon="ion-language" style="width: 20;height: 20;" />
        </div>
        <div class="min-w-[100px]">
          <NSelect v-model:value="languageValue" size="small" :options="languageOptions" @update-value="handleChangeLanuage" />
        </div>
      </div>

      <div class="login-title  ">
        <NGradientText :size="30" type="success" class="!font-bold">
          {{ $t('common.appName') }}
        </NGradientText>
      </div>
      <NForm :model="form" label-width="100px" @keydown.enter="handleSubmit">
        <NFormItem>
          <NInput v-model:value="form.username" :placeholder="$t('login.usernamePlaceholder')">
            <template #prefix>
              <SvgIcon icon="ph:user-bold" />
            </template>
          </NInput>
        </NFormItem>

        <NFormItem>
          <NInput v-model:value="form.password" type="password" :placeholder="$t('login.passwordPlaceholder')">
            <template #prefix>
              <SvgIcon icon="mdi:password-outline" />
            </template>
          </NInput>
        </NFormItem>

        <!-- <NFormItem v-if="isShowCaptcha">
          <div class="w-[120px] h-[34px] mr-[20px] rounded border flex cursor-pointer">
            <Captcha ref="captchaRef" src="/api/captcha/getImage" />
          </div>
          <NInput v-model:value="form.vcode" type="text" placeholder="请输入图像验证码" />
        </NFormItem> -->
        <NFormItem style="margin-top: 10px">
          <NButton type="primary" block :loading="loading" @click="handleSubmit">
            {{ $t('login.loginButton') }}
          </NButton>
        </NFormItem>

        <div v-if="ssoProviders.length > 0" class="flex flex-col items-center mt-2 mb-4">
          <div class="w-full flex items-center justify-between text-slate-400 mb-4 px-2">
            <div class="h-[1px] bg-slate-200 flex-1" />
            <span class="px-2 text-xs">OR</span>
            <div class="h-[1px] bg-slate-200 flex-1" />
          </div>
          <div class="flex gap-4 justify-center flex-wrap">
            <NButton
              v-for="provider in ssoProviders"
              :key="provider.provider"
              tertiary
              @click="goSsoLogin(provider.provider)"
            >
              <template #icon>
                <SvgIcon :icon="provider.provider === 'github' ? 'mdi:github' : provider.provider === 'google' ? 'mdi:google' : 'mdi:login'" />
              </template>
              {{ provider.name }}
            </NButton>
          </div>
        </div>

        <!-- <div class="flex justify-end">
          <NButton v-if="isShowRegister" quaternary type="info" class="flex" @click="$router.push({ path: '/register' })">
            注册
          </NButton>
          <NButton quaternary type="info" class="flex" @click="$router.push({ path: '/resetPassword' })">
            忘记密码?
          </NButton>
        </div> -->

        <div class="flex justify-center text-slate-300">
          Powered By <a href="https://github.com/hslr-s/sun-panel" target="_blank" class="ml-[5px] text-slate-500">Sun-Panel</a>
        </div>
      </NForm>
    </NCard>
  </div>
</template>

  <style>
    .login-container {
        padding: 20px;
        display: flex;
        justify-content: center;
        align-items: center;
        height: 100vh;
        background-color: #f2f6ff;
    }

    /* 夜间模式 */
    .dark .login-container{
      background-color: rgb(43, 43, 43);
    }

    @media (min-width: 600px) {
        .login-card {
            width: auto;
            margin: 0px 10px;
        }
        .login-button {
            width: 100%;
        }
    }

    .login-card {
        margin: 20px;
        min-width:400px;
    }

  .login-title{
    text-align: center;
    margin: 20px;
  }
  </style>
