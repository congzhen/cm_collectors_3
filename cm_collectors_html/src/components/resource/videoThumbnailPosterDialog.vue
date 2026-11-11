<template>
  <dialogCommon ref="dialogCommonRef" title="视频关键帧截图" width="915px" top="10vh" btnSubmitTitle="设置封面海报"
    @submit="submitHandle" @closed="onClosed">
    <div class="keyframe-toolbar">
      <span>优先使用最近预览，刷新可重新获取。</span>
      <el-button :loading="loading" :disabled="loading" @click="refreshHandle">刷新</el-button>
    </div>
    <div class="video-keyframe-poster-dialog" v-loading="loading">
      <button v-for="item, index in base64Images" :key="index" type="button" class="keyframe-option"
        :class="{ selected: selectedIndex === index }" :aria-label="`选择第 ${index + 1} 个视频关键帧`"
        :aria-pressed="selectedIndex === index" :disabled="loading" @click="selectImageHandle(index)">
        <el-image :src="item" fit="contain" />
        <span v-if="selectedIndex === index" class="selection-badge">
          <el-icon>
            <Check />
          </el-icon>
          已选择
        </span>
      </button>
    </div>
  </dialogCommon>
</template>
<script lang="ts" setup>
import { debounceNow } from '@/assets/debounce';
import dialogCommon from '@/components/com/dialog/dialog-common.vue';
import { ffmpegServer } from '@/server/ffmpeg.server';
import { Check } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { onBeforeUnmount, ref } from 'vue';

const emits = defineEmits(['selectImage'])

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();

const loading = ref(false);
const base64Images = ref<string[]>([]);
const selectedIndex = ref<number>(0);

let requestSerial = 0;
let videoPath = '';
let currentFrameCount = 32;

const getVKeyFramePosters = async (refresh = false) => {
  const serial = ++requestSerial;
  loading.value = true;
  dialogCommonRef.value?.disabledSubmit(true);
  try {
    const result = await ffmpegServer.getVideoThumbnails(videoPath, currentFrameCount, refresh);
    if (serial !== requestSerial) return;
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    base64Images.value = result.data;
    selectedIndex.value = 0;
  } catch (error) {
    if (serial === requestSerial) ElMessage.error('获取视频预览图失败，请重试');
  } finally {
    if (serial === requestSerial) {
      loading.value = false;
      dialogCommonRef.value?.disabledSubmit(base64Images.value.length === 0);
    }
  }
}

const refreshHandle = () => {
  if (!loading.value) void getVKeyFramePosters(true);
}
const onClosed = () => {
  requestSerial++;
  loading.value = false;
}

onBeforeUnmount(onClosed);

const selectImageHandle = (index: number) => {
  selectedIndex.value = index;
}

const submitHandle = debounceNow(() => {
  if (loading.value || !base64Images.value[selectedIndex.value]) return;
  close();
  emits('selectImage', base64Images.value[selectedIndex.value])
})

const open = (_videoPath: string, frameCount: number) => {
  dialogCommonRef.value?.open();
  videoPath = _videoPath;
  currentFrameCount = frameCount;
  base64Images.value = [];
  selectedIndex.value = 0;
  void getVKeyFramePosters();
}
const close = () => {
  dialogCommonRef.value?.close();
}

defineExpose({ open })
</script>
<style scoped lang="scss">
.keyframe-toolbar {
  display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 12px;
  span { color: var(--el-text-color-secondary); font-size: 12px; }
}

.video-keyframe-poster-dialog {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  width: 100%;
  height: 70vh;
  overflow-y: auto;

  .keyframe-option {
    position: relative;
    display: flex;
    align-self: flex-start;
    padding: 3px;
    overflow: hidden;
    background: var(--el-fill-color-light);
    border: 2px solid var(--el-border-color-darker);
    border-radius: 8px;
    cursor: pointer;
    transition: border-color 0.15s ease, box-shadow 0.15s ease, transform 0.15s ease;

    &:hover {
      border-color: var(--el-color-primary);
      transform: translateY(-1px);
    }

    &:focus-visible {
      outline: 3px solid var(--el-color-primary-light-5);
      outline-offset: 2px;
    }

    &.selected {
      background: var(--el-color-primary-light-9);
      border-color: var(--el-color-primary);
      box-shadow: 0 0 0 3px var(--el-color-primary-light-5), 0 4px 12px rgb(0 0 0 / 18%);
    }

    .el-image {
      display: block;
      max-width: 200px;
      border-radius: 4px;
    }

    .selection-badge {
      position: absolute;
      top: 8px;
      right: 8px;
      display: inline-flex;
      gap: 4px;
      align-items: center;
      padding: 4px 8px;
      color: #fff;
      font-size: 13px;
      font-weight: 600;
      line-height: 1;
      background: var(--el-color-primary);
      border: 1px solid rgb(255 255 255 / 80%);
      border-radius: 999px;
      box-shadow: 0 2px 8px rgb(0 0 0 / 35%);
      pointer-events: none;
    }
  }
}

:global(html.bright) .video-keyframe-poster-dialog .keyframe-option {
  background: #fff;
  border-color: #909399;
}

:global(html.bright) .video-keyframe-poster-dialog .keyframe-option.selected {
  background: #ecf5ff;
  border-color: #337ecc;
  box-shadow: 0 0 0 3px #a0cfff, 0 4px 12px rgb(0 0 0 / 20%);
}

:global(html.dark) .video-keyframe-poster-dialog .keyframe-option {
  background: #1d1e1f;
  border-color: #606266;
}

:global(html.dark) .video-keyframe-poster-dialog .keyframe-option:hover {
  border-color: #79bbff;
}

:global(html.dark) .video-keyframe-poster-dialog .keyframe-option.selected {
  background: #18222c;
  border-color: #79bbff;
  box-shadow: 0 0 0 3px rgb(64 158 255 / 55%), 0 4px 14px rgb(0 0 0 / 55%);
}
</style>
