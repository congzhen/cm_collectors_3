<template>
  <div class="content-view" v-loading="loading">
    <div class="list">
      <contentList ref="contentListRef" v-if="!store.appStoreData.adminResourceStatus" :data-list="dataList"
        @select-resources="selectResourcesHandle"></contentList>
      <contentListAdmin v-else :data-list="dataList" @select-resources="selectResourcesHandle"
        @update-data="init_DataList"></contentListAdmin>
    </div>
    <div class="paging">
      <el-pagination background layout="total, prev, pager, next, jumper" v-model:current-page="currentPage"
        :total="dataCount" :page-size="pageSize" :pager-count="5" size="small" @change="changePageHandle" />
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
import { ref, onMounted, watch } from 'vue'
import { appStoreData } from '@/storeData/app.storeData';
import { searchStoreData } from '@/storeData/search.storeData';
import type { I_resource } from '@/dataType/resource.dataType';
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
let fetchCount = true;
let durationRefreshTimer: number | undefined;
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

const init = async () => {
  isInitializing.value = true;
  dataList.value = [];
  dataCount.value = 0;
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
  });

}

const init_DataList = async (fn: () => void = () => { }, fetch: boolean = false) => {
  if (fetch) {
    fetchCount = true;
  }
  await getDataList(fn);
}

const getDataList = debounce(async (fn: () => void = () => { }) => {
  try {
    loading.value = true;
    const result = await resourceServer.dataList(store.appStoreData.currentFilesBases.id, fetchCount, currentPage.value, pageSize.value, store.searchStoreData.searchData);
    if (result && result.status) {
      dataList.value = result.data.dataList;
      if (fetchCount) {
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
    loading.value = false;
  }
}, 200)

// 资源列表接口只负责触发后端异步采集，不会等待 ffprobe 完成。
// 开启“显示视频时长”后，这里延迟重新拉取当前页一次，让刚写入数据库的时长能自动出现在封面角标上。
// 关闭开关时不做额外请求，避免影响默认浏览体验。
const scheduleDurationRefresh = () => {
  if (!store.appStoreData.currentConfigApp.showVideoDuration) {
    return;
  }
  window.clearTimeout(durationRefreshTimer);
  durationRefreshTimer = window.setTimeout(async () => {
    const result = await resourceServer.dataList(store.appStoreData.currentFilesBases.id, false, currentPage.value, pageSize.value, store.searchStoreData.searchData);
    if (result && result.status) {
      dataList.value = result.data.dataList;
    }
  }, 2500);
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

    .bottom-btns {
      display: flex;
      gap: 5px;
      align-items: center;
    }
  }
}
</style>
