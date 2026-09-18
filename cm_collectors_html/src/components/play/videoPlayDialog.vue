<template>
  <el-dialog v-model="dialogVisible" :width="dialogWidth_C" class="video-play-dialog" top="20px" :modal="false"
    :show-close="false" :modal-penetrable="true" :draggable="true" :fullscreen="fullScreenDisplay"
    :close-on-click-modal="false" append-to-body @close="closeHandle">
    <template #header="{ close }">
      <div class="video-play-dialog-header">
        <div class="title">
          <el-icon size="20">
            <VideoPlay />
          </el-icon>
          <h5>
            {{ title_C }}
          </h5>
        </div>
        <div class="btn">
          <el-icon @click="toggleVideoPlayer">
            <ArrowUp v-if="!videoPlayerVisible" />
            <ArrowDown v-else />
          </el-icon>
          <el-icon @click="toggleFullScreenDisplay">
            <FullScreen v-if="!fullScreenDisplay" />
            <BottomLeft v-else />
          </el-icon>

          <el-icon @click="close">
            <Close />
          </el-icon>
        </div>
      </div>
    </template>

    <div ref="videoPlayElementRef" v-loading="loading" class="video-play"
      :style="{ display: videoPlayerVisible ? 'block' : 'none' }">
      <ResourceVideoPlayer v-if="dialogVisible && resourceInfo && selectedDramaSeriesId" ref="videoPlayRef"
        :resource="resourceInfo" :drama-series-id="selectedDramaSeriesId" @dimensions="onDimensions" />
    </div>

  </el-dialog>
</template>
<script lang="ts" setup>
import { ref, computed, nextTick } from 'vue';
import ResourceVideoPlayer from './resourceVideoPlayer.vue';
import { resourceServer } from '@/server/resource.server';
import { ElMessage } from 'element-plus';
import type { I_resource } from '@/dataType/resource.dataType';


const videoPlayElementRef = ref<HTMLDivElement>();
const videoPlayRef = ref<InstanceType<typeof ResourceVideoPlayer>>();

const dialogVisible = ref(false);
const loading = ref(false);
const selectedDramaSeriesId = ref('');
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
  const resource = await getResourceInfo(resourceId);
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

const getResourceInfo = async (resourceId: string): Promise<I_resource | undefined> => {
  try {
    loading.value = true;
    const result = await resourceServer.info(resourceId);
    if (!result || !result.status) {
      ElMessage.error(result.msg);
      return undefined;
    }
    return result.data;
  } catch (error) {
    console.log(error)
  } finally {
    loading.value = false;
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

const open = (_resourceId: string, _dramaSeriesId: string) => {
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
  padding: 4px;
  max-width: calc(100vw - 16px);
  border: 1px solid #434344;
  display: flex;
  flex-direction: column;

  .el-dialog__header {
    flex-shrink: 0;
    padding-bottom: 2px !important;
    overflow: hidden;
  }

  .video-play-dialog-header {
    display: flex;
    gap: 10px;
    justify-content: space-between;
    padding-bottom: 0px;

    .title {
      flex-grow: 1;
      display: flex;
      gap: 5px;
      min-width: 0; // 添加此属性以允许 flex 项目收缩

      h5 {
        flex-grow: 1;
        font-size: 12px;
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
      gap: 5px;

      .el-icon {
        font-size: 20px;
        cursor: pointer;

        &:hover {
          color: var(--el-color-primary);
        }
      }
    }
  }

  .el-dialog__body {
    flex: 1;
    overflow: hidden;

    .video-play {
      width: 100%;
      height: 100%;
      overflow: hidden;
    }
  }
}
</style>
