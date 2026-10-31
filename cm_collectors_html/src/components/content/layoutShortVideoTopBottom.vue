<template>
  <div class="layout-short-video-top-bottom">
    <section class="video-section">
      <div ref="videoContainerRef" class="video-container">
        <videoPlay v-if="!selection?.active.value" ref="videoPlayRef" class="fit-video-player" />
        <div v-if="selection?.active.value" class="management-player-placeholder">正在管理资源，请在资源列表中勾选</div>
      </div>
    </section>

    <section class="material-section">
      <div ref="materialViewportRef" class="material-scroll" @wheel.prevent="handleMaterialWheel">
        <el-scrollbar ref="scrollbarRef">
          <div class="material-grid">
            <contentRightClickMenu v-for="(item, index) in dataList_C" :key="item.id"
              class="material-resource" :style="getMaterialResourceStyle(item.id)" :resource="props.dataList[index]">
              <div class="material-item"
                :class="{ active: index === currentPlayIndex, ['material-item-' + index]: true }"
                @click.stop="clickResourceHandle(index)">
                <ResourceGroupCover :resource="props.dataList[index]" :src="item.src" :title="item.title" fit="contain"
                  @load="onMaterialImageLoad(item.id, $event)" />
                <contentVideoDurationBadge :resource="props.dataList[index]" offset-bottom="8px" offset-right="4px"
                  :compact-text="isMobile()" :adaptive-mobile="isMobile()" />
                <div v-if="!isMobile()" class="play-icon" @click.stop="selectResourcesHandle(item)">
                  <el-icon><Setting /></el-icon>
                </div>
              </div>
            </contentRightClickMenu>
          </div>
        </el-scrollbar>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus';
import { computed, inject, provide, nextTick, onBeforeUnmount, onMounted, ref, watch, type PropType } from 'vue';
import type { ElScrollbar } from 'element-plus';
import { isMobile } from '@/assets/mobile';
import { getResourceCoverPoster } from '@/common/photo';
import { getPlayVideoURLAndType } from '@/common/play';
import type { I_resource } from '@/dataType/resource.dataType';
import { appStoreData } from '@/storeData/app.storeData';
import videoPlay from '@/components/play/videoPlay.vue';
import ResourceGroupCover from './ResourceGroupCover.vue';
import contentRightClickMenu from './contentRightClickMenu.vue';
import contentVideoDurationBadge from './contentVideoDurationBadge.vue';

import { resourceSelectionKey } from './resourceSelection';
import { resourceGroupKeepOpenKey, resourceGroupPlaybackList, resourceGroupSelectKey } from './resourceGroups';
provide(resourceGroupKeepOpenKey, true);
const selectedGroupResource = ref<I_resource | null>(null);
const selectedFromGroup = ref(false);
provide(resourceGroupSelectKey, resource => selectGroupResource(resource));
const selection = inject(resourceSelectionKey, undefined);
const store = {
  appStoreData: appStoreData(),
};

const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
});
const emits = defineEmits(['selectResources']);

const materialCardHeight = ref(200);
const materialAspectRatios = ref<Record<string, number>>({});
const currentPlayIndex = ref(-1);
const currentPlayingDramaSeriesId = ref('');
const videoPlayRef = ref<InstanceType<typeof videoPlay>>();
const videoContainerRef = ref<HTMLDivElement>();
const materialViewportRef = ref<HTMLDivElement>();
const scrollbarRef = ref<InstanceType<typeof ElScrollbar>>();
let sourceRequestVersion = 0;
let resizeObserver: ResizeObserver | undefined;
let resetScrollAfterFilesBaseChange = false;

const dataListWrapper = computed(() => resourceGroupPlaybackList(props.dataList, selectedFromGroup.value ? selectedGroupResource.value : null, store.appStoreData.currentFilesBases.id));
const dataList_C = computed(() => props.dataList.map(item => ({
  ...item,
  src: getResourceCoverPoster(item),
})));

watch(dataListWrapper, () => {
  syncCurrentPlayIndex();
  if (resetScrollAfterFilesBaseChange && dataListWrapper.value.length > 0) {
    resetScrollAfterFilesBaseChange = false;
    resetResourceScroll();
  }
}, { deep: true });
watch(() => store.appStoreData.currentFilesBases.id, resetPlaybackForFilesBaseChange, { flush: 'sync' });
watch(currentPlayIndex, newIndex => {
  if (newIndex < 0) return;
  if (!selectedFromGroup.value || dataListWrapper.value[newIndex]?.id !== selectedGroupResource.value?.id) scrollToCurrentItem();
  const dramaSeriesId = getResourceDramaSeriesId(newIndex);
  if (dramaSeriesId === currentPlayingDramaSeriesId.value) return;
  void setVideoSource(dramaSeriesId);
});

function syncCurrentPlayIndex() {
  if (selection?.active.value) {
    currentPlayIndex.value = -1;
    return;
  }
  if (dataListWrapper.value.length === 0) {
    currentPlayIndex.value = -1;
    return;
  }
  if (currentPlayingDramaSeriesId.value === '') {
    // No active video can mean a deliberately selected member without videos.
    // Keep that selection when durations or page resources refresh.
    currentPlayIndex.value = Math.max(0, dataListWrapper.value.findIndex(resource =>
      resource.id === selectedGroupResource.value?.id));
    return;
  }
  currentPlayIndex.value = dataListWrapper.value.findIndex(resource =>
    resource.dramaSeries.some(dramaSeries => dramaSeries.id === currentPlayingDramaSeriesId.value)
  );
}

function resetPlaybackForFilesBaseChange() {
  // 同库翻页保留播放状态；切库时必须立即隔离旧库的播放器和异步请求。
  sourceRequestVersion++;
  currentPlayingDramaSeriesId.value = '';
  selectedGroupResource.value = null;
  selectedFromGroup.value = false;
  currentPlayIndex.value = -1;
  videoPlayRef.value?.pause();
  resetScrollAfterFilesBaseChange = true;
  resetResourceScroll();
}

function resetResourceScroll() {
  nextTick(() => {
    scrollbarRef.value?.update();
    window.requestAnimationFrame(() => {
      scrollbarRef.value?.setScrollLeft(0);
      if (scrollbarRef.value?.wrapRef) scrollbarRef.value.wrapRef.scrollLeft = 0;
    });
  });
}

function getResourceDramaSeriesId(index: number) {
  return dataListWrapper.value[index]?.dramaSeries[0]?.id || '';
}

async function setVideoSource(dramaSeriesId: string) {
  if (selection?.active.value) return;
  const videoPlayer = videoPlayRef.value;
  if (!videoPlayer || dramaSeriesId === '') return;

  currentPlayingDramaSeriesId.value = dramaSeriesId;
  const requestVersion = ++sourceRequestVersion;
  const shouldResume = videoPlayer.isPlaying();
  videoPlayer.pause();
  const playData = await getPlayVideoURLAndType(dramaSeriesId).catch(() => null);
  if (requestVersion !== sourceRequestVersion) return;
  if (!playData?.playUrl) {
    currentPlayingDramaSeriesId.value = '';
    if (!playData) ElMessage.error('播放地址加载失败，请重试');
    return;
  }
  const { playUrl, playType } = playData;
  const resource = dataListWrapper.value.find((item) =>
    item.dramaSeries.some((dramaSeries) => dramaSeries.id === dramaSeriesId)
  );
  const dramaSeries = resource?.dramaSeries.find((item) => item.id === dramaSeriesId);

  videoPlayer.setVideoSource(playUrl, playType, () => {
    if (requestVersion !== sourceRequestVersion) return;
    videoPlayer.addTextTrack(`/api/video/subtitle/${dramaSeriesId}`, '默认字幕', 'zh', true);
    if (shouldResume) videoPlayer.play();
  }, dramaSeries?.src || resource?.title || '');
}

const selectGroupResource = async (resource: I_resource, keepOffPage = true) => {
  const previousRequestVersion = sourceRequestVersion;
  selectedFromGroup.value = keepOffPage;
  selectedGroupResource.value = resource;
  // 等列表同步当前播放索引后，再切换到组内成员；支持连续选择两个页外成员。
  await nextTick();
  if (selectedGroupResource.value?.id !== resource.id || store.appStoreData.currentFilesBases.id !== resource.filesBases_id) return;
  currentPlayIndex.value = dataListWrapper.value.findIndex(item => item.id === resource.id);
  await nextTick();
  if (selectedGroupResource.value?.id !== resource.id || store.appStoreData.currentFilesBases.id !== resource.filesBases_id) return;
  const dramaSeriesId = resource.dramaSeries[0]?.id || '';
  if (!dramaSeriesId) {
    sourceRequestVersion++;
    currentPlayingDramaSeriesId.value = '';
    videoPlayRef.value?.pause();
    return;
  }
  if (previousRequestVersion === sourceRequestVersion && dramaSeriesId !== currentPlayingDramaSeriesId.value) void setVideoSource(dramaSeriesId);
};
function clickResourceHandle(index: number) {
  const resource = props.dataList[index];
  if (resource) void selectGroupResource(resource, false);
}

function selectResourcesHandle(item: I_resource) {
  emits('selectResources', item);
}

function scrollToCurrentItem() {
  nextTick(() => {
    const container = scrollbarRef.value?.wrapRef;
    const item = container?.querySelector(`.material-item-${currentPlayIndex.value}`) as HTMLElement | null;
    item?.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'center' });
  });
}

function handleMaterialWheel(event: WheelEvent) {
  const scrollContainer = scrollbarRef.value?.wrapRef;
  if (!scrollContainer) return;
  const scrollDistance = Math.abs(event.deltaX) > Math.abs(event.deltaY) ? event.deltaX : event.deltaY;
  scrollContainer.scrollLeft += scrollDistance;
}

function getMaterialResourceStyle(resourceId: string) {
  const aspectRatio = materialAspectRatios.value[resourceId] || 2 / 3;
  const width = Math.round(materialCardHeight.value * aspectRatio);
  return {
    width: `${width}px`,
    flex: `0 0 ${width}px`,
  };
}

function onMaterialImageLoad(resourceId: string, event: Event) {
  const image = event.target as HTMLImageElement | null;
  if (!image?.naturalWidth || !image.naturalHeight) return;
  materialAspectRatios.value = {
    ...materialAspectRatios.value,
    [resourceId]: image.naturalWidth / image.naturalHeight,
  };
  nextTick(() => scrollbarRef.value?.update());
}

function updateLayoutSize() {
  const materialHeight = materialViewportRef.value?.clientHeight || 0;
  if (materialHeight > 0) materialCardHeight.value = materialHeight;
  scrollbarRef.value?.update();
}

function change() {
  nextTick(() => {
    scrollbarRef.value?.setScrollLeft(0);
    updateLayoutSize();
  });
}

onMounted(() => {
  syncCurrentPlayIndex();
  resizeObserver = new ResizeObserver(updateLayoutSize);
  if (videoContainerRef.value) resizeObserver.observe(videoContainerRef.value);
  if (materialViewportRef.value) resizeObserver.observe(materialViewportRef.value);
  nextTick(updateLayoutSize);
});

onBeforeUnmount(() => {
  sourceRequestVersion++;
  videoPlayRef.value?.pause();
  resizeObserver?.disconnect();
});

defineExpose({ change });
</script>

<style lang="scss" scoped>
.management-player-placeholder { display: grid; place-items: center; height: 100%; min-height: 120px; color: var(--el-text-color-secondary); font-size: 13px; }
.layout-short-video-top-bottom {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .video-section {
    flex: 3;
    min-height: 0;
    padding: 8px 10px 0;
    display: flex;
    flex-direction: column;
  }

  .video-container {
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  .fit-video-player {
    min-height: 0;
  }

  .video-container :deep(.video-player-container) {
    min-height: 0;
  }

  .video-container :deep(.video-player-windows) {
    min-height: 0;
    overflow: hidden;
  }

  // video.js 的 fluid 模式会通过 padding-top 强制宽高比；在固定高度区域中必须清除，
  // 否则视频窗口会超过父容器并把自定义控制条挤出可视区域。
  .video-container :deep(.video-js.vjs-fluid) {
    height: 100% !important;
    padding-top: 0 !important;
  }

  .video-container :deep(.video-controller) {
    flex-shrink: 0;
  }

  .material-section {
    flex: 0 0 200px;
    height: 200px;
    min-height: 200px;
    max-height: 200px;
    padding: 5px 10px;
    display: flex;
    flex-direction: column;
  }

  .material-scroll {
    flex: 0 0 200px;
    height: 200px;
    min-height: 200px;
    max-height: 200px;
    overflow: hidden;
  }

  .material-scroll :deep(.el-scrollbar),
  .material-scroll :deep(.el-scrollbar__wrap),
  .material-scroll :deep(.el-scrollbar__view) {
    height: 200px;
    min-height: 200px;
    max-height: 200px;
  }

  .material-scroll :deep(.el-scrollbar__wrap) {
    overflow-y: hidden;
  }

  .material-scroll :deep(.el-scrollbar__bar.is-vertical) {
    display: none;
  }

  .material-grid {
    height: 200px;
    min-height: 200px;
    max-height: 200px;
    min-width: 100%;
    width: max-content;
    padding: 0 8px;
    box-sizing: border-box;
    display: flex;
    align-items: stretch;
    gap: 10px;
  }

  :deep(.material-resource) {
    height: 100%;
    min-height: 0;
  }

  .material-item {
    position: relative;
    width: 100%;
    height: 100%;
    overflow: hidden;
    cursor: pointer;
    border: 2px solid transparent;
    border-radius: 4px;
    box-sizing: border-box;
    transition: border-color 0.2s ease, box-shadow 0.2s ease;
    container-type: inline-size;

    &.active {
      border-color: #409eff;
      box-shadow: 0 0 8px rgba(64, 158, 255, 0.5);
    }

    .el-image {
      width: 100%;
      height: 100%;
      transition: transform 0.2s ease;
    }

    &:hover {
      .el-image { transform: scale(1.02); }
      .play-icon { display: block; }
    }

    .play-icon {
      position: absolute;
      z-index: 10;
      left: 6px;
      bottom: 6px;
      display: none;
      color: #f3f3f3;
      font-size: 2.2em;
      opacity: 0.8;
    }
  }

}
</style>
