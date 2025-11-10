<template>
  <dialogCommon ref="dialogCommonRef" title="检测更新" width="860px" top="15vh" :footer="false">
    <div class="updateSoftView">
      <div class="updateSoftView-header">

        <div class="version-info">
          <div class="version-item">
            <span class="version-label">当前版本：</span>
            <span class="version-number">{{ currentVersion }}</span>
          </div>
          <div class="version-item">
            <span class="version-label">最新版本：</span>
            <span class="version-number" :class="{ 'update-available': hasUpdate }">{{ latestVersion || '检查中...'
            }}</span>
          </div>
        </div>
      </div>

      <div class="updateSoftView-content">
        <div v-if="checking" class="checking-status">
          正在检查更新...
        </div>

        <div v-else-if="errorMessage" class="error-status">
          <div>检测更新失败</div>
          <div class="error-message">{{ errorMessage }}</div>
        </div>
        <div v-else>
          <div v-if="hasUpdate">
            <div class="update-download">
              <p class="update-download-version">v{{ latestVersion }}</p>
              <p>windows 下载地址:</p>
              <p v-if="updateInfo?.windowsDownloadUrl">{{ updateInfo.windowsDownloadUrl }}</p>
              <p v-else>暂无</p>
            </div>
            <div class="update-download">
              <p class="update-download-version">v{{ latestVersion }}</p>
              <p>linux 下载地址:</p>
              <p v-if="updateInfo?.linuxDownloadUrl">{{ updateInfo.linuxDownloadUrl }}</p>
              <p v-else>暂无</p>
            </div>
            <div class="update-download">
              <p class="update-download-version">v{{ latestVersion }}</p>
              <p>docker:full 下载地址:</p>
              <p v-if="updateInfo?.dockerDownloadUrl">{{ updateInfo.dockerDownloadUrl }}</p>
              <p v-else>暂无</p>
            </div>
            <div class="update-download">
              <p class="update-download-version">v{{ latestVersion }}</p>
              <p>docker:minimal 下载地址:</p>
              <p v-if="updateInfo?.dockerDownloadUrl">{{ updateInfo.dockerMinimalDownloadUrl }}</p>
              <p v-else>暂无</p>
            </div>
            <div class="update-download">
              <p class="update-download-version">v{{ latestVersion }}</p>
              <p>云播脚本 下载地址:</p>
              <p v-if="updateInfo?.cloudPlayScriptUrl">{{ updateInfo.cloudPlayScriptUrl }}</p>
              <p v-else>暂无</p>
            </div>
          </div>
          <div v-else>
            您的软件是最新版本
          </div>
        </div>

        <div class="action-section">
          <button class="btn-refresh" @click="checkForUpdate" :disabled="checking">
            {{ checking ? '检查中...' : '重新检查' }}
          </button>
        </div>

        <div class="changelog-section" v-if="updateInfo?.changelog && updateInfo.changelog.length > 0">
          <div class="updateSoftView-content-title">更新日志：</div>
          <ul>
            <li v-for="(log, index) in updateInfo.changelog" :key="index">{{ log }}</li>
          </ul>
        </div>
      </div>
    </div>
  </dialogCommon>
</template>
<script setup lang="ts">
import request from '@/assets/request';
import dialogCommon from '@/components/com/dialog/dialog-common.vue';
import { computed, ref } from 'vue';
const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();

// 定义更新信息接口
interface UpdateInfo {
  version: string;
  changelog: string[];
  windowsDownloadUrl?: string;
  linuxDownloadUrl?: string;
  dockerDownloadUrl?: string;
  dockerMinimalDownloadUrl?: string;
  cloudPlayScriptUrl?: string;
}

const currentVersion = ref('3.0.5')
const latestVersion = ref('')
const updateInfo = ref<UpdateInfo | null>(null)
const checking = ref(false)
const errorMessage = ref('')

// 版本比较函数
const compareVersions = (v1: string, v2: string): number => {
  const parts1 = v1.split('.').map(Number)
  const parts2 = v2.split('.').map(Number)

  for (let i = 0; i < Math.max(parts1.length, parts2.length); i++) {
    const num1 = i < parts1.length ? parts1[i] : 0
    const num2 = i < parts2.length ? parts2[i] : 0

    if (num1 > num2) return 1
    if (num1 < num2) return -1
  }

  return 0
}

const hasUpdate = computed(() => {
  if (!latestVersion.value) return false
  return compareVersions(latestVersion.value, currentVersion.value) > 0
})

// 检查更新
const checkForUpdate = async () => {
  checking.value = true
  errorMessage.value = ''

  try {
    // 发送网络请求获取更新信息

    const response = await request<string>({
      url: '/updateSoftConfig',
      method: 'GET'
    })

    if (response.status) {
      const data = JSON.parse(response.data)
      updateInfo.value = data
      latestVersion.value = data.version
    } else {
      errorMessage.value = response.msg || '检查更新失败，请稍后重试'
    }
  } catch (error: unknown) {
    console.error('检查更新时发生错误:', error)
    errorMessage.value = '网络连接失败，请检查您的网络设置后重试'
  } finally {
    checking.value = false
  }
}


const open = () => {
  dialogCommonRef.value?.open()
  checkForUpdate()
}

defineExpose({
  open,
})
</script>
<style scoped lang="scss">
.updateSoftView {
  margin: 0 auto;
  padding: 20px;
  background-color: #1F1F1F;

  .updateSoftView-header {
    text-align: center;
    margin-bottom: 10px;
    padding-bottom: 20px;
  }

  .version-info {
    display: flex;
    justify-content: center;
    gap: 30px;
    flex-wrap: wrap;
  }

  .version-item {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .version-label {
    font-size: 16px;
  }

  .version-number {
    font-size: 18px;
    font-weight: 600;
    color: #fff;

    &.update-available {
      color: #e74c3c;
    }
  }

  &-content {
    border-radius: 10px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
    padding: 25px;
  }

  .update-download {
    margin: 10px 0;
    font-size: 14px;

    .update-download-version {
      font-size: 18px;
      font-weight: 700;
    }
  }

  .checking-status,
  .error-status {
    text-align: center;
    padding: 20px;
    font-size: 16px;
  }

  .error-message {
    color: #e74c3c;
    margin-top: 10px;
  }

  .action-section {
    margin: 25px 0;
    text-align: center;

    .btn-refresh {
      background-color: #42b983;
      color: white;
      border: none;
      padding: 12px 24px;
      border-radius: 6px;
      font-size: 16px;
      font-weight: 500;
      cursor: pointer;
      transition: background-color 0.3s;

      &:hover:not([disabled]) {
        background-color: #359c6d;
      }

      &[disabled] {
        opacity: 0.7;
        cursor: not-allowed;
      }
    }
  }

  &-content-title {
    font-weight: bold;
    margin: 15px 0 5px 0;
  }

  ul {
    li {
      margin: 8px 0;
    }
  }

  a {
    color: #42b983;
    text-decoration: none;

    &:hover {
      text-decoration: underline;
    }
  }
}
</style>
