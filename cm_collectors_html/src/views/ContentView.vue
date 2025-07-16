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
import { ref, onMounted } from 'vue'
import { appStoreData } from '@/storeData/app.storeData'
import type { I_resource } from '@/dataType/resource.dataType';
import { ElMessage } from 'element-plus';
import { resourceServer } from '@/server/resource.server';
const store = {
  appStoreData: appStoreData(),
}
const emits = defineEmits(['selectResources']);

const loading = ref(false);
const dataList = ref<I_resource[]>([]);
const dataCount = ref(0);
let fetchCount = true;
const currentPage = ref(1);
const pageSize = ref(store.appStoreData.currentConfigApp.pageLimit);

const init = async () => {

  dataList.value = [];
  dataCount.value = 0;
  fetchCount = true;
  currentPage.value = 1;
  pageSize.value = store.appStoreData.currentConfigApp.pageLimit;

  await init_DataList();
  if (dataList.value.length > 0) {
    emits('selectResources', dataList.value[0]);
  }
}

const init_DataList = async () => {
  await getDataList();
}

const getDataList = async () => {
  loading.value = true;
  const result = await resourceServer.dataList(store.appStoreData.currentFilesBases.id, fetchCount, currentPage.value, pageSize.value,);
  if (result && result.status) {
    dataList.value = result.data.dataList;
    if (fetchCount) {
      dataCount.value = result.data.total;
      fetchCount = false;
    }
  } else {
    ElMessage.error(result.msg);
  }
  loading.value = false;
}

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
