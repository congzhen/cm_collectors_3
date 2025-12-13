<template>
  <div class="module-content" v-loading="loading" v-if="dataList.length > 0">
    <div class="header-section">
      <div class="header-section-left">
        <label>{{ title_C }}</label>
        <el-icon class="refresh-icon" @click="getDataList">
          <Refresh />
        </el-icon>
      </div>
      <div class="header-section-right">
        <el-icon class="delete-icon" v-admin @click="closeModuleHandle">
          <Close />
        </el-icon>
      </div>
    </div>
    <ul class="list-ul" :style="contentLayoutStyle_C">
      <li v-for="(item, key) in dataList" :key="key">
        <contentStyleIndex :resources="item" :resourcesShowMode="props.resourcesShowMode"
          @selectResources="selectResourcesHandle(item)" />
      </li>
    </ul>
    <el-divider>
      <el-icon>
        <Upload />
      </el-icon>
      <label>{{ title_C }}</label>
    </el-divider>
  </div>
</template>
<script lang="ts" setup>
import { contentLayoutStyle_C } from '@/common/content';
import type { I_resource } from '@/dataType/resource.dataType';
import { resourceServer } from '@/server/resource.server';
import { appStoreData } from '@/storeData/app.storeData';
import { ref, onMounted, watch, computed, type PropType } from 'vue';
import contentStyleIndex from './contentStyleIndex.vue';
import type { T_resourcesShowMode } from '@/dataType/app.dataType';
import { filesBasesServer } from '@/server/filesBases.server';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  resourcesShowMode: {
    type: String as PropType<T_resourcesShowMode>,
    required: true,
  },
  moduleType: {
    type: String as PropType<'casualView' | 'history' | 'hot'>,
    required: true,
  },
  quantity: {
    type: Number,
    default: 10,
  },
})
const emits = defineEmits(['selectResources']);
const loading = ref(false);
const dataList = ref<I_resource[]>([]); // 谁便看看

const title_C = computed(() => {
  switch (props.moduleType) {
    case 'casualView':
      return '随便看看';
    case 'history':
      return '历史记录';
    case 'hot':
      return '热门资源';
    default:
      return '';
  }
})

const getRequest = () => {
  switch (props.moduleType) {
    case 'casualView':
      return resourceServer.dataListCasualView(store.appStoreData.currentFilesBases.id, props.quantity);
    case 'history':
      return resourceServer.dataListHistory(store.appStoreData.currentFilesBases.id, props.quantity);
    case 'hot':
      return resourceServer.dataListHot(store.appStoreData.currentFilesBases.id, props.quantity);
  }
}
const getDataList = async () => {
  try {
    loading.value = true;
    const result = await getRequest();
    if (result && result.status) {
      dataList.value = result.data;
    }
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }
}

const selectResourcesHandle = (item: I_resource) => {
  emits('selectResources', item)
}

const closeModuleHandle = () => {
  switch (props.moduleType) {
    case 'casualView':
      store.appStoreData.currentConfigApp.casualViewModule = false;
      break;
    case 'history':
      store.appStoreData.currentConfigApp.historyModule = false;
      break;
    case 'hot':
      store.appStoreData.currentConfigApp.hotModule = false;
      break;
    default:
  }
  filesBasesServer.setFilesBasesConfigById(store.appStoreData.currentFilesBases.id, store.appStoreData.currentConfigApp);
}

// 监听 currentFilesBases.id 的变化
watch(
  () => store.appStoreData.currentFilesBases.id,
  (newId, oldId) => {
    if (newId !== oldId) {
      getDataList();
    }
  }
);

onMounted(() => {
  getDataList();
})
</script>
<style lang="scss" scoped>
.header-section {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-section-left {
  display: flex;
  align-items: center;
}



.header-section label {
  font-size: 16px;
  font-weight: 600;
}

.refresh-icon {
  font-size: 20px;
  color: #666666;
  cursor: pointer;
  transition: transform 0.3s ease, color 0.3s ease;
  border-radius: 50%;
  padding: 6px;
}

.refresh-icon:hover {
  color: #409eff;
  transform: rotate(90deg);
}

.header-section-right {
  display: flex;
  padding-right: 10px;
  align-items: center;

  .delete-icon {
    font-size: 20px;
    color: #666666;
    cursor: pointer;

    &:hover {
      color: #409eff;
    }
  }
}

.list-ul {
  list-style-type: none;
  display: flex;
  flex-wrap: wrap;
  gap: 0.4em;
}

.el-divider {
  margin: 12px 0 30px 0;

  :deep(.el-divider__text) {
    font-size: 12px;
    background-color: #1F1F1F;
    padding: 0 10px;
    display: flex;
    align-items: center;
    gap: 5px;
    color: #4C4D4F;
  }
}
</style>
