<template>
  <div class="content-list" :style="contentListStyleObj_C">
    <layoutCoverPoster ref="layoutCoverPosterRef"
      v-if="isLayoutCoverPoster(store.appStoreData.currentConfigApp.resourcesShowMode)"
      :resourcesShowMode="store.appStoreData.currentConfigApp.resourcesShowMode" :data-list="props.dataList"
      @select-resources="selectResourcesHandle">
    </layoutCoverPoster>
    <layoutCoverPosterWaterfall ref="layoutCoverPosterWaterfallRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPosterWaterfall'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutCoverPosterWaterfall>
    <layoutCoverPosterMosaic ref="layoutCoverPosterMosaicRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPosterMosaic'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutCoverPosterMosaic>
    <layoutCoverPosterMosaicShortVideo ref="layoutCoverPosterMosaicShortVideoRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPosterMosaicShortVideo'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutCoverPosterMosaicShortVideo>
    <layoutShortVideo ref="layoutShortVideoRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'shortVideo'" :data-list="props.dataList"
      @select-resources="selectResourcesHandle">
    </layoutShortVideo>
    <layoutShortVideoTopBottom ref="layoutShortVideoTopBottomRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'shortVideoTopBottom'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutShortVideoTopBottom>
    <layoutTable ref="layoutTableRef" v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'table'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutTable>
  </div>
</template>
<script lang="ts" setup>
import layoutCoverPoster from './layoutCoverPoster.vue';
import layoutCoverPosterWaterfall from './layoutCoverPosterWaterfall.vue';
import layoutCoverPosterMosaic from './layoutCoverPosterMosaic.vue';
import layoutCoverPosterMosaicShortVideo from './layoutCoverPosterMosaicShortVideo.vue';
import layoutShortVideo from './layoutShortVideo.vue';
import layoutShortVideoTopBottom from './layoutShortVideoTopBottom.vue';
import layoutTable from './layoutTable.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { ref, computed, type PropType } from 'vue';
import { appStoreData } from '@/storeData/app.storeData';
import { A_layoutCoverPosterSlc, type T_resourcesShowMode } from '@/dataType/app.dataType';

const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
})
const emits = defineEmits(['selectResources']);

const layoutCoverPosterRef = ref<typeof layoutCoverPoster>();
const layoutCoverPosterWaterfallRef = ref<typeof layoutCoverPosterWaterfall>();
const layoutCoverPosterMosaicRef = ref<typeof layoutCoverPosterMosaic>();
const layoutCoverPosterMosaicShortVideoRef = ref<typeof layoutCoverPosterMosaicShortVideo>();
const layoutShortVideoRef = ref<typeof layoutShortVideo>();
const layoutShortVideoTopBottomRef = ref<typeof layoutShortVideoTopBottom>();
const layoutTableRef = ref<typeof layoutTable>();

const contentListStyleObj_C = computed(() => {
  if (store.appStoreData.currentConfigApp.contentPadding == 0) {
    return {};
  }
  return {
    width: `calc(100% - ${store.appStoreData.currentConfigApp.contentPadding * 2}%)`,
    padding: `0 ${store.appStoreData.currentConfigApp.contentPadding}%`
  }
})

const selectResourcesHandle = (item: I_resource) => {
  emits('selectResources', item)
}

const isLayoutCoverPoster = (resourcesShowMode: T_resourcesShowMode) => {
  return A_layoutCoverPosterSlc.includes(resourcesShowMode)
}

const change = () => {
  const resourcesShowMode = store.appStoreData.currentConfigApp.resourcesShowMode;
  if (isLayoutCoverPoster(resourcesShowMode)) {
    layoutCoverPosterRef.value?.change();
    return;
  }

  switch (resourcesShowMode) {
    case 'coverPosterWaterfall':
      layoutCoverPosterWaterfallRef.value?.change();
      break;
    case 'coverPosterMosaic':
      layoutCoverPosterMosaicRef.value?.change();
      break;
    case 'coverPosterMosaicShortVideo':
      layoutCoverPosterMosaicShortVideoRef.value?.change();
      break;
    case 'shortVideo':
      layoutShortVideoRef.value?.change();
      break;
    case 'shortVideoTopBottom':
      layoutShortVideoTopBottomRef.value?.change();
      break;
    case 'table':
      layoutTableRef.value?.change();
      break;
  }
}

defineExpose({ change })

</script>
<style lang="scss" scoped>
.content-list {
  width: 100%;
  height: 100%;
  overflow: hidden;

  :deep(.play-icon) {
    .el-icon {
      color: #f3f3f3;
      filter: drop-shadow(0 0 4px rgba(0, 0, 0, 0.9));
    }
  }
}
</style>
