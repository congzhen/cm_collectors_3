<template>
  <div class="index-view" v-loading="loading">
    <HeaderView class="header" @create-resouce-success="createResouceSuccessHandle"></HeaderView>
    <dataBaseMenuView class="menu" @select-files-base="selectFilesBaseHandle"></dataBaseMenuView>
    <div class="main">
      <TagView ref="tagViewRef" class="tag"></TagView>
      <ContentView ref="contentViewRef" class="content" @select-resources="selectResourcesHandle"></ContentView>
      <DetailsView ref="detailsViewRef" v-if="store.appStoreData.detailsViewStatus && resDetails" class="details"
        :resource="resDetails" @update-resouce-success="updateResouceSuccessHandle"
        @delete-resource-success="deleteResouceSuccessHandle">
      </DetailsView>
    </div>
    <videoPlayDialog ref="videoPlayDialogRef"></videoPlayDialog>
  </div>
</template>
<script setup lang="ts">
import HeaderView from './HeaderView.vue'
import dataBaseMenuView from './dataBaseMenuView.vue'
import TagView from './TagView.vue'
import ContentView from './ContentView.vue'
import DetailsView from './DetailsView.vue'
import videoPlayDialog from '@/components/play/videoPlayDialog.vue'
import type { I_resource } from '@/dataType/resource.dataType'
import { ref, onMounted } from 'vue'
import type { I_filesBases } from '@/dataType/filesBases.dataType'
import { appStoreData } from '@/storeData/app.storeData'
import { searchStoreData } from '@/storeData/search.storeData'
import { ElMessage } from 'element-plus'
import { eventBus } from '@/main'
const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
}
const tagViewRef = ref<InstanceType<typeof TagView>>();
const contentViewRef = ref<InstanceType<typeof ContentView>>();
const detailsViewRef = ref<InstanceType<typeof DetailsView>>();
const videoPlayDialogRef = ref<InstanceType<typeof videoPlayDialog>>();

const loading = ref(false);
const resDetails = ref<I_resource | undefined>(undefined);

const selectFilesBaseHandle = async (filesBases: I_filesBases) => {
  loading.value = true;
  resDetails.value = undefined;
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

const selectResourcesHandle = (resource: I_resource | undefined, isInit: boolean) => {
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
const updateResouceSuccessHandle = async (data: I_resource) => {
  await contentViewRef.value?.init_DataList();
  resDetails.value = data;
}

const deleteResouceSuccessHandle = () => {
  contentViewRef.value?.init_DataList(() => { }, true);
}


const resourceDialogPlayStartHandle = (event: unknown) => {
  const typedEvent = event as { resourceId: string; dramaSeriesId: string };
  videoPlayDialogRef.value?.open(typedEvent.resourceId, typedEvent.dramaSeriesId);
}

// 监听事件
onMounted(() => {
  eventBus.on('resource-dialog-play-start', resourceDialogPlayStartHandle);
})

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
    flex-shrink: 0;
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
      overflow: hidden;
    }

    .details {
      flex-shrink: 0;
      height: 100%;
    }
  }
}
</style>
