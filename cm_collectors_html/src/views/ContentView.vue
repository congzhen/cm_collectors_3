<template>
  <div class="content-view" v-loading="loading">
    <div class="list">
      <contentList ref="contentListRef" v-if="!store.appStoreData.adminResourceStatus" :data-list="dataList"
        @select-resources="selectResourcesHandle"></contentList>
      <contentListAdmin v-else :data-list="dataList" @select-resources="selectResourcesHandle"
        @update-data="init_DataList"></contentListAdmin>
    </div>
    <div class="paging">
      <div class="paging-main">
        <div class="paging-summary">
          <span>Total {{ dataCount }}</span>
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
        <playListBtn></playListBtn>
        <coverAdjuster v-admin></coverAdjuster>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import contentList from '@/components/content/contentList.vue'
import contentListAdmin from '@/components/content/contentListAdmin.vue';
import coverAdjuster from '@/components/setting/fileDatabaseSetting/coverAdjuster.vue';
import playListBtn from '@/components/playList/playListBtn.vue';
import { ref, onMounted, watch, nextTick } from 'vue'
import { appStoreData } from '@/storeData/app.storeData';
import { searchStoreData } from '@/storeData/search.storeData';
import type { I_resource, I_resourceFileSizeStats } from '@/dataType/resource.dataType';
import { ElMessage } from 'element-plus';
import { resourceServer } from '@/server/resource.server';
import { debounce } from '@/assets/debounce';
const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
}
const emits = defineEmits(['selectResources']);

const contentListRef = ref<InstanceType<typeof contentList>>();


const isInitializing = ref(false);
const loading = ref(false);
const dataList = ref<I_resource[]>([]);
const dataCount = ref(0);
const fileSizeStats = ref<I_resourceFileSizeStats | null>(null);
const fileSizeStatsLoading = ref(false);
const fileSizeStatsError = ref('');
let fileSizeStatsCacheKey = '';
let fileSizeStatsRequestVersion = 0;
let fetchCount = true;
let durationRefreshTimer: number | undefined;
let dataListRequestVersion = 0;
const currentPage = ref(1);
const pageSize = ref(store.appStoreData.currentConfigApp.pageLimit);

watch(
  () => store.searchStoreData.searchData,
  () => {
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
    fetchCount = true;
  }
  await getDataList(fn);
}

const executeGetDataList = debounce(async (requestVersion: number, fn: () => void) => {
  const filesBasesId = store.appStoreData.currentFilesBases.id;
  const shouldFetchCount = fetchCount;
  try {
    loading.value = true;
    const result = await resourceServer.dataList(filesBasesId, shouldFetchCount, currentPage.value, pageSize.value, store.searchStoreData.searchData);
    if (requestVersion !== dataListRequestVersion || filesBasesId !== store.appStoreData.currentFilesBases.id) return;
    if (result && result.status) {
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
    if (requestVersion === dataListRequestVersion) loading.value = false;
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
    if (requestVersion !== dataListRequestVersion || filesBasesId !== store.appStoreData.currentFilesBases.id) return;
    const result = await resourceServer.dataList(filesBasesId, false, currentPage.value, pageSize.value, store.searchStoreData.searchData);
    if (requestVersion === dataListRequestVersion && filesBasesId === store.appStoreData.currentFilesBases.id && result && result.status) {
      dataList.value = result.data.dataList;
    }
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
    getDataList();
    contentListRef.value?.change();
  }
}
const selectResourcesHandle = (item: I_resource) => {
  emits('selectResources', item, false)
}

onMounted(async () => {
  await init()
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
      display: flex;
      align-items: center;
      min-width: 0;
    }

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
      display: flex;
      gap: 5px;
      align-items: center;
    }
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

@keyframes file-size-stats-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
