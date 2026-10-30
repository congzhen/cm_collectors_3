<template>
  <div class="content-list-admin">
    <div class="management-toolbar">
      <div class="management-toolbar__header">
        <div class="management-title">
          <span class="management-title__icon"><el-icon><Operation /></el-icon></span>
          <div><strong>资源管理</strong><span class="management-hint">{{ managementView === 'layout' ? '点击卡片选择资源' : '勾选表格中的资源' }}</span></div>
        </div>
        <el-radio-group v-model="managementView" size="default" aria-label="管理视图" class="management-view-switch">
          <el-radio-button value="layout"><el-icon><Grid /></el-icon> 当前布局</el-radio-button>
          <el-radio-button value="table"><el-icon><List /></el-icon> 管理表格</el-radio-button>
        </el-radio-group>
      </div>
      <div class="management-toolbar__actions">
        <div class="management-selection">
          <el-checkbox :model-value="allSelected" :indeterminate="partiallySelected" :disabled="!props.dataList.length"
            @change="toggleAll">全选本页</el-checkbox>
          <span class="selection-count">已选 <strong>{{ selectedResources.length }}</strong> 项</span>
          <el-button link :disabled="!selectedResources.length" @click="clearSelection">清空选择</el-button>
          <el-popover v-model:visible="selectionPreviewVisible" trigger="hover" placement="bottom-start"
            :width="360" :show-after="180" :hide-after="180" :disabled="!selectedResources.length"
            :popper-style="{ maxWidth: 'calc(100vw - 24px)', padding: '0' }">
            <template #reference>
              <el-button class="selection-preview-trigger" link :disabled="!selectedResources.length">
                <el-icon><List /></el-icon><span>查看已选</span><el-icon><ArrowDown /></el-icon>
              </el-button>
            </template>
            <div class="selection-preview">
              <div class="selection-preview__header"><strong>已选资源</strong><span>{{ selectedResources.length }} 项</span></div>
              <el-scrollbar max-height="340px">
                <ul class="selection-preview__list">
                  <li v-for="resource in selectedResources" :key="resource.id" class="selection-preview__item">
                    <el-image class="selection-preview__cover" :src="getResourceCoverPoster(resource)" fit="cover">
                      <template #error><el-icon><Picture /></el-icon></template>
                    </el-image>
                    <div class="selection-preview__info">
                      <span class="selection-preview__title" :title="resource.title">{{ resource.title }}</span>
                      <span v-if="resource.subtitle || resource.issueNumber" class="selection-preview__subtitle"
                        :title="resource.subtitle || resource.issueNumber">{{ resource.subtitle || resource.issueNumber }}</span>
                    </div>
                    <el-button class="selection-preview__remove" link :aria-label="`取消选择：${resource.title}`"
                      title="取消选择" @click="removeSelection(resource.id)"><el-icon><Close /></el-icon></el-button>
                  </li>
                </ul>
              </el-scrollbar>
            </div>
          </el-popover>
        </div>
        <div class="batch-actions">
          <el-button icon="PriceTag" :disabled="!selectedResources.length" @click="batchAddTagHandle">添加标签</el-button>
          <el-button icon="Remove" :disabled="!selectedResources.length" @click="batchDeleteTagHandle">移除标签</el-button>
          <el-button icon="User" :disabled="!selectedResources.length" @click="batchAddPerformerHandle">添加演员</el-button>
          <el-button icon="Star" :disabled="!selectedResources.length" @click="batchSetStarsHandle">设置评星</el-button>
          <el-button icon="VideoCamera" :loading="addingTranscode" :disabled="!selectedResources.length" @click="batchAddTranscodeHandle">加入转码与剪辑</el-button>
          <el-button class="batch-delete" type="danger" plain icon="Delete" :disabled="!selectedResources.length" @click="batchDeleteHandle">删除资源</el-button>
        </div>
      </div>
    </div>
    <div class="table-container">
      <contentList ref="layoutRef" v-if="managementView === 'layout'" :data-list="props.dataList" />
      <el-table v-else ref="tableRef" :data="props.dataList" row-key="id" border height="100%" style="width: 100%"
        :row-class-name="selectionRowClass"
        @selection-change="tableSelectionChange">
        <el-table-column type="selection" width="55" />
        <el-table-column width="120" label="操作">
          <template #default="{ row }">
            <el-button-group>
              <el-button icon="VideoPlay" @click="playResourceHandle(row)" />
              <el-button icon="Folder" @click="playOpenResourceFolder(row.id)" />
              <el-button icon="Edit" @click="editResourceHandle(row)" />
              <el-button icon="Delete" @click="resourceDeleteHandle(row)" />
            </el-button-group>
          </template>
        </el-table-column>
        <el-table-column width="120" label="封面">
          <template #default="{ row }">
            <el-image :src="getResourceCoverPoster(row)" fit="cover" />
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="180" />
        <el-table-column prop="issueNumber" label="版号、番号、刊号" width="140" />
        <el-table-column prop="issuingDate" label="年份" width="100" />
        <el-table-column prop="country" label="国家" width="120">
          <template #default="{ row }">
            {{ appLang.country(row.country) }}
          </template>
        </el-table-column>
        <el-table-column prop="performers" label="演员" width="260">
          <template #default="{ row }">
            <div class="tag-container">
              <el-tag v-for="performer in row.performers" :key="performer.id" size="small">{{ performer.name }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="tags" label="标签" width="300">
          <template #default="{ row }">
            <div class="tag-container">
              <el-tag v-for="tag in row.tags" :key="tag.id" size="small">{{ tag.name }}</el-tag>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <resourceFormDrawer ref="resourceFormDrawerRef" @success="updateResourceSuccessHandle"></resourceFormDrawer>
      <resourceSetTagBatchDialog ref="resourceSetTagBatchDialogRef" @success="updateResourceSuccessHandle">
      </resourceSetTagBatchDialog>
      <resourceSetPerformerBatchDialog ref="resourceSetPerformerBatchDialogRef" @success="updateResourceSuccessHandle">
      </resourceSetPerformerBatchDialog>
      <resourceSetStarsBatchDialog ref="resourceSetStarsBatchDialogRef" @success="updateResourceSuccessHandle">
      </resourceSetStarsBatchDialog>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { ArrowDown, Close, Grid, List, Operation, Picture } from '@element-plus/icons-vue';
import type { I_resource } from '@/dataType/resource.dataType';
import resourceFormDrawer from '@/components/resource/resourceFormDrawer.vue'
import resourceSetTagBatchDialog from '../resource/resourceSetTagBatchDialog.vue';
import resourceSetPerformerBatchDialog from '../resource/resourceSetPerformerBatchDialog.vue';
import resourceSetStarsBatchDialog from '../resource/resourceSetStarsBatchDialog.vue';
import { getResourceCoverPoster } from '@/common/photo';
import { playResource, playOpenResourceFolder } from '@/common/play'
import { resourceDelete, resourceBatchDelete } from '@/common/resource'
import { computed, nextTick, provide, ref, watch, type PropType } from 'vue';
import contentList from './contentList.vue';
import { resourceSelectionKey } from './resourceSelection';
import { ElMessage, ElMessageBox, type ElTable } from 'element-plus';
import { AppLang } from '@/language/app.lang'
import { videoTranscodeServer } from '@/server/videoTranscode.server';
const appLang = AppLang()

const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
})
const emits = defineEmits(['updateData']);

const tableRef = ref<InstanceType<typeof ElTable>>()
const resourceFormDrawerRef = ref<InstanceType<typeof resourceFormDrawer>>()
const resourceSetTagBatchDialogRef = ref<InstanceType<typeof resourceSetTagBatchDialog>>()
const resourceSetPerformerBatchDialogRef = ref<InstanceType<typeof resourceSetPerformerBatchDialog>>()
const resourceSetStarsBatchDialogRef = ref<InstanceType<typeof resourceSetStarsBatchDialog>>()
const layoutRef = ref<InstanceType<typeof contentList>>();
defineExpose({ change: () => { layoutRef.value?.change(); tableRef.value?.setScrollTop(0); } });
const addingTranscode = ref(false);
const managementView = ref<'layout' | 'table'>('layout');
const selectedResources = ref<I_resource[]>([]);
const selectionPreviewVisible = ref(false);
const removeSelection = (id: string) => {
  selectedResources.value = selectedResources.value.filter(resource => resource.id !== id);
};
watch(() => selectedResources.value.length, count => {
  if (!count) selectionPreviewVisible.value = false;
});
const selectedIds = computed(() => new Set(selectedResources.value.map(resource => resource.id)));
const selectionRowClass = ({ row }: { row: I_resource }) =>
  selectedIds.value.has(row.id) ? 'management-selected-row' : '';
const allSelected = computed(() => props.dataList.length > 0 && props.dataList.every(resource => selectedIds.value.has(resource.id)));
const partiallySelected = computed(() => selectedResources.value.length > 0 && !allSelected.value);
let syncingTable = false;
let tableSyncVersion = 0;

provide(resourceSelectionKey, {
  active: computed(() => managementView.value === 'layout'),
  selectedIds,
  toggle(resource) {
    selectedResources.value = selectedIds.value.has(resource.id)
      ? selectedResources.value.filter(item => item.id !== resource.id)
      : [...selectedResources.value, resource];
  },
});

const clearSelection = () => { selectedResources.value = []; };
const toggleAll = () => { selectedResources.value = allSelected.value ? [] : [...props.dataList]; };
const tableSelectionChange = (rows: I_resource[]) => {
  if (!syncingTable) selectedResources.value = rows;
};
const syncTableSelection = async () => {
  const version = ++tableSyncVersion;
  syncingTable = true;
  await nextTick();
  // 切换视图和刷新数据可能同时排队，只由最新任务恢复勾选状态。
  if (version !== tableSyncVersion) return;
  try {
    tableRef.value?.clearSelection();
    props.dataList.forEach(resource => {
      if (selectedIds.value.has(resource.id)) tableRef.value?.toggleRowSelection(resource, true);
    });
  } finally {
    syncingTable = false;
  }
};
watch([selectedResources, managementView], syncTableSelection);
// 本页选择：翻页或筛选移除不可见项，刷新后使用最新资源数据。
watch(() => props.dataList, resources => {
  selectedResources.value = resources.filter(resource => selectedIds.value.has(resource.id));
});

const playResourceHandle = (item: I_resource) => {
  playResource(item)
}
const editResourceHandle = (item: I_resource) => {
  resourceFormDrawerRef.value?.open('edit', item)
}
const updateResourceSuccessHandle = (data: I_resource) => {
  emits('updateData', data)
}
const resourceDeleteHandle = (item: I_resource) => {
  resourceDelete(item, () => {
    emits('updateData')
  })
}

const batchDeleteHandle = () => {
  const resources = [...selectedResources.value];
  resourceBatchDelete(resources, () => {
    emits('updateData')
  })
}
const batchAddTagHandle = () => {
  const resources = [...selectedResources.value];
  if (resources.length == 0) {
    ElMessage.error('请选择要添加标签的资源');
  } else {
    resourceSetTagBatchDialogRef.value?.open(resources, 'add')
  }
}
const batchDeleteTagHandle = () => {
  const resources = [...selectedResources.value];
  if (resources.length == 0) {
    ElMessage.error('请选择要删除标签的资源');
  } else {
    resourceSetTagBatchDialogRef.value?.open(resources, 'remove')
  }
}

const batchAddPerformerHandle = () => {
  const resources = [...selectedResources.value];
  if (resources.length == 0) {
    ElMessage.error('请选择要添加演员的资源');
  } else {
    resourceSetPerformerBatchDialogRef.value?.open(resources)
  }
}

const batchSetStarsHandle = () => {
  const resources = [...selectedResources.value];
  if (resources.length == 0) {
    ElMessage.error('请选择要设置评星的资源');
  } else {
    resourceSetStarsBatchDialogRef.value?.open(resources)
  }
}

const batchAddTranscodeHandle = async () => {
  if (addingTranscode.value) return;
  const resources = [...selectedResources.value];
  if (resources.length === 0) {
    ElMessage.error('请选择要加入转码与剪辑列表的资源');
    return;
  }
  addingTranscode.value = true;
  try {
    await ElMessageBox.confirm(
      `将选中的 ${resources.length} 个资源下的全部视频加入转码与剪辑列表，重复项和不存在的文件会自动跳过。`,
      '批量加入转码与剪辑列表',
      { type: 'warning' },
    );
    const result = await videoTranscodeServer.add({
      resourceIds: resources.map(item => item.id),
    });
    if (!result.status) {
      ElMessage.error(result.msg || '批量加入转码与剪辑列表失败');
      return;
    }
    const skipped = result.data.skippedDuplicate + result.data.skippedMissing;
    if (result.data.added > 0) {
      ElMessage.success(`已加入 ${result.data.added} 个视频，跳过 ${skipped} 个`);
    } else if (skipped > 0) {
      ElMessage.info(`没有新增视频，已跳过 ${skipped} 个重复或不存在的文件`);
    } else {
      ElMessage.warning('所选资源下没有可加入的视频');
    }
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error('批量加入转码与剪辑列表失败，请稍后重试');
    }
  } finally {
    addingTranscode.value = false;
  }
}

</script>
<style lang="scss" scoped>
.selection-preview {
  color: var(--el-text-color-primary);
  &__header {
    display: flex; align-items: center; justify-content: space-between;
    margin: 0 14px; padding: 13px 0 11px; border-bottom: 1px solid var(--el-border-color-lighter);
    strong { font-size: 13px; font-weight: 600; }
    > span { font-size: 12px; color: var(--el-text-color-secondary); }
  }
  &__list { list-style: none; margin: 0; padding: 6px; }
  &__item {
    display: flex; align-items: center; gap: 10px; padding: 8px; border-radius: 8px;
    &:hover { background: var(--el-fill-color-light); }
  }
  &__cover {
    flex: 0 0 36px; width: 36px; height: 48px; border-radius: 4px; background: var(--el-fill-color);
    :deep(.el-image__error) { color: var(--el-text-color-placeholder); font-size: 18px; }
  }
  &__info { flex: 1; min-width: 0; display: grid; gap: 4px; }
  &__title { font-size: 13px; line-height: 19px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  &__subtitle { color: var(--el-text-color-secondary); font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  &__remove { flex-shrink: 0; width: 28px; height: 28px; margin: 0; border-radius: 6px; color: var(--el-text-color-secondary); }
}
:deep(.management-selected-row) {
  --el-table-tr-bg-color: color-mix(in srgb, var(--el-color-primary) 12%, var(--el-bg-color));
  --el-table-row-hover-bg-color: color-mix(in srgb, var(--el-color-primary) 18%, var(--el-bg-color));
}
:deep(.management-selected-row > td:first-child) { box-shadow: inset 2px 0 var(--el-color-primary); }
.content-list-admin {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;

  .management-toolbar {
    flex-shrink: 0;
    margin: 4px 6px 12px;
    padding: 14px 16px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 12px;
    background: var(--el-bg-color-overlay);
    box-shadow: 0 3px 14px rgba(0, 0, 0, .04);
  }

  .management-toolbar__header, .management-toolbar__actions, .management-selection, .batch-actions {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 10px;
  }
  .management-toolbar__header { justify-content: space-between; gap: 12px; }
  .management-title { display: flex; align-items: center; gap: 10px; color: var(--el-text-color-primary); }
  .management-title strong { font-size: 14px; }
  .management-title__icon {
    display: grid; place-items: center; width: 34px; height: 34px; border-radius: 10px;
    background: var(--el-color-primary-light-9); color: var(--el-color-primary); font-size: 18px;
  }
  .management-hint { display: block; margin-top: 3px; color: var(--el-text-color-secondary); font-size: 12px; }
  .management-view-switch :deep(.el-radio-button__inner) { display: inline-flex; align-items: center; gap: 6px; }
  .management-toolbar__actions { margin-top: 12px; padding-top: 12px; border-top: 1px solid var(--el-border-color-lighter); }
  .management-selection { align-items: flex-end; gap: 12px; margin-right: auto; font-size: 14px; }
  .management-selection :deep(.el-checkbox),
  .management-selection :deep(.el-button) { height: 32px; margin: 0; font-size: 14px; line-height: 20px; }
  .management-selection :deep(.el-checkbox__label) { font-size: 14px; line-height: 20px; }
  .selection-count {
    display: inline-flex; align-items: center; gap: 3px; height: 32px; line-height: 20px;
    color: var(--el-text-color-regular); white-space: nowrap;
  }
  .selection-count strong { color: var(--el-color-primary); font-size: inherit; }
  .selection-preview-trigger { gap: 5px; }
  .selection-preview-trigger :deep(> span) { display: inline-flex; align-items: center; gap: 5px; }
  .batch-actions { gap: 8px; }
  .batch-actions :deep(.el-button) { margin-left: 0; border-radius: 8px; height: 32px; padding: 0 11px; }
  .batch-actions .batch-delete { margin-left: 6px; }

  .table-container {
    flex: 1;
    min-height: 0;
    overflow: hidden;

    .tag-container {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
    }
  }
}
</style>
