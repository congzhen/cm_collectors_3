<template>
  <el-dialog v-model="dialogVisible" :title="title_C" width="458px" class="video-play-dialog" top="20px"
    header-class="video-play-dialog-header" :modal="false" :modal-penetrable="true" :draggable="true" :z-index="99999"
    :close-on-click-modal="false" append-to-body @close="closeHandle">
    <div v-loading="loading">
      <div class="video-play">
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
    padding-bottom: 0px !important;

    .el-dialog__title {
      font-size: 12px;

      display: inline-block;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      max-width: 100%;
      box-sizing: border-box;
    }

    .el-dialog__headerbtn {
      width: 32px;
      height: 32px;

      .el-icon {
        margin-top: 4px;
      }
    }
  }
}
</style>
