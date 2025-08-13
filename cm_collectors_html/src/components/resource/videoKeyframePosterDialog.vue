<template>
  <dialogCommon ref="dialogCommonRef" title="视频关键帧截图" width="880px">
    <div class="video-keyframe-poster-dialog" v-loading="loading">
      <el-image v-for="item, index in base64Images" :key="index" :src="item" fit="contain" />
    </div>
  </dialogCommon>
</template>
<script lang="ts" setup>
import dialogCommon from '@/components/com/dialog/dialog-common.vue';
import { ffmpegServer } from '@/server/ffmpeg.server';
import { ElMessage } from 'element-plus';
import { nextTick, ref } from 'vue';

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();

const loading = ref(false);
const base64Images = ref<string[]>([]);

const init = async (_videoPath: string) => {
  base64Images.value = [];
  await getVKeyFramePosters(_videoPath)
}

const getVKeyFramePosters = async (_videoPath: string) => {
  try {
    loading.value = true;
    const result = await ffmpegServer.getVideoKeyFramePosters(_videoPath, 30);
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    base64Images.value = result.data;
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }
}


const open = (_videoPath: string) => {
  dialogCommonRef.value?.open();
  nextTick(() => {
    init(_videoPath);
  });
}

defineExpose({ open })
</script>
<style scoped lang="scss">
.video-keyframe-poster-dialog {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  width: 100%;
  height: 400px;
  overflow-y: auto;

  .el-image {
    max-width: 200px;
  }
}
</style>
