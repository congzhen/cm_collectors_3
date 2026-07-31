<template>
  <header class="modern-header" :class="{ 'is-search-wrapped': searchWrapped }" data-header-style="modern">
    <div class="modern-header-brand">
      <img src="/public/icon32.png" alt="" />
      <span>{{ store.appStoreData.getLogoName }}</span>
      <switchMobile />
    </div>

    <template v-if="props.mode === E_headerMode.Index">
      <div ref="modernHeaderSearchRef" class="modern-header-search">
        <searchFavoritePopover class="modern-header-favorite" />
        <searchInputTagByStore class="modern-header-search-input" />
      </div>

      <nav class="modern-header-actions" aria-label="主工具栏">
        <button v-admin class="modern-header-action" type="button" title="添加资源" @click="openResourceForm">
          <el-icon><Plus /></el-icon>
          <span>添加资源</span>
        </button>

        <button v-admin class="modern-header-action" :class="{ active: store.appStoreData.adminResourceStatus }"
          type="button" :title="store.appStoreData.adminResourceStatus ? '关闭管理模式' : '开启管理模式'"
          @click="switchAdminStatus">
          <el-icon><Operation /></el-icon>
          <span>管理模式</span>
        </button>

        <button v-admin class="modern-header-action" type="button" title="标签管理" @click="openTagList">
          <el-icon><PriceTag /></el-icon>
          <span>标签管理</span>
        </button>

        <button class="modern-header-action" type="button" :title="appLang.performer()" @click="goToPerformer">
          <el-icon><User /></el-icon>
          <span>{{ appLang.performer() }}</span>
        </button>

        <button v-admin class="modern-header-action" type="button" title="导入与刮削" @click="openImportResource">
          <el-icon><Magnet /></el-icon>
          <span>导入刮削</span>
        </button>

        <button v-admin class="modern-header-action" type="button" title="检测更新" @click="openCheckUpdateSoft">
          <el-icon><Refresh /></el-icon>
          <span>检测更新</span>
        </button>

        <el-dropdown class="modern-header-home-mode" trigger="click" @command="changeHomeMode">
          <button class="modern-header-action modern-header-action--dropdown" type="button"
            :title="`主页样式：${currentHomeModeLabel}`">
            <el-icon><Monitor /></el-icon>
            <span>{{ currentHomeModeLabel }}</span>
            <el-icon class="modern-header-chevron"><ArrowDown /></el-icon>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-for="item in homeModeOptions" :key="item.value" :command="item.value"
                :disabled="item.value === currentHomeMode">
                {{ item.label }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>

        <button v-admin class="modern-header-action" type="button" title="设置" @click="goToSetting">
          <el-icon><Setting /></el-icon>
          <span>设置</span>
        </button>

        <button v-if="store.appStoreData.isAdminLogin && !store.appStoreData.isAdminLoginStatus"
          class="modern-header-action" type="button" title="管理员登录" @click="goToAdminLogin">
          <el-icon><UserFilled /></el-icon>
          <span>登录</span>
        </button>
      </nav>
    </template>

    <nav v-else class="modern-header-actions modern-header-actions--sub" aria-label="页面导航">
      <button class="modern-header-action" type="button" title="主页" @click="router.push('/')">
        <el-icon><HomeFilled /></el-icon>
        <span>主页</span>
      </button>
      <button class="modern-header-action" type="button" title="播放列表" @click="openPlayListHandle">
        <el-icon><Memo /></el-icon>
        <span>播放列表</span>
      </button>
      <button class="modern-header-action" type="button" title="返回" @click="router.go(-1)">
        <el-icon><Back /></el-icon>
        <span>返回</span>
      </button>
    </nav>

    <tagListDrawer ref="tagListDrawerRef" />
    <resourceFormDrawer ref="resourceFormDrawerRef" @success="createResouceSuccessHandle" />
    <importResourceDrawer ref="importResourceDrawerRef" @success="createResouceSuccessHandle" />
    <updateSoftDialog ref="updateSoftDialogRef" />
    <playListDrawer ref="playListDrawerRef" />
  </header>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, type PropType } from 'vue'
import { useRouter } from 'vue-router'
import { E_headerMode } from '@/dataType/app.dataType'
import type { I_resource } from '@/dataType/resource.dataType'
import { appStoreData } from '@/storeData/app.storeData'
import { AppLang } from '@/language/app.lang'
import { useHomeMode, type T_homeMode } from '@/common/homeMode'
import switchMobile from '@/components/com/form/switchMobile.vue'
import searchFavoritePopover from '@/components/search/searchFavoritePopover.vue'
import searchInputTagByStore from '@/components/com/form/searchInputTagByStore.vue'
import tagListDrawer from '@/components/tag/tagListDrawer.vue'
import resourceFormDrawer from '@/components/resource/resourceFormDrawer.vue'
import importResourceDrawer from '@/components/importResource/importResourceDrawer.vue'
import updateSoftDialog from '@/components/setting/updateSoft/updateSoftDialog.vue'
import playListDrawer from '@/components/playList/playListDrawer.vue'

const props = defineProps({
  mode: {
    type: String as PropType<E_headerMode>,
    default: E_headerMode.Index,
  },
})

const emits = defineEmits<{
  createResouceSuccess: [resource: I_resource]
}>()

const router = useRouter()
const appLang = AppLang()
const store = {
  appStoreData: appStoreData(),
}
const { currentHomeMode, currentHomeModeLabel, homeModeOptions, setHomeMode } = useHomeMode()

const tagListDrawerRef = ref<InstanceType<typeof tagListDrawer>>()
const resourceFormDrawerRef = ref<InstanceType<typeof resourceFormDrawer>>()
const importResourceDrawerRef = ref<InstanceType<typeof importResourceDrawer>>()
const updateSoftDialogRef = ref<InstanceType<typeof updateSoftDialog>>()
const playListDrawerRef = ref<InstanceType<typeof playListDrawer>>()
const modernHeaderSearchRef = ref<HTMLElement>()
const searchWrapped = ref(false)
let searchResizeObserver: ResizeObserver | undefined

const updateSearchWrapped = () => {
  const inputTag = modernHeaderSearchRef.value?.querySelector<HTMLElement>('.el-input-tag__wrapper')
  searchWrapped.value = (inputTag?.getBoundingClientRect().height || 0) > 42
}

const openResourceForm = () => resourceFormDrawerRef.value?.open('add')
const switchAdminStatus = () => {
  store.appStoreData.adminResourceStatus = !store.appStoreData.adminResourceStatus
}
const openTagList = () => tagListDrawerRef.value?.open()
const openImportResource = () => importResourceDrawerRef.value?.open()
const openCheckUpdateSoft = () => updateSoftDialogRef.value?.open()
const openPlayListHandle = () => playListDrawerRef.value?.open()
const changeHomeMode = (mode: T_homeMode) => setHomeMode(mode)

const goToPerformer = () => {
  if (store.appStoreData.currentPerformerBasesIds.length > 1) {
    router.push(`/performer/basesList/${store.appStoreData.currentFilesBases.id}`)
    return
  }
  router.push(`/performer/${store.appStoreData.currentMainPerformerBasesId}`)
}

const goToAdminLogin = () => router.push('/adminLogin')
const goToSetting = () => router.push('/setting')
const createResouceSuccessHandle = (resource: I_resource) => {
  emits('createResouceSuccess', resource)
}

onMounted(async () => {
  await nextTick()
  const inputTag = modernHeaderSearchRef.value?.querySelector<HTMLElement>('.el-input-tag__wrapper')
  if (!inputTag) return
  searchResizeObserver = new ResizeObserver(updateSearchWrapped)
  searchResizeObserver.observe(inputTag)
  updateSearchWrapped()
})

onBeforeUnmount(() => {
  searchResizeObserver?.disconnect()
})
</script>

<style scoped lang="scss">
.modern-header {
  --header-panel: #0d1d29;
  --header-panel-hover: #122a38;
  --header-border: rgba(137, 174, 196, 0.15);
  --header-text: #edf6fb;
  --header-muted: #9eb0bc;
  --header-accent: #25b5b3;
  width: 100%;
  min-width: 0;
  min-height: 54px;
  height: auto;
  padding: 5px 10px !important;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  align-content: center;
  gap: 16px;
  box-sizing: border-box;
  border: 0 !important;
  border-bottom: 1px solid var(--header-border) !important;
  border-radius: 0 !important;
  color: var(--header-text);
  background: transparent !important;
}

.modern-header-brand {
  width: 188px;
  min-height: 42px;
  min-width: 0;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
  color: var(--header-text);
  font-size: 15px;
  font-weight: 650;
  white-space: nowrap;

  img {
    width: 24px;
    height: 24px;
    flex-shrink: 0;
  }

  span {
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.modern-header-search {
  min-width: 220px;
  min-height: 42px;
  margin-left: auto;
  flex: 1 1 460px;
  display: flex;
  align-items: center;
  gap: 7px;
}

.modern-header-favorite {
  height: 32px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.modern-header-search-input {
  min-width: 0;
  flex: 1;
}

.modern-header.is-search-wrapped {
  align-items: flex-start;
  align-content: flex-start;

  .modern-header-search {
    margin-top: -3px;
    align-items: flex-start;
  }
}

.modern-header-actions {
  min-width: 0;
  display: flex;
  align-items: stretch;
  justify-content: flex-end;
  gap: 2px;
}

.modern-header-actions--sub {
  flex: 1;
}

.modern-header-action {
  min-width: 54px;
  height: 42px;
  padding: 3px 6px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  border: 0;
  border-radius: 5px;
  color: var(--header-muted);
  background: transparent;
  font: inherit;
  cursor: pointer;
  transition: color 0.16s ease, background-color 0.16s ease;

  > .el-icon:first-child {
    font-size: 17px;
  }

  > span {
    max-width: 66px;
    overflow: hidden;
    font-size: 10px;
    line-height: 13px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &:hover,
  &:focus-visible {
    color: var(--header-text);
    background: var(--header-panel-hover);
    outline: none;
  }

  &.active {
    color: var(--header-accent);
    background: rgba(37, 181, 179, 0.11);
  }
}

.modern-header-action--dropdown {
  min-width: 66px;
}

.modern-header-chevron {
  position: absolute;
  right: 2px;
  bottom: 5px;
  font-size: 9px;
}

.modern-header-home-mode {
  flex-shrink: 0;

  .modern-header-action {
    position: relative;
  }
}

:global(.bright .modern-header) {
  --header-panel: #f5f8fa;
  --header-panel-hover: #eef4f7;
  --header-border: #dfe7ec;
  --header-text: #263540;
  --header-muted: #667883;
  --header-accent: #159a9a;
  box-shadow: none;
}

:global(.studio-home .modern-header) {
  border: 1px solid var(--home-border) !important;
  border-radius: 8px !important;
  background: var(--home-panel-bg) !important;
  box-shadow: none;
}

@media (max-width: 1280px) {
  .modern-header {
    gap: 9px;
  }

  .modern-header-brand {
    width: 158px;
  }

  .modern-header-action {
    min-width: 36px;
    width: 36px;

    > span {
      display: none;
    }
  }

  .modern-header-action--dropdown {
    min-width: 40px;
    width: 40px;
  }
}

@media (max-width: 820px) {
  .modern-header-brand {
    width: auto;

    > span {
      display: none;
    }
  }

  .modern-header-search {
    min-width: 120px;
  }
}
</style>
