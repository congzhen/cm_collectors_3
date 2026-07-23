<template>
  <el-dialog v-model="visible" :title="`从头像库选择头像 - ${performer?.name || ''}`" width="860px" destroy-on-close>
    <div v-loading="loading">
      <el-empty v-if="!loading && candidates.length === 0" description="头像库中没有匹配到该演员" />
      <el-scrollbar v-else height="520px">
        <div class="candidate-grid">
          <button v-for="candidate in candidates" :key="candidate.id" type="button" class="candidate"
            :class="{ selected: selectedCandidateId === candidate.id }" @click="selectCandidate(candidate.id)">
            <div class="candidate-image">
              <el-image v-if="previewStates[candidate.id] === 'loaded'" :src="previewImages[candidate.id]" fit="cover"
                @error="previewStates[candidate.id] = 'error'" />
              <div v-else-if="previewStates[candidate.id] === 'loading'" class="image-state image-state--loading">
                <span class="loading-dot" />
                正在加载
              </div>
              <div v-else class="image-state image-state--error" :title="previewErrors[candidate.id]"
                @click.stop="retryPreviewImage(candidate.id)">
                加载失败，点击重试
              </div>
            </div>
            <div class="candidate-info">
              <span :title="candidate.source">{{ candidate.source }}</span>
              <el-tag v-if="candidate.aiFixed" size="small" type="warning">AI</el-tag>
              <el-tag v-else size="small" type="success">原图</el-tag>
            </div>
          </button>
        </div>
      </el-scrollbar>
    </div>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="saving" :disabled="!selectedCandidateId" @click="submit">使用此头像</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { I_performer } from '@/dataType/performer.dataType';
import type { I_performerAvatarCandidate } from '@/dataType/performerAvatarLibrary.dataType';
import { performerAvatarLibraryServer as server } from '@/server/performerAvatarLibrary.server';
import { ensurePerformerAvatarLibraryReady } from '@/common/performerAvatarLibrary';

const emits = defineEmits(['success']);
const visible = ref(false);
const loading = ref(false);
const saving = ref(false);
const performer = ref<I_performer>();
const candidates = ref<I_performerAvatarCandidate[]>([]);
const selectedCandidateId = ref('');
const previewImages = ref<Record<string, string>>({});
const previewStates = ref<Record<string, 'loading' | 'loaded' | 'error'>>({});
const previewErrors = ref<Record<string, string>>({});
let previewGeneration = 0;
const previewConcurrency = 4;

const clearPreviewImages = () => {
  previewGeneration += 1;
  Object.values(previewImages.value).forEach(value => URL.revokeObjectURL(value));
  previewImages.value = {};
  previewStates.value = {};
  previewErrors.value = {};
};

const loadPreviewImage = async (candidateId: string, generation: number) => {
  if (!performer.value) return;
  previewStates.value[candidateId] = 'loading';
  const result = await server.previewImage(performer.value.id, candidateId);
  if (generation !== previewGeneration) return;
  if (!result.status) {
    previewErrors.value[candidateId] = result.msg;
    previewStates.value[candidateId] = 'error';
    return;
  }
  const previousImage = previewImages.value[candidateId];
  if (previousImage) URL.revokeObjectURL(previousImage);
  previewImages.value[candidateId] = URL.createObjectURL(result.data);
  delete previewErrors.value[candidateId];
  previewStates.value[candidateId] = 'loaded';
};

const loadPreviewImages = () => {
  previewGeneration += 1;
  const generation = previewGeneration;
  const queue = candidates.value.map(candidate => candidate.id);
  const worker = async () => {
    while (queue.length > 0 && generation === previewGeneration) {
      const candidateId = queue.shift();
      if (candidateId) await loadPreviewImage(candidateId, generation);
    }
  };
  for (let index = 0; index < Math.min(previewConcurrency, queue.length); index += 1) void worker();
};

const retryPreviewImage = (candidateId: string) => {
  void loadPreviewImage(candidateId, previewGeneration);
};

const open = async (value: I_performer) => {
  performer.value = value;
  visible.value = true;
  loading.value = true;
  candidates.value = [];
  selectedCandidateId.value = '';
  clearPreviewImages();
  try {
    const libraryStatus = await ensurePerformerAvatarLibraryReady();
    if (!libraryStatus) {
      visible.value = false;
      return;
    }
    const candidateResult = await server.candidates(value.id);
    if (!candidateResult.status) {
      ElMessage.error(candidateResult.msg);
      return;
    }
    candidates.value = candidateResult.data;
    loadPreviewImages();
  } finally {
    loading.value = false;
  }
};

const selectCandidate = (candidateId: string) => {
  selectedCandidateId.value = candidateId;
};

const submit = async () => {
  if (!performer.value || !selectedCandidateId.value) return;
  const overwrite = performer.value.photo !== '';
  try {
    await ElMessageBox.confirm(
      overwrite ? `将替换【${performer.value.name}】现有头像，是否继续？` : `确定为【${performer.value.name}】使用所选头像吗？`,
      '确认更新头像', { type: overwrite ? 'warning' : 'info' },
    );
  } catch {
    return;
  }
  saving.value = true;
  try {
    const result = await server.apply(performer.value.id, selectedCandidateId.value, overwrite);
    if (result.status) {
      ElMessage.success('演员头像已更新');
      visible.value = false;
      emits('success');
    } else ElMessage.error(result.msg);
  } finally {
    saving.value = false;
  }
};

defineExpose({ open });

watch(visible, value => { if (!value) clearPreviewImages(); });
onBeforeUnmount(clearPreviewImages);
</script>

<style scoped lang="scss">
.candidate-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); gap: 12px; padding-right: 10px; }
.candidate { padding: 6px; border: 2px solid transparent; border-radius: 6px; background: var(--el-fill-color-light); color: inherit; cursor: pointer; text-align: left; }
.candidate.selected { border-color: var(--el-color-primary); }
.candidate-image { width: 100%; aspect-ratio: 2 / 3; overflow: hidden; border-radius: 4px; background: var(--el-fill-color-darker); }
.candidate-image .el-image { width: 100%; height: 100%; }
.candidate-info { display: flex; align-items: center; justify-content: space-between; gap: 6px; margin-top: 5px; }
.candidate-info span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.image-state { display: flex; align-items: center; justify-content: center; gap: 7px; width: 100%; height: 100%; color: var(--el-text-color-secondary); }
.image-state--error { cursor: pointer; }
.image-state--error:hover { color: var(--el-color-primary); }
.loading-dot { width: 12px; height: 12px; border: 2px solid var(--el-border-color); border-top-color: var(--el-color-primary); border-radius: 50%; animation: preview-loading 0.8s linear infinite; }
@keyframes preview-loading { to { transform: rotate(360deg); } }
</style>
