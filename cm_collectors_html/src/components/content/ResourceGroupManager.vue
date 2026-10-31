<template>
  <div class="resource-group-manager">
    <template v-if="targetGroup">
      <span class="group-target-hint">请选择要加入的资源</span>
      <el-button type="primary" plain :disabled="!canAddTarget || busy" :loading="busy" @click="addTo(targetGroup)">加入此组</el-button>
      <el-button link :disabled="busy" @click="targetGroup = ''">取消添加</el-button>
    </template>
    <template v-else>
      <el-button v-if="!selectedGroupIds.length" icon="Collection" :disabled="selected.length < 2 || busy" :loading="busy" @click="create">创建关联组</el-button>
      <el-button v-if="selectedGroupIds.length === 1 && ungrouped.length" type="primary" plain :disabled="busy" @click="addTo(selectedGroupIds[0])">加入此组</el-button>
      <el-button v-if="!selectedGroupIds.length" icon="FolderAdd" :disabled="!selected.length || busy" @click="openPicker">加入已有组</el-button>
      <el-popover v-if="selectedGroupIds.length === 1" v-model:visible="panelVisible" trigger="click" placement="bottom-start" :width="520" :show-arrow="false"
        :popper-style="{ maxWidth: 'calc(100vw - 24px)', padding: '0', borderRadius: '14px', overflow: 'hidden' }">
        <template #reference><el-button icon="Collection" :disabled="busy">查看分组</el-button></template>
        <ResourceGroupPanel v-if="panelVisible" :files-bases-id="filesBasesId" :group-id="selectedGroupIds[0]" editable
          @changed="emit('changed')" @closed="panelVisible = false" @add="beginAdd" />
      </el-popover>
      <el-button v-if="selectedGroupIds.length === 1" :disabled="busy" @click="removeSelected">移出分组</el-button>
      <span v-if="selectedGroupIds.length > 1" class="group-target-hint">请选择同一组的成员，暂不支持合并多个组</span>
    </template>
    <el-dialog v-model="pickerVisible" title="加入已有组" width="560px" append-to-body :close-on-click-modal="!busy">
      <el-radio-group v-model="pickerScope" class="group-picker-scope" size="default" aria-label="分组显示范围" @change="changeScope">
        <el-radio-button value="page">本页关联组</el-radio-button>
        <el-radio-button value="all">同库全部组</el-radio-button>
      </el-radio-group>
      <el-input v-model="search" placeholder="按资源名称、副标题或番号查找分组" clearable @input="scheduleSearch" />
      <div v-loading="pickerLoading || busy" class="group-picker" :aria-busy="pickerLoading || busy">
        <div v-if="pickerError" class="picker-state"><span>{{ pickerError }}</span><el-button link @click="loadGroups">重试</el-button></div>
        <el-empty v-else-if="!groups.length && !pickerLoading" :description="pickerScope === 'page' ? '本页没有匹配的关联组，可切换到同库全部组' : '没有匹配的关联组'" />
        <button v-for="group in groups" :key="group.id" class="group-choice" type="button" :disabled="busy || pickerLoading" @click="addTo(group.id)">
          <span class="group-choice__covers"><img loading="lazy" decoding="async" v-for="(resource, index) in group.preview.slice(0, 3)" :key="resource.id" :src="getResourceCoverPoster(resource) || undefined" :style="{ left: `${index * 9}px`, top: `${index * 3}px` }" alt="" /></span>
          <span class="group-choice__info"><strong>{{ group.preview[0]?.title || '关联资源' }}</strong><span class="group-choice__meta"><ResourceGroupMark v-if="pageGroupIds.includes(group.id)" :group-id="group.id" /><span>{{ group.count }} 个资源 · {{ pageGroupIds.includes(group.id) ? '本页' : '其他页' }}</span></span></span>
          <el-icon><Plus /></el-icon>
        </button>
      </div>
      <template #footer><el-pagination v-if="pickerScope === 'all' && total > 20" v-model:current-page="page" :total="total" :page-size="20" :disabled="busy || pickerLoading" layout="prev, pager, next" @current-change="loadGroups" /></template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { Plus } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import type { I_resource, I_resourceGroupSummary } from '@/dataType/resource.dataType';
import { resourceGroupServer } from '@/server/resourceGroup.server';
import { getResourceCoverPoster } from '@/common/photo';
import ResourceGroupPanel from './ResourceGroupPanel.vue';
import ResourceGroupMark from './ResourceGroupMark.vue';
const props = withDefaults(defineProps<{ selected: I_resource[]; filesBasesId: string; pageResources?: I_resource[] }>(), { pageResources: () => [] });
const emit = defineEmits<{ changed: []; clearSelection: [] }>();
const selectedGroupIds = computed(() => [...new Set(props.selected.map(resource => resource.associationGroupId).filter((id): id is string => !!id))]);
const ungrouped = computed(() => props.selected.filter(resource => !resource.associationGroupId));
const busy = ref(false);
const panelVisible = ref(false);
const pickerVisible = ref(false);
const pickerLoading = ref(false);
const pickerError = ref('');
const groups = ref<I_resourceGroupSummary[]>([]);
const total = ref(0);
const page = ref(1);
const search = ref('');
const pickerScope = ref<'page' | 'all'>('page');
const pageGroupIds = computed(() => [...new Set(props.pageResources.map(resource => resource.associationGroupId).filter((id): id is string => !!id))]);
const targetGroup = ref('');
const canAddTarget = computed(() => props.selected.length > 0 && props.selected.some(resource => !resource.associationGroupId)
  && props.selected.every(resource => !resource.associationGroupId || resource.associationGroupId === targetGroup.value));
let requestVersion = 0;
let timer: ReturnType<typeof setTimeout> | undefined;
let alive = true;
watch(() => props.filesBasesId, () => {
  panelVisible.value = false; pickerVisible.value = false; targetGroup.value = ''; groups.value = [];
  requestVersion++; clearTimeout(timer);
});
watch(selectedGroupIds, () => { if (selectedGroupIds.value.length !== 1) panelVisible.value = false; });
onBeforeUnmount(() => { alive = false; requestVersion++; clearTimeout(timer); });
const mutate = async (action: 'create' | 'add' | 'remove', groupId = '') => {
  if (busy.value || !props.selected.length) return;
  const library = props.filesBasesId;
  const ids = props.selected.filter(resource => action !== 'remove' || resource.associationGroupId === groupId).map(resource => resource.id);
  busy.value = true;
  try {
    const result = await resourceGroupServer.mutate(action, library, ids, groupId);
    if (!alive || library !== props.filesBasesId) return;
    if (!result.status) { ElMessage.error(result.msg || '分组操作失败'); return; }
    pickerVisible.value = false; targetGroup.value = '';
    ElMessage.success(action === 'create' ? '已创建关联组' : action === 'add' ? '已加入关联组' : '已移出分组');
    emit('changed');
  } catch { if (alive && library === props.filesBasesId) ElMessage.error('分组操作失败，请重试'); }
  finally { busy.value = false; }
};
const create = () => mutate('create');
const addTo = (groupId: string) => mutate('add', groupId);
const removeSelected = () => mutate('remove', selectedGroupIds.value[0]);
const beginAdd = (groupId: string) => { panelVisible.value = false; targetGroup.value = groupId; emit('clearSelection'); };
const loadGroups = async () => {
  const current = ++requestVersion;
  const library = props.filesBasesId;
  const scope = pickerScope.value;
  const ids = scope === 'page' ? pageGroupIds.value : [];
  pickerLoading.value = true; pickerError.value = ''; groups.value = [];
  // 空数组在摘要接口中代表全库；本页无组时直接展示空态，不能误查全库。
  if (scope === 'page' && !ids.length) { groups.value = []; total.value = 0; pickerLoading.value = false; return; }
  try {
    const result = await resourceGroupServer.summaries(library, ids, search.value.trim(), page.value);
    if (current !== requestVersion) return;
    if (!result.status) { pickerError.value = result.msg || '分组加载失败'; return; }
    groups.value = scope === 'page'
      ? [...result.data.dataList].sort((a, b) => ids.indexOf(a.id) - ids.indexOf(b.id))
      : result.data.dataList;
    total.value = result.data.total;
  } catch { if (current === requestVersion) pickerError.value = '分组加载失败，请重试'; }
  finally { if (current === requestVersion) pickerLoading.value = false; }
};
const changeScope = () => { clearTimeout(timer); page.value = 1; groups.value = []; void loadGroups(); };
const openPicker = () => { clearTimeout(timer); search.value = ''; pickerScope.value = 'page'; page.value = 1; groups.value = []; pickerVisible.value = true; void loadGroups(); };
const scheduleSearch = () => { clearTimeout(timer); requestVersion++; groups.value = []; pickerLoading.value = true; timer = setTimeout(() => { page.value = 1; void loadGroups(); }, 250); };
watch(() => pageGroupIds.value.join('|'), () => {
  if (pickerVisible.value && pickerScope.value === 'page') changeScope();
});
</script>
<style scoped lang="scss">
.resource-group-manager { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.resource-group-manager > .el-button { margin: 0; border-radius: 8px; height: 32px; }
.group-target-hint { color: var(--el-text-color-secondary); font-size: 12px; }
.group-picker-scope { margin-bottom: 14px; }
.group-picker { display: grid; gap: 8px; min-height: 100px; max-height: 50vh; overflow: auto; margin-top: 14px; }
.group-choice { display: flex; gap: 14px; align-items: center; padding: 10px; border: 1px solid var(--el-border-color-lighter); border-radius: 9px; background: var(--el-bg-color); color: var(--el-text-color-primary); text-align: left; cursor: pointer; }
.group-choice:hover { border-color: var(--el-color-primary); background: var(--el-fill-color-light); }
.group-choice:disabled { cursor: wait; opacity: .6; }
.group-choice__covers { position: relative; flex: 0 0 64px; width: 64px; height: 65px; }
.group-choice__covers img { position: absolute; width: 42px; height: 58px; object-fit: cover; background: var(--el-fill-color-dark); border: 1px solid var(--el-border-color); border-radius: 4px; }
.group-choice__info { display: grid; flex: 1; min-width: 0; gap: 7px; }
.group-choice__info strong { font-size: 13px; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.group-choice__info > span { color: var(--el-text-color-secondary); font-size: 12px; }
.group-choice__meta { display: flex; align-items: center; gap: 8px; }
.picker-state { display: grid; gap: 10px; text-align: center; color: var(--el-text-color-secondary); padding: 25px; }
</style>
