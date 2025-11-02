<template>
  <div class="content-list" :style="contentListStyleObj_C">
    <layoutCoverPoster ref="layoutCoverPosterRef"
      v-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPoster'" :data-list="props.dataList"
      @select-resources="selectResourcesHandle">
    </layoutCoverPoster>
    <layoutCoverPosterBox ref="layoutCoverPosterBoxRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPosterBox'" :data-list="props.dataList"
      @select-resources="selectResourcesHandle">
    </layoutCoverPosterBox>
    <layoutCoverPosterBoxWideSeparate ref="layoutCoverPosterBoxWideSeparateRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPosterBoxWideSeparate'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutCoverPosterBoxWideSeparate>
    <layoutCoverPosterWaterfall ref="layoutCoverPosterWaterfallRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPosterWaterfall'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutCoverPosterWaterfall>
    <layoutShortVideo ref="layoutShortVideoRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'shortVideo'" :data-list="props.dataList"
      @select-resources="selectResourcesHandle">
    </layoutShortVideo>
    <layoutTable ref="layoutTableRef" v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'table'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutTable>
  </div>
</template>
<script lang="ts" setup>
import layoutCoverPoster from './layoutCoverPoster.vue';
import layoutCoverPosterBox from './layoutCoverPosterBox.vue';
import layoutCoverPosterBoxWideSeparate from './layoutCoverPosterBoxWideSeparate.vue';
import layoutCoverPosterWaterfall from './layoutCoverPosterWaterfall.vue';
import layoutShortVideo from './layoutShortVideo.vue';
import layoutTable from './layoutTable.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { ref, computed, type PropType } from 'vue';
import { appStoreData } from '@/storeData/app.storeData';
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
const layoutCoverPosterBoxRef = ref<typeof layoutCoverPosterBox>();
const layoutCoverPosterBoxWideSeparateRef = ref<typeof layoutCoverPosterBoxWideSeparate>();
const layoutCoverPosterWaterfallRef = ref<typeof layoutCoverPosterWaterfall>();
const layoutShortVideoRef = ref<typeof layoutShortVideo>();
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

const change = () => {
  switch (store.appStoreData.currentConfigApp.resourcesShowMode) {
    case 'coverPoster':
      layoutCoverPosterRef.value?.change();
      break;
    case 'coverPosterBox':
      layoutCoverPosterBoxRef.value?.change();
      break;
    case 'coverPosterBoxWideSeparate':
      layoutCoverPosterBoxWideSeparateRef.value?.change();
      break;
    case 'coverPosterWaterfall':
      layoutCoverPosterWaterfallRef.value?.change();
      break;
    case 'shortVideo':
      layoutShortVideoRef.value?.change();
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
