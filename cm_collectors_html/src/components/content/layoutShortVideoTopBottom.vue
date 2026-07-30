<template>
  <div class="layout-short-video-top-bottom">
    <section class="video-section">
      <div ref="videoContainerRef" class="video-container">
        <videoPlay ref="videoPlayRef" class="fit-video-player" />
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
                <el-image :src="item.src" :title="item.title" fit="contain"
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
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch, type PropType } from 'vue';
import type { ElScrollbar } from 'element-plus';
import { isMobile } from '@/assets/mobile';
import { getResourceCoverPoster } from '@/common/photo';
import { getPlayVideoURLAndType } from '@/common/play';
import type { I_resource } from '@/dataType/resource.dataType';
import { appStoreData } from '@/storeData/app.storeData';
import videoPlay from '@/components/play/videoPlay.vue';
import contentRightClickMenu from './contentRightClickMenu.vue';
import contentVideoDurationBadge from './contentVideoDurationBadge.vue';

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

const dataListWrapper = computed(() => props.dataList);
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
  scrollToCurrentItem();
  const dramaSeriesId = getResourceDramaSeriesId(newIndex);
  if (dramaSeriesId === currentPlayingDramaSeriesId.value) return;
  void setVideoSource(dramaSeriesId);
});

function syncCurrentPlayIndex() {
  if (dataListWrapper.value.length === 0) {
    currentPlayIndex.value = -1;
    return;
  }
  if (currentPlayingDramaSeriesId.value === '') {
    currentPlayIndex.value = 0;
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
  return props.dataList[index]?.dramaSeries[0]?.id || '';
}

async function setVideoSource(dramaSeriesId: string) {
  const videoPlayer = videoPlayRef.value;
  if (!videoPlayer || dramaSeriesId === '') return;

  currentPlayingDramaSeriesId.value = dramaSeriesId;
  const requestVersion = ++sourceRequestVersion;
  const shouldResume = videoPlayer.isPlaying();
  videoPlayer.pause();
  const { playUrl, playType } = await getPlayVideoURLAndType(dramaSeriesId);
  if (requestVersion !== sourceRequestVersion) return;
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

function clickResourceHandle(index: number) {
  currentPlayIndex.value = index;
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
