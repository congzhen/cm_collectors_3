<template>
  <el-select class="resource-group-style-select" :model-value="modelValue === 'strip' || modelValue === 'fold' || modelValue === 'spine' || modelValue === 'magazine' ? modelValue : 'fan'" aria-label="关联组展开样式" @update:model-value="update">
    <template #prefix><el-icon><Collection /></el-icon></template>
    <el-option label="扇形手牌" value="fan" />
    <el-option label="横向封面带" value="strip" />
    <el-option label="封面折屏" value="fold" />
    <el-option label="纵向封面书脊" value="spine" />
    <el-option label="杂志拼版" value="magazine" />
    <template #footer>
      <div class="ratio-option" @click.stop @pointerdown.stop>
        <span title="展开时按封面原始比例调整显示；具体效果随展开样式变化">适应封面比例</span>
        <el-switch :model-value="adaptCoverRatio === true" aria-label="适应封面比例" @update:model-value="emit('update:adaptCoverRatio', $event === true)" />
      </div>
    </template>
  </el-select>
</template>
<script setup lang="ts">
import { Collection } from '@element-plus/icons-vue';
defineProps<{ adaptCoverRatio?: boolean; modelValue?: 'fan' | 'strip' | 'fold' | 'spine' | 'magazine' }>();
const emit = defineEmits<{ 'update:adaptCoverRatio': [value: boolean]; 'update:modelValue': [value: 'fan' | 'strip' | 'fold' | 'spine' | 'magazine'] }>();
const update = (value: string | number | boolean | undefined) => {
  if (value === 'fan' || value === 'strip' || value === 'fold' || value === 'spine' || value === 'magazine') emit('update:modelValue', value);
};
</script>
<style scoped>
.resource-group-style-select { width: 100%; min-width: 0; }
.resource-group-style-select :deep(.el-select__prefix) { color: var(--el-text-color-secondary); }
.ratio-option { display: flex; align-items: center; justify-content: space-between; gap: 10px; min-width: 150px; font-size: 13px; color: var(--el-text-color-regular); }
.ratio-option > span { white-space: nowrap; }
.ratio-option .el-switch { flex-shrink: 0; }
</style>
