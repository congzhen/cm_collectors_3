<template>
  <div class="layout-cover-poster-mosaic-short-video">
    <layoutCoverPosterMosaic
      ref="mosaicRef"
      class="mosaic-panel"
      :data-list="props.dataList"
      player-mode
      :active-resource-id="currentPlayResourceId"
      storage-key-prefix="layout-cover-poster-mosaic-short-video-column"
      :default-column="4"
      :max-column="8"
      :limit-columns-by-width="false"
      @play-resource="clickResourceHandle"
      @select-resources="selectResourcesHandle"
    />

    <div class="short-video-play" :class="{ 'mobile-layout': isMobile() }">
      <div ref="videoPlayContainerRef" class="short-video-play-container">
        <videoPlay ref="videoPlayRef" />
        <div v-if="showUnavailableMessage" class="unavailable-message">
          暂无可播放资源
        </div>
      </div>
      <div class="tool">
        <el-button-group :size="isMobile() ? 'small' : 'default'">
          <el-button
            icon="ArrowLeft"
            :disabled="currentPlayIndex <= 0"
            @click="prevHandle"
          >
            上一个
          </el-button>
          <el-button
            :disabled="
              currentPlayIndex < 0 ||
              currentPlayIndex >= dataListWrapper.length - 1
            "
            @click="nextHandle"
          >
            下一个
            <el-icon class="el-icon--right">
              <ArrowRight />
            </el-icon>
          </el-button>
        </el-button-group>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
  type PropType,
} from 'vue';

import { isMobile } from '@/assets/mobile';
import { getPlayVideoURLAndType } from '@/common/play';
import type { I_resource } from '@/dataType/resource.dataType';
import { appStoreData } from '@/storeData/app.storeData';
import videoPlay from '@/components/play/videoPlay.vue';

import layoutCoverPosterMosaic from './layoutCoverPosterMosaic.vue';

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

const mosaicRef = ref<InstanceType<typeof layoutCoverPosterMosaic>>();
const videoPlayContainerRef = ref<HTMLDivElement>();
const videoPlayRef = ref<InstanceType<typeof videoPlay>>();
const currentPlayIndex = ref(-1);
const currentPlayingDramaSeriesId = ref('');
let sourceRequestVersion = 0;
let resetScrollAfterFilesBaseChange = false;
let resizeObserver: ResizeObserver | undefined;

const dataListWrapper = computed(() => props.dataList);
const currentPlayResourceId = computed(
  () => dataListWrapper.value[currentPlayIndex.value]?.id || '',
);
const currentPlayDramaSeriesId = computed(() =>
  getResourceDramaSeriesId(currentPlayIndex.value),
);
const showUnavailableMessage = computed(
  () =>
    currentPlayIndex.value >= 0 && !currentPlayDramaSeriesId.value,
);

watch(
  dataListWrapper,
  () => {
    syncCurrentPlayIndex();
    if (resetScrollAfterFilesBaseChange && dataListWrapper.value.length > 0) {
      resetScrollAfterFilesBaseChange = false;
      resetResourceScroll();
    }
  },
  { deep: true },
);

watch(
  () => store.appStoreData.currentFilesBases.id,
  () => {
    resetPlaybackForFilesBaseChange();
  },
  { flush: 'sync' },
);

watch(currentPlayIndex, (newIndex) => {
  if (newIndex < 0) return;
  mosaicRef.value?.scrollToResource(currentPlayResourceId.value);
  const dramaSeriesId = currentPlayDramaSeriesId.value;
  if (!dramaSeriesId) {
    clearUnavailablePlayback();
    return;
  }
  if (dramaSeriesId === currentPlayingDramaSeriesId.value) return;
  setVideoSource(dramaSeriesId);
});

const syncCurrentPlayIndex = () => {
  if (dataListWrapper.value.length === 0) {
    currentPlayIndex.value = -1;
    return;
  }

  if (!currentPlayingDramaSeriesId.value) {
    currentPlayIndex.value = 0;
    return;
  }

  currentPlayIndex.value = dataListWrapper.value.findIndex((resource) =>
    resource.dramaSeries.some(
      (dramaSeries) => dramaSeries.id === currentPlayingDramaSeriesId.value,
    ),
  );
};

const resetPlaybackForFilesBaseChange = () => {
  sourceRequestVersion++;
  currentPlayingDramaSeriesId.value = '';
  currentPlayIndex.value = -1;
  videoPlayRef.value?.pause();
  resetScrollAfterFilesBaseChange = true;
  resetResourceScroll();
};

const resetResourceScroll = () => {
  nextTick(() => {
    mosaicRef.value?.change();
  });
};

const selectResourcesHandle = (resource: I_resource) => {
  emits('selectResources', resource);
};

const clickResourceHandle = (resource: I_resource) => {
  currentPlayIndex.value = dataListWrapper.value.findIndex(
    (item) => item.id === resource.id,
  );
};

const clearUnavailablePlayback = () => {
  sourceRequestVersion++;
  currentPlayingDramaSeriesId.value = '';
  videoPlayRef.value?.pause();
  videoPlayRef.value?.resetPlayer();
};

const setVideoSource = async (dramaSeriesId: string) => {
  const player = videoPlayRef.value;
  if (!player || !dramaSeriesId) return;

  currentPlayingDramaSeriesId.value = dramaSeriesId;
  const requestVersion = ++sourceRequestVersion;
  const shouldResume = player.isPlaying() || false;
  player.pause();

  const { playUrl, playType } =
    await getPlayVideoURLAndType(dramaSeriesId);
  if (requestVersion !== sourceRequestVersion) return;

  const resource = dataListWrapper.value.find((item) =>
    item.dramaSeries.some(
      (dramaSeries) => dramaSeries.id === dramaSeriesId,
    ),
  );
  const dramaSeries = resource?.dramaSeries.find(
    (item) => item.id === dramaSeriesId,
  );

  player.setVideoSource(
    playUrl,
    playType,
    () => {
      if (requestVersion !== sourceRequestVersion) return;
      player.addTextTrack(
        `/api/video/subtitle/${dramaSeriesId}`,
        '默认字幕',
        'zh',
        true,
      );
      if (shouldResume) player.play();
    },
    dramaSeries?.src || resource?.title || '',
  );
};

const setVideoPlaySize = () => {
  const width = videoPlayContainerRef.value?.clientWidth;
  const height = videoPlayContainerRef.value?.clientHeight;
  const controllerHeight = videoPlayRef.value?.getControllerHeight() || 0;
  if (!width || !height) return;
  videoPlayRef.value?.setAspectRatio(
    `${width}:${Math.max(1, height - controllerHeight)}`,
  );
};

const getResourceDramaSeriesId = (index: number) => {
  return dataListWrapper.value[index]?.dramaSeries[0]?.id || '';
};

const prevHandle = () => {
  if (currentPlayIndex.value <= 0) return;
  currentPlayIndex.value--;
};

const nextHandle = () => {
  if (
    currentPlayIndex.value < 0 ||
    currentPlayIndex.value >= dataListWrapper.value.length - 1
  ) {
    return;
  }
  currentPlayIndex.value++;
};

const isInteractiveTarget = (target: EventTarget | null) => {
  if (!(target instanceof Element)) return false;
  return Boolean(
    target.closest(
      [
        'input',
        'textarea',
        'select',
        'button',
        'a',
        'video',
        'audio',
        '[contenteditable="true"]',
        '[role="slider"]',
        '[role="dialog"]',
        '[role="menu"]',
        '.el-dialog',
        '.el-drawer',
        '.el-popper',
      ].join(','),
    ),
  );
};

const hasOpenFloatingLayer = () => {
  return Array.from(
    document.querySelectorAll<HTMLElement>(
      '.el-overlay, .el-popper, .el-message-box',
    ),
  ).some(
    (layer) =>
      layer.getClientRects().length > 0 &&
      getComputedStyle(layer).visibility !== 'hidden',
  );
};

const handleKeyDown = (event: KeyboardEvent) => {
  if (
    event.altKey ||
    event.ctrlKey ||
    event.metaKey ||
    event.shiftKey ||
    isInteractiveTarget(event.target) ||
    hasOpenFloatingLayer()
  ) {
    return;
  }
  if (event.key === 'ArrowUp' || event.key === 'ArrowLeft') {
    event.preventDefault();
    prevHandle();
  } else if (event.key === 'ArrowDown' || event.key === 'ArrowRight') {
    event.preventDefault();
    nextHandle();
  }
};

onMounted(() => {
  syncCurrentPlayIndex();
  nextTick(() => {
    setVideoPlaySize();
    if (!videoPlayContainerRef.value) return;
    resizeObserver = new ResizeObserver(setVideoPlaySize);
    resizeObserver.observe(videoPlayContainerRef.value);
  });
  window.addEventListener('keydown', handleKeyDown);
});

onBeforeUnmount(() => {
  sourceRequestVersion++;
  videoPlayRef.value?.pause();
  resizeObserver?.disconnect();
  window.removeEventListener('keydown', handleKeyDown);
});

const change = () => {
  mosaicRef.value?.change();
};

defineExpose({ change });
</script>

<style lang="scss" scoped>
.layout-cover-poster-mosaic-short-video {
  display: flex;
  width: 100%;
  height: 100%;
  overflow: hidden;

  .mosaic-panel {
    min-width: 0;
    flex: 1;
  }

  .short-video-play {
    display: flex;
    width: 50%;
    padding: 10px;
    flex-shrink: 0;
    flex-direction: column;

    &.mobile-layout {
      width: 70%;
    }
  }

  .short-video-play-container {
    position: relative;
    min-height: 0;
    flex: 1;
  }

  .unavailable-message {
    position: absolute;
    z-index: 20;
    display: flex;
    color: var(--el-text-color-secondary);
    background: var(--el-bg-color);
    inset: 0;
    align-items: center;
    justify-content: center;
    font-size: 14px;
  }

  .short-video-play > .tool {
    display: flex;
    padding-top: 10px;
    flex-shrink: 0;
    justify-content: center;
  }
}
</style>
