<template>
  <div class="duplicate-detection" v-loading="loading">
    <div class="summary">
      <div class="summary-item">
        <span class="summary-label">当前指纹</span>
        <strong>{{ stats.done }}</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">待计算</span>
        <strong>{{ stats.pending }}</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">失败</span>
        <strong>{{ stats.failed }}</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">覆盖率</span>
        <strong>{{ coverageText }}</strong>
      </div>
    </div>

    <el-alert
      class="task-tip"
      type="warning"
      show-icon
      :closable="false"
      title="需要建立“视频指纹”计划任务，系统才会持续为新增或更新后的分集生成指纹。"
    >
      <template #default>
        <el-button type="warning" link @click="emit('switchTab', 'cronJobs')">前往计划任务</el-button>
      </template>
    </el-alert>

    <el-alert
      v-if="taskStatus.running || taskStatus.last_error"
      class="task-tip"
      :type="taskStatus.last_error ? 'error' : 'info'"
      show-icon
      :closable="false"
      :title="taskStatus.running ? `视频指纹任务执行中，文件库：${taskStatus.files_bases_id || '全部'}` : `上次任务错误：${taskStatus.last_error}`"
    />

    <el-form class="filters" :model="query" label-width="100px">
      <el-form-item label="文件库">
        <selectFilesBases v-model="query.files_bases_id" width="260px" @change="handleFilesBasesChange" />
      </el-form-item>
      <el-form-item label="匹配模式">
        <el-radio-group v-model="query.match_mode">
          <el-radio-button label="全采样" value="full" />
          <el-radio-button label="高度" value="high" />
          <el-radio-button label="宽松" value="loose" />
          <el-radio-button label="最小" value="minimal" />
        </el-radio-group>
      </el-form-item>
      <el-form-item label="汉明阈值">
        <div class="threshold">
          <el-slider v-model="query.threshold" :min="0" :max="30" :step="1" />
          <span>{{ query.threshold }}</span>
        </div>
      </el-form-item>
      <el-form-item label="时长优先">
        <el-switch v-model="query.duration_first" />
        <el-text class="delete-tip" type="info" size="small">
          开启后先按时长容差缩小候选范围，速度更快；关闭后只按指纹匹配，适合查找剪辑或时长变化的视频。
        </el-text>
      </el-form-item>
      <el-form-item label="时长容差">
        <div class="threshold">
          <el-slider v-model="query.duration_tolerance" :disabled="!query.duration_first" :min="0" :max="30" :step="1" />
          <span>{{ query.duration_tolerance }}秒</span>
        </div>
      </el-form-item>
      <el-form-item label="删除方式">
        <el-checkbox v-model="deleteLocalFile">同时删除本地视频文件</el-checkbox>
        <el-text class="delete-tip" type="warning" size="small">
          不勾选时只删除数据库资源/分集记录，本地视频文件会保留。
        </el-text>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" icon="Search" @click="loadDuplicates">筛选</el-button>
        <el-button icon="Refresh" @click="refreshAll">刷新统计</el-button>
        <el-button type="success" icon="VideoPlay" :disabled="taskStatus.running" @click="triggerCompute">
          立即计算待处理
        </el-button>
        <el-button type="warning" icon="Plus" @click="reScan">补录待计算</el-button>
        <el-button type="warning" icon="RefreshLeft" @click="resetFailed">重置失败</el-button>
      </el-form-item>
    </el-form>

    <div class="groups">
      <el-empty v-if="groups.length === 0" description="暂无重复结果" />
      <div v-for="(group, groupIndex) in groups" :key="groupIndex" class="group">
        <div class="group-header">
          <div>
            <strong>重复组 #{{ (query.page - 1) * query.limit + groupIndex + 1 }}</strong>
            <span>共 {{ group.items.length }} 条，匹配 {{ group.matched_count }} 个采样点，平均距离 {{ group.avg_distance.toFixed(2) }}</span>
          </div>
          <el-button type="danger" size="small" plain @click="deleteSelectedInGroup(group)">删除本组已选分集</el-button>
        </div>
        <el-table :data="group.items" row-key="drama_series_id" border>
          <el-table-column width="52">
            <template #default="{ row }">
              <el-checkbox :model-value="selectedKeys.has(rowKey(row))" @change="toggleRow(row)" />
            </template>
          </el-table-column>
          <el-table-column label="资源标题" prop="resource_title" min-width="180" show-overflow-tooltip />
          <el-table-column label="分集路径" prop="src" min-width="360" show-overflow-tooltip />
          <el-table-column label="时长" width="110">
            <template #default="{ row }">{{ formatDuration(row.duration) }}</template>
          </el-table-column>
          <el-table-column label="资源ID" prop="resources_id" width="180" show-overflow-tooltip />
        </el-table>
      </div>
    </div>

    <el-pagination
      v-if="total > 0"
      class="pagination"
      background
      layout="prev, pager, next, sizes, total"
      :total="total"
      v-model:current-page="query.page"
      v-model:page-size="query.limit"
      :page-sizes="[10, 20, 50, 100]"
      @change="loadDuplicates"
    />
  </div>
</template>

<script setup lang="ts">
import selectFilesBases from '@/components/com/form/selectFilesBases.vue';
import type {
  I_DuplicateGroup,
  I_DuplicateItem,
  I_VideoFingerprintStats,
  I_VideoFingerprintTaskStatus,
} from '@/dataType/videoFingerprint.dataType';
import { videoFingerprintServer } from '@/server/videoFingerprint.server';
import { ElMessage, ElMessageBox } from 'element-plus';
import { computed, onMounted, reactive, ref } from 'vue';

const emit = defineEmits<{
  (e: 'switchTab', tabName: string): void;
}>();

const loading = ref(false);
const deleteLocalFile = ref(false);
const groups = ref<I_DuplicateGroup[]>([]);
const total = ref(0);
const selectedKeys = ref(new Set<string>());
const stats = reactive<I_VideoFingerprintStats>({
  total: 0,
  pending: 0,
  done: 0,
  failed: 0,
  drama_total: 0,
});
const taskStatus = reactive<I_VideoFingerprintTaskStatus>({
  running: false,
  files_bases_id: '',
  started_at: '',
  last_finished_at: '',
  last_error: '',
  last_success: 0,
  last_failed: 0,
});

const query = reactive({
  files_bases_id: '',
  match_mode: 'high',
  threshold: 8,
  duration_first: true,
  duration_tolerance: 3,
  page: 1,
  limit: 20,
});

const coverageText = computed(() => {
  if (!stats.drama_total) return '0%';
  return `${((stats.done / stats.drama_total) * 100).toFixed(1)}%`;
});

const rowKey = (row: I_DuplicateItem) => `${row.resources_id}_${row.drama_series_id}`;

const assignStats = (data: I_VideoFingerprintStats) => {
  stats.total = data.total || 0;
  stats.pending = data.pending || 0;
  stats.done = data.done || 0;
  stats.failed = data.failed || 0;
  stats.drama_total = data.drama_total || 0;
};

const assignTaskStatus = (data: I_VideoFingerprintTaskStatus) => {
  taskStatus.running = data.running || false;
  taskStatus.files_bases_id = data.files_bases_id || '';
  taskStatus.started_at = data.started_at || '';
  taskStatus.last_finished_at = data.last_finished_at || '';
  taskStatus.last_error = data.last_error || '';
  taskStatus.last_success = data.last_success || 0;
  taskStatus.last_failed = data.last_failed || 0;
};

const loadStats = async () => {
  const result = await videoFingerprintServer.stats(query.files_bases_id);
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }
  assignStats(result.data);
};

const loadTaskStatus = async () => {
  const result = await videoFingerprintServer.taskStatus();
  if (result.status) {
    assignTaskStatus(result.data);
  }
};

const loadDuplicates = async () => {
  try {
    loading.value = true;
    const result = await videoFingerprintServer.queryDuplicates(query);
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    groups.value = result.data.dataList;
    total.value = result.data.total;
    selectedKeys.value.clear();
  } finally {
    loading.value = false;
  }
};

const refreshAll = async () => {
  await loadStats();
  await loadTaskStatus();
  await loadDuplicates();
};

const handleFilesBasesChange = async () => {
  query.page = 1;
  await refreshAll();
};

const triggerCompute = async () => {
  const result = await videoFingerprintServer.triggerCompute(50, query.files_bases_id);
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }
  ElMessage.success('已开始后台计算待处理指纹');
  await loadTaskStatus();
};

const reScan = async () => {
  const result = await videoFingerprintServer.reScan(query.files_bases_id);
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }
  ElMessage.success(`已补录 ${result.data.added} 条待计算指纹`);
  await loadStats();
};

const resetFailed = async () => {
  const result = await videoFingerprintServer.resetFailed(query.files_bases_id);
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }
  ElMessage.success(`已重置 ${result.data.reset} 条失败记录`);
  await loadStats();
};

const toggleRow = (row: I_DuplicateItem) => {
  const next = new Set(selectedKeys.value);
  const key = rowKey(row);
  if (next.has(key)) {
    next.delete(key);
  } else {
    next.add(key);
  }
  selectedKeys.value = next;
};

const deleteSelectedInGroup = async (group: I_DuplicateGroup) => {
  const selectedRows = group.items.filter((item) => selectedKeys.value.has(rowKey(item)));
  if (selectedRows.length === 0) {
    ElMessage.warning('请选择要删除的分集');
    return;
  }
  if (selectedRows.length >= group.items.length) {
    ElMessage.warning('同一重复组内至少保留一个分集');
    return;
  }

  await ElMessageBox.confirm(
    deleteLocalFile.value
      ? `将删除 ${selectedRows.length} 个分集记录，并尝试删除对应本地视频文件。若某个资源只剩一个分集，会同时删除该资源。是否继续？`
      : `将删除 ${selectedRows.length} 个分集记录，本地视频文件会保留。若某个资源只剩一个分集，会同时删除该资源。是否继续？`,
    '确认删除',
    {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  );

  try {
    loading.value = true;
    const result = await videoFingerprintServer.deleteDramaSeries(
      selectedRows.map((item) => item.drama_series_id),
      deleteLocalFile.value
    );
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    ElMessage.success('删除成功');
    await refreshAll();
  } finally {
    loading.value = false;
  }
};

const formatDuration = (duration: number) => {
  if (!duration) return '-';
  const totalSeconds = Math.round(duration);
  const h = Math.floor(totalSeconds / 3600);
  const m = Math.floor((totalSeconds % 3600) / 60);
  const s = totalSeconds % 60;
  return h > 0 ? `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}` : `${m}:${String(s).padStart(2, '0')}`;
};

onMounted(refreshAll);
</script>

<style lang="scss" scoped>
.duplicate-detection {
  height: 100%;
  overflow: auto;
  padding: 10px;

  .summary {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 10px;
    margin-bottom: 10px;
  }

  .summary-item {
    border: 1px solid var(--el-border-color);
    border-radius: 6px;
    padding: 12px;
    background: var(--el-bg-color);

    .summary-label {
      display: block;
      color: var(--el-text-color-secondary);
      font-size: 12px;
      margin-bottom: 6px;
    }

    strong {
      font-size: 22px;
      font-weight: 600;
    }
  }

  .task-tip,
  .filters {
    margin-bottom: 12px;
  }

  .threshold {
    display: grid;
    grid-template-columns: minmax(160px, 320px) 56px;
    gap: 12px;
    align-items: center;
    width: 400px;
  }

  .delete-tip {
    margin-left: 12px;
  }

  .group {
    border: 1px solid var(--el-border-color);
    border-radius: 6px;
    margin-bottom: 12px;
    overflow: hidden;
  }

  .group-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    background: var(--el-fill-color-lighter);

    span {
      margin-left: 12px;
      color: var(--el-text-color-secondary);
      font-size: 13px;
    }
  }

  .pagination {
    justify-content: flex-end;
    margin-top: 12px;
  }
}
</style>
