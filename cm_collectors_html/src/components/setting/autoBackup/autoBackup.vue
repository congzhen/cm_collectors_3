<template>
  <div class="auto-backup" v-loading="loading">
    <el-form v-model="formData" label-width="170px">
      <el-form-item label="自动备份">
        <div class="field-content">
          <el-switch v-model="formData.autoBackup.enabled" />
          <el-text class="help-text" type="info" size="small">
            开启后按下方时间或资源变化数量静默备份；两个触发条件都为 0 时不会自动执行。
          </el-text>
        </div>
      </el-form-item>
      <el-form-item label="备份路径">
        <div class="field-content">
          <el-input v-model="formData.autoBackup.backupPath">
            <template #append>
              <el-button icon="FolderOpened" @click="openServerFileManagement" />
            </template>
          </el-input>
          <el-text class="help-text" type="info" size="small">
            默认保存到程序根目录的 auto_backup 文件夹。
          </el-text>
        </div>
      </el-form-item>
      <el-form-item label="每隔多少小时备份">
        <div class="field-content">
          <el-input-number v-model="formData.autoBackup.intervalHours" :min="0" :max="8760" />
          <el-text class="help-text" type="info" size="small">
            填 0 表示关闭时间触发；长时间未打开程序时，下次启动会按最近成功备份时间补一次检查。
          </el-text>
        </div>
      </el-form-item>
      <el-form-item label="资源变化数备份">
        <div class="field-content">
          <el-input-number v-model="formData.autoBackup.resourceChangeThreshold" :min="0" :max="100000" />
          <el-text class="help-text" type="info" size="small">
            填 0 表示关闭数量触发；新增、编辑、批量设置资源后会累计变化数，达到后低优先级备份。
          </el-text>
        </div>
      </el-form-item>
      <el-form-item label="最多保留份数">
        <div class="field-content">
          <el-input-number v-model="formData.autoBackup.maxBackups" :min="1" :max="1000" />
          <el-text class="help-text" type="info" size="small">
            超出后自动删除最旧的自动备份，只影响自动备份目录中的 zip 文件。
          </el-text>
        </div>
      </el-form-item>
      <el-form-item>
        <div class="field-content">
          <div class="button-row">
            <el-button type="primary" icon="Edit" @click="saveHandle">保存</el-button>
            <el-button icon="Refresh" @click="refreshHandle">刷新</el-button>
            <el-button type="success" icon="Download" @click="runBackupHandle">立即备份</el-button>
          </div>
          <el-text class="help-text" type="info" size="small">
            立即备份会马上执行一次，不需要开启自动备份开关。
          </el-text>
        </div>
      </el-form-item>
    </el-form>

    <el-alert
      v-if="formData.autoBackup.enabled && formData.autoBackup.intervalHours <= 0 && formData.autoBackup.resourceChangeThreshold <= 0"
      title="自动备份已开启，但尚未设置触发条件" type="warning" :closable="false" />

    <div class="state-section">
      <h3>自动备份状态</h3>
      <div class="state-grid">
        <div class="state-item">
          <span class="label">最近成功</span>
          <span>{{ stateData?.state.last_success_backup_at || '-' }}</span>
        </div>
        <div class="state-item">
          <span class="label">待备份变化</span>
          <span>{{ stateData?.state.pending_resource_change_count ?? 0 }}</span>
        </div>
        <div class="state-item">
          <span class="label">最近资源变化</span>
          <span>{{ stateData?.state.last_resource_change_at || '-' }}</span>
        </div>
        <div class="state-item">
          <span class="label">运行状态</span>
          <span>{{ stateData?.state.running ? '备份中' : '空闲' }}</span>
        </div>
        <div class="state-item full">
          <span class="label">最近备份文件</span>
          <span>{{ stateData?.state.last_backup_path || '-' }}</span>
        </div>
        <div class="state-item full error">
          <span class="label">最近错误</span>
          <span>{{ stateData?.state.last_error || '-' }}</span>
        </div>
      </div>
    </div>

    <div class="backup-section">
      <h3>自动备份列表</h3>
      <div class="backup-list">
        <div v-if="backupList.length === 0" class="empty-backup">暂无自动备份</div>
        <div v-for="(backup, index) in backupList" :key="backup" class="backup-item">
          <span class="backup-name">{{ backup }}</span>
          <el-button type="danger" size="small" plain @click="deleteBackup(backup, index)">删除</el-button>
        </div>
      </div>
    </div>

    <serverFileManagementDialog ref="serverFileManagementDialogRef" @selectedFiles="selectedFilesHandle" />
  </div>
</template>

<script lang="ts" setup>
import { debounceNow } from '@/assets/debounce';
import { getFileNameFromPath } from '@/assets/path';
import type { I_autoBackupStateData, I_appSystemConfig } from '@/dataType/app.dataType';
import { appDataServer } from '@/server/app.server';
import serverFileManagementDialog from '@/components/serverFileManagement/serverFileManagementDialog.vue';
import type { I_sfm_FileEntry } from '@/components/serverFileManagement/com/dataType';
import dataset from '@/assets/dataset';
import { ElMessage, ElMessageBox } from 'element-plus';
import { onMounted, ref } from 'vue';

const serverFileManagementDialogRef = ref<InstanceType<typeof serverFileManagementDialog>>();
const loading = ref(false);
const backupList = ref<string[]>([]);
const stateData = ref<I_autoBackupStateData>();

const formData = ref<I_appSystemConfig>({
  logoName: 'CM File Collectors',
  isAdminLogin: false,
  adminPassword: '',
  language: 'zhCn',
  notAllowServerOpenFile: false,
  allowAppCloseServer: false,
  theme: 'default',
  homeMode: 'classic',
  closeMobileDisplay: false,
  closePlayCloud: true,
  closePlayCloudDialog: true,
  playCloudMode: 'm3u8',
  playVideoFormats: dataset.playVideoFormats,
  playAudioFormats: dataset.playAudioFormats,
  serverFileManagementRootPath: dataset.serverFileManagementRootPath,
  windowsStartNotRunApp: false,
  tvBoxEnabled: true,
  videoRateLimit: {
    enabled: false,
    requestsPerSecond: 5,
    burst: 10,
  },
  scraper: {
    useBrowserPath: false,
    browserPath: '',
  },
  autoBackup: {
    enabled: false,
    backupPath: './auto_backup',
    intervalHours: 0,
    resourceChangeThreshold: 0,
    maxBackups: 5,
  },
  taryMenu: [],
});

const getAppConfig = async () => {
  const result = await appDataServer.getAppConfig();
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }
  formData.value = {
    ...result.data,
    autoBackup: {
      enabled: result.data.autoBackup?.enabled === true,
      backupPath: result.data.autoBackup?.backupPath || './auto_backup',
      intervalHours: result.data.autoBackup?.intervalHours || 0,
      resourceChangeThreshold: result.data.autoBackup?.resourceChangeThreshold || 0,
      maxBackups: result.data.autoBackup?.maxBackups || 5,
    },
  };
};

const getAutoBackupState = async () => {
  const result = await appDataServer.autoBackupState();
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }
  stateData.value = result.data;
};

const getAutoBackupList = async () => {
  const result = await appDataServer.autoBackupList();
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }
  backupList.value = result.data;
};

const refreshHandle = async () => {
  try {
    loading.value = true;
    await getAppConfig();
    await getAutoBackupState();
    await getAutoBackupList();
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }
};

const saveHandle = debounceNow(async () => {
  try {
    loading.value = true;
    const result = await appDataServer.setAppConfig(formData.value);
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    ElMessage.success('保存成功');
    await getAutoBackupState();
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }
});

const runBackupHandle = debounceNow(async () => {
  try {
    loading.value = true;
    const result = await appDataServer.runAutoBackup();
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    ElMessage.success('备份完成');
    await getAutoBackupState();
    await getAutoBackupList();
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }
});

const deleteBackup = (backupPath: string, index: number) => {
  const backupName = getFileNameFromPath(backupPath, true);
  ElMessageBox.confirm(`确定要删除备份 "${backupName}" 吗？此操作不可恢复。`, '确认删除', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      loading.value = true;
      const result = await appDataServer.deleteAutoBackup(backupName);
      if (!result.status) {
        ElMessage.error(result.msg);
        return;
      }
      backupList.value.splice(index, 1);
      ElMessage.success('删除成功');
    } catch (error) {
      console.log(error);
      ElMessage.error('删除失败');
    } finally {
      loading.value = false;
    }
  });
};

const openServerFileManagement = () => {
  serverFileManagementDialogRef.value?.open();
};

const selectedFilesHandle = (slc: I_sfm_FileEntry[]) => {
  if (slc.length === 0) {
    return;
  }
  formData.value.autoBackup.backupPath = slc[slc.length - 1].path;
};

onMounted(() => {
  refreshHandle();
});
</script>

<style lang="scss" scoped>
.auto-backup {
  max-width: 960px;
  height: 100%;
  box-sizing: border-box;
  overflow-y: auto;
  padding: 20px;

  .el-form {
    max-width: 760px;
  }

  .field-content {
    width: 100%;
  }

  .help-text {
    display: block;
    margin-top: 6px;
    line-height: 1.4;
  }

  .button-row {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .state-section,
  .backup-section {
    margin-top: 24px;

    h3 {
      margin: 0 0 14px;
      font-size: 18px;
      font-weight: 600;
    }
  }

  .state-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .state-item {
    display: flex;
    gap: 10px;
    min-height: 36px;
    align-items: center;
    padding: 8px 10px;
    color: var(--el-text-color-regular);
    background: var(--el-fill-color-light);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    word-break: break-all;

    &.full {
      grid-column: 1 / -1;
    }

    &.error span:last-child {
      color: #f56c6c;
    }

    .label {
      width: 110px;
      flex-shrink: 0;
      color: var(--el-text-color-secondary);
    }
  }

  .backup-list {
    max-height: 300px;
    overflow-y: auto;
  }

  .backup-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    margin-bottom: 8px;
    color: var(--el-text-color-regular);
    background: var(--el-fill-color-light);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;

    .backup-name {
      flex: 1;
      word-break: break-all;
    }
  }

  .empty-backup {
    padding: 24px 0;
    text-align: center;
    color: var(--el-text-color-secondary);
  }
}
</style>
