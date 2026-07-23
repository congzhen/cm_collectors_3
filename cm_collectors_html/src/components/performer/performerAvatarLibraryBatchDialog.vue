<template>
  <el-dialog v-model="visible" title="批量匹配演员头像" width="820px" destroy-on-close
    :close-on-click-modal="!applying" :close-on-press-escape="!applying" :show-close="!applying">
    <div v-loading="loading || actorLoading">
      <el-form label-width="120px" :class="{ 'batch-form--disabled': applying }">
        <el-form-item label="演员范围">
          <el-radio-group v-model="scopeMode" @change="scopeChanged">
            <el-radio-button value="all">全部演员</el-radio-button>
            <el-radio-button value="selected">指定演员</el-radio-button>
          </el-radio-group>
          <span class="scope-summary">
            {{ scopeMode === 'all' ? `共 ${allActorTotal} 人` : `已选择 ${selectedActorIds.length} 人` }}
          </span>
        </el-form-item>
        <el-form-item v-if="scopeMode === 'selected'" label="选择演员" class="actor-selector-form-item">
          <div class="actor-selector">
            <div class="selector-filters">
              <el-input v-model="actorSearch" clearable prefix-icon="Search" placeholder="搜索演员姓名或别名" />
              <el-segmented v-model="photoFilter" :options="photoFilterOptions" />
            </div>
            <div class="selector-actions">
              <el-button size="small" @click="selectCurrentPageActors">选择本页</el-button>
              <el-button size="small" @click="selectFilteredActors">选择全部筛选结果</el-button>
              <el-button size="small" @click="selectActorsWithoutPhoto">仅选无头像</el-button>
              <el-button size="small" @click="invertCurrentPageActors">反选本页</el-button>
              <el-button size="small" @click="clearSelectedActors">清空</el-button>
              <span>筛选结果 {{ actorTotal }} 人</span>
            </div>
            <el-scrollbar height="290px" class="actor-scrollbar">
              <el-checkbox-group v-model="selectedActorIds" class="actor-grid" @change="schedulePreview">
                <div v-for="actor in actors" :key="actor.id" class="actor-card"
                  :class="{ selected: selectedActorIds.includes(actor.id) }" @click="toggleActor(actor.id)">
                  <el-checkbox :value="actor.id" @click.stop />
                  <el-image class="actor-card__photo" :src="actorPhoto(actor)" fit="cover">
                    <template #error><el-image :src="getPerformerEmptyPhoto()" fit="cover" /></template>
                  </el-image>
                  <span class="actor-card__content">
                    <span class="actor-card__name" :title="actor.name">{{ actor.name || '未设置主姓名' }}</span>
                    <span v-if="actor.aliasName" class="actor-card__alias" :title="actor.aliasName">别名：{{ actor.aliasName }}</span>
                  </span>
                  <el-tag :type="actor.hasPhoto ? 'info' : 'success'" size="small">
                    {{ actor.hasPhoto ? '已有头像' : '无头像' }}
                  </el-tag>
                </div>
              </el-checkbox-group>
              <el-empty v-if="actors.length === 0" description="没有符合条件的演员" :image-size="60" />
            </el-scrollbar>
            <el-pagination v-if="actorTotal > pageSize" v-model:current-page="currentActorPage"
              class="actor-pagination" background layout="total, prev, pager, next, jumper"
              :page-size="pageSize" :total="actorTotal" small @current-change="loadActors" />
          </div>
        </el-form-item>
        <el-form-item label="头像策略">
          <el-radio-group v-model="strategy" @change="loadPreview">
            <el-radio-button value="recommended">推荐头像</el-radio-button>
            <el-radio-button value="original">原图优先</el-radio-button>
            <el-radio-button value="aiFix">AI 优化图优先</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="处理范围">
          <el-radio-group v-model="overwrite" @change="loadPreview">
            <el-radio-button :value="false">仅无头像演员</el-radio-button>
            <el-radio-button :value="true">包括已有头像</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <el-alert v-if="overwrite" title="已有头像将被替换，批量执行前只会统一确认一次。" type="warning" :closable="false" show-icon />
      <el-empty v-if="scopeMode === 'selected' && selectedActorIds.length === 0" description="请至少选择一位演员" :image-size="70" />
      <template v-else-if="preview">
        <el-alert v-if="preview.matched === 0" class="no-match-alert" type="warning" :closable="false" show-icon
          title="所选演员当前均未在头像库中命中，继续执行也不会更新头像。可以重新选择演员，或检查演员姓名和别名。" />
        <el-descriptions :column="2" border class="preview">
          <el-descriptions-item label="演员总数">{{ preview.total }}</el-descriptions-item>
          <el-descriptions-item label="可匹配">{{ preview.matched }}</el-descriptions-item>
          <el-descriptions-item label="未匹配">{{ preview.unmatched }}</el-descriptions-item>
          <el-descriptions-item label="跳过已有头像">{{ preview.skippedExisting }}</el-descriptions-item>
          <el-descriptions-item label="有多个候选" :span="2">{{ preview.multipleCandidates }}</el-descriptions-item>
        </el-descriptions>
      </template>
      <section v-if="applying && batchProgress" class="batch-progress">
        <div class="batch-progress__header">
          <strong>正在匹配演员头像</strong>
          <span>{{ batchProgress.completed }} / {{ batchProgress.total }}</span>
        </div>
        <el-progress :percentage="progressPercentage" :stroke-width="12" />
        <div class="batch-progress__stats">
          <span>成功 {{ batchProgress.success }}</span>
          <span>失败 {{ batchProgress.failed }}</span>
          <span>未匹配 {{ batchProgress.unmatched }}</span>
          <span>跳过 {{ batchProgress.skippedExisting }}</span>
        </div>
        <div v-if="batchProgress.currentActors.length" class="batch-progress__current">
          正在处理：{{ batchProgress.currentActors.join('、') }}
        </div>
        <div v-else class="batch-progress__current">正在准备匹配任务…</div>
      </section>
    </div>
    <template #footer>
      <el-button :disabled="applying" @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="applying"
        :disabled="!preview || (scopeMode === 'selected' && selectedActorIds.length === 0)"
        @click="submit">开始批量匹配</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { I_performerAvatarBatchActor, I_performerAvatarBatchPreview, I_performerAvatarBatchProgress, PerformerAvatarStrategy } from '@/dataType/performerAvatarLibrary.dataType';
import { performerAvatarLibraryServer as server } from '@/server/performerAvatarLibrary.server';
import { ensurePerformerAvatarLibraryReady } from '@/common/performerAvatarLibrary';
import { getPerformerEmptyPhoto } from '@/common/photo';

const emits = defineEmits(['success']);
const props = withDefaults(defineProps<{ pageSize?: number }>(), { pageSize: 90 });
const visible = ref(false);
const loading = ref(false);
const actorLoading = ref(false);
const applying = ref(false);
const performerBasesId = ref('');
const strategy = ref<PerformerAvatarStrategy>('recommended');
const overwrite = ref(false);
const preview = ref<I_performerAvatarBatchPreview>();
const batchProgress = ref<I_performerAvatarBatchProgress>();
const actors = ref<I_performerAvatarBatchActor[]>([]);
const actorTotal = ref(0);
const allActorTotal = ref(0);
const scopeMode = ref<'all' | 'selected'>('all');
const selectedActorIds = ref<string[]>([]);
const actorSearch = ref('');
const photoFilter = ref<'all' | 'missing' | 'existing'>('all');
const photoFilterOptions = [
  { label: '全部', value: 'all' },
  { label: '无头像', value: 'missing' },
  { label: '已有头像', value: 'existing' },
];
const pageSize = computed(() => props.pageSize);
const currentActorPage = ref(1);
let previewTimer: ReturnType<typeof setTimeout> | undefined;
let actorListTimer: ReturnType<typeof setTimeout> | undefined;
let progressTimer: ReturnType<typeof setTimeout> | undefined;
let previewRequestId = 0;
let activeBatchId = '';
const progressPollInterval = 3000;

const progressPercentage = computed(() => {
  if (!batchProgress.value?.total) return 0;
  return Math.min(100, Math.round(batchProgress.value.completed / batchProgress.value.total * 100));
});

const currentPerformerIds = () => scopeMode.value === 'all' ? [] : selectedActorIds.value;
const actorPhoto = (actor: I_performerAvatarBatchActor) => actor.photo
  ? `/api/performerFace/${actor.performerBasesId}/${actor.photo}`
  : getPerformerEmptyPhoto();

const open = async (id: string) => {
  performerBasesId.value = id;
  visible.value = true;
  overwrite.value = false;
  loading.value = true;
  const libraryStatus = await ensurePerformerAvatarLibraryReady();
  loading.value = false;
  if (!libraryStatus) {
    visible.value = false;
    return;
  }
  strategy.value = libraryStatus.setting.defaultStrategy;
  scopeMode.value = 'all';
  selectedActorIds.value = [];
  actorSearch.value = '';
  photoFilter.value = 'all';
  currentActorPage.value = 1;
  if (!await loadActors()) {
    visible.value = false;
    return;
  }
  allActorTotal.value = actorTotal.value;
  await loadPreview();
};

const loadActors = async () => {
  if (!performerBasesId.value) return false;
  actorLoading.value = true;
  try {
    const result = await server.batchActors(
      performerBasesId.value, currentActorPage.value, pageSize.value,
      actorSearch.value.trim(), photoFilter.value,
    );
    if (!result.status) {
      ElMessage.error(result.msg);
      return false;
    }
    actors.value = result.data.dataList;
    actorTotal.value = result.data.total;
    return true;
  } finally {
    actorLoading.value = false;
  }
};

const loadPreview = async () => {
  const requestId = ++previewRequestId;
  if (!performerBasesId.value) return;
  if (scopeMode.value === 'selected' && selectedActorIds.value.length === 0) {
    preview.value = undefined;
    loading.value = false;
    return;
  }
  loading.value = true;
  preview.value = undefined;
  const allPerformers = scopeMode.value === 'all';
  const performerIds = [...currentPerformerIds()];
  try {
    const result = await server.batchPreview(performerBasesId.value, allPerformers, performerIds, strategy.value, overwrite.value);
    if (requestId !== previewRequestId) return;
    if (result.status) preview.value = result.data;
    else ElMessage.error(result.msg);
  } finally {
    if (requestId === previewRequestId) loading.value = false;
  }
};

const schedulePreview = () => {
  if (previewTimer) clearTimeout(previewTimer);
  previewTimer = setTimeout(loadPreview, 250);
};

const scopeChanged = () => {
  preview.value = undefined;
  void loadPreview();
};

const selectFilteredActors = async () => {
  const result = await server.batchActorIds(performerBasesId.value, actorSearch.value.trim(), photoFilter.value);
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }
  selectedActorIds.value = Array.from(new Set([...selectedActorIds.value, ...result.data]));
  schedulePreview();
};

const selectCurrentPageActors = () => {
  selectedActorIds.value = Array.from(new Set([...selectedActorIds.value, ...actors.value.map(actor => actor.id)]));
  schedulePreview();
};

const selectActorsWithoutPhoto = async () => {
  const result = await server.batchActorIds(performerBasesId.value, actorSearch.value.trim(), 'missing');
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }
  selectedActorIds.value = result.data;
  schedulePreview();
};

const invertCurrentPageActors = () => {
  const filteredIds = new Set(actors.value.map(actor => actor.id));
  const selectedIds = new Set(selectedActorIds.value);
  filteredIds.forEach(id => {
    if (selectedIds.has(id)) selectedIds.delete(id);
    else selectedIds.add(id);
  });
  selectedActorIds.value = Array.from(selectedIds);
  if (selectedActorIds.value.length === 0) {
    previewRequestId++;
    preview.value = undefined;
    loading.value = false;
  }
  else schedulePreview();
};

const toggleActor = (actorId: string) => {
  if (selectedActorIds.value.includes(actorId)) {
    selectedActorIds.value = selectedActorIds.value.filter(id => id !== actorId);
  } else {
    selectedActorIds.value = [...selectedActorIds.value, actorId];
  }
  if (selectedActorIds.value.length === 0) {
    previewRequestId++;
    preview.value = undefined;
    loading.value = false;
  } else {
    schedulePreview();
  }
};

const clearSelectedActors = () => {
  selectedActorIds.value = [];
  previewRequestId++;
  preview.value = undefined;
  loading.value = false;
  if (previewTimer) clearTimeout(previewTimer);
};

const submit = async () => {
  if (!preview.value) return;
  const description = preview.value.matched === 0
    ? `所选 ${preview.value.total} 位演员均未匹配到头像，执行后不会更新任何头像。仍要继续吗？`
    : overwrite.value
      ? `将为 ${preview.value.matched} 位演员匹配头像，并可能替换现有头像。是否继续？`
      : `将为 ${preview.value.matched} 位无头像演员匹配头像。是否继续？`;
  try {
    await ElMessageBox.confirm(description, '确认批量匹配', { type: overwrite.value ? 'warning' : 'info' });
  } catch {
    return;
  }
  applying.value = true;
  try {
    const startResult = await server.batchApply(
      performerBasesId.value, scopeMode.value === 'all', currentPerformerIds(), strategy.value, overwrite.value,
    );
    if (startResult.status) {
      activeBatchId = startResult.data.batchId;
      batchProgress.value = startResult.data;
      const finalProgress = startResult.data.done
        ? startResult.data
        : await waitForBatchCompletion(startResult.data.batchId);
      batchProgress.value = finalProgress;
      const failures = finalProgress.failures ?? [];
      const failedNames = failures.slice(0, 5).map(item => item.name).join('、');
      const failureSuffix = failedNames
        ? `失败演员：${failedNames}${failures.length > 5 ? '等' : ''}。`
        : '';
      await ElMessageBox.alert(
        `处理完成：成功 ${finalProgress.success}，失败 ${finalProgress.failed}，未匹配 ${finalProgress.unmatched}，跳过已有头像 ${finalProgress.skippedExisting}。${failureSuffix}`,
        finalProgress.failed > 0 ? '批量匹配完成（存在失败）' : '批量匹配完成',
        {
          type: finalProgress.failed > 0 ? 'warning' : 'success',
          confirmButtonText: '确定',
          showClose: false,
          closeOnClickModal: false,
          closeOnPressEscape: false,
        },
      );
      visible.value = false;
      emits('success');
    } else ElMessage.error(startResult.msg);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '读取批量匹配进度失败');
  } finally {
    stopProgressPoll();
    applying.value = false;
  }
};

const stopProgressPoll = () => {
  activeBatchId = '';
  if (progressTimer) clearTimeout(progressTimer);
  progressTimer = undefined;
};

const waitForBatchCompletion = (batchId: string) => new Promise<I_performerAvatarBatchProgress>((resolve, reject) => {
  let consecutiveFailures = 0;
  const poll = (delay: number) => {
    if (progressTimer) clearTimeout(progressTimer);
    progressTimer = setTimeout(async () => {
      if (activeBatchId !== batchId) return;
      const result = await server.batchProgress(batchId);
      if (activeBatchId !== batchId) return;
      if (result.status) {
        consecutiveFailures = 0;
        batchProgress.value = result.data;
        if (result.data.done) {
          resolve(result.data);
          return;
        }
      } else {
        consecutiveFailures += 1;
        if (consecutiveFailures >= 3) {
          reject(new Error(result.msg || '读取批量匹配进度失败'));
          return;
        }
      }
      poll(progressPollInterval);
    }, delay);
  };
  poll(1000);
});

defineExpose({ open });
watch([actorSearch, photoFilter], () => {
  currentActorPage.value = 1;
  if (actorListTimer) clearTimeout(actorListTimer);
  actorListTimer = setTimeout(loadActors, 300);
});
onBeforeUnmount(() => {
  if (previewTimer) clearTimeout(previewTimer);
  if (actorListTimer) clearTimeout(actorListTimer);
  stopProgressPoll();
});
</script>

<style scoped>
.scope-summary { margin-left: 12px; color: var(--el-text-color-secondary); }
.batch-form--disabled { pointer-events: none; opacity: 0.65; }
.actor-selector-form-item :deep(.el-form-item__content) { display: block; }
.actor-selector { width: 100%; }
.selector-filters { display: flex; gap: 10px; margin-bottom: 8px; }
.selector-filters .el-input { flex: 1; }
.selector-actions { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.selector-actions .el-button + .el-button { margin-left: 0; }
.selector-actions span { margin-left: auto; color: var(--el-text-color-secondary); font-size: 12px; }
.actor-scrollbar { border: 1px solid var(--el-border-color); border-radius: 7px; background: var(--el-fill-color-blank); }
.actor-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; padding: 10px; }
.actor-card { display: flex; align-items: center; gap: 8px; min-width: 0; padding: 8px 10px; border: 1px solid var(--el-border-color-lighter); border-radius: 6px; background: var(--el-bg-color); color: var(--el-text-color-primary); cursor: pointer; transition: 0.15s ease; }
.actor-card:hover { border-color: var(--el-color-primary-light-5); background: var(--el-color-primary-light-9); }
.actor-card.selected { border-color: var(--el-color-primary); background: var(--el-color-primary-light-9); }
.actor-card__photo { flex: 0 0 42px; width: 42px; height: 42px; border: 1px solid var(--el-border-color-lighter); border-radius: 50%; background: var(--el-fill-color-light); }
.actor-card__photo :deep(.el-image) { width: 100%; height: 100%; }
.actor-card__content { display: flex; flex: 1; min-width: 0; flex-direction: column; font-size: 14px; line-height: 1.3; }
.actor-card__name, .actor-card__alias { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.actor-card__name { display: block; color: var(--el-text-color-primary); font-size: 14px; font-weight: 600; line-height: 20px; }
.actor-card__alias { color: var(--el-text-color-secondary); font-size: 12px; }
.actor-pagination { justify-content: flex-end; margin-top: 10px; }
.no-match-alert { margin-top: 14px; }
.preview { margin-top: 16px; }
.batch-progress { margin-top: 16px; padding: 14px 16px; border: 1px solid var(--el-border-color); border-radius: 8px; background: var(--el-fill-color-light); }
.batch-progress__header { display: flex; justify-content: space-between; margin-bottom: 10px; color: var(--el-text-color-primary); }
.batch-progress__stats { display: flex; flex-wrap: wrap; gap: 18px; margin-top: 9px; color: var(--el-text-color-regular); font-size: 13px; }
.batch-progress__current { min-height: 20px; margin-top: 8px; overflow: hidden; color: var(--el-text-color-secondary); font-size: 13px; text-overflow: ellipsis; white-space: nowrap; }
:global(html.dark) .actor-card { border-color: #414243; background: #1d1e1f; color: #e5eaf3; }
:global(html.dark) .actor-card__name { color: #e5eaf3; }
:global(html.dark) .actor-card__alias { color: #a3a6ad; }
:global(html.dark) .actor-card:hover,
:global(html.dark) .actor-card.selected { border-color: #409eff; background: #18222c; }
:global(html.bright) .actor-card { border-color: #dcdfe6; background: #ffffff; color: #303133; }
:global(html.bright) .actor-card__name { color: #303133; }
:global(html.bright) .actor-card__alias { color: #909399; }
:global(html.bright) .actor-card:hover,
:global(html.bright) .actor-card.selected { border-color: #409eff; background: #ecf5ff; }
</style>
