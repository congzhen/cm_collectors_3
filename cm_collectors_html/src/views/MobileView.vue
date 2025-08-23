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
    <div class="content-list-wrapper">
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
  width: calc(100% - 10px);
  height: calc(100% - 10px);
  display: flex;
  flex-direction: column;
  background-color: #1f1f1f;
  color: #f3f3f3;
  padding: 5px;
  overflow: hidden;
}

.header {
  flex-shrink: 0;
  text-align: center;
  padding: 5px 0;
  border-bottom: 1px solid #444;
  margin-bottom: 5px;

  .database-selector {
    select {
      background-color: #333;
      color: #f3f3f3;
      border: 1px solid #555;
      padding: 8px 12px;
      border-radius: 6px;
      font-size: 1em;
      width: 100%;
      appearance: none;
      background-image: url("data:image/svg+xml;charset=UTF-8,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23f3f3f3' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3e%3cpolyline points='6 9 12 15 18 9'%3e%3c/polyline%3e%3c/svg%3e");
      background-repeat: no-repeat;
      background-position: right 12px center;
      background-size: 16px;
      padding-right: 40px;
      cursor: pointer;
      transition: all 0.2s ease;

      &:focus {
        outline: none;
        border-color: #409eff;
        box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
      }

      option {
        background-color: #333;
        color: #f3f3f3;
        padding: 8px;
      }
    }
  }
}

.content-list-wrapper {
  flex: 1;
  overflow-y: auto;

  :deep(.content-style1) {
    width: 120px !important;
    height: auto !important;
  }
}

.pagination {
  flex-shrink: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 0;
  min-height: 28px; // 确保分页区域始终有一定高度

  button {
    background-color: #333;
    color: #f3f3f3;
    border: 1px solid #555;
    padding: 4px 12px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9em;
    min-width: 60px; // 确保按钮有最小宽度
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s ease;

    &:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }

    &:not(:disabled):hover {
      background-color: #444;
      border-color: #666;
    }

    &:focus {
      outline: none;
      box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.5);
    }
  }

  span {
    margin: 0 10px;
    font-size: 0.9em;
    white-space: nowrap;
  }
}
</style>
