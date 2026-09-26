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
      <template #header>视频采集排除规则（全部文件库）</template>
      <el-form label-width="230px" label-position="left">
        <el-form-item label="自动排除常见非视频文件">
          <el-switch v-model="settingData.setting.autoExcludeNonVideo" />
        </el-form-item>
        <el-form-item label="系统默认排除后缀">
          <details><summary>查看图片、网页、文本和字幕格式</summary><p class="help">{{ defaultExcludedExtensions.join('、') }}</p></details>
        </el-form-item>
        <el-form-item label="自定义排除后缀">
          <el-input v-model="settingData.setting.excludedExtensions" type="textarea" :rows="2" placeholder="例如 pdf, zip；支持空格、逗号、分号分隔，不区分大小写" />
        </el-form-item>
      </el-form>
      <el-alert type="info" :closable="false" title="仅影响视频信息采集，不删除文件或资源。匹配文件不采集、不重试，也不计入采集失败。保存后生效。" />
      <p class="help">删除排除后缀后，文件将在下一次采集任务中重新参与。人工标记的非视频不会自动恢复；可在“已排除文件”中标记为视频并重试。</p>
    </el-card>

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

    <el-card class="collection-result-card" shadow="never">
      <el-tabs v-model="collectionResultTab">
        <el-tab-pane label="文件库采集概况" name="stats">
          <el-table :data="stats" stripe>
            <el-table-column prop="name" label="文件库" min-width="160" />
            <el-table-column prop="total" label="视频分集" width="100" align="center" />
            <el-table-column prop="completed" label="已完成" width="100" align="center" />
            <el-table-column prop="pending" label="待补齐" width="100" align="center" />
            <el-table-column prop="stale" label="需更新" width="100" align="center" />
            <el-table-column prop="failed" label="失败" width="90" align="center" />
            <el-table-column prop="processing" label="处理中" width="90" align="center" />
            <el-table-column prop="manual" label="人工" width="90" align="center" />
            <el-table-column prop="excluded" label="已排除" width="90" align="center" />
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="采集失败文件" name="failures">
          <div class="failure-tools">
            <el-radio-group v-model="failureQuery.excluded" @change="searchFailures">
              <el-radio-button :value="false">采集失败</el-radio-button>
              <el-radio-button :value="true">已排除文件</el-radio-button>
            </el-radio-group>
            <el-button :loading="cleanup.status === 'running'" @click="previewCleanup">按已保存规则整理历史记录</el-button>
          </div>
          <el-alert v-if="cleanup.status !== 'idle'" :type="cleanup.status === 'failed' ? 'error' : 'info'" :closable="false"
            :title="cleanup.status === 'running' ? `正在整理：已处理 ${cleanup.processed} 条 / 预览 ${cleanup.total} 条` : cleanup.status === 'failed' ? cleanup.error : `整理完成，共处理 ${cleanup.processed} 条`" />
          <div class="failure-tools">
            <span>已选 {{ selectedFailures.length }} 项（表头勾选全选当前页）</span>
            <el-button v-if="!failureQuery.excluded" :disabled="!selectedFailures.length || bulkBusy" @click="bulkAction('retry')">批量重试</el-button>
            <el-button v-if="!failureQuery.excluded" :disabled="!selectedFailures.length || bulkBusy" @click="bulkAction('nonvideo')">批量标记非视频</el-button>
            <el-button v-else :disabled="!selectedFailures.length || bulkBusy" @click="bulkAction('video')">标记为视频并重试</el-button>
          </div>
          <div class="failure-tools">
            <el-input v-model="failureQuery.keyword" clearable placeholder="搜索资源、路径或错误"
              @keyup.enter="searchFailures" @clear="searchFailures" />
            <el-button type="primary" plain @click="searchFailures">查询</el-button>
          </div>
          <el-table v-if="failures.length" :key="String(failureQuery.excluded)" :data="failures" stripe max-height="430" @selection-change="selectedFailures = $event">
            <el-table-column type="selection" width="48" />
            <el-table-column v-if="failureQuery.excluded" label="排除原因" width="130">
              <template #default="scope">{{ exclusionReasonText(scope.row.exclusionReason) }}</template>
            </el-table-column>
            <el-table-column prop="filesBasesName" label="文件库" width="120" />
            <el-table-column prop="resourceTitle" label="资源" min-width="150" show-overflow-tooltip />
            <el-table-column prop="src" label="文件路径" min-width="260" show-overflow-tooltip />
            <el-table-column label="文件大小" width="110" align="right">
              <template #default="scope">{{ formatFileSize(scope.row.fileSize) }}</template>
            </el-table-column>
            <el-table-column v-if="!failureQuery.excluded" prop="errorMessage" label="失败原因" min-width="220" show-overflow-tooltip />
            <el-table-column prop="retryCount" label="重试" width="70" align="center" />
            <el-table-column label="操作" width="250" fixed="right">
              <template #default="scope">
                <el-button v-if="!failureQuery.excluded" link type="primary" @click="retryFailure(scope.row)">重试</el-button>
                <el-button v-if="!failureQuery.excluded" link type="success" @click="openManualDialog(scope.row)">手工补录</el-button>
                <el-button v-if="!failureQuery.excluded" link type="warning" @click="markNonVideo(scope.row)">标记非视频</el-button>
                <el-button v-if="failureQuery.excluded" link type="primary" @click="bulkAction('video', [scope.row])">标记为视频并重试</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination v-if="failureTotal > 0" class="failure-pagination" background
            layout="total, sizes, prev, pager, next" v-model:current-page="failureQuery.page"
            v-model:page-size="failureQuery.limit" :page-sizes="[10, 20, 50, 100]" :total="failureTotal"
            @current-change="loadFailures" @size-change="changeFailurePageSize" />
          <el-empty v-if="!failures.length" :description="failureQuery.excluded ? '当前没有已排除文件' : '当前没有采集失败的视频'" />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-dialog v-model="manualDialogVisible" title="手工补录视频信息" width="520px" append-to-body>
      <div class="manual-path" :title="manualItem?.src">{{ manualItem?.src }}</div>
      <el-form label-width="110px">
        <el-form-item label="时长（秒）"><el-input-number v-model="manualForm.durationSeconds" :min="0" /></el-form-item>
        <el-form-item label="宽度"><el-input-number v-model="manualForm.width" :min="0" /></el-form-item>
        <el-form-item label="高度"><el-input-number v-model="manualForm.height" :min="0" /></el-form-item>
        <el-form-item label="视频编码"><el-input v-model="manualForm.videoCodec" placeholder="例如 h264" /></el-form-item>
        <el-form-item label="音频编码"><el-input v-model="manualForm.audioCodec" placeholder="例如 aac" /></el-form-item>
      </el-form>
      <el-alert title="可以只填写已知项目；文件大小会自动读取。人工数据不会被普通自动补齐覆盖。"
        type="info" :closable="false" />
      <template #footer>
        <el-button @click="manualDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="manualSaving" @click="saveManualMetadata">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import selectFilesBases from '@/components/com/form/selectFilesBases.vue';
import { videoMetadataServer } from '@/server/videoMetadata.server';
import type {
  I_videoMetadataBatchTask,
  I_videoMetadataFailureItem,
  I_videoMetadataFailureQuery,
  I_videoMetadataManualRequest,
  I_videoMetadataRunRequest,
  I_videoMetadataSettingData,
  I_videoMetadataStats,
} from '@/dataType/videoMetadata.dataType';

const defaultExcludedExtensions = ['jpg','jpeg','png','gif','bmp','webp','svg','avif','heic','html','htm','nfo','txt','json','xml','url','srt','ass','ssa','vtt','sub'];
const selectedFailures = ref<I_videoMetadataFailureItem[]>([]);
const bulkBusy = ref(false);
const cleanup = reactive({ status: 'idle', processed: 0, total: 0, error: '' });
const savedRules = ref('');
const ruleKey = () => JSON.stringify([settingData.setting.autoExcludeNonVideo, settingData.setting.excludedExtensions]);
const exclusionReasonText = (reason: string) => ({ system: '系统规则', custom: '自定义后缀', manual: '人工标记非视频', legacy: '历史排除（来源未知）' }[reason] || reason);
const bulkAction = async (action: 'retry' | 'nonvideo' | 'video', rows = selectedFailures.value) => {
  if (bulkBusy.value || !rows.length) return;
  try {
    if (action !== 'retry') await ElMessageBox.confirm(action === 'video' ? `将 ${rows.length} 项指定为视频并重试，人工指定优先于后缀排除规则。` : `将 ${rows.length} 项标记为非视频并清除其视频采集信息，不删除文件或资源。`, '批量操作', { type: 'warning' });
    bulkBusy.value = true;
    const result = await videoMetadataServer.bulk(rows.map(x => x.dramaSeriesId), action);
    if (!result.status) { ElMessage.error(result.msg); return; }
    const errors = Object.values(result.data.failed);
    if (errors.length) ElMessage.warning(`成功 ${result.data.succeeded} 项，失败 ${errors.length} 项：${errors[0]}`);
    else ElMessage.success(`已处理 ${result.data.succeeded} 项`);
    selectedFailures.value = [];
    failureQuery.page = 1;
    await refreshAll();
  } catch (error) { if (error instanceof Error) ElMessage.error(error.message); }
  finally { bulkBusy.value = false; }
};
const previewCleanup = async () => {
  if (ruleKey() !== savedRules.value) { ElMessage.warning('排除规则尚未保存，请先保存设置再整理。'); return; }
  try {
    const preview = await videoMetadataServer.cleanupPreview();
    if (!preview.status) { ElMessage.error(preview.msg); return; }
    if (!preview.data.total) { ElMessage.info('没有需要整理的历史采集记录。'); return; }
    const counts = Object.entries(preview.data.byExtension).map(([ext,n]) => `${ext || '无后缀'}：${n} 条`).join('；');
    await ElMessageBox.confirm(`将整理全部文件库中 ${preview.data.total} 条匹配记录（${counts}）。只清理视频采集信息，不删除文件、资源或人工补录数据。`, '预览历史记录整理', { type: 'warning' });
    const result = await videoMetadataServer.startCleanup(preview.data.token);
    if (!result.status) { ElMessage.error(result.msg); return; }
    Object.assign(cleanup, { status: 'running', processed: 0, total: preview.data.total, error: '' });
    await refreshAll();
  } catch (error) { if (error instanceof Error) ElMessage.error(error.message); }
};

const defaultSettingData = (): I_videoMetadataSettingData => ({
  setting: {
    id: 'default',
    autoExcludeNonVideo: true,
    excludedExtensions: '',
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
const collectionResultTab = ref<'stats' | 'failures'>('stats');
const failures = ref<I_videoMetadataFailureItem[]>([]);
const failureTotal = ref(0);
const failureQuery = reactive<I_videoMetadataFailureQuery>({ page: 1, limit: 20, keyword: '', excluded: false });
const manualDialogVisible = ref(false);
const manualSaving = ref(false);
const manualItem = ref<I_videoMetadataFailureItem>();
const manualForm = reactive<{
  durationSeconds?: number;
  width?: number;
  height?: number;
  videoCodec: string;
  audioCodec: string;
}>({ durationSeconds: undefined, width: undefined, height: undefined, videoCodec: '', audioCodec: '' });
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
  Object.assign(settingData.setting, data.setting, { autoExcludeNonVideo: data.setting.autoExcludeNonVideo !== false, excludedExtensions: data.setting.excludedExtensions || '' });
  savedRules.value = ruleKey();
  settingData.filesBasesIds = [...(data.filesBasesIds || [])];
};

const assignTask = (data?: I_videoMetadataBatchTask) => {
  Object.assign(task, data || emptyTask());
};

let failureRequest = 0;
const loadFailures = async () => {
  const request = ++failureRequest;
  selectedFailures.value = [];
  const result = await videoMetadataServer.failures({ ...failureQuery });
  if (result.status && request === failureRequest) {
    failures.value = result.data?.dataList || [];
    failureTotal.value = result.data?.total || 0;
  }
};

const refreshAll = async () => {
  const request = ++failureRequest;
  const [statsResult, taskResult, cleanupResult, failureResult] = await Promise.all([
    videoMetadataServer.stats(),
    videoMetadataServer.taskStatus(),
    videoMetadataServer.cleanupStatus(),
    videoMetadataServer.failures({ ...failureQuery }),
  ]);
  if (cleanupResult.status) Object.assign(cleanup, cleanupResult.data);
  if (statsResult.status) stats.value = statsResult.data || [];
  if (failureResult.status && request === failureRequest && !selectedFailures.value.length) {
    failures.value = failureResult.data?.dataList || [];
    failureTotal.value = failureResult.data?.total || 0;
  }
  if (taskResult.status) {
    if (taskResult.data && ['running', 'paused'].includes(taskResult.data.status)) {
      taskPanelVisible.value = true;
    }
    assignTask(taskResult.data);
  }
};

const searchFailures = () => {
  failureQuery.page = 1;
  loadFailures();
};

const changeFailurePageSize = () => {
  failureQuery.page = 1;
  loadFailures();
};

const retryFailure = async (item: I_videoMetadataFailureItem) => {
  const result = await videoMetadataServer.retryFailure(item.dramaSeriesId);
  if (result.status) {
    selectedFailures.value = [];
    ElMessage.success('已加入重新采集队列');
    window.setTimeout(() => refreshAll(), 1200);
  } else {
    ElMessage.error(result.msg);
  }
};

const markNonVideo = async (item: I_videoMetadataFailureItem) => {
  await ElMessageBox.confirm('标记后将清除该文件的错误视频元数据，但不会删除文件或资源，是否继续？', '标记为非视频', {
    type: 'warning',
  });
  const result = await videoMetadataServer.setClassification(item.dramaSeriesId, false);
  if (result.status) {
    selectedFailures.value = [];
    ElMessage.success('已标记为非视频');
    await refreshAll();
  } else {
    ElMessage.error(result.msg);
  }
};

const openManualDialog = (item: I_videoMetadataFailureItem) => {
  manualItem.value = item;
  Object.assign(manualForm, {
    durationSeconds: undefined,
    width: undefined,
    height: undefined,
    videoCodec: '',
    audioCodec: '',
  });
  manualDialogVisible.value = true;
};

const saveManualMetadata = async () => {
  if (!manualItem.value) return;
  if ((manualForm.durationSeconds ?? 0) <= 0 && (manualForm.width ?? 0) <= 0 && (manualForm.height ?? 0) <= 0 &&
    !manualForm.videoCodec.trim() && !manualForm.audioCodec.trim()) {
    ElMessage.warning('请至少填写一项视频信息');
    return;
  }
  manualSaving.value = true;
  try {
    const request: I_videoMetadataManualRequest = {
      dramaSeriesId: manualItem.value.dramaSeriesId,
    };
    if ((manualForm.durationSeconds ?? 0) > 0) request.durationSeconds = manualForm.durationSeconds;
    if ((manualForm.width ?? 0) > 0) request.width = manualForm.width;
    if ((manualForm.height ?? 0) > 0) request.height = manualForm.height;
    if (manualForm.videoCodec.trim()) request.videoCodec = manualForm.videoCodec.trim();
    if (manualForm.audioCodec.trim()) request.audioCodec = manualForm.audioCodec.trim();
    const result = await videoMetadataServer.saveManual(request);
    if (result.status) {
      manualDialogVisible.value = false;
      selectedFailures.value = [];
      ElMessage.success('人工视频信息已保存');
      await refreshAll();
    } else {
      ElMessage.error(result.msg);
    }
  } finally {
    manualSaving.value = false;
  }
};

const formatFileSize = (size: number) => {
  if (!size || size < 0) return '-';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = size;
  let index = 0;
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index++;
  }
  return `${value.toFixed(index >= 3 ? 2 : 1)} ${units[index]}`;
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
    const risky = (settingData.setting.excludedExtensions || '').toLowerCase().split(/[\s,，;；]+/).map(x => x.replace(/^\./, '')).filter(x => ['mp4','mkv','avi','mov','wmv','flv','webm','ts','m2ts','m4v','mpg','mpeg','vob','rmvb','3gp','m3u8'].includes(x));
    if (risky.length) await ElMessageBox.confirm(`你正在排除视频格式 ${risky.join('、')}。损坏或未下载完整的同类视频也会被隐藏，确认保存？`, '排除视频格式', { type: 'warning' });
    const result = await videoMetadataServer.saveSetting({
      setting: { ...settingData.setting },
      filesBasesIds: [...settingData.filesBasesIds],
    });
    if (result.status) {
      assignSetting(result.data);
      ElMessage.success('视频信息采集设置已保存');
      selectedFailures.value = [];
      failureQuery.page = 1;
      await refreshAll();
    } else {
      ElMessage.error(result.msg);
    }
  } catch (error) {
    if (error instanceof Error) ElMessage.error(error.message);
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
    if (taskRunning.value || taskPaused.value || cleanup.status === 'running') refreshAll();
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

  .failure-tools {
    width: 100%;
    flex-wrap: wrap;
    align-items: center;
    margin: 0 0 12px;
    display: flex;
    gap: 8px;
  }

  .failure-pagination {
    margin-top: 14px;
    justify-content: flex-end;
  }

  .manual-path {
    margin-bottom: 16px;
    padding: 10px 12px;
    overflow: hidden;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-light);
    border-radius: 4px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
