<template>
  <el-dialog v-model="dialogVisible" width="458px" class="video-play-dialog" top="20px" :modal="false"
    :show-close="false" :modal-penetrable="true" :draggable="true" :z-index="99999" :close-on-click-modal="false"
    append-to-body @close="closeHandle">
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
          <el-icon @click="close">
            <Close />
          </el-icon>
        </div>
      </div>
    </template>
    <div v-loading="loading">
      <div class="video-play" :style="{ display: videoPlayerVisible ? 'block' : 'none' }">
        <videoPlay ref="videoPlayRef" />
      </div>
    </div>
  </el-dialog>
</template>
<script lang="ts" setup>
import { ref, computed } from 'vue';
import videoPlay from './videoPlay.vue';
import { resourceServer } from '@/server/resource.server';
import { ElMessage } from 'element-plus';
import type { I_resource } from '@/dataType/resource.dataType';

const videoPlayRef = ref<InstanceType<typeof videoPlay>>();

const dialogVisible = ref(false);
const loading = ref(false);
const isPlaying = ref(false);

const videoPlayerVisible = ref(true); // 控制视频播放器显示状态

const resourceInfo = ref<I_resource>();

const title_C = computed(() => {
  return resourceInfo.value?.title || '视频播放器';
})


const init = async (resourceId: string, _dramaSeriesId: string) => {
  isPlaying.value = videoPlayRef.value?.isPlaying() || false;
  resourceInfo.value = await getResourceInfo(resourceId);
  if (!resourceInfo.value) {
    return;
  }
  const playerDramaSeriesId = _dramaSeriesId || (resourceInfo.value.dramaSeries.length > 0 ? resourceInfo.value.dramaSeries[0].id : '');
  if (playerDramaSeriesId) {
    await setVideoSource(playerDramaSeriesId);
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

const setVideoSource = (dramaSeriesId: string) => {
  const vp = videoPlayRef.value;
  if (!vp) return;
  vp.setVideoSource('/api/video/mp4/' + dramaSeriesId, 'mp4', () => {
    vp.addTextTrack(
      `/api/video/subtitle/${dramaSeriesId}`,
      '默认字幕',
      'zh',
      true // 设为默认字幕
    )
    const dimensions = vp.getVideoDimensions();
    if (dimensions) {
      vp.setAspectRatio(dimensions.width + ':' + dimensions.height);
    }
    if (isPlaying.value) {
      vp.play();
    }
  });
}

// 切换视频播放器显示状态
const toggleVideoPlayer = () => {
  videoPlayerVisible.value = !videoPlayerVisible.value;
};

const closeHandle = () => {
  videoPlayRef.value?.resetPlayer();
}

const open = (_resourceId: string, _dramaSeriesId: string) => {
  init(_resourceId, _dramaSeriesId)
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
  border: 1px solid #434344;

  .el-dialog__header {
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
}
</style>
