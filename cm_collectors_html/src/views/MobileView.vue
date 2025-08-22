<template>
  <div class="mobile-view">
    <div class="header">
      <div class="database-selector">
        <select v-model="selectedDataBase" @change="changeDataBase">
          <option v-for="filesBases in store.filesBasesStoreData.filesBasesStatus" :key="filesBases.id"
            :value="filesBases.id">
            {{ filesBases.name }}
          </option>
        </select>
      </div>
    </div>
    <div class="content-list">
      <contentList ref="contentListRef" :data-list="dataList" @select-resources="selectResourcesHandle">
      </contentList>
    </div>
    <div class="pagination">
      <button @click="prevPage" :disabled="currentPage <= 1">上一页</button>
      <span>{{ currentPage }} / {{ Math.ceil(dataCount / pageSize) }}</span>
      <button @click="nextPage" :disabled="currentPage >= Math.ceil(dataCount / pageSize)">下一页</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import contentList from '@/components/content/contentList.vue'
import { appStoreData } from '@/storeData/app.storeData'
import { filesBasesStoreData } from '@/storeData/filesBases.storeData'
import { searchStoreData } from '@/storeData/search.storeData'
import type { I_resource } from '@/dataType/resource.dataType'
import type { I_filesBases } from '@/dataType/filesBases.dataType'
import { resourceServer } from '@/server/resource.server'
import { debounce } from '@/assets/debounce'
import { ElMessage } from 'element-plus'
import { playResource } from '@/common/play';
const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
  filesBasesStoreData: filesBasesStoreData()
}


const contentListRef = ref<InstanceType<typeof contentList>>()

let fetchCount = true;
const loading = ref(false);
const dataList = ref<I_resource[]>([])
const dataCount = ref(0);
const currentPage = ref(1)
const pageSize = ref(store.appStoreData.currentConfigApp.pageLimit);

// 计算属性，获取选中的数据库
const selectedDataBase = computed({
  get: () => store.appStoreData.currentFilesBases.id,
  set: () => {
    // 这里只是设置值，实际切换在 changeDataBase 方法中处理
  }
})

// 初始化数据
const init = async () => {
  await getDataList()
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

// 上一页
const prevPage = async () => {
  if (currentPage.value > 1) {
    currentPage.value--
    await getDataList(() => {
      contentListRef.value?.change();
    });

  }
}

// 下一页
const nextPage = async () => {
  if (currentPage.value < Math.ceil(dataCount.value / pageSize.value)) {
    await currentPage.value++
    await getDataList(() => {
      contentListRef.value?.change();
    });
  }
}

// 选择资源
const selectResourcesHandle = (item: I_resource) => {
  playResource(item);
}

// 切换数据库
const changeDataBase = async (event: Event) => {
  const target = event.target as HTMLSelectElement
  const selectedId = target.value
  const selectedFilesBases = store.filesBasesStoreData.filesBasesStatus.find(
    (filesBases: I_filesBases) => filesBases.id === selectedId
  )

  if (selectedFilesBases) {
    await store.appStoreData.initCurrentFilesBases(selectedFilesBases.id)
    currentPage.value = 1
    await getDataList()
  }
}

onMounted(async () => {
  await init()
})
</script>

<style lang="scss" scoped>
.mobile-view {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: #1f1f1f;
  color: #f3f3f3;
  padding: 5px;
  box-sizing: border-box;
}

.header {
  text-align: center;
  padding: 5px 0;
  border-bottom: 1px solid #444;
  margin-bottom: 5px;

  .database-selector {
    select {
      background-color: #333;
      color: #f3f3f3;
      border: 1px solid #444;
      padding: 4px 8px;
      border-radius: 4px;
      font-size: 0.9em;
      width: 100%;
    }
  }
}

.content-list {
  flex: 1;
  overflow-y: auto;

  :deep(.content-style1) {
    width: 120px !important;
    height: auto !important;
  }
}

.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 0;

  button {
    background-color: #333;
    color: #f3f3f3;
    border: none;
    padding: 6px 12px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9em;

    &:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
  }

  span {
    margin: 0 5px;
    font-size: 0.9em;
  }
}
</style>
