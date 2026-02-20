<script setup lang="ts">
import { ref, watch } from 'vue'
import type { UploadFileInfo } from 'naive-ui'
import { NAvatar, NButton, NCard, NColorPicker, NGrid, NGridItem, NInput, NInputGroup, NModal, NPopconfirm, NSelect, NSlider, NSwitch, NUpload, NUploadDragger, useMessage } from 'naive-ui'
import { useAuthStore, usePanelState } from '@/store'
import { set as setUserConfig } from '@/api/panel/userConfig'
import { PanelPanelConfigStyleEnum } from '@/enums/panel'
import { t } from '@/locales'
import { useModuleConfig } from '@/store/modules'
import { SvgIcon } from '@/components/common'

import SvgSrcBaidu from '@/assets/search_engine_svg/baidu.svg'
import SvgSrcBing from '@/assets/search_engine_svg/bing.svg'
import SvgSrcGoogle from '@/assets/search_engine_svg/google.svg'

const authStore = useAuthStore()
const panelState = usePanelState()
const ms = useMessage()
const showWallpaperInput = ref(false)
const bingLoading = ref(false)

const isSaveing = ref(false)

// Custom Search Engine State
const moduleConfigName = 'deskModuleSearchBox'
const moduleConfig = useModuleConfig()
const showManageSearchEngineModal = ref(false)

const defaultSearchEngineList = ref<DeskModule.SearchBox.SearchEngine[]>([
  { iconSrc: SvgSrcGoogle, title: 'Google', url: 'https://www.google.com/search?q=%s' },
  { iconSrc: SvgSrcBaidu, title: 'Baidu', url: 'https://www.baidu.com/s?wd=%s' },
  { iconSrc: SvgSrcBing, title: 'Bing', url: 'https://www.bing.com/search?q=%s' },
])

const searchBoxState = ref<any>({
  currentSearchEngine: defaultSearchEngineList.value[0],
  searchEngineList: [...defaultSearchEngineList.value],
  newWindowOpen: false,
})

const newEngine = ref<DeskModule.SearchBox.SearchEngine>({
  title: '',
  url: '',
  iconSrc: '',
  isCustom: true,
})

const editingIndex = ref<number>(-1)

async function handleUseBingWallpaper() {
  bingLoading.value = true
  try {
    const res = await fetch('/api/openness/bingWallpaper')
    const data = await res.json()
    if (data.code === 0 && data.data?.url) {
      panelState.panelConfig.backgroundImageSrc = data.data.url
      ms.success('已设置 Bing 今日壁纸')
    }
    else {
      ms.error(`获取 Bing 壁纸失败：${data.msg || '未知错误'}`)
    }
  }
  catch (e) {
    ms.error('获取 Bing 壁纸失败')
  }
  finally {
    bingLoading.value = false
  }
}

const iconTypeOptions = [
  {
    label: t('apps.baseSettings.detailIcon'),
    value: PanelPanelConfigStyleEnum.info,
  },
  {
    label: t('apps.baseSettings.smallIcon'),
    value: PanelPanelConfigStyleEnum.icon,
  },
]

const maxWidthUnitOption = [
  {
    label: 'px',
    value: 'px',
  },
  {
    label: '%',
    value: '%',
  },
]

watch(panelState.panelConfig, () => {
  if (!isSaveing.value) {
    isSaveing.value = true

    setTimeout(() => {
      panelState.recordState()// 本地记录
      isSaveing.value = false
      uploadCloud()
    }, 1000)
  }
})

function handleUploadBackgroundFinish({
  file,
  event,
}: {
  file: UploadFileInfo
  event?: ProgressEvent
}) {
  const res = JSON.parse((event?.target as XMLHttpRequest).response)
  panelState.panelConfig.backgroundImageSrc = res.data.imageUrl
  return file
}

function uploadCloud() {
  setUserConfig({ panel: panelState.panelConfig }).then((res) => {
    if (res.code === 0)
      ms.success(t('apps.baseSettings.configSaved'))
    else
      ms.error(t('apps.baseSettings.configFailed', { message: res.msg }))
  })
}

function resetPanelConfig() {
  panelState.resetPanelConfig()
  uploadCloud()
}

// Custom Search Engine Logic
function openManageSearchEngines() {
  moduleConfig.getValueByNameFromCloud<any>(moduleConfigName).then(({ code, data }) => {
    if (code === 0 && data) {
      const savedEngines = data.searchEngineList || []
      const customEngines = savedEngines.filter((e: any) => e.isCustom)
      data.searchEngineList = [...defaultSearchEngineList.value, ...customEngines]
      searchBoxState.value = data
    }
  })
  showManageSearchEngineModal.value = true
}

function handleAddCustomEngine() {
  if (!newEngine.value.title || !newEngine.value.url) {
    ms.error('请输入名称和网址')
    return
  }

  if (editingIndex.value !== -1) {
    // Edit existing engine
    searchBoxState.value.searchEngineList[editingIndex.value] = { ...newEngine.value, isCustom: true }
    ms.success('修改成功')
  }
  else {
    // Add new engine
    searchBoxState.value.searchEngineList.push({ ...newEngine.value, isCustom: true })
    ms.success('添加成功')
  }

  moduleConfig.saveToCloud(moduleConfigName, searchBoxState.value)
  handleCancelEditEngine()
}

function handleEditCustomEngine(index: number) {
  editingIndex.value = index
  newEngine.value = { ...searchBoxState.value.searchEngineList[index] }
}

function handleCancelEditEngine() {
  editingIndex.value = -1
  newEngine.value = { title: '', url: '', iconSrc: '', isCustom: true }
}

function handleSetDefaultEngine(item: any) {
  searchBoxState.value.currentSearchEngine = item
  moduleConfig.saveToCloud(moduleConfigName, searchBoxState.value)
  ms.success('已设为默认搜索引擎')
}

function handleDeleteCustomEngine(index: number) {
  searchBoxState.value.searchEngineList.splice(index, 1)

  const isCurrentDeleted = !searchBoxState.value.searchEngineList.find((e: any) => e.title === searchBoxState.value.currentSearchEngine.title)
  if (isCurrentDeleted && defaultSearchEngineList.value.length > 0)
    searchBoxState.value.currentSearchEngine = defaultSearchEngineList.value[0]

  moduleConfig.saveToCloud(moduleConfigName, searchBoxState.value)
}
</script>

<template>
  <div class="bg-slate-200 dark:bg-zinc-900 rounded-[10px] p-[8px] overflow-auto">
    <NCard style="border-radius:10px" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        LOGO
      </div>

      <div>
        <div>
          {{ $t('apps.baseSettings.textContent') }}
        </div>
        <div class="flex items-center mt-[5px]">
          <NInput v-model:value="panelState.panelConfig.logoText" type="text" show-count :maxlength="20" placeholder="请输入文字" />
        </div>
      </div>
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('apps.baseSettings.clock') }}
      </div>
      <div class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.clockSecondShow') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.clockShowSecond" />
      </div>
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('apps.baseSettings.searchBar') }}
      </div>
      <div class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('common.show') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.searchBoxShow" />
      </div>
      <div v-if="panelState.panelConfig.searchBoxShow" class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.searchBarSearchItem') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.searchBoxSearchIcon" />
      </div>
      <div v-if="panelState.panelConfig.searchBoxShow" class="mt-[10px]">
        <NButton size="small" type="primary" secondary @click="openManageSearchEngines">
          <template #icon>
            <SvgIcon icon="mdi:cog-outline" />
          </template>
          管理搜索引擎
        </NButton>
      </div>
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('apps.baseSettings.systemMonitorStatus') }}
      </div>
      <div class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('common.show') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.systemMonitorShow" />
      </div>
      <div v-if="panelState.panelConfig.systemMonitorShow" class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.showTitle') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.systemMonitorShowTitle" />
      </div>
      <div v-if="panelState.panelConfig.systemMonitorShow" class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.publicVisitModeShow') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.systemMonitorPublicVisitModeShow" />
      </div>
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('common.icon') }}
      </div>
      <div class="mt-[5px]">
        <div>
          {{ $t('common.style') }}
        </div>
        <div class="flex items-center mt-[5px]">
          <NSelect v-model:value="panelState.panelConfig.iconStyle" :options="iconTypeOptions" />
        </div>
      </div>

      <div v-if="panelState.panelConfig.iconStyle === PanelPanelConfigStyleEnum.info" class="mt-[5px]">
        <div>
          {{ $t('apps.baseSettings.hideDescription') }}
        </div>
        <div class="flex items-center mt-[5px]">
          <NSwitch v-model:value="panelState.panelConfig.iconTextInfoHideDescription" />
        </div>
      </div>

      <div v-if="panelState.panelConfig.iconStyle === PanelPanelConfigStyleEnum.icon" class="mt-[5px]">
        <div>
          {{ $t('apps.baseSettings.hideTitle') }}
        </div>
        <div class="flex items-center mt-[5px]">
          <NSwitch v-model:value="panelState.panelConfig.iconTextIconHideTitle" />
        </div>
      </div>

      <div class="mt-[5px]">
        <div>
          {{ $t('common.textColor') }}
        </div>
        <div class="flex items-center mt-[5px]">
          <NColorPicker
            v-model:value="panelState.panelConfig.iconTextColor"
            :show-alpha="false"
            size="small"
            :modes="['hex']"
            :swatches="[
              '#000000',
              '#ffffff',
              '#18A058',
              '#2080F0',
              '#F0A020',
            ]"
          />
        </div>
      </div>
    </NCard>
    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('apps.baseSettings.wallpaper') }}
      </div>
      <NUpload
        action="/api/file/uploadImg"
        :show-file-list="false"
        name="imgfile"
        :headers="{
          token: authStore.token as string,
        }"
        :directory-dnd="true"
        @finish="handleUploadBackgroundFinish"
      >
        <NUploadDragger style="width: 100%;">
          <div
            class="h-[200px] w-full border bg-slate-100 flex justify-center items-center cursor-pointer rounded-[10px]"
            :style="{ background: `url(${panelState.panelConfig.backgroundImageSrc}) no-repeat`, backgroundSize: 'cover' }"
          >
            <div class="text-shadow text-white">
              {{ $t('apps.baseSettings.uploadOrDragText') }}
            </div>
          </div>
        </NUploadDragger>
      </NUpload>

      <div class="flex items-center mt-[5px] gap-[8px]">
        <NButton size="small" :loading="bingLoading" @click="handleUseBingWallpaper">
          🌄 Bing 今日壁纸
        </NButton>
        <NButton v-if="panelState.panelConfig.backgroundImageSrc" size="small" quaternary type="error" @click="panelState.panelConfig.backgroundImageSrc = ''">
          清除壁纸
        </NButton>
      </div>

      <div class="flex items-center mt-[5px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.customImageAddress') }}</span>
        <NSwitch v-model:value="showWallpaperInput" />
      </div>
      <div v-if="showWallpaperInput" class="mt-1">
        <NInput v-model:value="panelState.panelConfig.backgroundImageSrc" type="text" size="small" clearable />
      </div>

      <div class="flex items-center mt-[10px]">
        <span class="mr-[10px]">Bing 自动轮换</span>
        <NSwitch v-model:value="panelState.panelConfig.bingAutoRotate" />
      </div>

      <div v-if="panelState.panelConfig.bingAutoRotate" class="flex items-center mt-[10px]">
        <span class="mr-[10px]">轮换间隔 (分钟)</span>
        <NInput :value="String(panelState.panelConfig.bingAutoRotateInterval || '')" type="text" size="small" :style="{ width: '100px' }" placeholder="30" @update:value="(val: string) => panelState.panelConfig.bingAutoRotateInterval = Number(val)" />
      </div>

      <div class="flex items-center mt-[10px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.vague') }}</span>
        <NSlider v-model:value="panelState.panelConfig.backgroundBlur" class="max-w-[200px]" :step="2" :max="20" />
      </div>

      <div class="flex items-center mt-[10px]">
        <span class="mr-[10px]">{{ $t('apps.baseSettings.mask') }}</span>
        <NSlider v-model:value="panelState.panelConfig.backgroundMaskNumber" class="max-w-[200px]" :step="0.1" :max="1" />
      </div>
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('apps.baseSettings.contentArea') }}
      </div>

      <NGrid cols="2">
        <NGridItem span="12 400:12">
          <div class="flex items-center mt-[5px]">
            <span class="mr-[10px]">{{ $t('apps.baseSettings.netModeChangeButtonShow') }}</span>
            <NSwitch v-model:value="panelState.panelConfig.netModeChangeButtonShow" />
          </div>
        </NGridItem>

        <NGridItem span="12 400:12">
          <div class="flex items-center mt-[10px]">
            <span class="mr-[10px]">{{ $t('apps.baseSettings.maxWidth') }}</span>
            <div class="flex">
              <NInputGroup>
                <NInput :value="String(panelState.panelConfig.maxWidth || '')" size="small" type="text" :maxlength="10" :style="{ width: '100px' }" placeholder="1200" @update:value="(val: string) => panelState.panelConfig.maxWidth = Number(val)" />
                <NSelect v-model:value="panelState.panelConfig.maxWidthUnit" :style="{ width: '80px' }" :options="maxWidthUnitOption" size="small" />
              </NInputGroup>
            </div>
          </div>
        </NGridItem>
        <NGridItem span="12 400:12">
          <div class="flex items-center mt-[10px]">
            <span class="mr-[10px]">{{ $t('apps.baseSettings.leftRightMargin') }}</span>
            <NSlider v-model:value="panelState.panelConfig.marginX" class="max-w-[200px]" :step="1" :max="100" />
          </div>
        </NGridItem>
        <NGridItem span="12 400:12">
          <div class="flex items-center mt-[10px]">
            <span class="mr-[10px]">{{ $t('apps.baseSettings.topMargin') }} (%)</span>
            <NSlider v-model:value="panelState.panelConfig.marginTop" class="max-w-[200px]" :step="1" :max="50" />
          </div>
        </NGridItem>
        <NGridItem span="12 400:6">
          <div class="flex items-center mt-[10px]">
            <span class="mr-[10px]">{{ $t('apps.baseSettings.bottomMargin') }} (%)</span>
            <NSlider v-model:value="panelState.panelConfig.marginBottom" class="max-w-[200px]" :step="1" :max="50" />
          </div>
        </NGridItem>
      </NGrid>
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <div class="text-slate-500 mb-[5px] font-bold">
        {{ $t('apps.baseSettings.customFooter') }}
      </div>

      <NInput
        v-model:value="panelState.panelConfig.footerHtml"
        type="textarea"
        clearable
      />
    </NCard>

    <NCard style="border-radius:10px" class="mt-[10px]" size="small">
      <NPopconfirm
        @positive-click="resetPanelConfig"
      >
        <template #trigger>
          <NButton size="small" quaternary type="error">
            {{ $t('common.reset') }}
          </NButton>
        </template>
        {{ $t('apps.baseSettings.resetWarnText') }}
      </NPopconfirm>

      <NButton size="small" quaternary type="success" class="ml-[10px]" @click="uploadCloud">
        {{ $t('common.save') }}
      </NButton>
    </NCard>

    <!-- 管理搜索引擎模态框 -->
    <NModal v-model:show="showManageSearchEngineModal" preset="card" style="width: 500px" title="管理自定义搜索引擎" :bordered="false">
      <!-- 已有搜索引擎列表 -->
      <div class="mb-4">
        <label class="text-sm font-bold text-slate-500 mb-2 block">当前搜索引擎列表</label>
        <div class="flex flex-col gap-2 max-h-[150px] overflow-y-auto pr-2 pb-2">
          <div
            v-for="(item, index) in searchBoxState.searchEngineList"
            :key="index"
            class="flex items-center justify-between p-2 rounded-lg border bg-slate-50 dark:bg-zinc-800"
            :class="{ 'border-blue-500 bg-blue-50 dark:bg-blue-900/20': searchBoxState.currentSearchEngine?.title === item.title }"
          >
            <div class="flex items-center gap-2">
              <NAvatar :src="item.iconSrc || SvgSrcGoogle" :size="20" style="background-color: transparent;" />
              <span class="text-sm font-medium">{{ item.title }}</span>
              <span v-if="searchBoxState.currentSearchEngine?.title === item.title" class="text-[10px] bg-blue-100 dark:bg-blue-800 text-blue-600 dark:text-blue-200 px-1 py-0.5 rounded ml-1">当前默认</span>
            </div>
            <div class="flex items-center gap-1">
              <NButton
                v-if="searchBoxState.currentSearchEngine?.title !== item.title"
                size="tiny"
                type="info"
                quaternary
                @click="handleSetDefaultEngine(item)"
              >
                设为默认
              </NButton>

              <NButton v-if="item.isCustom" size="tiny" type="warning" quaternary @click="handleEditCustomEngine(index)">
                编辑
              </NButton>

              <span v-if="!item.isCustom" class="text-xs text-slate-400 mx-1">内置</span>
              <NButton v-else size="tiny" type="error" quaternary @click="handleDeleteCustomEngine(Number(index))">
                移除
              </NButton>
            </div>
          </div>
        </div>
      </div>
      <div class="h-[1px] bg-slate-200 dark:bg-zinc-700 w-full my-4" />

      <!-- 添加/修改新引擎表单 -->
      <div class="flex flex-col gap-4">
        <div>
          <label class="text-sm font-bold text-slate-500 mb-1 block">{{ editingIndex !== -1 ? '修改引擎名称' : '添加新引擎名称' }}</label>
          <NInput v-model:value="newEngine.title" placeholder="如: SearXNG、Bing" />
        </div>
        <div>
          <label class="text-sm font-bold text-slate-500 mb-1 block">搜索 URL (用 %s 代替关键词)</label>
          <NInput v-model:value="newEngine.url" placeholder="http://searxng.local/search?q=%s" />
        </div>
        <div>
          <label class="text-sm font-bold text-slate-500 mb-1 block">图标 URL (可选)</label>
          <NInput v-model:value="newEngine.iconSrc" placeholder="https://..." />
        </div>
        <div class="flex justify-end gap-2 mt-2">
          <NButton v-if="editingIndex !== -1" @click="handleCancelEditEngine">
            取消修改
          </NButton>
          <NButton v-else @click="showManageSearchEngineModal = false">
            关闭
          </NButton>
          <NButton type="primary" @click="handleAddCustomEngine">
            {{ editingIndex !== -1 ? '保存修改' : '添加并保存' }}
          </NButton>
        </div>
      </div>
    </NModal>
  </div>
</template>

<style scoped>
.text-shadow{
  text-shadow: 0px 0px 5px gray;
}
</style>
