<template>
  <div class="index-view" v-loading="loading">
    <HeaderView class="header"></HeaderView>
    <dataBaseMenuView @select-files-base="selectFilesBaseHandle"></dataBaseMenuView>
    <div class="main">
      <TagView ref="tagViewRef" class="tag"></TagView>
      <ContentView ref="contentViewRef" class="content" @select-resources="selectResourcesHandle"></ContentView>
      <DetailsView class="details" :resource="resDetails"></DetailsView>
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
import { ElMessage } from 'element-plus'
const store = {
  appStoreData: appStoreData(),
}
const tagViewRef = ref<InstanceType<typeof TagView>>();
const contentViewRef = ref<InstanceType<typeof ContentView>>();

const loading = ref(false);
const resDetails = ref<I_resource | undefined>(undefined);

const selectFilesBaseHandle = async (filesBases: I_filesBases) => {
  loading.value = true;
  const result = await store.appStoreData.init(filesBases.id)
  if (result && !result.status) {
    ElMessage.error(result.message);
    return
  }
  //tagViewRef.value?.init();
  contentViewRef.value?.init();
  loading.value = false;
};

const selectResourcesHandle = (resource: I_resource) => {
  resDetails.value = resource;
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
