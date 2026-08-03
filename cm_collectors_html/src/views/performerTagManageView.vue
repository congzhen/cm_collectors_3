<template>
  <div class="performer-tag-manage-view">
    <HeaderView :mode="E_headerMode.GoBack" />
    <div class="page-header">
      <div>
        <h2>演员标签管理</h2>
        <span>管理当前演员库的标签分类与标签名称</span>
      </div>
      <div class="header-actions">
        <el-input v-model="keyword" clearable placeholder="搜索标签名称" prefix-icon="Search" />
        <el-button type="primary" plain icon="FolderAdd" @click="createClass">新增标签分类</el-button>
        <el-button type="primary" icon="CirclePlus" :disabled="!currentClass" @click="createTag">新增标签</el-button>
        <el-button @click="router.push(`/performer/${props.performerBasesId}`)">返回演员管理</el-button>
      </div>
    </div>

    <div class="manage-layout" v-loading="loading">
      <section class="panel class-panel">
        <div class="panel-title"><strong>标签分类</strong><span>{{ classes.length }}</span></div>
        <draggable v-model="classes" item-key="id" class="class-list" handle=".drag-handle" @end="saveSort">
          <template #item="{ element }">
            <div :class="['class-row', { active: currentClass?.id === element.id }]" @click="selectClass(element)">
              <el-icon class="drag-handle"><Rank /></el-icon>
              <span class="class-name">{{ element.name }}</span>
              <span class="class-count">{{ tags.filter(tag => tag.performerTagClass_id === element.id).length }}</span>
              <el-button link icon="Edit" @click.stop="editClass(element)" />
              <el-button link type="danger" icon="Delete" @click.stop="deleteClass(element)" />
            </div>
          </template>
        </draggable>
        <el-button class="add-class" plain icon="Plus" @click="createClass">新增分类</el-button>
      </section>

      <section class="panel tag-panel">
        <div class="panel-title">
          <strong>{{ currentClass?.name || '标签' }}</strong>
          <el-button link type="primary" :disabled="!currentClass" @click="createTag">新增标签</el-button>
        </div>
        <div class="tag-table-header">
          <span>排序</span><span>标签名称</span><span>演员数量</span><span>状态</span><span>操作</span>
        </div>
        <el-scrollbar>
          <draggable v-model="currentTags" item-key="id" handle=".drag-handle" :disabled="!!keyword.trim()" @end="saveTagOrder">
            <template #item="{ element }">
              <div class="tag-row">
                <el-icon class="drag-handle"><Rank /></el-icon>
                <strong>{{ element.name }}</strong>
                <span class="performer-count">{{ element.performerCount }} 人</span>
                <el-tag size="small" :type="element.status ? 'success' : 'info'">{{ element.status ? '启用' : '停用' }}</el-tag>
                <div class="row-actions">
                  <el-button link @click.stop="editTag(element)">编辑</el-button>
                  <el-button link @click.stop="toggleTag(element)">{{ element.status ? '停用' : '启用' }}</el-button>
                  <el-button link type="danger" @click.stop="deleteTag(element)">删除</el-button>
                </div>
              </div>
            </template>
          </draggable>
          <el-empty v-if="!currentTags.length" description="当前分类还没有标签" />
        </el-scrollbar>
        <div class="delete-hint"><el-icon><Warning /></el-icon>删除标签只会解除演员关联，不会删除演员</div>
      </section>

    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import draggable from 'vuedraggable';
import { ElMessage, ElMessageBox } from 'element-plus';
import HeaderView from './HeaderView.vue';
import { E_headerMode } from '@/dataType/app.dataType';
import { performerTagServer } from '@/server/performerTag.server';
import type { I_performerTag, I_performerTagClass } from '@/dataType/performerTag.dataType';

const props = defineProps<{ performerBasesId: string }>();
const router = useRouter();
const loading = ref(false);
const keyword = ref('');
const classes = ref<I_performerTagClass[]>([]);
const tags = ref<I_performerTag[]>([]);
const currentClass = ref<I_performerTagClass>();

const currentTags = computed({
  get: () => tags.value.filter(tag => tag.performerTagClass_id === currentClass.value?.id && tag.name.toLowerCase().includes(keyword.value.trim().toLowerCase())),
  set: value => {
    const visibleIDs = new Set(value.map(tag => tag.id));
    const other = tags.value.filter(tag => tag.performerTagClass_id !== currentClass.value?.id || !visibleIDs.has(tag.id));
    tags.value = [...other, ...value];
  },
});
const currentClassTagCount = () => tags.value.filter(tag => tag.performerTagClass_id === currentClass.value?.id).length;
const showRequestError = (result: { status: boolean; msg?: string } | undefined, fallback: string) => {
  if (result?.status) return false;
  ElMessage.error(result?.msg || fallback);
  return true;
};

const load = async (preserveClassId = '') => {
  loading.value = true;
  try {
    const result = await performerTagServer.data(props.performerBasesId, true);
    if (!result?.status) return ElMessage.error(result?.msg || '加载演员标签失败');
    classes.value = result.data.tagClasses;
    tags.value = result.data.tags;
    currentClass.value = classes.value.find(item => item.id === preserveClassId) || classes.value[0];
  } finally { loading.value = false; }
};
const selectClass = (item: I_performerTagClass) => { currentClass.value = item; };
const promptName = async (title: string, value = '') => {
  const result = await ElMessageBox.prompt('请输入名称', title, { inputValue: value, inputPattern: /\S+/, inputErrorMessage: '名称不能为空' });
  return result.value.trim();
};
const createClass = async () => {
  try {
    const name = await promptName('新增标签分类');
    const result = await performerTagServer.createClass({ performerBases_id: props.performerBasesId, name, sort: classes.value.length, status: true });
    if (!showRequestError(result, '新增标签分类失败')) await load(result.data.id);
  } catch { /* 用户取消 */ }
};
const editClass = async (item: I_performerTagClass) => {
  try { const name = await promptName('编辑标签分类', item.name); const result = await performerTagServer.updateClass({ ...item, name }); if (!showRequestError(result, '编辑标签分类失败')) await load(item.id); } catch { /* 用户取消 */ }
};
const deleteClass = async (item: I_performerTagClass) => {
  try {
    await ElMessageBox.confirm(`删除分类“${item.name}”会同时删除其中的标签，是否继续？`, '删除标签分类', { type: 'warning' });
    const result = await performerTagServer.deleteClass(item.id); if (!showRequestError(result, '删除标签分类失败')) await load();
  } catch { /* 用户取消 */ }
};
const createTag = async () => {
  if (!currentClass.value) return;
  try {
    const name = await promptName('新增演员标签');
    const result = await performerTagServer.createTag({ performerTagClass_id: currentClass.value.id, name, sort: currentClassTagCount(), status: true });
    if (!showRequestError(result, '新增演员标签失败')) await load(currentClass.value.id);
  } catch { /* 用户取消 */ }
};
const editTag = async (item: I_performerTag) => {
  try { const name = await promptName('编辑演员标签', item.name); const result = await performerTagServer.updateTag({ ...item, name }); if (!showRequestError(result, '编辑演员标签失败')) await load(currentClass.value?.id); } catch { /* 用户取消 */ }
};
const toggleTag = async (item: I_performerTag) => { const result = await performerTagServer.updateTag({ ...item, status: !item.status }); if (!showRequestError(result, '更新标签状态失败')) await load(currentClass.value?.id); };
const deleteTag = async (item: I_performerTag) => {
  try { await ElMessageBox.confirm(`删除标签“${item.name}”只会解除演员关联，是否继续？`, '删除演员标签', { type: 'warning' }); const result = await performerTagServer.deleteTag(item.id); if (!showRequestError(result, '删除演员标签失败')) await load(currentClass.value?.id); } catch { /* 用户取消 */ }
};
const saveSort = async () => { const result = await performerTagServer.updateSort(classes.value.map((item, sort) => ({ id: item.id, sort })), []); if (showRequestError(result, '保存分类排序失败')) await load(currentClass.value?.id); };
const saveTagOrder = async () => { const result = await performerTagServer.updateSort([], currentTags.value.map((item, sort) => ({ id: item.id, sort }))); if (showRequestError(result, '保存标签排序失败')) await load(currentClass.value?.id); };

onMounted(() => load());
</script>

<style scoped lang="scss">
.performer-tag-manage-view { height: 100%; display: flex; flex-direction: column; overflow: hidden; background: var(--el-bg-color-page); }
.page-header { display: flex; justify-content: space-between; align-items: center; gap: 20px; padding: 14px 20px; border-bottom: 1px solid var(--el-border-color); }
.page-header h2 { margin: 0 0 4px; font-size: 20px; }
.page-header span { color: var(--el-text-color-secondary); }
.header-actions { display: flex; gap: 8px; align-items: center; }
.header-actions .el-input { width: 260px; }
.manage-layout { flex: 1; min-height: 0; display: grid; grid-template-columns: 280px minmax(580px, 1fr); gap: 12px; padding: 12px; }
.panel { min-height: 0; display: flex; flex-direction: column; border: 1px solid var(--el-border-color); border-radius: 4px; background: var(--el-bg-color); overflow: hidden; }
.panel-title { min-height: 48px; box-sizing: border-box; padding: 0 16px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid var(--el-border-color); }
.class-list { flex: 1; padding: 8px; }
.class-row { display: grid; grid-template-columns: 24px 1fr auto 28px 28px; align-items: center; gap: 6px; min-height: 44px; padding: 0 8px; border-radius: 3px; cursor: pointer; }
.class-row:hover { background: var(--el-fill-color-light); }
.class-row.active { color: var(--el-color-primary); background: var(--el-color-primary-light-9); }
.class-count { min-width: 24px; text-align: center; padding: 2px 5px; border-radius: 3px; background: var(--el-fill-color); }
.drag-handle { cursor: move; color: var(--el-text-color-secondary); }
.add-class { margin: 12px; }
.tag-table-header, .tag-row { display: grid; grid-template-columns: 70px minmax(130px, 1fr) 110px 90px 220px; align-items: center; gap: 8px; }
.tag-table-header { min-height: 42px; padding: 0 16px; color: var(--el-text-color-secondary); border-bottom: 1px solid var(--el-border-color); }
.tag-row { min-height: 58px; padding: 0 16px; border-bottom: 1px solid var(--el-border-color-lighter); }
.tag-row:hover { background: var(--el-fill-color-light); }
.performer-count { justify-self: start; min-width: 42px; padding: 3px 9px; border-radius: 10px; color: var(--el-text-color-secondary); background: var(--el-fill-color-light); font-size: 12px; line-height: 16px; text-align: center; font-variant-numeric: tabular-nums; }
.row-actions { display: flex; gap: 10px; }
.delete-hint { padding: 12px 16px; color: var(--el-color-warning); border-top: 1px solid var(--el-border-color); }
@media (max-width: 1250px) { .manage-layout { grid-template-columns: 230px minmax(540px, 1fr); } }
</style>
