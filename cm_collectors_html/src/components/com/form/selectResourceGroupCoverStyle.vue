<template>
  <el-select class="resource-group-cover-select" :model-value="modelValue === 'triptych' || modelValue === 'book' || modelValue === 'adaptive' ? modelValue : 'stack'" aria-label="关联组封面样式" @update:model-value="update">
    <template #prefix><el-icon><Collection /></el-icon></template>
    <el-option label="叠牌封面" value="stack" />
    <el-option label="三联切片" value="triptych" />
    <el-option label="侧翻书页" value="book" />
    <el-option label="自适应拼封" value="adaptive" />
  </el-select>
</template>
<script setup lang="ts">
import { Collection } from '@element-plus/icons-vue';
defineProps<{ modelValue?: 'stack' | 'triptych' | 'book' | 'adaptive' }>();
const emit = defineEmits<{ 'update:modelValue': [value: 'stack' | 'triptych' | 'book' | 'adaptive'] }>();
const update = (value: string | number | boolean | undefined) => {
  if (value === 'stack' || value === 'triptych' || value === 'book' || value === 'adaptive') emit('update:modelValue', value);
};
</script>
<style scoped>
.resource-group-cover-select { width: 100%; min-width: 0; }
.resource-group-cover-select :deep(.el-select__prefix) { color: var(--el-text-color-secondary); }
</style>
