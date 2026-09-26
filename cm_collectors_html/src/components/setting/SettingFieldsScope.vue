<template>
  <div class="setting-fields-scope" :class="{ 'is-readonly': disabled }">
    <div v-if="disabled" class="readonly-indicator">
      <el-icon aria-hidden="true"><Lock /></el-icon>
      <span>公共配置 · 只读</span>
    </div>
    <div class="fields-content"><slot /></div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, provide, reactive, toRefs } from 'vue'
import { ElIcon, formContextKey } from 'element-plus'
import { Lock } from '@element-plus/icons-vue'

const props = defineProps<{ disabled?: boolean }>()
const form = inject(formContextKey, undefined)
// 只影响本组控件，继承外层表单的布局和校验，不创建嵌套 HTML 表单。
if (form) {
  provide(formContextKey, reactive({
    ...toRefs(form),
    disabled: computed(() => !!props.disabled || form.disabled),
  }))
}
</script>

<style scoped>
.is-readonly {
  --readonly-label: #9098a5;
  --readonly-value: #858e9c;
  --readonly-border: #dce1e8;
  --readonly-surface: #f7f8fa;
  --readonly-badge: #e9edf2;
  margin-bottom: 18px;
  padding: 12px 14px 0;
  border: 1px dashed var(--readonly-border);
  border-radius: 8px;
  background: var(--readonly-surface);
}

:global(html.dark .setting-fields-scope.is-readonly) {
  --readonly-label: #858b95;
  --readonly-value: #9aa1ac;
  --readonly-border: #494e57;
  --readonly-surface: rgba(255, 255, 255, 0.045);
  --readonly-badge: #363b43;
}

.readonly-indicator {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 5px;
  width: fit-content;
  margin: 0 0 14px auto;
  padding: 3px 9px;
  border-radius: 4px;
  background: var(--readonly-badge);
  color: var(--el-text-color-regular);
  font-size: 12px;
  line-height: 1.5;
}

.is-readonly .fields-content {
  filter: grayscale(1);
  cursor: not-allowed;
}

.is-readonly :deep(.el-form-item__label) {
  color: var(--readonly-label);
  font-weight: normal;
  text-shadow: none;
}

.is-readonly :deep(.el-form-item__content) {
  text-shadow: none;
}

.is-readonly :deep(.el-input__wrapper),
.is-readonly :deep(.el-select__wrapper) {
  background: transparent;
  box-shadow: 0 0 0 1px var(--readonly-border) inset;
}

.is-readonly :deep(.el-input__inner),
.is-readonly :deep(.el-select__selected-item),
.is-readonly :deep(.el-checkbox__label) {
  color: var(--readonly-value);
  -webkit-text-fill-color: var(--readonly-value);
}

.is-readonly :deep(.el-switch),
.is-readonly :deep(.el-checkbox__input) {
  opacity: 0.5;
}
</style>
