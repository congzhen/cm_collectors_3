<template>
  <div class="mobile-view" :class="{ 'mobile-view--bright': isBrightTheme }" v-loading="loading">
    <header class="mobile-toolbar">
      <div class="mobile-toolbar-row mobile-database-row">
        <el-select v-model="selectedDataBaseId" class="mobile-database-select" aria-label="选择资源库"
          @change="changeDataBase">
          <el-option v-for="filesBase in store.filesBasesStoreData.filesBasesStatus" :key="filesBase.id"
            :label="filesBase.name" :value="filesBase.id" />
        </el-select>
        <el-button class="mobile-desktop-button" @click="switchToDesktop">
          <el-icon><Monitor /></el-icon>
          <span>桌面版</span>
        </el-button>
      </div>

      <div class="mobile-toolbar-row mobile-search-row">
        <el-input v-model="searchText" class="mobile-search-input" clearable placeholder="搜索资源、标签、演员…"
          aria-label="搜索资源" @keyup.enter="applySearch" @clear="applySearch">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button class="mobile-filter-button" :class="{ active: activeFilterCount > 0 }"
          @click="filterDrawerVisible = true">
          <el-icon><Filter /></el-icon>
          <span>筛选</span>
          <span v-if="activeFilterCount > 0" class="mobile-filter-count">{{ activeFilterCount }}</span>
        </el-button>
      </div>
    </header>

    <main ref="contentScrollerRef" class="mobile-content"
      :class="{ 'mobile-content--short-video': isShortVideoMode }">
      <MobileResourceLayout :data-list="dataList" :mode="resourceShowMode"
        @select-resource="selectResourceHandle" />
    </main>

    <footer class="mobile-pagination" aria-label="资源分页">
      <el-button class="mobile-page-button" circle :disabled="currentPage <= 1" aria-label="上一页"
        @click="prevPage">
        <el-icon><ArrowLeft /></el-icon>
      </el-button>
      <div class="mobile-page-status">
        <strong>{{ currentPage }}</strong>
        <span>/ {{ totalPages }}</span>
      </div>
      <el-button class="mobile-page-button" circle :disabled="currentPage >= totalPages" aria-label="下一页"
        @click="nextPage">
        <el-icon><ArrowRight /></el-icon>
      </el-button>
      <span class="mobile-total-count">共 {{ dataCount }} 项</span>
    </footer>

    <el-drawer v-model="filterDrawerVisible" title="筛选" direction="rtl" size="88%" append-to-body
      class="mobile-filter-drawer">
      <TagView :show-modern-header="false" />
      <template #footer>
        <div class="mobile-filter-actions">
          <el-button @click="clearFilters">清除筛选</el-button>
          <el-button type="primary" @click="filterDrawerVisible = false">查看结果</el-button>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ArrowLeft, ArrowRight, Filter, Monitor, Search } from '@element-plus/icons-vue'
import MobileResourceLayout from '@/components/mobile/MobileResourceLayout.vue'
import TagView from './TagView.vue'
import { appStoreData } from '@/storeData/app.storeData'
import { filesBasesStoreData } from '@/storeData/filesBases.storeData'
import { searchStoreData } from '@/storeData/search.storeData'
import type { I_resource } from '@/dataType/resource.dataType'
import type { T_resourcesShowMode } from '@/dataType/app.dataType'
import { resourceServer } from '@/server/resource.server'
import { debounce } from '@/assets/debounce'
import { ElMessage } from 'element-plus'
import { playResource } from '@/common/play'
import { goToMobileOrPC, setMobileShow } from '@/assets/mobile'

const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
  filesBasesStoreData: filesBasesStoreData(),
}

let fetchCount = true
let durationRefreshTimer: number | undefined
let dataListRequestVersion = 0

const loading = ref(false)
const dataList = ref<I_resource[]>([])
const dataCount = ref(0)
const currentPage = ref(1)
const normalizedPageSize = () => Math.max(1, Number(store.appStoreData.currentConfigApp.pageLimit) || 20)
const pageSize = ref(normalizedPageSize())
const selectedDataBaseId = ref(store.appStoreData.currentFilesBases.id)
const searchText = ref(store.searchStoreData.searchData.searchTextSlc.join(' '))
const filterDrawerVisible = ref(false)
const contentScrollerRef = ref<HTMLElement>()

const isBrightTheme = computed(() => store.appStoreData.appConfig.theme === 'bright')
const resourceShowMode = computed<T_resourcesShowMode>(() =>
  store.appStoreData.currentConfigApp.resourcesShowMode || 'coverPoster'
)
const isShortVideoMode = computed(() => [
  'coverPosterMosaicShortVideo',
  'shortVideo',
  'shortVideoTopBottom',
].includes(resourceShowMode.value))
const totalPages = computed(() => Math.max(1, Math.ceil(dataCount.value / pageSize.value)))
const activeFilterCount = computed(() => {
  const searchData = store.searchStoreData.searchData
  const fixedGroups = [
    searchData.country,
    searchData.definition,
    searchData.videoCodec,
    searchData.year,
    searchData.star,
    searchData.performer,
    searchData.cup,
  ]
  const fixedCount = fixedGroups.reduce((total, group) => total + group.options.length, 0)
  const tagCount = Object.values(searchData.tag)
    .reduce((total, group) => total + group.options.length, 0)
  return fixedCount + tagCount
})

watch(
  () => store.appStoreData.currentFilesBases.id,
  (id) => {
    selectedDataBaseId.value = id
    dataListRequestVersion++
    window.clearTimeout(durationRefreshTimer)
    dataList.value = []
    dataCount.value = 0
  },
  { flush: 'sync' },
)

watch(
  () => store.searchStoreData.searchData,
  () => {
    fetchCount = true
    currentPage.value = 1
    getDataList(scrollContentToTop)
  },
  { deep: true },
)

const executeGetDataList = debounce(async (requestVersion: number, callback: () => void) => {
  const filesBasesId = store.appStoreData.currentFilesBases.id
  const shouldFetchCount = fetchCount
  try {
    loading.value = true
    const result = await resourceServer.dataList(
      filesBasesId,
      shouldFetchCount,
      currentPage.value,
      pageSize.value,
      store.searchStoreData.searchData,
    )
    if (requestVersion !== dataListRequestVersion || filesBasesId !== store.appStoreData.currentFilesBases.id) return
    if (result && result.status) {
      dataList.value = result.data.dataList
      if (shouldFetchCount) {
        dataCount.value = result.data.total
        fetchCount = false
      }
      scheduleDurationRefresh()
      callback()
    } else {
      ElMessage.error(result?.msg || '资源加载失败')
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('资源加载失败')
  } finally {
    if (requestVersion === dataListRequestVersion) loading.value = false
  }
}, 200)

const getDataList = (callback: () => void = () => undefined) => {
  const requestVersion = ++dataListRequestVersion
  executeGetDataList(requestVersion, callback)
}

const scheduleDurationRefresh = () => {
  if (!store.appStoreData.currentConfigApp.showVideoDuration) return
  window.clearTimeout(durationRefreshTimer)
  const requestVersion = dataListRequestVersion
  const filesBasesId = store.appStoreData.currentFilesBases.id
  durationRefreshTimer = window.setTimeout(async () => {
    if (requestVersion !== dataListRequestVersion || filesBasesId !== store.appStoreData.currentFilesBases.id) return
    const result = await resourceServer.dataList(
      filesBasesId,
      false,
      currentPage.value,
      pageSize.value,
      store.searchStoreData.searchData,
    )
    if (requestVersion === dataListRequestVersion
      && filesBasesId === store.appStoreData.currentFilesBases.id
      && result?.status) {
      dataList.value = result.data.dataList
    }
  }, 2500)
}

const scrollContentToTop = () => contentScrollerRef.value?.scrollTo({ top: 0, behavior: 'smooth' })

const prevPage = () => {
  if (currentPage.value <= 1) return
  currentPage.value--
  getDataList(scrollContentToTop)
}

const nextPage = () => {
  if (currentPage.value >= totalPages.value) return
  currentPage.value++
  getDataList(scrollContentToTop)
}

const selectResourceHandle = (resource: I_resource) => playResource(resource)

const changeDataBase = async (selectedId: string) => {
  if (!selectedId || selectedId === store.appStoreData.currentFilesBases.id) return
  loading.value = true
  try {
    const result = await store.appStoreData.initCurrentFilesBases(selectedId)
    if (result && !result.status) {
      selectedDataBaseId.value = store.appStoreData.currentFilesBases.id
      ElMessage.error(result.message)
      return
    }
    store.searchStoreData.init()
    searchText.value = ''
    fetchCount = true
    currentPage.value = 1
    pageSize.value = normalizedPageSize()
    getDataList(scrollContentToTop)
  } finally {
    loading.value = false
  }
}

const applySearch = () => {
  const text = searchText.value.trim()
  const nextValue = text ? [text] : []
  if (JSON.stringify(nextValue) === JSON.stringify(store.searchStoreData.searchData.searchTextSlc)) return
  store.searchStoreData.searchData.searchTextSlc = nextValue
}

const clearFilters = () => {
  store.searchStoreData.clear()
  searchText.value = ''
}

const switchToDesktop = () => {
  setMobileShow(true)
  goToMobileOrPC()
}

onMounted(() => getDataList())

onBeforeUnmount(() => {
  dataListRequestVersion++
  window.clearTimeout(durationRefreshTimer)
})
</script>

<style scoped lang="scss">
.mobile-view {
  --mobile-bg: #121619;
  --mobile-surface: #1b2024;
  --mobile-surface-raised: #242a2f;
  --mobile-border: rgba(255, 255, 255, 0.11);
  --mobile-text: #edf1f3;
  --mobile-muted: #9ba5ab;
  --mobile-accent: #27bfc1;
  width: calc(100% + 10px);
  height: calc(100% + 10px);
  margin: -5px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: var(--mobile-text);
  background: var(--mobile-bg);
}

.mobile-view--bright {
  --mobile-bg: #f5f7f8;
  --mobile-surface: #ffffff;
  --mobile-surface-raised: #f3f6f7;
  --mobile-border: #dce3e7;
  --mobile-text: #23323b;
  --mobile-muted: #75838c;
  --mobile-accent: #159fa3;
}

.mobile-toolbar {
  z-index: 2;
  flex: 0 0 auto;
  display: grid;
  gap: 9px;
  padding: calc(10px + env(safe-area-inset-top)) 12px 10px;
  background: var(--mobile-surface);
  border-bottom: 1px solid var(--mobile-border);
  box-shadow: 0 5px 18px rgba(0, 0, 0, 0.12);
}

.mobile-toolbar-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.mobile-database-select,
.mobile-search-input {
  min-width: 0;
  flex: 1;
}

.mobile-desktop-button,
.mobile-filter-button {
  flex: 0 0 auto;
  height: 36px;
  color: var(--mobile-text);
  background: var(--mobile-surface-raised);
  border-color: var(--mobile-border);
}

.mobile-filter-button.active {
  color: var(--mobile-accent);
  border-color: color-mix(in srgb, var(--mobile-accent) 65%, transparent);
  background: color-mix(in srgb, var(--mobile-accent) 12%, var(--mobile-surface));
}

.mobile-filter-count {
  display: inline-grid;
  min-width: 18px;
  height: 18px;
  place-items: center;
  padding: 0 5px;
  margin-left: 1px;
  color: #07383a;
  font-size: 11px;
  font-weight: 700;
  line-height: 18px;
  background: var(--mobile-accent);
  border-radius: 9px;
}

.mobile-content {
  min-height: 0;
  flex: 1;
  overflow: auto;
  padding: 12px;
  overscroll-behavior: contain;
  scrollbar-width: none;
  -webkit-overflow-scrolling: touch;
}

.mobile-content::-webkit-scrollbar {
  display: none;
}

.mobile-content--short-video {
  overflow: hidden;
  padding: 8px 10px;

  :deep(.mobile-resource-layout) {
    height: 100%;
    min-height: 0;
  }
}

.mobile-pagination {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 48px;
  padding: 6px 12px calc(6px + env(safe-area-inset-bottom));
  background: var(--mobile-surface);
  border-top: 1px solid var(--mobile-border);
}

.mobile-page-button {
  color: var(--mobile-text);
  background: var(--mobile-surface-raised);
  border-color: var(--mobile-border);
}

.mobile-page-status {
  display: flex;
  align-items: baseline;
  gap: 4px;
  min-width: 58px;
  justify-content: center;

  strong {
    color: var(--mobile-accent);
    font-size: 16px;
  }

  span {
    color: var(--mobile-muted);
    font-size: 12px;
  }
}

.mobile-total-count {
  position: absolute;
  right: 12px;
  color: var(--mobile-muted);
  font-size: 11px;
}

:deep(.mobile-database-select .el-select__wrapper),
:deep(.mobile-search-input .el-input__wrapper) {
  min-height: 36px;
  color: var(--mobile-text);
  background: var(--mobile-surface-raised);
  box-shadow: 0 0 0 1px var(--mobile-border) inset;
}

:deep(.mobile-database-select .el-select__selected-item),
:deep(.mobile-search-input .el-input__inner) {
  color: var(--mobile-text);
}

:deep(.mobile-search-input .el-input__inner::placeholder) {
  color: var(--mobile-muted);
}

:global(.mobile-filter-drawer) {
  --el-drawer-bg-color: #1b2024;
  --el-text-color-primary: #edf1f3;
  --el-border-color-lighter: rgba(255, 255, 255, 0.11);
}

:global(.bright .mobile-filter-drawer) {
  --el-drawer-bg-color: #ffffff;
  --el-text-color-primary: #23323b;
  --el-border-color-lighter: #dce3e7;
}

:global(.mobile-filter-drawer .el-drawer__header) {
  padding: calc(15px + env(safe-area-inset-top)) 16px 12px;
  margin: 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

:global(.mobile-filter-drawer .el-drawer__body) {
  padding: 0;
  overflow: hidden;
}

:global(.mobile-filter-drawer .el-drawer__footer) {
  padding: 10px 12px calc(10px + env(safe-area-inset-bottom));
  border-top: 1px solid var(--el-border-color-lighter);
}

:global(.mobile-filter-drawer .tag-container) {
  position: relative !important;
  left: auto !important;
  width: 100% !important;
  height: 100% !important;
  margin: 0 !important;
  border: 0 !important;
  border-radius: 0 !important;
  box-shadow: none !important;
}

:global(.mobile-filter-drawer .tag-container .arrow) {
  display: none !important;
}

.mobile-filter-actions {
  display: grid;
  grid-template-columns: 1fr 1.4fr;
  gap: 8px;

  .el-button {
    width: 100%;
    margin: 0;
  }
}

@media (max-width: 360px) {
  .mobile-total-count {
    display: none;
  }

  .mobile-desktop-button,
  .mobile-filter-button {
    padding-inline: 11px;
  }
}
</style>
