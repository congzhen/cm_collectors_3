<template>
  <div class="content-list" :style="contentListStyleObj_C">
    <layoutCoverPoster ref="layoutCoverPosterRef"
      v-if="isLayoutCoverPoster(store.appStoreData.currentConfigApp.resourcesShowMode)"
      :resourcesShowMode="store.appStoreData.currentConfigApp.resourcesShowMode" :data-list="renderedList"
      @select-resources="selectResourcesHandle">
    </layoutCoverPoster>
    <layoutCoverPosterWaterfall ref="layoutCoverPosterWaterfallRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPosterWaterfall'"
      :data-list="renderedList" @select-resources="selectResourcesHandle">
    </layoutCoverPosterWaterfall>
    <layoutCoverPosterMosaic ref="layoutCoverPosterMosaicRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPosterMosaic'"
      :data-list="renderedList" @select-resources="selectResourcesHandle">
    </layoutCoverPosterMosaic>
    <layoutCoverPosterMosaicShortVideo ref="layoutCoverPosterMosaicShortVideoRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPosterMosaicShortVideo'"
      :data-list="renderedList" @select-resources="selectResourcesHandle">
    </layoutCoverPosterMosaicShortVideo>
    <layoutCoverPosterCompactWall ref="layoutCoverPosterCompactWallRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPosterCompactWall'"
      :data-list="renderedList" @select-resources="selectResourcesHandle">
    </layoutCoverPosterCompactWall>
    <layoutShortVideo ref="layoutShortVideoRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'shortVideo'" :data-list="renderedList"
      @select-resources="selectResourcesHandle">
    </layoutShortVideo>
    <layoutShortVideoTopBottom ref="layoutShortVideoTopBottomRef"
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'shortVideoTopBottom'"
      :data-list="renderedList" @select-resources="selectResourcesHandle">
    </layoutShortVideoTopBottom>
    <layoutTable ref="layoutTableRef" v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'table'"
      :data-list="renderedList" @select-resources="selectResourcesHandle">
    </layoutTable>
  </div>
</template>
<script lang="ts" setup>
import layoutCoverPoster from './layoutCoverPoster.vue';
import layoutCoverPosterWaterfall from './layoutCoverPosterWaterfall.vue';
import layoutCoverPosterMosaic from './layoutCoverPosterMosaic.vue';
import layoutCoverPosterMosaicShortVideo from './layoutCoverPosterMosaicShortVideo.vue';
import layoutCoverPosterCompactWall from './layoutCoverPosterCompactWall.vue';
import layoutShortVideo from './layoutShortVideo.vue';
import layoutShortVideoTopBottom from './layoutShortVideoTopBottom.vue';
import layoutTable from './layoutTable.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { ref, computed, provide, inject, watch, onBeforeUnmount, type PropType } from 'vue';
import { appStoreData } from '@/storeData/app.storeData';
import { A_layoutCoverPosterSlc, type T_resourcesShowMode } from '@/dataType/app.dataType';

import { isMobile } from '@/assets/mobile';
import { resourceGroupServer } from '@/server/resourceGroup.server';
import type { I_resourceGroupSummary } from '@/dataType/resource.dataType';
import { mergeResourceGroupPage, resourceGroupContextKey, resourceGroupPageKey, type ResourceGroupPage } from './resourceGroups';
import { resourceSelectionKey } from './resourceSelection';
const selection = inject(resourceSelectionKey, undefined);
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  groupPage: Object as PropType<ResourceGroupPage>,
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
})
provide('resource-list-page', computed(() => props.dataList));
const groupingEnabled = computed(() => !isMobile() && !selection?.active.value
  && store.appStoreData.currentConfigApp.resourcesShowMode !== 'table'
  && store.appStoreData.currentConfigApp.mergeResourceGroups !== false);
const renderedList = computed(() => mergeResourceGroupPage(props.dataList, groupingEnabled.value));
const groupSummaries = ref(new Map<string, I_resourceGroupSummary>());
provide(resourceGroupContextKey, {
  enabled: groupingEnabled,
  pageIds: computed(() => new Set(props.dataList.map(resource => resource.id))),
  summaries: groupSummaries,
});
let groupRequestVersion = 0;
let summaryLibrary = '';
watch(() => [props.dataList, props.groupPage, groupingEnabled.value, store.appStoreData.currentFilesBases.id] as const, async () => {
  const version = ++groupRequestVersion;
  const library = store.appStoreData.currentFilesBases.id;
  if (!groupingEnabled.value) { groupSummaries.value = new Map(); return; }
  if (props.groupPage?.key === resourceGroupPageKey(library, props.dataList)) {
    summaryLibrary = library;
    groupSummaries.value = new Map(props.groupPage.summaries.map(group => [group.id, group]));
    return;
  }
  const groups = [...new Set(props.dataList.map(resource => resource.associationGroupId).filter((id): id is string => !!id))];
  // Keep known current-page covers steady while refreshing; never carry them across libraries.
  groupSummaries.value = new Map(summaryLibrary === library
    ? [...groupSummaries.value].filter(([id]) => groups.includes(id)) : []);
  summaryLibrary = library;
  if (!groups.length) return;
  try {
    const result = await resourceGroupServer.summaries(library, groups);
    if (version === groupRequestVersion && result.status) groupSummaries.value = new Map(result.data.dataList.map(group => [group.id, group]));
  } catch { /* 摘要失败仍保留资源卡片，悬停面板可重试获取整组。 */ }
}, { immediate: true });
onBeforeUnmount(() => { groupRequestVersion++; });
const emits = defineEmits(['selectResources']);

const layoutCoverPosterRef = ref<typeof layoutCoverPoster>();
const layoutCoverPosterWaterfallRef = ref<typeof layoutCoverPosterWaterfall>();
const layoutCoverPosterMosaicRef = ref<typeof layoutCoverPosterMosaic>();
const layoutCoverPosterMosaicShortVideoRef = ref<typeof layoutCoverPosterMosaicShortVideo>();
const layoutCoverPosterCompactWallRef = ref<typeof layoutCoverPosterCompactWall>();
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
    case 'coverPosterCompactWall':
      layoutCoverPosterCompactWallRef.value?.change();
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
