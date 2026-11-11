<template>
  <el-select class="resource-group-cover-select" :model-value="modelValue === 'triptych' || modelValue === 'book' || modelValue === 'adaptive' ? modelValue : 'stack'" aria-label="关联组封面样式" @update:model-value="update">
    <template #prefix><el-icon><Collection /></el-icon></template>
    <el-option label="叠牌封面" value="stack" />
    <el-option label="三联切片" value="triptych" />
    <el-option label="侧翻书页" value="book" />
    <el-option label="自适应拼封" value="adaptive" />
    <template v-if="showSingleToggle" #footer>
      <div v-if="showSingleToggle" class="single-option" @click.stop @pointerdown.stop>
        <span title="只有一个资源时不显示悬停预览；点击封面仍可进入合集">单资源不展开</span>
        <el-switch :model-value="hideSingle === true" aria-label="单资源不展开" @update:model-value="emit('update:hideSingle', $event === true)" />
      </div>
    </template>
  </el-select>
</template>
<script setup lang="ts">
import { Collection } from '@element-plus/icons-vue';
defineProps<{ hideSingle?: boolean; showSingleToggle?: boolean; modelValue?: 'stack' | 'triptych' | 'book' | 'adaptive' }>();
const emit = defineEmits<{ 'update:hideSingle': [value: boolean]; 'update:modelValue': [value: 'stack' | 'triptych' | 'book' | 'adaptive'] }>();
const update = (value: string | number | boolean | undefined) => {
  if (value === 'stack' || value === 'triptych' || value === 'book' || value === 'adaptive') emit('update:modelValue', value);
};
</script>
<style scoped>
.resource-group-cover-select { width: 100%; min-width: 0; }
.resource-group-cover-select :deep(.el-select__prefix) { color: var(--el-text-color-secondary); }
.single-option { display: flex; align-items: center; justify-content: space-between; gap: 10px; min-width: 150px; font-size: 13px; color: var(--el-text-color-regular); }
.single-option > span { white-space: nowrap; }
.single-option .el-switch { flex-shrink: 0; }
</style>
