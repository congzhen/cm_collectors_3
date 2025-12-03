<template>
  <dialogCommon ref="dialogCommonRef" title="云播提示" @submit="handleConfirm" width="500px" :footer="true"
    btnSubmitTitle="确定" btnCloseTitle="取消">
    <div class="play-cloud-check-content">
      <p>即将进行云播，请确保已安装云播插件。</p>
      <p>云播插件下载地址：<a @click="handleDownloadClick">{{ downloadUrl }}</a></p>

      <div class="play-cloud-mode">
        <label>云播方式：</label>
        <el-radio-group v-model="playCloudMode" size="small" @change="setPlayCloudMode">
          <el-radio-button label="原始流" value="mp4" />
          <el-radio-button label="m3u8" value="m3u8" />
        </el-radio-group>
      </div>
      <!--
      <div class="play-cloud-check-checkbox">
        <el-checkbox v-model="noPromptChecked" label="不再提示" />
        <el-text type="warning">恢复提示请清理缓存</el-text>
      </div>
      -->
    </div>
  </dialogCommon>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import dialogCommon from '../com/dialog/dialog-common.vue';
import { setPlayCloudCheckComplete } from '../../common/play';
import type { T_VideoPlayMode } from '@/dataType/app.dataType';
import { playCloud } from './playCloud';

// 定义本地存储的键名
const CLOUD_PLAYER_MODE_KEY = 'cloud-player-mode';

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();
const noPromptChecked = ref(false);

const downloadUrl = window.location.origin + '/video_caller.zip';

const dramaSeriesId = ref('');
const callbackFunction = ref<((mode: T_VideoPlayMode) => void) | null>(null);

const playCloudMode = ref<T_VideoPlayMode>('m3u8')

const init = (_dramaSeriesId: string) => {
  dramaSeriesId.value = _dramaSeriesId;
  playCloudMode.value = getPlayCloudMode()
};

const setPlayCloudMode = (mode: T_VideoPlayMode) => {
  localStorage.setItem(CLOUD_PLAYER_MODE_KEY, mode)

}
const getPlayCloudMode = (): T_VideoPlayMode => {
  const mode = localStorage.getItem(CLOUD_PLAYER_MODE_KEY)
  if (mode) {
    return mode as T_VideoPlayMode
  }
  return 'm3u8'
}

const handleConfirm = async () => {
  if (noPromptChecked.value) {
    setPlayCloudCheckComplete();
  }
  await playCloud(dramaSeriesId.value, playCloudMode.value);
  if (callbackFunction.value) {
    callbackFunction.value(playCloudMode.value);
  }
  dialogCommonRef.value?.close();
};

const handleDownloadClick = async (event: Event) => {
  // 阻止默认行为，我们自己处理下载
  event.preventDefault();

  try {
    // 检查是否有访问父窗口的权限（即是否在 iframe 中）
    let hasParentAccess = false;
    try {
      hasParentAccess = !!(window.top && window.top !== window.self);
    } catch {
      // 如果访问被拒绝（安全错误），则认为在 iframe 中但无访问权限
      hasParentAccess = false;
    }

    if (hasParentAccess) {
      // 在 iframe 中且可以访问父窗口，通过父窗口打开下载
      if (window.top) {
        window.top.location.href = downloadUrl;
      } else {
        window.location.href = downloadUrl;
      }
    } else {
      // 不在 iframe 中或无法访问父窗口，直接打开下载
      window.location.href = downloadUrl;
    }
  } catch (error) {
    console.error('下载插件失败:', error);
    alert('无法下载插件，请检查网络连接或手动下载。');
  }
};

const open = (_dramaSeriesId: string, fn?: (mode: T_VideoPlayMode) => void) => {
  init(_dramaSeriesId);
  if (fn) {
    callbackFunction.value = fn;
  }
  dialogCommonRef.value?.open();
  /*
  if (!playCloudCheck()) {
    dialogCommonRef.value?.open();
  } else {
    if (callbackFunction.value) {
      callbackFunction.value(playCloudMode.value);
    }
  }
  */
};

defineExpose({
  open,
});
</script>

<style scoped>
.play-cloud-check-content {
  line-height: 1.6;
}

.play-cloud-check-content p {
  margin-bottom: 15px;
}

.play-cloud-check-content a {
  color: #409eff;
  text-decoration: underline;
  cursor: pointer;
}

.play-cloud-check-checkbox {
  display: flex;
  gap: 20px;

  .el-checkbox {
    display: block;
    margin-top: 9px;
  }
}

.play-cloud-mode {
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
