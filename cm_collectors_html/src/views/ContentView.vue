<template>
  <div class="content-view" v-loading="loading">
    <div class="list">
      <contentList ref="contentListRef" v-if="!store.appStoreData.adminResourceStatus" :data-list="dataList" :group-page="groupPage"
        @select-resources="selectResourcesHandle"></contentList>
      <contentListAdmin ref="contentListAdminRef" v-else :data-list="dataList"
        @update-data="init_DataList(() => {}, true)"></contentListAdmin>
    </div>
    <div class="paging" :class="{
      'paging--modern': isModernAppearance,
      'paging--bright': isModernAppearance && isBrightTheme,
    }">
      <div class="paging-main">
        <div class="paging-summary">
          <span>{{ isModernAppearance ? `共 ${dataCount} 个资源` : `Total ${dataCount}` }}</span>
          <span v-if="groupingActive" class="group-page-count">本页展示 {{ groupedPageCount }} 项</span>
          <el-popover trigger="hover" :width="260" :show-after="250" @before-enter="handleFileSizeStatsShow">
            <template #reference>
              <el-icon class="file-size-stats-trigger" title="当前筛选结果的视频文件统计">
                <InfoFilled />
              </el-icon>
            </template>
            <div v-if="fileSizeStatsLoading" class="file-size-stats-state">
              <span class="file-size-stats-spinner"></span>
              <span>正在统计当前筛选结果…</span>
            </div>
            <div v-else-if="fileSizeStatsError" class="file-size-stats-state file-size-stats-state--error">
              <span>{{ fileSizeStatsError }}</span>
              <el-button link type="primary" size="small" @click="loadFileSizeStats(true)">重新统计</el-button>
            </div>
            <div v-else-if="fileSizeStats" class="file-size-stats">
              <div class="file-size-stats__title">当前筛选结果</div>
              <div>
                <span>已统计文件大小</span>
                <strong>{{ formatFileSize(fileSizeStats.totalSize) }}</strong>
              </div>
              <div>
                <span>视频文件</span>
                <strong>{{ fileSizeStats.totalFiles }}</strong>
              </div>
              <div>
                <span>已统计</span>
                <strong>{{ fileSizeStats.countedFiles }}</strong>
              </div>
              <div>
                <span>未统计 / 待更新</span>
                <strong>{{ fileSizeStats.uncountedFiles }}</strong>
              </div>
            </div>
            <div v-else class="file-size-stats-state">将鼠标停留片刻后开始统计</div>
          </el-popover>
        </div>
        <el-pagination background layout="prev, pager, next, jumper" v-model:current-page="currentPage"
          :total="dataCount" :page-size="pageSize" :pager-count="5" size="small" @change="changePageHandle" />
      </div>
      <div class="bottom-btns">
        <el-tooltip v-if="!isMobile()" :content="store.appStoreData.adminResourceStatus ? '管理模式按单个资源展示' : store.appStoreData.currentConfigApp.resourcesShowMode === 'table' ? '表格模式按单个资源展示' : '只合并当前页的同组资源，分页仍按资源计算'">
          <div v-admin class="group-display" role="group" aria-label="资源分组显示方式" :aria-busy="savingGroupDisplay">
            <button type="button" :class="{ 'is-active': store.appStoreData.currentConfigApp.mergeResourceGroups === false }"
              :aria-pressed="store.appStoreData.currentConfigApp.mergeResourceGroups === false" :disabled="groupDisplayDisabled" @click="setGroupDisplay(false)">
              <el-icon><Grid /></el-icon><span>独立显示</span>
            </button>
            <button type="button" :class="{ 'is-active': store.appStoreData.currentConfigApp.mergeResourceGroups !== false }"
              :aria-pressed="store.appStoreData.currentConfigApp.mergeResourceGroups !== false" :disabled="groupDisplayDisabled" @click="setGroupDisplay(true)">
              <el-icon :class="{ 'is-loading': savingGroupDisplay }"><Loading v-if="savingGroupDisplay" /><Collection v-else /></el-icon><span>合并同组</span>
            </button>
          </div>
        </el-tooltip>
        <LibraryMomentsButton :files-bases-id="store.appStoreData.currentFilesBases.id || ''" />
        <playListBtn :modern="isModernAppearance"></playListBtn>
        <coverAdjuster v-admin></coverAdjuster>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import contentList from '@/components/content/contentList.vue'
import LibraryMomentsButton from '@/components/play/LibraryMomentsButton.vue'
import contentListAdmin from '@/components/content/contentListAdmin.vue';
import coverAdjuster from '@/components/setting/fileDatabaseSetting/coverAdjuster.vue';
import playListBtn from '@/components/playList/playListBtn.vue';
import { ref, onMounted, onUnmounted, watch, nextTick, computed } from 'vue'
import { appStoreData } from '@/storeData/app.storeData';
import { searchStoreData } from '@/storeData/search.storeData';
import type { I_resource, I_resourceFileSizeStats } from '@/dataType/resource.dataType';
import { ElMessage } from 'element-plus';
import { Collection, Grid, Loading } from '@element-plus/icons-vue';
import { resourceServer } from '@/server/resource.server';
import { debounce } from '@/assets/debounce';
import { E_searchSort } from '@/dataType/search.dataType';
import { eventBus } from '@/main';
import { isMobile } from '@/assets/mobile';
import { resourceGroupServer } from '@/server/resourceGroup.server';
import { mergeResourceGroupPage, resourceGroupPageKey, type ResourceGroupPage } from '@/components/content/resourceGroups';
const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
}
const isModernAppearance = computed(() =>
  (store.appStoreData.appConfig.appearanceStyle || store.appStoreData.appConfig.headerStyle || 'modern') === 'modern'
);
const isBrightTheme = computed(() => store.appStoreData.appConfig.theme === 'bright');
const emits = defineEmits(['selectResources']);

const contentListAdminRef = ref<InstanceType<typeof contentListAdmin>>();
const contentListRef = ref<InstanceType<typeof contentList>>();


const isInitializing = ref(false);
const loading = ref(false);
const dataList = ref<I_resource[]>([]);
const groupPage = ref<ResourceGroupPage>();
const dataCount = ref(0);
const savingGroupDisplay = ref(false);
const groupDisplayDisabled = computed(() => store.appStoreData.adminResourceStatus
  || store.appStoreData.currentConfigApp.resourcesShowMode === 'table' || savingGroupDisplay.value);
const groupingActive = computed(() => !isMobile() && !store.appStoreData.adminResourceStatus
  && store.appStoreData.currentConfigApp.resourcesShowMode !== 'table' && store.appStoreData.currentConfigApp.mergeResourceGroups !== false);
const groupedPageCount = computed(() => mergeResourceGroupPage(dataList.value, groupingActive.value).length);
const setGroupDisplay = async (value: string | number | boolean) => {
  if (groupDisplayDisabled.value || !!value === (store.appStoreData.currentConfigApp.mergeResourceGroups !== false)) return;
  const library = store.appStoreData.currentFilesBases.id;
  const config = store.appStoreData.currentConfigApp;
  const enabled = !!value;
  savingGroupDisplay.value = true;
  try {
    const result = await resourceGroupServer.setDisplay(library, enabled);
    if (!result.status) { ElMessage.error(result.msg || '分组显示设置保存失败'); return; }
    if (library === store.appStoreData.currentFilesBases.id && config === store.appStoreData.currentConfigApp) config.mergeResourceGroups = enabled;
  } catch { ElMessage.error('分组显示设置保存失败，请重试'); }
  finally { savingGroupDisplay.value = false; }
};
const fileSizeStats = ref<I_resourceFileSizeStats | null>(null);
const fileSizeStatsLoading = ref(false);
const fileSizeStatsError = ref('');
let fileSizeStatsCacheKey = '';
let fileSizeStatsRequestVersion = 0;
let fetchCount = true;
let durationRefreshTimer: number | undefined;
let dataListRequestVersion = 0;
let randomSeed = '';
const currentPage = ref(1);
const pageSize = ref(store.appStoreData.currentConfigApp.pageLimit);

watch(
  () => store.searchStoreData.searchData,
  () => {
    randomSeed = store.searchStoreData.searchData.sort === E_searchSort.Random ? createRandomSeed() : '';
    init();
  },
  { deep: true }
)

watch(
  () => store.appStoreData.currentConfigApp.pageLimit,
  () => {
    init();
  },
)

watch(
  () => store.appStoreData.currentFilesBases.id,
  () => {
    dataListRequestVersion++;
    randomSeed = '';
    window.clearTimeout(durationRefreshTimer);
    dataList.value = [];
    dataCount.value = 0;
    resetFileSizeStats();
  },
  { flush: 'sync' }
)

const init = async () => {
  isInitializing.value = true;
  dataList.value = [];
  dataCount.value = 0;
  resetFileSizeStats();
  fetchCount = true;
  currentPage.value = 1;
  pageSize.value = store.appStoreData.currentConfigApp.pageLimit;

  await init_DataList(() => {
    let firstData = undefined;
    if (dataList.value.length > 0) {
      firstData = dataList.value[0];
    }
    emits('selectResources', firstData, true);
    isInitializing.value = false;
    nextTick(() => contentListRef.value?.change());
  });

}

const init_DataList = async (fn: () => void = () => { }, fetch: boolean = false) => {
  if (fetch) {
    resetFileSizeStats();
    fetchCount = true;
  }
  await getDataList(fn);
}

const createRandomSeed = () => `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;

const currentRandomSeed = () => {
  if (store.searchStoreData.searchData.sort !== E_searchSort.Random) return '';
  if (!randomSeed) randomSeed = createRandomSeed();
  return randomSeed;
}

const prepareGroupPage = async (resources: I_resource[], library: string): Promise<ResourceGroupPage | undefined> => {
  if (!groupingActive.value) return undefined;
  const page: ResourceGroupPage = { key: resourceGroupPageKey(library, resources), summaries: [] };
  const ids = [...new Set(resources.map(resource => resource.associationGroupId).filter((id): id is string => !!id))];
  if (!ids.length) return page;
  try {
    const result = await resourceGroupServer.summaries(library, ids);
    if (result.status) page.summaries = result.data.dataList;
  } catch { /* 摘要不可用时仍展示资源，组内成员可通过悬停加载。 */ }
  return page;
};

const executeGetDataList = debounce(async (requestVersion: number, fn: () => void) => {
  if (requestVersion !== dataListRequestVersion) return;
  const filesBasesId = store.appStoreData.currentFilesBases.id;
  const shouldFetchCount = fetchCount;
  try {
    loading.value = true;
    const result = await resourceServer.dataList(filesBasesId, shouldFetchCount, currentPage.value, pageSize.value, store.searchStoreData.searchData, currentRandomSeed());
    if (requestVersion !== dataListRequestVersion || filesBasesId !== store.appStoreData.currentFilesBases.id) return;
    if (result && result.status) {
      const nextGroupPage = await prepareGroupPage(result.data.dataList, filesBasesId);
      if (requestVersion !== dataListRequestVersion || filesBasesId !== store.appStoreData.currentFilesBases.id) return;
      groupPage.value = nextGroupPage;
      dataList.value = result.data.dataList;
      if (shouldFetchCount) {
        dataCount.value = result.data.total;
        fetchCount = false;
      }
      scheduleDurationRefresh();
      fn();
    } else {
      ElMessage.error(result.msg);
    }
  } catch (error) {
    console.log(error);
  } finally {
    if (requestVersion === dataListRequestVersion) {
      loading.value = false;
      // Failure or a newer refresh may bypass init's success callback.
      // The latest completed request must always release pagination.
      isInitializing.value = false;
    }
  }
}, 200)

const getDataList = (fn: () => void = () => { }) => {
  const requestVersion = ++dataListRequestVersion;
  executeGetDataList(requestVersion, fn);
}

// 资源列表接口只负责触发后端异步采集，不会等待 ffprobe 完成。
// 开启“显示视频时长”后，这里延迟重新拉取当前页一次，让刚写入数据库的时长能自动出现在封面角标上。
// 关闭开关时不做额外请求，避免影响默认浏览体验。
const scheduleDurationRefresh = () => {
  if (!store.appStoreData.currentConfigApp.showVideoDuration) {
    return;
  }
  window.clearTimeout(durationRefreshTimer);
  const requestVersion = dataListRequestVersion;
  const filesBasesId = store.appStoreData.currentFilesBases.id;
  durationRefreshTimer = window.setTimeout(async () => {
    if (!store.appStoreData.currentConfigApp.showVideoDuration
      || requestVersion !== dataListRequestVersion || filesBasesId !== store.appStoreData.currentFilesBases.id) return;
    try {
      const result = await resourceServer.dataList(filesBasesId, false, currentPage.value, pageSize.value, store.searchStoreData.searchData, currentRandomSeed());
      if (requestVersion === dataListRequestVersion && filesBasesId === store.appStoreData.currentFilesBases.id && result && result.status) {
        // This poll only fills in asynchronously collected durations. Replacing the
        // whole list remounts covers and clears group summaries 2.5s after paging.
        const refreshed = new Map(result.data.dataList.map(resource => [resource.id, resource]));
        for (const resource of dataList.value) {
          const durations = new Map((refreshed.get(resource.id)?.dramaSeries || []).map(series => [series.id, series.durationSeconds]));
          for (const series of resource.dramaSeries || []) {
            const duration = durations.get(series.id);
            if (typeof duration === 'number') series.durationSeconds = duration;
          }
        }
      }
    } catch { /* 后台补采失败保留当前画面，正常刷新时再尝试采集时长。 */ }
  }, 2500);
}

const resetFileSizeStats = () => {
  fileSizeStatsRequestVersion++;
  fileSizeStats.value = null;
  fileSizeStatsLoading.value = false;
  fileSizeStatsError.value = '';
  fileSizeStatsCacheKey = '';
}

const currentFileSizeStatsKey = () => JSON.stringify({
  filesBasesId: store.appStoreData.currentFilesBases.id,
  searchData: store.searchStoreData.searchData,
});

const handleFileSizeStatsShow = () => {
  loadFileSizeStats();
}

const loadFileSizeStats = async (force = false) => {
  const filesBasesId = store.appStoreData.currentFilesBases.id;
  if (!filesBasesId) return;
  const cacheKey = currentFileSizeStatsKey();
  if (!force && (fileSizeStatsLoading.value || (fileSizeStatsCacheKey === cacheKey && fileSizeStats.value))) {
    return;
  }

  const requestVersion = ++fileSizeStatsRequestVersion;
  fileSizeStatsLoading.value = true;
  fileSizeStatsError.value = '';
  try {
    const result = await resourceServer.fileSizeStats(filesBasesId, store.searchStoreData.searchData);
    if (requestVersion !== fileSizeStatsRequestVersion || cacheKey !== currentFileSizeStatsKey()) return;
    if (result && result.status) {
      fileSizeStats.value = result.data;
      fileSizeStatsCacheKey = cacheKey;
    } else {
      fileSizeStatsError.value = result?.msg || '统计失败，请稍后重试';
    }
  } catch (error) {
    if (requestVersion === fileSizeStatsRequestVersion) {
      fileSizeStatsError.value = '统计失败，请稍后重试';
    }
    console.log(error);
  } finally {
    if (requestVersion === fileSizeStatsRequestVersion) {
      fileSizeStatsLoading.value = false;
    }
  }
}

const formatFileSize = (size: number) => {
  if (!size || size < 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  let value = size;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex++;
  }
  const digits = unitIndex >= 3 ? 2 : 1;
  return `${value.toFixed(digits)} ${units[unitIndex]}`;
}

const showDataList = () => {
  return dataList.value;
}

const changePageHandle = () => {
  if (!isInitializing.value) {
    // 管理模式按本页选择，发起翻页时立即移除旧页，避免慢请求或失败时操作上一页。
    if (store.appStoreData.adminResourceStatus) dataList.value = [];
    getDataList();
    contentListRef.value?.change();
    contentListAdminRef.value?.change();
  }
}
const selectResourcesHandle = (item: I_resource) => {
  emits('selectResources', item, false)
}

const resourceAddTimeExchangedHandle = () => {
  init_DataList();
}

onMounted(async () => {
  eventBus.on('resource-add-time-exchanged', resourceAddTimeExchangedHandle);
  await init()
})
onUnmounted(() => {
  dataListRequestVersion++;
  fileSizeStatsRequestVersion++;
  window.clearTimeout(durationRefreshTimer);
  eventBus.off('resource-add-time-exchanged', resourceAddTimeExchangedHandle);
})
defineExpose({ init, init_DataList, showDataList });
</script>
<style lang="scss" scoped>
.content-view {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .list {
    flex-grow: 1;
    overflow: hidden;
  }

  .paging {
    width: calc(100% - 10px);
    padding-top: 5px;
    padding-right: 10px;
    flex-shrink: 0;
    display: flex;
    justify-content: space-between;

    .paging-main {
      display: flex; flex-wrap: wrap;
      align-items: center;
      min-width: 0;
    }

    .group-page-count { color: var(--el-text-color-secondary); }
    .paging-summary {
      display: flex;
      align-items: center;
      gap: 4px;
      padding-right: 8px;
      color: var(--el-text-color-regular);
      font-size: 12px;
      white-space: nowrap;
    }

    .file-size-stats-trigger {
      color: inherit;
      cursor: help;
    }

    .bottom-btns {
      display: flex; flex-wrap: wrap;
      gap: 5px;
      align-items: center;
    }
  }
}

.group-display {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  gap: 2px;
  padding: 3px;
  border: 1px solid var(--paging-border, var(--el-border-color));
  border-radius: 8px;
  background: var(--paging-control-bg, var(--el-fill-color-light));

  button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    min-height: 26px;
    padding: 0 10px;
    border: 1px solid transparent;
    border-radius: 5px;
    background: transparent;
    color: var(--el-text-color-secondary);
    font: inherit;
    font-size: 12px;
    white-space: nowrap;
    cursor: pointer;
    transition: background .15s, color .15s, border-color .15s;

    &:hover:not(:disabled) { color: var(--paging-accent, var(--el-color-primary)); }
    &.is-active {
      color: var(--paging-accent, var(--el-color-primary));
      border-color: color-mix(in srgb, var(--paging-accent, var(--el-color-primary)) 25%, transparent);
      background: var(--paging-accent-soft, var(--el-color-primary-light-9));
      font-weight: 500;
    }
    &:focus-visible { outline: 2px solid var(--paging-accent, var(--el-color-primary)); outline-offset: 1px; }
    &:disabled { opacity: .45; cursor: not-allowed; }
  }
}

.file-size-stats {
  display: grid;
  gap: 9px;

  .file-size-stats__title {
    padding-bottom: 7px;
    border-bottom: 1px solid var(--el-border-color-lighter);
    color: var(--el-text-color-primary);
    font-weight: 600;
  }

  > div:not(.file-size-stats__title) {
    display: flex;
    justify-content: space-between;
    gap: 16px;

    span {
      color: var(--el-text-color-secondary);
    }

    strong {
      color: var(--el-text-color-primary);
      font-weight: 500;
      white-space: nowrap;
    }
  }
}

.file-size-stats-state {
  min-height: 54px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--el-text-color-secondary);
}

.file-size-stats-state--error {
  flex-direction: column;
  color: var(--el-color-danger);
}

.file-size-stats-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid var(--el-border-color);
  border-top-color: var(--el-text-color-secondary);
  border-radius: 50%;
  animation: file-size-stats-spin 0.8s linear infinite;
}

.paging.paging--modern {
  --paging-accent: #20aaa9;
  --paging-accent-soft: rgba(32, 170, 169, 0.15);
  --paging-bg: rgba(31, 31, 31, 0.96);
  --paging-control-bg: rgba(255, 255, 255, 0.045);
  --paging-border: rgba(255, 255, 255, 0.1);
  --paging-text: #d9dce1;
  width: 100%;
  min-height: 44px;
  padding: 6px 10px;
  box-sizing: border-box;
  align-items: center;
  border-top: 1px solid var(--paging-border);
  color: var(--paging-text);
  background: var(--paging-bg);

  .paging-main {
    gap: 11px;
  }

  .paging-summary {
    height: 30px;
    padding: 0 10px;
    gap: 5px;
    border: 1px solid var(--paging-border);
    border-radius: 7px;
    color: var(--paging-text);
    background: var(--paging-control-bg);
  }

  .file-size-stats-trigger {
    color: var(--paging-accent);
  }

  :deep(.el-pagination) {
    --el-pagination-bg-color: var(--paging-control-bg);
    --el-pagination-button-bg-color: var(--paging-control-bg);
    --el-pagination-hover-color: var(--paging-accent);
    --el-pagination-button-color: var(--paging-text);
    gap: 6px;
  }

  :deep(.el-pagination .el-pager) {
    gap: 5px;
  }

  :deep(.el-pagination.is-background .btn-prev),
  :deep(.el-pagination.is-background .btn-next),
  :deep(.el-pagination.is-background .el-pager li) {
    min-width: 30px;
    height: 30px;
    margin: 0;
    border: 1px solid var(--paging-border);
    border-radius: 7px;
    color: var(--paging-text);
    background: var(--paging-control-bg);
  }

  :deep(.el-pagination.is-background .el-pager li.is-active) {
    color: var(--paging-accent);
    border-color: var(--paging-accent);
    background: var(--paging-accent-soft);
    box-shadow: inset 0 0 0 1px rgba(32, 170, 169, 0.12);
  }

  :deep(.el-pagination__jump) {
    height: 30px;
    margin-left: 5px;
    display: inline-flex;
    align-items: center;
    gap: 5px;
    color: var(--paging-text);
    font-size: 0;
  }

  :deep(.el-pagination__goto),
  :deep(.el-pagination__classifier) {
    display: none;
  }

  :deep(.el-pagination__jump::before) {
    content: '前往';
    font-size: 12px;
  }

  :deep(.el-pagination__jump::after) {
    content: '页';
    font-size: 12px;
  }

  :deep(.el-pagination__editor) {
    width: 54px;
    height: 30px;
    margin: 0 2px;
    font-size: 12px;
  }

  :deep(.el-pagination__editor .el-input__wrapper) {
    border-radius: 7px;
    background: var(--paging-control-bg);
    box-shadow: 0 0 0 1px var(--paging-border) inset;
  }

  .bottom-btns {
    gap: 7px;

    :deep(.el-button) {
      min-width: 92px;
      height: 30px;
      margin: 0;
      padding: 0 11px;
      border-color: var(--paging-border);
      border-radius: 7px;
      color: var(--paging-text);
      background: var(--paging-control-bg);
    }

    :deep(.el-button:hover) {
      color: var(--paging-accent);
      border-color: var(--paging-accent);
      background: var(--paging-accent-soft);
    }
  }
}

.paging.paging--modern.paging--bright,
:global(.bright) .paging.paging--modern {
  --paging-accent: #159a9a;
  --paging-accent-soft: rgba(21, 154, 154, 0.11);
  --paging-bg: rgba(255, 255, 255, 0.97);
  --paging-control-bg: #f7f9fa;
  --paging-border: #dfe5e9;
  --paging-text: #4d5c66;
}

@keyframes file-size-stats-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
