<template>
  <template v-if="!editable">
    <ResourceGroupMagazine v-if="displayActive !== false && expandStyle === 'magazine'" :expected-count="expectedCount" :source-resource="sourceResource" :members="members" :loading="loading" :error="error" @retry="load" @select="emit('select', $event)" />
    <ResourceGroupFold v-else-if="displayActive !== false && (expandStyle === 'fold' || expandStyle === 'spine')" :vertical="expandStyle === 'spine'" :expected-count="expectedCount" :source-resource="sourceResource" :members="members" :loading="loading" :error="error" @retry="load" @select="emit('select', $event)" />
    <ResourceGroupFan v-else-if="displayActive !== false" :expand-style="expandStyle === 'strip' ? 'strip' : 'fan'" :members="members" :loading="loading" :error="error" :expected-count="expectedCount" :matched-ids="matchedIds" @retry="load" @select="emit('select', $event)" />
  </template>
  <section v-else class="resource-group-panel resource-group-panel--editable" v-loading="saving" :aria-busy="loading || saving">
    <header>
      <div class="panel-heading">
        <span class="panel-heading__icon"><el-icon><Collection /></el-icon></span>
        <div><strong>{{ editable ? '管理关联组' : '关联资源' }}</strong><span>{{ editable ? '调整顺序或维护组内成员' : '点击卡片打开对应资源' }}</span></div>
      </div>
      <span class="panel-count">{{ loading ? '加载中…' : `${members.length} 个资源` }}<template v-if="!loading && !editable && matchedIds"> · 本页 {{ matchedCount }} 个</template></span>
    </header>
    <div v-if="error" class="panel-state"><span>{{ error }}</span><el-button link @click="load">重新加载</el-button></div>
    <div v-else-if="loading" class="group-members" aria-hidden="true">
      <div v-for="index in 4" :key="index" class="member-skeleton"><span></span><span></span><span></span></div>
    </div>
    <el-scrollbar v-else class="panel-scrollbar" max-height="min(440px, 58vh)">
      <div class="group-members">
        <article v-for="(resource, index) in members" :key="resource.id" class="group-member">
          <button class="member-main" type="button" :title="resource.title" :disabled="editable" @click="emit('select', resource)">
            <span class="member-cover">
              <el-image :src="getResourceCoverPoster(resource)" :fit="editable ? 'cover' : 'contain'">
                <template #error><span class="member-cover__empty"><el-icon><Picture /></el-icon></span></template>
              </el-image>
              <span v-if="!editable" class="member-open" aria-hidden="true"><el-icon><ArrowRight /></el-icon></span>
            </span>
            <span class="member-info">
              <span class="member-title">{{ resource.title }}</span>
              <span v-if="resource.subtitle || resource.issueNumber || (!editable && matchedIds?.has(resource.id))" class="member-meta">
                <span class="member-subtitle">{{ resource.issueNumber || resource.subtitle }}</span>
                <span v-if="!editable && matchedIds?.has(resource.id)" class="member-match">本页</span>
              </span>
            </span>
          </button>
          <div v-if="editable" class="member-actions">
            <el-button link :disabled="saving || index === 0" :aria-label="`上移：${resource.title}`" @click="reorder(index, -1)"><el-icon><ArrowUp /></el-icon></el-button>
            <el-button link :disabled="saving || index === members.length - 1" :aria-label="`下移：${resource.title}`" @click="reorder(index, 1)"><el-icon><ArrowDown /></el-icon></el-button>
            <el-button link :disabled="saving" @click="remove(resource)">移出分组</el-button>
          </div>
        </article>
      </div>
    </el-scrollbar>
    <footer v-if="editable">
      <el-button icon="Plus" :disabled="saving || !!error || !members.length" @click="emit('add', groupId)">添加成员</el-button>
      <el-button type="danger" plain :disabled="saving || !!error || !members.length" @click="dissolve">解散分组</el-button>
    </footer>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { ArrowDown, ArrowRight, ArrowUp, Collection, Picture } from '@element-plus/icons-vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { resourceGroupServer } from '@/server/resourceGroup.server';
import { getResourceCoverPoster } from '@/common/photo';
import type { I_resource } from '@/dataType/resource.dataType';
import ResourceGroupFan from './ResourceGroupFan.vue';
import ResourceGroupFold from './ResourceGroupFold.vue';
import ResourceGroupMagazine from './ResourceGroupMagazine.vue';
const props = defineProps<{ filesBasesId: string; groupId: string; editable?: boolean; displayActive?: boolean; expectedCount?: number; sourceResource?: I_resource; expandStyle?: 'fan' | 'strip' | 'fold' | 'spine' | 'magazine'; matchedIds?: Set<string> }>();
const emit = defineEmits<{ select: [resource: I_resource]; changed: []; closed: []; add: [groupId: string] }>();
const members = ref<I_resource[]>([]);
const matchedCount = computed(() => members.value.filter(resource => props.matchedIds?.has(resource.id)).length);
const loading = ref(false);
const saving = ref(false);
const error = ref('');
let version = 0;
let alive = true;
let needsRefresh = false;
const load = async () => {
  needsRefresh = false;
  const current = ++version;
  loading.value = true;
  error.value = '';
  members.value = [];
  try {
    const result = await resourceGroupServer.info(props.filesBasesId, props.groupId);
    if (current !== version) return;
    if (!result.status) { error.value = result.msg || '分组加载失败'; return; }
    members.value = result.data;
  } catch { if (current === version) error.value = '分组加载失败，请重试'; }
  finally { if (current === version) loading.value = false; }
};
watch(() => [props.filesBasesId, props.groupId], load, { immediate: true });
// Keep the visible deck steady during list refresh; refresh its cache on the next open.
watch(() => [props.sourceResource, props.expectedCount], () => { needsRefresh = true; });
watch(() => props.displayActive, active => { if (active && needsRefresh) void load(); });
onBeforeUnmount(() => { alive = false; version++; });
const save = async (action: 'remove' | 'reorder' | 'dissolve', ids: string[], next: I_resource[]) => {
  if (!alive || saving.value || !props.editable) return;
  saving.value = true;
  const library = props.filesBasesId;
  const group = props.groupId;
  try {
    const result = await resourceGroupServer.mutate(action, library, ids, group);
    if (!alive || library !== props.filesBasesId || group !== props.groupId) return;
    if (!result.status) { ElMessage.error(result.msg || '分组操作失败'); return; }
    members.value = next;
    emit('changed');
    if (next.length < 2) emit('closed');
    ElMessage.success(action === 'dissolve' ? '已解散分组，资源保持不变' : action === 'remove' ? '已移出分组' : '已更新组内顺序');
  } catch { if (alive && library === props.filesBasesId && group === props.groupId) ElMessage.error('分组操作失败，请重试'); }
  finally { saving.value = false; }
};
const remove = (resource: I_resource) => save('remove', [resource.id], members.value.filter(item => item.id !== resource.id));
const reorder = (index: number, step: number) => {
  const next = [...members.value];
  const target = index + step;
  if (target < 0 || target >= next.length) return;
  [next[index], next[target]] = [next[target], next[index]];
  return save('reorder', next.map(resource => resource.id), next);
};
const dissolve = async () => {
  const library = props.filesBasesId;
  const group = props.groupId;
  try { await ElMessageBox.confirm('只解除本组所有资源之间的关联，不会删除资源或文件。', '解散关联组', { type: 'warning' }); }
  catch { return; }
  if (!alive || library !== props.filesBasesId || group !== props.groupId) return;
  await save('dissolve', [], []);
};
</script>

<style scoped lang="scss">
.resource-group-panel { min-height: 120px; color: var(--el-text-color-primary); background: var(--el-bg-color-overlay); }
.panel-scrollbar { height: auto; }
header { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 12px; padding: 16px 18px; border-bottom: 1px solid var(--el-border-color-lighter); }
.panel-heading { display: flex; align-items: center; gap: 10px; min-width: 0; }
.panel-heading__icon { display: grid; place-items: center; width: 34px; height: 34px; border-radius: 10px; color: var(--el-color-primary); background: color-mix(in srgb, var(--el-color-primary) 12%, transparent); font-size: 18px; }
.panel-heading > div { display: grid; gap: 4px; }
.panel-heading strong { font-size: 14px; font-weight: 600; line-height: 20px; }
.panel-heading > div > span { font-size: 12px; color: var(--el-text-color-secondary); }
.panel-count { padding: 5px 8px; border-radius: 6px; background: var(--el-fill-color-light); color: var(--el-text-color-secondary); font-size: 12px; white-space: nowrap; }
.group-members { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); align-items: start; gap: 12px; padding: 14px; }
.group-member { min-width: 0; }
.member-main { width: 100%; display: grid; gap: 10px; padding: 8px; box-sizing: border-box; border: 1px solid var(--el-border-color-lighter); border-radius: 10px; background: var(--el-bg-color); text-align: left; font: inherit; color: inherit; cursor: pointer; transition: border-color .16s ease, box-shadow .16s ease; }
.member-main:not(:disabled):hover, .member-main:focus-visible { border-color: color-mix(in srgb, var(--el-color-primary) 65%, transparent); box-shadow: 0 3px 12px rgba(0,0,0,.08); }
.member-main:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: 2px; }
.member-main:disabled { cursor: default; opacity: 1; }
.member-cover { position: relative; display: block; width: 100%; aspect-ratio: 8 / 5; overflow: hidden; border-radius: 6px; background: var(--el-fill-color-light); }
.member-cover .el-image { display: block; width: 100%; height: 100%; }
.member-cover__empty { display: grid; place-items: center; width: 100%; height: 100%; font-size: 28px; color: var(--el-text-color-placeholder); }
.member-open { position: absolute; display: grid; place-items: center; right: 8px; bottom: 8px; width: 26px; height: 26px; border-radius: 8px; background: rgba(18,24,30,.85); color: white; opacity: 0; transform: translateY(3px); transition: opacity .16s ease, transform .16s ease; }
.member-main:hover .member-open, .member-main:focus-visible .member-open { opacity: 1; transform: translateY(0); }
.member-info { display: grid; gap: 6px; min-width: 0; padding: 0 2px 3px; }
.member-title { min-height: 40px; font-size: 13px; font-weight: 500; line-height: 20px; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; overflow-wrap: anywhere; }
.member-meta { display: flex; align-items: center; justify-content: space-between; gap: 8px; min-height: 20px; }
.member-subtitle { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 12px; color: var(--el-text-color-secondary); }
.member-match { flex-shrink: 0; padding: 1px 6px; border-radius: 4px; color: var(--el-color-primary); background: color-mix(in srgb, var(--el-color-primary) 10%, transparent); font-size: 11px; line-height: 18px; }
.member-actions { display: flex; align-items: center; justify-content: space-between; gap: 3px; margin-top: 7px; }
.member-actions .el-button { margin: 0; font-size: 12px; }
.resource-group-panel--editable .group-members { grid-template-columns: repeat(3, minmax(0, 1fr)); }
.resource-group-panel--editable .member-main { padding: 0; border: 0; background: transparent; }
.resource-group-panel--editable .member-cover { aspect-ratio: 2 / 3; }
footer { display: flex; justify-content: space-between; gap: 10px; padding: 12px 14px; border-top: 1px solid var(--el-border-color-lighter); }
.panel-state { display: grid; gap: 10px; padding: 25px; text-align: center; color: var(--el-text-color-secondary); }
.member-skeleton { display: grid; gap: 10px; padding: 8px; border: 1px solid var(--el-border-color-lighter); border-radius: 10px; }
.member-skeleton span { display: block; height: 12px; border-radius: 5px; background: var(--el-fill-color); }
.member-skeleton span:first-child { height: auto; aspect-ratio: 8 / 5; }
.member-skeleton span:last-child { width: 60%; }
</style>
