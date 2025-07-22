<template>
  <div class="content-view" v-loading="loading">
    <div class="list">
      <el-scrollbar>
        <ul class="list-ul">
          <li v-for="(item, key) in dataList" :key="key">
            <contentItem :resource="item" @click="emits('selectResources', item)"></contentItem>
          </li>
        </ul>
      </el-scrollbar>
    </div>
    <div class="paging">
      <el-pagination background layout="total, prev, pager, next, jumper" v-model:current-page="currentPage"
        :total="dataCount" :page-size="pageSize" @change="changePageHandle" />
    </div>
  </div>
</template>
<script setup lang="ts">
import contentItem from '@/components/content/contentItem.vue'
import { ref, onMounted, watch } from 'vue'
import { appStoreData } from '@/storeData/app.storeData';
import { searchStoreData } from '@/storeData/search.storeData';
import type { I_resource } from '@/dataType/resource.dataType';
import { ElMessage } from 'element-plus';
import { resourceServer } from '@/server/resource.server';
import { debounce } from '@/assets/debounce';
import { ca } from 'element-plus/es/locales.mjs';
const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
}
const emits = defineEmits(['selectResources']);

const loading = ref(false);
const dataList = ref<I_resource[]>([]);
const dataCount = ref(0);
let fetchCount = true;
const currentPage = ref(1);
const pageSize = ref(store.appStoreData.currentConfigApp.pageLimit);

watch(
  () => store.searchStoreData.searchData,
  () => {
    init();
  },
  { deep: true }
)

const init = async () => {
  dataList.value = [];
  dataCount.value = 0;
  fetchCount = true;
  currentPage.value = 1;
  pageSize.value = store.appStoreData.currentConfigApp.pageLimit;

  await init_DataList(() => {
    if (dataList.value.length > 0) {
      emits('selectResources', dataList.value[0]);
    }
  });

}

const init_DataList = async (fn: Function) => {
  await getDataList(fn);
}

const getDataList = debounce(async (fn: Function) => {
  try {
    loading.value = true;
    const result = await resourceServer.dataList(store.appStoreData.currentFilesBases.id, fetchCount, currentPage.value, pageSize.value, store.searchStoreData.searchData);
    if (result && result.status) {
      dataList.value = result.data.dataList;
      if (fetchCount) {
        dataCount.value = result.data.total;
        fetchCount = false;
      }
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

const changePageHandle = () => {
  getDataList();
}


onMounted(async () => {
  await init()
})
defineExpose({ init, init_DataList });
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

    .list-ul {
      list-style-type: none;
      display: flex;
      flex-wrap: wrap;
      gap: 0.4em;
      padding-bottom: 1em;
    }
  }

  .paging {
    width: 100%;
    padding-top: 5px;
  }
}
</style>
