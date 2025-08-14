<template>
  <dialogCommon ref="dialogCommonRef" title="视频关键帧截图" width="900px" top="10vh" btnSubmitTitle="设置封面海报"
    @submit="submitHandle">
    <div class="video-keyframe-poster-dialog" v-loading="loading">
      <el-image :class="[selectedIndex == index ? 'selected' : 'noSelected']" v-for="item, index in base64Images"
        :key="index" :src="item" fit="contain" @click="selectImageHandle(index)" />
    </div>
  </dialogCommon>
</template>
<script lang="ts" setup>
import { debounceNow } from '@/assets/debounce';
import dialogCommon from '@/components/com/dialog/dialog-common.vue';
import { ffmpegServer } from '@/server/ffmpeg.server';
import { ElMessage } from 'element-plus';
import { nextTick, ref } from 'vue';

const emits = defineEmits(['selectImage'])

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();

const loading = ref(false);
const base64Images = ref<string[]>([]);
const selectedIndex = ref<number>(0);

const init = async (_videoPath: string) => {
  base64Images.value = [];
  await getVKeyFramePosters(_videoPath)
}

const getVKeyFramePosters = async (_videoPath: string) => {
  try {
    loading.value = true;
    const result = await ffmpegServer.getVideoThumbnails(_videoPath, 32);
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

const selectImageHandle = (index: number) => {
  selectedIndex.value = index;
}

const submitHandle = debounceNow(() => {
  close();
  emits('selectImage', base64Images.value[selectedIndex.value])
})

const open = (_videoPath: string) => {
  dialogCommonRef.value?.open();
  nextTick(() => {
    init(_videoPath);
  });
}
const close = () => {
  dialogCommonRef.value?.close();
}

defineExpose({ open })
</script>
<style scoped lang="scss">
.video-keyframe-poster-dialog {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  width: 100%;
  height: 70vh;
  overflow-y: auto;

  .el-image {
    max-width: 200px;
    border-radius: 5px;
  }

  .selected {
    border: 1px solid #E6A23C;

  }

  .noSelected {
    border: 1px solid #303133;
  }
}
</style>
