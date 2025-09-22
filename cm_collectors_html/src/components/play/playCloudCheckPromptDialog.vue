<template>
  <dialogCommon ref="dialogCommonRef" title="云播提示" @submit="handleConfirm" width="500px" :footer="true"
    btnSubmitTitle="确定" btnCloseTitle="取消">
    <div class="play-cloud-check-content">
      <p>即将进行云播，请确保已安装云播插件。</p>
      <p>云播插件下载地址：<a @click="handleDownloadClick">{{ downloadUrl }}</a></p>
      <div class="play-cloud-check-checkbox">
        <el-checkbox v-model="noPromptChecked" label="不再提示" />
        <el-text type="warning">恢复提示请清理缓存</el-text>
      </div>

    </div>
  </dialogCommon>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import dialogCommon from '../com/dialog/dialog-common.vue';
import { playCloudCheck, setPlayCloudCheckComplete } from '../../common/play';


const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();
const noPromptChecked = ref(false);

const downloadUrl = window.location.origin + '/video_caller.zip';

const callbackFunction = ref<(() => void) | null>(null);

const handleConfirm = () => {
  if (noPromptChecked.value) {
    setPlayCloudCheckComplete();
  }
  if (callbackFunction.value) {
    callbackFunction.value();
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

const open = (fn: () => void) => {
  callbackFunction.value = fn;
  if (!playCloudCheck()) {
    dialogCommonRef.value?.open();
  } else {
    fn();
  }
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
</style>
