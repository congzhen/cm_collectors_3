<template>
  <el-dialog v-model="dialogVisible" :width="dialogWidth_C" class="video-play-dialog" top="20px" :modal="false"
    :show-close="false" :modal-penetrable="true" :draggable="true" :fullscreen="fullScreenDisplay"
    :close-on-click-modal="false" append-to-body @close="closeHandle">
    <template #header="{ close }">
      <div class="video-play-dialog-header">
        <div class="title">
          <el-icon class="player-title-icon" size="16">
            <VideoPlay />
          </el-icon>
          <h5 :title="title_C">
            {{ title_C }}
          </h5>
        </div>
        <div class="btn" @mousedown.stop>
          <button class="player-window-action" type="button"
            :title="videoPlayerVisible ? '收起播放器' : '展开播放器'"
            :aria-label="videoPlayerVisible ? '收起播放器' : '展开播放器'" @click="toggleVideoPlayer">
            <el-icon>
            <ArrowUp v-if="!videoPlayerVisible" />
            <ArrowDown v-else />
            </el-icon>
          </button>
          <button class="player-window-action" type="button"
            :title="fullScreenDisplay ? '恢复窗口' : '最大化窗口'"
            :aria-label="fullScreenDisplay ? '恢复窗口' : '最大化窗口'" @click="toggleFullScreenDisplay">
            <el-icon>
            <FullScreen v-if="!fullScreenDisplay" />
            <BottomLeft v-else />
            </el-icon>
          </button>
          <button class="player-window-action player-window-action--close" type="button"
            title="关闭" aria-label="关闭播放器" @click="close">
            <el-icon><Close /></el-icon>
          </button>
        </div>
      </div>
    </template>

    <div ref="videoPlayElementRef" v-loading="loading" class="video-play"
      :style="{ display: videoPlayerVisible ? 'block' : 'none' }">
      <ResourceVideoPlayer v-if="dialogVisible && resourceInfo && selectedDramaSeriesId" ref="videoPlayRef"
        :key="playbackSession" :resource="resourceInfo" :drama-series-id="selectedDramaSeriesId" :initial-moment="initialMoment" @dimensions="onDimensions" />
    </div>

  </el-dialog>
</template>
<script lang="ts" setup>
import { ref, computed, nextTick } from 'vue';
import ResourceVideoPlayer from './resourceVideoPlayer.vue';
import { resourceServer } from '@/server/resource.server';
import { ElMessage } from 'element-plus';
import type { I_resource } from '@/dataType/resource.dataType';
import type { MomentPlayback } from '@/server/videoMoments.server';


const videoPlayElementRef = ref<HTMLDivElement>();
const videoPlayRef = ref<InstanceType<typeof ResourceVideoPlayer>>();

const dialogVisible = ref(false);
const loading = ref(false);
const selectedDramaSeriesId = ref('');
const initialMoment = ref<MomentPlayback>();
const playbackSession = ref(0);
let requestVersion = 0;

const videoPlayerVisible = ref(true); // 控制视频播放器显示状态
const fullScreenDisplay = ref(false); // 控制全屏显示状态

const videoVertical = ref(false); // 视频为竖屏否

const resourceInfo = ref<I_resource>();

const title_C = computed(() => {
  return resourceInfo.value?.title || '视频播放器';
})
const dialogWidth_C = computed(() => {
  return videoVertical.value ? '458px' : '640px';
})


const init = async (resourceId: string, _dramaSeriesId: string) => {
  const request = ++requestVersion;
  const resource = await getResourceInfo(resourceId, request);
  if (request !== requestVersion) return;
  resourceInfo.value = resource;
  if (!resourceInfo.value) {
    return;
  }
  const playerDramaSeriesId = _dramaSeriesId || (resourceInfo.value.dramaSeries.length > 0 ? resourceInfo.value.dramaSeries[0].id : '');
  if (playerDramaSeriesId) {
    selectedDramaSeriesId.value = playerDramaSeriesId;
  }
};

const getResourceInfo = async (resourceId: string, request: number): Promise<I_resource | undefined> => {
  try {
    loading.value = true;
    const result = await resourceServer.info(resourceId);
    if (request !== requestVersion || !dialogVisible.value) return undefined;
    if (!result || !result.status) {
      ElMessage.error(result?.msg || '资源加载失败');
      return undefined;
    }
    return result.data;
  } catch (error) {
    if (request === requestVersion && dialogVisible.value) ElMessage.error('资源加载失败，请重试')
  } finally {
    if (request === requestVersion) loading.value = false;
  }
};

const onDimensions = (width: number, height: number) => {
  videoVertical.value = height > width;
  setVideoDimensions(width, height);
};

const setVideoDimensions = (w: number, h: number) => {
  videoPlayRef.value?.setAspectRatio(w + ':' + h)
}

// 切换视频播放器显示状态
const toggleVideoPlayer = () => {
  videoPlayerVisible.value = !videoPlayerVisible.value;
};
// 切换全屏显示状态
const toggleFullScreenDisplay = () => {
  fullScreenDisplay.value = !fullScreenDisplay.value;
  nextTick(() => {
    if (fullScreenDisplay.value) {
      const ep = videoPlayElementRef.value || undefined;
      if (ep) {
        // 获取html的宽高
        const { width, height } = ep.getBoundingClientRect();
        const controllerHeight = videoPlayRef.value?.getControllerHeight() || 0;
        setVideoDimensions(width, height - controllerHeight);
      }
    } else {
      const vp = videoPlayRef.value || undefined;
      if (vp) {
        const dimensions = vp.getVideoDimensions();
        if (dimensions) {
          setVideoDimensions(dimensions.width, dimensions.height);
        }
      }
    }
  });
};

const closeHandle = () => {
  fullScreenDisplay.value = false;
  requestVersion++;
  selectedDramaSeriesId.value = '';
  resourceInfo.value = undefined;
}

const open = (_resourceId: string, _dramaSeriesId: string, moment?: MomentPlayback) => {
  playbackSession.value++;
  initialMoment.value = moment;
  selectedDramaSeriesId.value = '';
  videoPlayerVisible.value = true;
  void init(_resourceId, _dramaSeriesId)
  dialogVisible.value = true
}
const close = () => {

  dialogVisible.value = false
}

defineExpose({ open, close })
</script>
<style lang="scss">
.video-play-dialog {
  padding: 0;
  max-width: calc(100vw - 16px);
  border: 1px solid var(--el-border-color);
  border-radius: 12px;
  background: var(--el-bg-color-overlay);
  box-shadow: 0 18px 64px #0006;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  &.is-fullscreen {
    max-width: none;
    border-radius: 0;
    box-shadow: none;
  }

  .el-dialog__header {
    flex-shrink: 0;
    margin: 0;
    padding: 7px 10px;
    border-bottom: 1px solid var(--el-border-color-lighter);
    overflow: hidden;
  }

  .video-play-dialog-header {
    display: flex;
    align-items: center;
    min-height: 30px;
    gap: 10px;
    justify-content: space-between;

    .title {
      flex: 1;
      display: flex;
      align-items: center;
      gap: 9px;
      min-width: 0;

      .player-title-icon {
        flex: 0 0 28px;
        width: 28px;
        height: 28px;
        border-radius: 8px;
        color: var(--el-color-primary);
        background: var(--el-color-primary-light-9);
      }

      h5 {
        flex: 1;
        min-width: 0;
        margin: 0;
        font-size: 13px;
        font-weight: 500;
        line-height: 20px;
        color: var(--el-text-color-primary);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 100%;
        box-sizing: border-box;
      }
    }


    .btn {
      flex-shrink: 0;
      display: flex;
      align-items: center;
      gap: 4px;

      .player-window-action {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 30px;
        height: 30px;
        padding: 0;
        border: 0;
        border-radius: 7px;
        background: transparent;
        color: var(--el-text-color-secondary);
        cursor: pointer;
        transition: background-color 0.15s, color 0.15s;

        .el-icon { font-size: 16px; }

        &:hover {
          color: var(--el-text-color-primary);
          background: var(--el-fill-color);
        }

        &:focus-visible {
          outline: 2px solid var(--el-color-primary);
          outline-offset: 1px;
        }

        &--close:hover {
          color: var(--el-color-danger);
          background: var(--el-color-danger-light-9);
        }
      }
    }
  }

  .el-dialog__body {
    flex: 1;
    min-height: 0;
    padding: 4px;
    overflow: hidden;

    .video-play {
      width: 100%;
      height: 100%;
      overflow: hidden;
    }
  }
}
</style>
