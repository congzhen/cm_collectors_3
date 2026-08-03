<template>
  <div class="video-metadata-setting" v-loading="loading">
    <div class="page-header">
      <div>
        <h2>视频信息采集</h2>
        <p>统一保存时长、分辨率、帧率、视频/音频编码等信息。历史列表补齐和空闲补齐默认关闭。</p>
      </div>
      <el-button type="primary" :loading="saving" @click="saveSetting">保存设置</el-button>
    </div>

    <el-card shadow="never">
      <template #header>自动触发策略</template>
      <el-form label-width="230px" label-position="left">
        <el-form-item label="新增或文件变化时采集">
          <el-switch v-model="settingData.setting.collectOnNewOrChanged" />
          <span class="help">建议开启，保证以后新增的视频信息完整。</span>
        </el-form-item>
        <el-form-item label="打开详情或播放时采集">
          <el-switch v-model="settingData.setting.collectOnDetailOrPlay" />
          <span class="help">只优先补齐用户正在使用的视频。</span>
        </el-form-item>
        <el-form-item label="浏览列表页时补齐">
          <el-switch v-model="settingData.setting.collectOnList" />
          <span class="help warning">NAS 用户建议关闭；开启后当前页视频会进入后台队列。</span>
        </el-form-item>
        <el-form-item label="空闲时补齐历史数据">
          <el-switch v-model="settingData.setting.idleBackfillEnabled" />
          <span class="help warning">默认关闭，开启后系统会低速处理选择的文件库。</span>
        </el-form-item>

        <template v-if="settingData.setting.idleBackfillEnabled">
          <el-form-item label="空闲补齐范围">
            <el-radio-group v-model="settingData.setting.idleScopeMode">
              <el-radio-button value="selected">指定文件库</el-radio-button>
              <el-radio-button value="all">全部文件库</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item v-if="settingData.setting.idleScopeMode === 'selected'" label="参与补齐的文件库">
            <selectFilesBases v-model="settingData.filesBasesIds" multiple />
          </el-form-item>
          <el-form-item label="启动等待时间">
            <el-input-number v-model="settingData.setting.idleWaitMinutes" :min="1" :max="1440" />
            <span class="unit">分钟</span>
          </el-form-item>
          <el-form-item label="文件处理间隔">
            <el-input-number v-model="settingData.setting.probeIntervalMilliseconds" :min="0" :max="60000" :step="100" />
            <span class="unit">毫秒</span>
          </el-form-item>
          <el-form-item label="每批查询数量">
            <el-input-number v-model="settingData.setting.idleBatchSize" :min="1" :max="1000" />
          </el-form-item>
        </template>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>历史数据补齐</span>
          <el-button link type="primary" @click="refreshAll">刷新状态</el-button>
        </div>
      </template>

      <el-form class="manual-form" label-width="110px" label-position="left">
        <el-form-item label="执行范围">
          <el-radio-group v-model="manual.scopeMode">
            <el-radio-button value="selected">指定文件库</el-radio-button>
            <el-radio-button value="all">全部文件库</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="manual.scopeMode === 'selected'" label="文件库">
          <selectFilesBases v-model="manual.filesBasesIds" multiple />
        </el-form-item>
        <el-form-item label="处理方式">
          <el-select v-model="manual.runMode">
            <el-option label="补齐缺失并更新失效项" value="missing_stale" />
            <el-option label="只补齐完全缺失的信息" value="missing" />
            <el-option label="重试已到期的失败项" value="failed" />
            <el-option label="强制重试失败项" value="failed_force" />
            <el-option label="强制重新采集全部" value="all" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="success" :disabled="taskRunning || taskPaused" @click="startTask">开始补齐</el-button>
          <el-button v-if="taskRunning" type="warning" @click="pauseTask">暂停</el-button>
          <el-button v-if="taskPaused" type="success" plain @click="resumeTask">继续</el-button>
          <el-button v-if="taskRunning || taskPaused" type="danger" plain @click="stopTask">停止</el-button>
        </el-form-item>
      </el-form>

      <el-alert v-if="task.id && taskPanelVisible" class="task-alert" :title="taskStatusText"
        :type="task.lastError ? 'error' : 'info'" :closable="false">
        <template #default>
          <el-progress v-if="task.total > 0" :percentage="taskProgress"
            :color="task.failed > 0 ? 'var(--el-color-warning)' : undefined" />
          <div class="task-summary">
            <div>成功 {{ task.success }}，失败 {{ task.failed }}，跳过 {{ task.skipped }}</div>
            <div v-if="task.currentSrc" class="task-current" :title="task.currentSrc">当前：{{ task.currentSrc }}</div>
          </div>
          <div v-if="task.lastError" class="task-error">{{ task.lastError }}</div>
        </template>
      </el-alert>
    </el-card>

    <el-card shadow="never">
      <template #header>文件库采集概况</template>
      <el-table :data="stats" stripe>
        <el-table-column prop="name" label="文件库" min-width="160" />
        <el-table-column prop="total" label="视频分集" width="100" align="center" />
        <el-table-column prop="completed" label="已完成" width="100" align="center" />
        <el-table-column prop="pending" label="待补齐" width="100" align="center" />
        <el-table-column prop="stale" label="需更新" width="100" align="center" />
        <el-table-column prop="failed" label="失败" width="90" align="center" />
        <el-table-column prop="processing" label="处理中" width="90" align="center" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import selectFilesBases from '@/components/com/form/selectFilesBases.vue';
import { videoMetadataServer } from '@/server/videoMetadata.server';
import type {
  I_videoMetadataBatchTask,
  I_videoMetadataRunRequest,
  I_videoMetadataSettingData,
  I_videoMetadataStats,
} from '@/dataType/videoMetadata.dataType';

const defaultSettingData = (): I_videoMetadataSettingData => ({
  setting: {
    id: 'default',
    collectOnNewOrChanged: true,
    collectOnDetailOrPlay: true,
    collectOnList: false,
    idleBackfillEnabled: false,
    idleScopeMode: 'selected',
    idleWaitMinutes: 5,
    probeIntervalMilliseconds: 1500,
    idleBatchSize: 20,
    paused: false,
  },
  filesBasesIds: [],
});

const emptyTask = (): I_videoMetadataBatchTask => ({
  id: '',
  scopeMode: 'selected',
  runMode: 'missing_stale',
  status: '',
  total: 0,
  success: 0,
  failed: 0,
  skipped: 0,
  currentSrc: '',
  lastError: '',
  createdAt: '',
  startedAt: '',
  finishedAt: '',
});

const loading = ref(false);
const saving = ref(false);
const taskPanelVisible = ref(false);
const settingData = reactive<I_videoMetadataSettingData>(defaultSettingData());
const stats = ref<I_videoMetadataStats[]>([]);
const task = reactive<I_videoMetadataBatchTask>(emptyTask());
const manual = reactive<I_videoMetadataRunRequest>({
  scopeMode: 'selected',
  filesBasesIds: [],
  runMode: 'missing_stale',
});
let refreshTimer: ReturnType<typeof setInterval> | undefined;

const taskRunning = computed(() => task.status === 'running');
const taskPaused = computed(() => task.status === 'paused');
const taskStatusText = computed(() => {
  const labels: Record<string, string> = {
    running: '视频信息补齐正在运行',
    paused: '视频信息补齐已暂停',
    stopped: '视频信息补齐已停止',
    completed: '视频信息补齐已完成',
  };
  return labels[task.status] || '暂无补齐任务';
});
const taskProgress = computed(() => {
  if (task.total <= 0) return 0;
  return Math.min(100, Math.round(((task.success + task.failed + task.skipped) / task.total) * 100));
});

const assignSetting = (data: I_videoMetadataSettingData) => {
  Object.assign(settingData.setting, data.setting);
  settingData.filesBasesIds = [...(data.filesBasesIds || [])];
};

const assignTask = (data?: I_videoMetadataBatchTask) => {
  Object.assign(task, data || emptyTask());
};

const refreshAll = async () => {
  const [statsResult, taskResult] = await Promise.all([
    videoMetadataServer.stats(),
    videoMetadataServer.taskStatus(),
  ]);
  if (statsResult.status) stats.value = statsResult.data || [];
  if (taskResult.status) {
    if (taskResult.data && ['running', 'paused'].includes(taskResult.data.status)) {
      taskPanelVisible.value = true;
    }
    assignTask(taskResult.data);
  }
};

const load = async () => {
  loading.value = true;
  try {
    const result = await videoMetadataServer.setting();
    if (result.status) assignSetting(result.data);
    else ElMessage.error(result.msg);
    await refreshAll();
  } finally {
    loading.value = false;
  }
};

const saveSetting = async () => {
  if (settingData.setting.idleBackfillEnabled &&
    settingData.setting.idleScopeMode === 'selected' &&
    settingData.filesBasesIds.length === 0) {
    ElMessage.warning('请选择空闲补齐的文件库');
    return;
  }
  saving.value = true;
  try {
    const result = await videoMetadataServer.saveSetting({
      setting: { ...settingData.setting },
      filesBasesIds: [...settingData.filesBasesIds],
    });
    if (result.status) {
      assignSetting(result.data);
      ElMessage.success('视频信息采集设置已保存');
    } else {
      ElMessage.error(result.msg);
    }
  } finally {
    saving.value = false;
  }
};

const startTask = async () => {
  if (manual.scopeMode === 'selected' && manual.filesBasesIds.length === 0) {
    ElMessage.warning('请选择需要补齐的文件库');
    return;
  }
  if (manual.runMode === 'all') {
    await ElMessageBox.confirm('强制重新采集会重新读取所选范围内的全部视频，是否继续？', '确认重新采集', {
      type: 'warning',
    });
  } else if (manual.runMode === 'failed_force') {
    await ElMessageBox.confirm('将忽略下次重试时间，立即重试所选范围内的全部失败项，是否继续？', '确认强制重试', {
      type: 'warning',
    });
  }
  const result = await videoMetadataServer.run({
    scopeMode: manual.scopeMode,
    filesBasesIds: [...manual.filesBasesIds],
    runMode: manual.runMode,
  });
  if (result.status) {
    taskPanelVisible.value = true;
    assignTask(result.data);
    ElMessage.success('补齐任务已开始');
  } else {
    ElMessage.error(result.msg);
  }
};

const pauseTask = async () => {
  const result = await videoMetadataServer.pause();
  if (result.status) await refreshAll();
};
const resumeTask = async () => {
  const result = await videoMetadataServer.resume();
  if (result.status) await refreshAll();
};
const stopTask = async () => {
  await ElMessageBox.confirm('停止后已保存的视频信息不会删除，确定停止吗？', '停止补齐', { type: 'warning' });
  const result = await videoMetadataServer.stop();
  if (result.status) await refreshAll();
};

onMounted(async () => {
  await load();
  refreshTimer = setInterval(() => {
    if (taskRunning.value || taskPaused.value) refreshAll();
  }, 2000);
});
onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer);
});
</script>

<style scoped lang="scss">
.video-metadata-setting {
  height: 100%;
  min-height: 0;
  box-sizing: border-box;
  overflow: auto;
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 16px;

  > .el-card {
    flex: 0 0 auto;
  }

  .page-header,
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 16px;
  }

  h2 {
    margin: 0 0 6px;
  }

  p {
    margin: 0;
    color: var(--el-text-color-secondary);
  }

  .help {
    margin-left: 12px;
    color: var(--el-text-color-secondary);
  }

  .warning {
    color: var(--el-color-warning);
  }

  .unit {
    margin-left: 8px;
  }

  .manual-form {
    width: min(100%, 920px);
  }

  .manual-form :deep(.el-form-item:last-child) {
    margin-bottom: 0;
  }

  .task-alert {
    width: min(100%, 920px);
    box-sizing: border-box;

    :deep(.el-alert__content) {
      flex: 1;
      min-width: 0;
    }

    :deep(.el-progress) {
      width: 100%;
    }
  }

  .task-summary {
    min-width: 0;
  }

  .task-current {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .task-error {
    margin-top: 8px;
    color: var(--el-color-danger);
  }
}
</style>
