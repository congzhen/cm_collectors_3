<template>
  <div class="transcode-setting" v-loading="loading">
    <div class="toolbar">
      <el-button type="primary" :disabled="!canApplyConfig" @click="applyConfig">
        应用参数
      </el-button>
      <el-button
        type="danger"
        plain
        :disabled="!selectedDeletableTasks.length"
        @click="deleteSelectedTasks"
      >
        移除选中<span v-if="selectedDeletableTasks.length">（{{ selectedDeletableTasks.length }}）</span>
      </el-button>
      <el-button
        :icon="RefreshLeft"
        :disabled="!selectedResettableTasks.length"
        @click="resetSelectedTasks"
      >
        重置选中<span v-if="selectedResettableTasks.length">（{{ selectedResettableTasks.length }}）</span>
      </el-button>
      <el-button type="success" :disabled="!canStartTasks" @click="startTasks">
        开始执行
      </el-button>
      <el-button v-if="!queueStatus.paused" type="warning" @click="pauseQueue">暂停队列</el-button>
      <el-button v-else type="success" @click="resumeQueue">继续队列</el-button>
      <el-button :icon="Refresh" @click="refresh()">刷新</el-button>
      <span class="queue-summary">
        共 {{ tasks.length }} 项
        <template v-if="queueStatus.currentId"> · 正在执行 1 项</template>
        <template v-if="queueStatus.paused"> · 队列已暂停</template>
      </span>
    </div>

    <el-table
      ref="tableRef"
      :data="tasks"
      row-key="id"
      class="task-table"
      @selection-change="selection = $event"
    >
      <el-table-column type="selection" width="42" :reserve-selection="true" />
      <el-table-column label="资源/源视频" min-width="280">
        <template #default="{ row }">
          <div class="resource-cell">
            <el-image
              v-if="coverURL(row)"
              class="resource-cover"
              :src="coverURL(row)"
              fit="cover"
              lazy
              preview-teleported
              :preview-src-list="[coverURL(row)]"
            >
              <template #error><div class="cover-placeholder"><el-icon><VideoCamera /></el-icon></div></template>
            </el-image>
            <div v-else class="resource-cover cover-placeholder"><el-icon><VideoCamera /></el-icon></div>
            <div class="resource-main">
              <div class="resource-title">{{ row.resourceTitle }}</div>
              <div class="file-path" :title="row.sourcePath">{{ fileName(row.sourcePath) }}</div>
              <div class="source-details">
                <span>{{ formatSize(row.sourceSize) }}</span>
                <span
                  v-for="item in sourceInfoItems(row)"
                  :key="item"
                  class="source-chip"
                >{{ item }}</span>
              </div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="150">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          <div v-if="row.speed" class="speed">{{ row.speed }}</div>
        </template>
      </el-table-column>
      <el-table-column label="进度" min-width="390">
        <template #default="{ row }">
          <el-progress :percentage="Math.round(row.progress || 0)" :status="progressStatus(row.status)" />
          <div class="applied-config">
            <el-tag
              v-for="item in appliedConfigItems(row.config)"
              :key="item"
              size="small"
              type="info"
              effect="plain"
            >
              {{ item }}
            </el-tag>
          </div>
          <el-tooltip v-if="row.errorMessage || row.warningMessage" :content="row.errorMessage || row.warningMessage">
            <div :class="row.errorMessage ? 'task-error' : 'task-warning'">
              {{ row.errorMessage || row.warningMessage }}
            </div>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column label="输出" min-width="330">
        <template #default="{ row }">
          <div class="output-file" :title="row.outputPath">
            {{ row.outputPath ? fileName(row.outputPath) : `等待生成 .${row.config.container}` }}
          </div>
          <div v-if="row.outputSize" class="source-details">
            <span>{{ formatSize(row.outputSize) }}</span>
            <span
              v-for="item in outputInfoItems(row)"
              :key="item"
              class="source-chip"
            >{{ item }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="135" fixed="right">
        <template #default="{ row }">
          <el-button
            v-if="canCancel(row.status)"
            type="warning"
            link
            @click="cancelTask(row)"
          >取消</el-button>
          <el-button
            v-if="canDelete(row.status)"
            type="danger"
            link
            @click="deleteTask(row)"
          >移除</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="转码列表为空，可在资源或分集上点击右键加入" />
      </template>
    </el-table>

    <el-card class="config-card" shadow="never">
      <template #header>
        <div class="config-title">
          <span>转码参数</span>
          <span>应用于选中的待转码任务；未选择时应用于全部待转码任务</span>
        </div>
      </template>
      <el-form :model="config" label-width="82px" class="config-form">
        <section class="config-section main-section">
          <div class="section-title">常用参数</div>
          <div class="section-grid main-grid">
            <el-form-item label="输出容器">
              <el-select v-model="config.container">
                <el-option label="MP4" value="mp4" />
                <el-option label="MKV" value="mkv" />
              </el-select>
            </el-form-item>
            <el-form-item label="视频编码">
              <el-select v-model="config.videoCodec">
                <el-option label="复制视频流" value="copy" />
                <el-option label="H.264 (CPU)" value="h264" />
                <el-option label="H.265 (CPU)" value="h265" />
              </el-select>
            </el-form-item>
            <el-form-item label="GPU加速">
              <div class="gpu-control">
                <el-switch v-model="useGPU" :disabled="config.videoCodec === 'copy' || !compatibleGPUEncoders.length" />
                <span v-if="!compatibleGPUEncoders.length" class="unit">
                  {{ config.videoCodec === 'copy' ? '复制模式无需编码' : '未检测到可用 GPU' }}
                </span>
                <el-select v-else-if="useGPU" v-model="config.gpuEncoder">
                  <el-option
                    v-for="item in compatibleGPUEncoders"
                    :key="item.id"
                    :label="item.label"
                    :value="item.id"
                  />
                </el-select>
              </div>
            </el-form-item>
            <el-form-item label="质量模式">
              <el-select v-model="config.qualityMode" :disabled="config.videoCodec === 'copy'">
                <el-option label="恒定质量 CRF" value="crf" />
                <el-option label="目标码率" value="bitrate" />
              </el-select>
            </el-form-item>
            <el-form-item label="质量参数">
              <el-input-number
                v-if="config.qualityMode === 'crf'"
                v-model="config.crf"
                :min="0"
                :max="51"
                :disabled="config.videoCodec === 'copy'"
              />
              <el-select v-else v-model="config.videoBitrateKbps" :disabled="config.videoCodec === 'copy'">
                <el-option
                  v-for="item in videoBitrateOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="音频编码">
              <el-select v-model="config.audioCodec">
                <el-option label="复制音频流" value="copy" />
                <el-option label="AAC" value="aac" />
              </el-select>
            </el-form-item>
          </div>
        </section>

        <button type="button" class="advanced-toggle" @click="advancedExpanded = !advancedExpanded">
          <span class="advanced-toggle-title">
            高级参数
            <el-icon>
              <ArrowUp v-if="advancedExpanded" />
              <ArrowDown v-else />
            </el-icon>
          </span>
        </button>

        <el-collapse-transition>
          <div v-show="advancedExpanded" class="advanced-panel">
            <div class="section-grid advanced-grid">
                <el-form-item label="编码速度">
                  <el-select v-model="config.preset" :disabled="config.videoCodec === 'copy'">
                    <el-option label="最快" value="ultrafast" />
                    <el-option label="快速" value="fast" />
                    <el-option label="均衡" value="medium" />
                    <el-option label="高质量" value="slow" />
                  </el-select>
                </el-form-item>
                <el-form-item label="视频尺寸">
                  <el-select v-model="config.resolutionHeight" :disabled="config.videoCodec === 'copy'">
                    <el-option label="保持原尺寸" :value="0" />
                    <el-option label="2160P（不放大）" :value="2160" />
                    <el-option label="1440P（不放大）" :value="1440" />
                    <el-option label="1080P（不放大）" :value="1080" />
                    <el-option label="720P（不放大）" :value="720" />
                    <el-option label="480P（不放大）" :value="480" />
                  </el-select>
                </el-form-item>
                <el-form-item label="视频帧率">
                  <el-select v-model="config.frameRate" :disabled="config.videoCodec === 'copy'">
                    <el-option
                      v-for="item in frameRateOptions"
                      :key="item.value"
                      :label="item.label"
                      :value="item.value"
                    />
                  </el-select>
                </el-form-item>
                <el-form-item label="音频码率">
                  <el-input-number
                    v-model="config.audioBitrateKbps"
                    :min="32"
                    :max="1024"
                    :step="32"
                    :disabled="config.audioCodec === 'copy'"
                  />
                  <span class="unit">Kbps</span>
                </el-form-item>
                <el-form-item label="CPU线程">
                  <el-input-number v-model="config.threads" :min="0" :max="128" />
                  <span class="unit">0 为自动</span>
                </el-form-item>
                <el-form-item label="保留旧文件">
                  <el-switch v-model="config.keepBackup" />
                </el-form-item>
            </div>
          </div>
        </el-collapse-transition>
      </el-form>
      <el-alert
        title="输出校验成功后才会替换源文件。替换时会先创建临时回滚文件；未开启“保留旧文件”时会在全部成功后删除。"
        type="warning"
        :closable="false"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { ElMessage, ElMessageBox, ElNotification, type TableInstance } from 'element-plus';
import { ArrowDown, ArrowUp, Refresh, RefreshLeft } from '@element-plus/icons-vue';
import type {
  I_videoTranscodeConfig,
  I_videoTranscodeGPUEncoder,
  I_videoTranscodeQueueStatus,
  I_videoTranscodeTask,
  VideoTranscodeStatus,
} from '@/dataType/videoTranscode.dataType';
import { videoTranscodeServer } from '@/server/videoTranscode.server';

const defaultConfig = (): I_videoTranscodeConfig => ({
  container: 'mp4',
  videoCodec: 'copy',
  qualityMode: 'crf',
  crf: 23,
  videoBitrateKbps: 4000,
  preset: 'medium',
  resolutionHeight: 0,
  frameRate: 0,
  audioCodec: 'copy',
  audioBitrateKbps: 192,
  threads: 0,
  keepBackup: false,
  gpuEncoder: '',
});
const configStorageKey = 'cm-video-transcode-config-v2';
const loadStoredConfig = (): I_videoTranscodeConfig => {
  const defaults = defaultConfig();
  try {
    const stored = window.localStorage.getItem(configStorageKey);
    if (!stored) return defaults;
    const parsed = JSON.parse(stored);
    return parsed && typeof parsed === 'object' ? { ...defaults, ...parsed } : defaults;
  } catch {
    return defaults;
  }
};

const videoBitrateOptions = [
  { label: '256 Kbps', value: 256 },
  { label: '512 Kbps', value: 512 },
  { label: '768 Kbps', value: 768 },
  { label: '1 Mbps', value: 1024 },
  { label: '1.5 Mbps', value: 1536 },
  { label: '2 Mbps', value: 2048 },
  { label: '2.5 Mbps', value: 2560 },
  { label: '3 Mbps', value: 3072 },
  { label: '4 Mbps', value: 4096 },
  { label: '5 Mbps', value: 5120 },
  { label: '6 Mbps', value: 6144 },
  { label: '7 Mbps', value: 7168 },
  { label: '8 Mbps', value: 8192 },
  { label: '9 Mbps', value: 9216 },
  { label: '10 Mbps', value: 10240 },
  { label: '12 Mbps', value: 12288 },
  { label: '20 Mbps', value: 20480 },
];
const frameRateOptions = [
  { label: '保持原帧率', value: 0 },
  { label: '23.976 fps（影视）', value: 23.976 },
  { label: '24 fps（电影）', value: 24 },
  { label: '25 fps（PAL）', value: 25 },
  { label: '29.97 fps（NTSC）', value: 29.97 },
  { label: '30 fps', value: 30 },
  { label: '50 fps', value: 50 },
  { label: '59.94 fps（NTSC）', value: 59.94 },
  { label: '60 fps', value: 60 },
  { label: '120 fps', value: 120 },
];

const loading = ref(false);
const tasks = ref<I_videoTranscodeTask[]>([]);
const selection = ref<I_videoTranscodeTask[]>([]);
const tableRef = ref<TableInstance>();
const config = reactive<I_videoTranscodeConfig>(loadStoredConfig());
const advancedExpanded = ref(false);
const queueStatus = reactive<I_videoTranscodeQueueStatus>({ paused: false, currentId: '' });
const gpuEncoders = ref<I_videoTranscodeGPUEncoder[]>([]);
let timer: number | undefined;
let mounted = false;
let refreshing = false;

const editableTasks = computed(() => tasks.value.filter(item => item.status === 'draft'));
const startableTasks = computed(() => tasks.value.filter(item =>
  ['draft', 'failed', 'cancelled', 'interrupted'].includes(item.status)));
const selectedEditableTasks = computed(() => selection.value.filter(item => item.status === 'draft'));
const selectedStartableTasks = computed(() => selection.value.filter(item =>
  ['draft', 'failed', 'cancelled', 'interrupted'].includes(item.status)));
const canApplyConfig = computed(() =>
  selection.value.length ? selectedEditableTasks.value.length > 0 : editableTasks.value.length > 0);
const canStartTasks = computed(() =>
  selection.value.length ? selectedStartableTasks.value.length > 0 : startableTasks.value.length > 0);
const selectedDeletableTasks = computed(() => selection.value.filter(item =>
  ['draft', 'success', 'failed', 'cancelled', 'interrupted', 'rollback_failed'].includes(item.status)));
const selectedResettableTasks = computed(() => selection.value.filter(item =>
  ['success', 'failed', 'cancelled', 'interrupted'].includes(item.status)));
const compatibleGPUEncoders = computed(() => gpuEncoders.value.filter(item =>
  config.videoCodec !== 'copy' && item.videoCodecs.includes(config.videoCodec)));
const useGPU = computed({
  get: () => config.gpuEncoder !== '',
  set: (enabled: boolean) => {
    config.gpuEncoder = enabled ? (compatibleGPUEncoders.value[0]?.id || '') : '';
  },
});

const activeStatuses: VideoTranscodeStatus[] = [
  'queued', 'probing', 'transcoding', 'verifying', 'replacing', 'refreshing_metadata',
];
const scheduleRefresh = () => {
  if (!mounted) return;
  window.clearTimeout(timer);
  const hasActiveTask = tasks.value.some(item => activeStatuses.includes(item.status));
  timer = window.setTimeout(() => refresh(false), hasActiveTask ? 2000 : 10000);
};
const refresh = async (showLoading = true) => {
  if (refreshing) return;
  refreshing = true;
  if (showLoading) loading.value = true;
  try {
    const [listResult, statusResult] = await Promise.all([
      videoTranscodeServer.list(),
      videoTranscodeServer.status(),
    ]);
    if (listResult.status) tasks.value = listResult.data;
    if (statusResult.status) Object.assign(queueStatus, statusResult.data);
  } finally {
    refreshing = false;
    if (showLoading) loading.value = false;
    scheduleRefresh();
  }
};

const loadCapabilities = async () => {
  const result = await videoTranscodeServer.capabilities();
  if (result.status) {
    gpuEncoders.value = result.data.gpuEncoders || [];
    if (config.gpuEncoder && !compatibleGPUEncoders.value.some(item => item.id === config.gpuEncoder)) {
      config.gpuEncoder = '';
    }
  }
};

const applyConfig = async () => {
  const ids = selection.value.length ? selectedEditableTasks.value.map(item => item.id) : [];
  if (selection.value.length && !ids.length) {
    return ElMessage.warning('选中的任务中没有待转码任务');
  }
  const result = await videoTranscodeServer.updateConfig(ids, { ...config });
  if (!result.status) return ElMessage.error(result.msg || '保存参数失败');
  ElNotification.success({
    title: ids.length ? `已应用到 ${ids.length} 个待转码任务` : '已应用到全部待转码任务',
    message: appliedConfigSummary(config),
    duration: 5000,
  });
  await refresh(false);
};

const startTasks = async () => {
  const ids = selection.value.length ? selectedStartableTasks.value.map(item => item.id) : [];
  if (selection.value.length && !ids.length) {
    return ElMessage.warning('选中的任务当前都不可执行');
  }
  await ElMessageBox.confirm(
    ids.length ? `确定执行选中的 ${ids.length} 个任务吗？` : '确定执行全部可开始任务吗？',
    '开始转码',
    { type: 'warning' },
  );
  const result = await videoTranscodeServer.start(ids);
  if (!result.status) return ElMessage.error(result.msg || '启动失败');
  ElMessage.success('任务已进入执行队列');
  await refresh(false);
};

const pauseQueue = async () => {
  const result = await videoTranscodeServer.pause();
  if (result.status) {
    queueStatus.paused = true;
    ElMessage.success('队列已暂停，当前任务会继续完成');
  }
};

const resumeQueue = async () => {
  const result = await videoTranscodeServer.resume();
  if (result.status) {
    queueStatus.paused = false;
    ElMessage.success('队列已继续');
  }
};

const cancelTask = async (task: I_videoTranscodeTask) => {
  await ElMessageBox.confirm('取消后不会替换源文件，确定继续吗？', '取消任务', { type: 'warning' });
  const result = await videoTranscodeServer.cancel(task.id);
  if (!result.status) return ElMessage.error(result.msg || '取消失败');
  await refresh(false);
};

const deleteTask = async (task: I_videoTranscodeTask) => {
  const result = await videoTranscodeServer.delete(task.id);
  if (!result.status) return ElMessage.error(result.msg || '移除失败');
  await refresh(false);
};

const deleteSelectedTasks = async () => {
  const deletable = selectedDeletableTasks.value;
  if (!deletable.length) return;
  const ignoredCount = selection.value.length - deletable.length;
  const confirmMessage = ignoredCount
    ? `确定移除选中的 ${deletable.length} 个任务吗？另有 ${ignoredCount} 个执行中任务不会被移除。`
    : `确定移除选中的 ${deletable.length} 个任务吗？`;
  await ElMessageBox.confirm(
    confirmMessage,
    '批量移除任务',
    { type: 'warning' },
  );
  const result = await videoTranscodeServer.deleteBatch(deletable.map(item => item.id));
  if (!result.status) return ElMessage.error(result.msg || '批量移除失败');
  tableRef.value?.clearSelection();
  selection.value = [];
  ElMessage.success(`已移除 ${result.data} 个任务`);
  await refresh(false);
};

const resetSelectedTasks = async () => {
  const resettable = selectedResettableTasks.value;
  if (!resettable.length) return;
  const ignoredCount = selection.value.length - resettable.length;
  const confirmMessage = ignoredCount
    ? `确定将选中的 ${resettable.length} 个任务重置为待转码吗？另有 ${ignoredCount} 个当前状态不允许重置。`
    : `确定将选中的 ${resettable.length} 个任务重置为待转码吗？`;
  await ElMessageBox.confirm(confirmMessage, '批量重置任务', { type: 'warning' });
  const result = await videoTranscodeServer.resetBatch(resettable.map(item => item.id));
  if (!result.status) return ElMessage.error(result.msg || '批量重置失败');
  tableRef.value?.clearSelection();
  selection.value = [];
  const skippedTip = result.data.skipped ? `，跳过 ${result.data.skipped} 个` : '';
  ElMessage.success(`已重置 ${result.data.reset} 个任务${skippedTip}`);
  await refresh(false);
};

const fileName = (path: string) => path.replace(/\\/g, '/').split('/').pop() || path;
const formatSize = (size: number) => {
  if (!size) return '-';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = size;
  let index = 0;
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index++;
  }
  return `${value.toFixed(index >= 3 ? 2 : 1)} ${units[index]}`;
};
const formatDuration = (seconds: number) => {
  const total = Math.max(0, Math.round(seconds));
  const hours = Math.floor(total / 3600);
  const minutes = Math.floor((total % 3600) / 60);
  const remaining = total % 60;
  return [hours, minutes, remaining].map(value => String(value).padStart(2, '0')).join(':');
};
const formatCodec = (codec: string) => {
  const normalized = codec.toLowerCase();
  const names: Record<string, string> = {
    h264: 'H.264',
    avc: 'H.264',
    h265: 'H.265',
    hevc: 'H.265',
    aac: 'AAC',
    av1: 'AV1',
    vp9: 'VP9',
    opus: 'Opus',
  };
  return names[normalized] || codec.toUpperCase();
};
const sourceInfoItems = (task: I_videoTranscodeTask) => {
  const items: string[] = [];
  if (task.sourceWidth && task.sourceHeight) items.push(`${task.sourceWidth}×${task.sourceHeight}`);
  if (task.sourceDuration) items.push(formatDuration(task.sourceDuration));
  if (task.sourceFrameRate) items.push(`${Number(task.sourceFrameRate.toFixed(3))} fps`);
  if (task.sourceVideoBitRate) items.push(formatBitrate(Math.round(task.sourceVideoBitRate / 1000)));
  if (task.sourceVideoCodec) items.push(formatCodec(task.sourceVideoCodec));
  if (task.sourceAudioCodec) items.push(formatCodec(task.sourceAudioCodec));
  return items;
};
const outputInfoItems = (task: I_videoTranscodeTask) => {
  const items: string[] = [];
  if (task.outputWidth && task.outputHeight) items.push(`${task.outputWidth}×${task.outputHeight}`);
  if (task.outputDuration) items.push(formatDuration(task.outputDuration));
  if (task.outputFrameRate) items.push(`${Number(task.outputFrameRate.toFixed(3))} fps`);
  if (task.outputVideoBitRate) items.push(formatBitrate(Math.round(task.outputVideoBitRate / 1000)));
  if (task.outputVideoCodec) items.push(formatCodec(task.outputVideoCodec));
  if (task.outputAudioCodec) items.push(formatCodec(task.outputAudioCodec));
  return items;
};
const presetNames: Record<string, string> = {
  ultrafast: '最快',
  fast: '快速',
  medium: '均衡',
  slow: '高质量',
};
const formatBitrate = (kbps: number) =>
  kbps >= 1024 ? `${Number((kbps / 1024).toFixed(2))} Mbps` : `${kbps} Kbps`;
const appliedConfigItems = (item: I_videoTranscodeConfig) => {
  const parts = [`输出：${item.container.toUpperCase()}`];
  if (item.videoCodec === 'copy') {
    parts.push('视频：复制');
  } else {
    const videoCodec = item.videoCodec === 'h265' ? 'H.265' : 'H.264';
    parts.push(`视频：${videoCodec} ${item.gpuEncoder ? item.gpuEncoder.toUpperCase() : 'CPU'}`);
    parts.push(item.qualityMode === 'bitrate'
      ? `视频码率：${formatBitrate(item.videoBitrateKbps)}`
      : `质量：CRF ${item.crf}`);
    parts.push(`速度：${presetNames[item.preset] || item.preset}`);
    if (item.resolutionHeight) parts.push(`尺寸：${item.resolutionHeight}P`);
    if (item.frameRate) parts.push(`帧率：${item.frameRate} fps`);
  }
  parts.push(item.audioCodec === 'copy' ? '音频：复制' : `音频：AAC ${item.audioBitrateKbps} Kbps`);
  if (item.threads) parts.push(`线程：${item.threads}`);
  if (item.keepBackup) parts.push('保留旧文件');
  return parts;
};
const appliedConfigSummary = (item: I_videoTranscodeConfig) => appliedConfigItems(item).join('；');
const coverURL = (task: I_videoTranscodeTask) => {
  if (!task.filesBasesId || !task.coverPoster) return '';
  return `/api/resCoverPoster/${encodeURIComponent(task.filesBasesId)}/${encodeURIComponent(task.coverPoster)}`;
};

const statusNames: Record<VideoTranscodeStatus, string> = {
  draft: '待转码',
  queued: '等待中',
  probing: '分析源文件',
  transcoding: '正在转码',
  verifying: '校验输出',
  replacing: '替换源文件',
  refreshing_metadata: '刷新元数据',
  success: '已完成',
  failed: '失败',
  cancelled: '已取消',
  interrupted: '已中断',
  rollback_failed: '回滚失败',
};
const statusText = (status: VideoTranscodeStatus) => statusNames[status] || status;
const statusType = (status: VideoTranscodeStatus) => {
  if (status === 'success') return 'success';
  if (['failed', 'rollback_failed'].includes(status)) return 'danger';
  if (['transcoding', 'replacing', 'refreshing_metadata'].includes(status)) return 'warning';
  if (status === 'queued' || status === 'probing' || status === 'verifying') return 'primary';
  return 'info';
};
const progressStatus = (status: VideoTranscodeStatus) => {
  if (status === 'success') return 'success';
  if (['failed', 'rollback_failed'].includes(status)) return 'exception';
  return undefined;
};
const canCancel = (status: VideoTranscodeStatus) =>
  ['draft', 'queued', 'probing', 'transcoding', 'verifying'].includes(status);
const canDelete = (status: VideoTranscodeStatus) =>
  ['draft', 'success', 'failed', 'cancelled', 'interrupted', 'rollback_failed'].includes(status);

onMounted(async () => {
  mounted = true;
  await Promise.all([refresh(), loadCapabilities()]);
});
onBeforeUnmount(() => {
  mounted = false;
  window.clearTimeout(timer);
});

watch(
  () => config.videoCodec,
  () => {
    if (config.videoCodec === 'copy') {
      config.gpuEncoder = '';
      return;
    }
    if (config.gpuEncoder && !compatibleGPUEncoders.value.some(item => item.id === config.gpuEncoder)) {
      config.gpuEncoder = compatibleGPUEncoders.value[0]?.id || '';
    }
  },
);
watch(
  config,
  value => {
    try {
      window.localStorage.setItem(configStorageKey, JSON.stringify(value));
    } catch {
      // 本地存储不可用时不影响转码参数的正常使用。
    }
  },
  { deep: true },
);
</script>

<style scoped lang="scss">
.transcode-setting {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.toolbar {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}
.queue-summary {
  margin-left: auto;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.task-table {
  flex: 1 1 46%;
  min-height: 220px;
}
.resource-title {
  font-weight: 600;
}
.resource-cell {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}
.resource-main {
  min-width: 0;
}
.resource-cover {
  width: 62px;
  height: 62px;
  flex: 0 0 62px;
  border-radius: 5px;
  overflow: hidden;
  background: var(--el-fill-color-light);
}
.cover-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: var(--el-text-color-placeholder);
  font-size: 22px;
}
.file-path, .output-file, .task-error, .task-warning {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.file-path, .source-info, .speed {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  margin-top: 3px;
}
.source-details {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 4px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.source-details > span {
  white-space: nowrap;
}
.source-chip {
  padding: 1px 5px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 3px;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-regular);
  line-height: 17px;
}
.applied-config {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 6px;
}
.task-error {
  color: var(--el-color-danger);
  font-size: 12px;
}
.task-warning {
  color: var(--el-color-warning);
  font-size: 12px;
}
.config-card {
  flex: 0 0 auto;
  max-height: 410px;
  overflow: auto;
}
.config-title {
  display: flex;
  justify-content: space-between;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.config-title span:first-child {
  color: var(--el-text-color-primary);
  font-size: 14px;
  font-weight: 600;
}
.config-form {
  display: flex;
  flex-direction: column;
}
.config-section {
  display: grid;
  grid-template-columns: 120px minmax(0, 1fr);
  padding: 12px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.config-section:first-child {
  padding-top: 0;
}
.config-section:last-child {
  border-bottom: 0;
}
.section-title {
  padding: 8px 16px 0 4px;
  color: var(--el-text-color-primary);
  font-size: 13px;
  font-weight: 600;
}
.section-grid {
  display: grid;
  grid-template-columns: repeat(3, 340px);
  justify-content: start;
  column-gap: 24px;
  row-gap: 2px;
}
.config-form :deep(.el-form-item) {
  margin-bottom: 10px;
}
.config-form :deep(.el-form-item__content) {
  min-width: 0;
  flex-wrap: nowrap;
}
.config-form :deep(.el-select) {
  width: 100%;
}
.config-form :deep(.el-input-number) {
  width: auto;
  min-width: 0;
  flex: 1;
}
.unit {
  margin-left: 6px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  white-space: nowrap;
}
.gpu-control {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
}
.gpu-control .el-select {
  flex: 1;
}
.advanced-toggle {
  width: 100%;
  min-height: 34px;
  display: flex;
  align-items: center;
  padding: 0 4px;
  border: 0;
  background: transparent;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  cursor: pointer;
}
.advanced-toggle:hover {
  background: var(--el-fill-color-extra-light);
}
.advanced-toggle-title {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: var(--el-text-color-primary);
  font-size: 13px;
  font-weight: 600;
}
.advanced-toggle-title .el-icon {
  color: var(--el-color-primary);
}
.advanced-panel {
  padding: 8px 0 2px 120px;
  border-top: 1px solid var(--el-border-color-lighter);
}
@media (max-width: 1400px) {
  .section-grid, .main-grid, .advanced-grid {
    grid-template-columns: repeat(2, 340px);
  }
}
@media (max-width: 800px) {
  .config-section { grid-template-columns: 1fr; }
  .section-title { padding-bottom: 8px; }
  .section-grid, .main-grid, .advanced-grid {
    grid-template-columns: minmax(0, 340px);
  }
  .advanced-panel { padding-left: 0; }
}
</style>
