<template>
  <div class="avatar-library-setting" v-loading="loading">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <div>
            <h3>Gfriends 演员头像库</h3>
            <p>本功能只下载和更新演员头像，不修改演员的其他资料。</p>
          </div>
          <div class="card-actions">
            <el-button type="danger" plain icon="Delete" :loading="clearing" :disabled="status.cachedImages === 0"
              @click="clearImageCache">清理头像缓存</el-button>
            <el-button type="success" icon="Refresh" :loading="updating" @click="updateDataFile">
              {{ status.ready ? '更新数据文件' : '下载数据文件' }}
            </el-button>
          </div>
        </div>
      </template>

      <el-result v-if="!status.ready" icon="warning" title="尚未下载头像库数据">
        <template #sub-title>点击“下载数据文件”后即可查询和批量匹配头像。</template>
      </el-result>
      <el-descriptions v-if="!status.ready" :column="2" border class="cache-summary">
        <el-descriptions-item label="已缓存头像">{{ status.cachedImages }} 张</el-descriptions-item>
        <el-descriptions-item label="头像缓存大小">{{ formatBytes(status.cacheSize) }}</el-descriptions-item>
      </el-descriptions>
      <el-descriptions v-else :column="2" border>
        <el-descriptions-item label="状态"><el-tag type="success">可用</el-tag></el-descriptions-item>
        <el-descriptions-item label="本地文件">{{ formatBytes(status.fileSize) }}</el-descriptions-item>
        <el-descriptions-item label="本地更新时间">{{ formatDate(status.updatedAt) }}</el-descriptions-item>
        <el-descriptions-item label="文件树时间戳">{{ status.dataTimestamp || '-' }}</el-descriptions-item>
        <el-descriptions-item label="头像数量">{{ status.totalNum || '-' }}</el-descriptions-item>
        <el-descriptions-item label="上游标记大小">{{ status.totalSize || '-' }}</el-descriptions-item>
        <el-descriptions-item label="已缓存头像">{{ status.cachedImages }} 张</el-descriptions-item>
        <el-descriptions-item label="头像缓存大小">{{ formatBytes(status.cacheSize) }}</el-descriptions-item>
        <el-descriptions-item label="当前数据源" :span="2">{{ status.activeBaseUrl }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card shadow="never">
      <template #header><h3>默认设置</h3></template>
      <el-form :model="setting" label-width="150px" label-position="left">
        <el-form-item label="默认头像策略">
          <el-radio-group v-model="setting.defaultStrategy">
            <el-radio-button value="recommended">推荐头像</el-radio-button>
            <el-radio-button value="original">原图优先</el-radio-button>
            <el-radio-button value="aiFix">AI 优化图优先</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="自定义镜像地址">
          <el-input v-model="setting.customBaseUrl" clearable placeholder="留空时使用 GitHub Raw，并在失败时尝试备用镜像" />
          <div class="field-help">地址需要指向包含 Filetree.json 和 Content 目录的仓库根路径。</div>
        </el-form-item>
        <el-form-item label="启动时清理缓存">
          <el-switch v-model="setting.clearCacheOnStartup" />
          <div class="field-help field-help--inline">开启后，每次软件启动都会清理已下载的头像缓存，不影响头像库数据文件和演员已使用的头像。</div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Check" @click="saveSetting">保存设置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { performerAvatarLibraryServer } from '@/server/performerAvatarLibrary.server';
import type { I_performerAvatarLibrarySetting, I_performerAvatarLibraryStatus } from '@/dataType/performerAvatarLibrary.dataType';
import { appDataServer } from '@/server/app.server';
import type { I_appSystemConfig } from '@/dataType/app.dataType';

const emptyStatus = (): I_performerAvatarLibraryStatus => ({
  ready: false, updating: false, fileSize: 0, cachedImages: 0, cacheSize: 0,
  updatedAt: '', dataTimestamp: '', totalNum: '', totalSize: '', activeBaseUrl: '',
  setting: { customBaseUrl: '', defaultStrategy: 'recommended', clearCacheOnStartup: false },
});
const status = reactive<I_performerAvatarLibraryStatus>(emptyStatus());
const setting = reactive<I_performerAvatarLibrarySetting>({
  customBaseUrl: '',
  defaultStrategy: 'recommended',
  clearCacheOnStartup: false,
});
const appConfig = ref<I_appSystemConfig>();
const loading = ref(false);
const updating = ref(false);
const clearing = ref(false);

const applyStatus = (value: I_performerAvatarLibraryStatus) => {
  Object.assign(status, emptyStatus(), value);
};

const loadStatus = async () => {
  loading.value = true;
  try {
    const [statusResult, configResult] = await Promise.all([
      performerAvatarLibraryServer.status(),
      appDataServer.getAppConfig(),
    ]);
    if (!statusResult.status) {
      ElMessage.error(statusResult.msg);
      return;
    }
    if (!configResult.status) {
      ElMessage.error(configResult.msg);
      return;
    }
    applyStatus(statusResult.data);
    appConfig.value = configResult.data;
    Object.assign(setting, configResult.data.performerAvatarLibrary);
  } finally {
    loading.value = false;
  }
};

const persistSetting = async (showSuccess: boolean) => {
  if (!appConfig.value) return false;
  const nextConfig: I_appSystemConfig = {
    ...appConfig.value,
    performerAvatarLibrary: {
      ...appConfig.value.performerAvatarLibrary,
      ...setting,
    },
  };
  const result = await appDataServer.setAppConfig(nextConfig);
  if (result.status) {
    appConfig.value = nextConfig;
    if (showSuccess) ElMessage.success('头像库设置已保存');
    return true;
  }
  ElMessage.error(result.msg);
  return false;
};

const saveSetting = async () => {
  await persistSetting(true);
};

const updateDataFile = async () => {
  updating.value = true;
  try {
    if (!await persistSetting(false)) return;
    const result = await performerAvatarLibraryServer.updateDataFile();
    if (result.status) {
      applyStatus(result.data);
      ElMessage.success('头像库数据文件更新完成');
    } else ElMessage.error(result.msg);
  } finally {
    updating.value = false;
  }
};

const clearImageCache = async () => {
  try {
    await ElMessageBox.confirm(
      `确定清理当前缓存的 ${status.cachedImages} 张头像吗？这不会删除头像库数据文件，也不会影响演员已经使用的头像。`,
      '确认清理头像缓存',
      { type: 'warning', confirmButtonText: '清理缓存' },
    );
  } catch {
    return;
  }
  clearing.value = true;
  try {
    const result = await performerAvatarLibraryServer.clearImageCache();
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    applyStatus(result.data.status);
    ElMessage.success(`已清理 ${result.data.clearedImages} 张缓存头像，释放 ${formatBytes(result.data.freedSize)}`);
  } finally {
    clearing.value = false;
  }
};

const formatBytes = (value: number) => {
  if (!value) return '0 B';
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${(value / 1024 / 1024).toFixed(1)} MB`;
};
const formatDate = (value: string) => value ? new Date(value).toLocaleString() : '-';

onMounted(loadStatus);
</script>

<style scoped lang="scss">
.avatar-library-setting { display: flex; flex-direction: column; gap: 12px; height: 100%; overflow: auto; }
.card-header { display: flex; align-items: center; justify-content: space-between; gap: 20px; }
.card-actions { display: flex; gap: 10px; }
h3 { margin: 0; }
p { margin: 6px 0 0; color: var(--el-text-color-secondary); }
.field-help { margin-top: 6px; color: var(--el-text-color-secondary); font-size: 12px; }
.field-help--inline { margin: 0 0 0 10px; }
.cache-summary { margin-top: 12px; }
</style>
