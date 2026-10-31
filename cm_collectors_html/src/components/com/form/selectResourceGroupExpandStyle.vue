<template>
  <el-select class="resource-group-style-select" :model-value="modelValue === 'strip' || modelValue === 'fold' || modelValue === 'spine' || modelValue === 'magazine' ? modelValue : 'fan'" aria-label="关联组展开样式" @update:model-value="update">
    <template #prefix><el-icon><Collection /></el-icon></template>
    <el-option label="扇形手牌" value="fan" />
    <el-option label="横向封面带" value="strip" />
    <el-option label="封面折屏" value="fold" />
    <el-option label="纵向封面书脊" value="spine" />
    <el-option label="杂志拼版" value="magazine" />
  </el-select>
</template>
<script setup lang="ts">
import { Collection } from '@element-plus/icons-vue';
defineProps<{ modelValue?: 'fan' | 'strip' | 'fold' | 'spine' | 'magazine' }>();
const emit = defineEmits<{ 'update:modelValue': [value: 'fan' | 'strip' | 'fold' | 'spine' | 'magazine'] }>();
const update = (value: string | number | boolean | undefined) => {
  if (value === 'fan' || value === 'strip' || value === 'fold' || value === 'spine' || value === 'magazine') emit('update:modelValue', value);
};
</script>
<style scoped>
.resource-group-style-select { width: 100%; min-width: 0; }
.resource-group-style-select :deep(.el-select__prefix) { color: var(--el-text-color-secondary); }
</style>
