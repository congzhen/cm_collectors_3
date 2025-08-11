<template>
  <div class="index-view" v-loading="loading">
    <HeaderView class="header" @create-resouce-success="createResouceSuccessHandle"></HeaderView>
    <dataBaseMenuView class="menu" @select-files-base="selectFilesBaseHandle"></dataBaseMenuView>
    <div class="main">
      <TagView ref="tagViewRef" class="tag"></TagView>
      <ContentView ref="contentViewRef" class="content" @select-resources="selectResourcesHandle"></ContentView>
      <DetailsView ref="detailsViewRef" v-if="store.appStoreData.detailsViewStatus" class="details"
        :resource="resDetails" @update-resouce-success="updateResouceSuccessHandle"
        @delete-resource-success="deleteResouceSuccessHandle">
      </DetailsView>
    </div>
  </div>
</template>
<script setup lang="ts">
import HeaderView from './HeaderView.vue'
import dataBaseMenuView from './dataBaseMenuView.vue'
import TagView from './TagView.vue'
import ContentView from './ContentView.vue'
import DetailsView from './DetailsView.vue'
import type { I_resource } from '@/dataType/resource.dataType'
import { ref } from 'vue'
import type { I_filesBases } from '@/dataType/filesBases.dataType'
import { appStoreData } from '@/storeData/app.storeData'
import { searchStoreData } from '@/storeData/search.storeData'
import { ElMessage } from 'element-plus'
const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
}
const tagViewRef = ref<InstanceType<typeof TagView>>();
const contentViewRef = ref<InstanceType<typeof ContentView>>();
const detailsViewRef = ref<InstanceType<typeof DetailsView>>();

const loading = ref(false);
const resDetails = ref<I_resource | undefined>(undefined);

const selectFilesBaseHandle = async (filesBases: I_filesBases) => {
  loading.value = true;
  const result = await store.appStoreData.initCurrentFilesBases(filesBases.id)
  if (result && !result.status) {
    ElMessage.error(result.message);
    return
  }
  store.searchStoreData.init();
  //tagViewRef.value?.init();
  contentViewRef.value?.init();
  loading.value = false;
};

const selectResourcesHandle = (resource: I_resource, isInit: boolean) => {
  resDetails.value = resource;
  if (!isInit) {
    detailsViewRef.value?.init();
  }
}
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const createResouceSuccessHandle = (data: I_resource) => {
  contentViewRef.value?.init();
}
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const updateResouceSuccessHandle = (data: I_resource) => {
  contentViewRef.value?.init_DataList();
}
const deleteResouceSuccessHandle = () => {
  contentViewRef.value?.init_DataList(() => { }, true);
}

</script>
<style lang="scss" scoped>
.index-view {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .header {
    width: 100%;
  }

  .menu {
    overflow: hidden;
    flex-shrink: 0;
  }

  .main {
    flex-grow: 1;
    display: flex;
    overflow: hidden;

    .tag {
      flex-shrink: 0;
      height: 100%;
    }

    .content {
      flex-grow: 1;
      overflow: auto;
    }

    .details {
      flex-shrink: 0;
      height: 100%;
    }
  }
}
</style>
