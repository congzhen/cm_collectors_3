<template>
  <span v-if="groupId" class="resource-group-mark" :style="colorStyle" :title="`${label}：本页标记相同的资源属于同一关联组`">
    <el-icon><Collection /></el-icon><span>{{ label }}</span>
  </span>
</template>
<script setup lang="ts">
import { computed, inject } from 'vue';
import { Collection } from '@element-plus/icons-vue';
import { RESOURCE_GROUP_MARK_COLORS, resourceSelectionKey } from './resourceSelection';
const props = defineProps<{ groupId?: string }>();
const selection = inject(resourceSelectionKey, undefined);
const label = computed(() => selection?.groupLabels?.value.get(props.groupId || '') || '已分组');
const colorStyle = computed(() => {
  const ids = [...(selection?.groupLabels?.value.keys() || [])];
  const index = Math.max(0, ids.indexOf(props.groupId || ''));
  const color = RESOURCE_GROUP_MARK_COLORS[index % RESOURCE_GROUP_MARK_COLORS.length];
  return { '--resource-group-bg': color.background, '--resource-group-text': color.text, '--resource-group-border': color.border };
});
</script>
<style scoped>
.resource-group-mark { display: inline-flex; align-items: center; gap: 4px; max-width: 100%; box-sizing: border-box; padding: 2px 5px; border: 1px solid var(--resource-group-border); border-radius: 5px; background: var(--resource-group-bg); color: var(--resource-group-text); font-size: 12px; line-height: 18px; white-space: nowrap; vertical-align: middle; }
.el-icon { flex-shrink: 0; font-size: 13px; }
.resource-group-mark > span { overflow: hidden; text-overflow: ellipsis; }
</style>
