<template>
  <div class="category-quick-settings" :class="{ 'is-horizontal': horizontal }">
    <el-form :disabled="disabled">
      <CategoryBrowseFields
        ref="fields"
        :config="draft"
        :files-bases-id="store.currentFilesBases.id || ''"
        @ready="applyPreview"
        always-show
      />
    </el-form>

  </div>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue'
import { appStoreData } from '@/storeData/app.storeData'
import { categorySettings } from '@/server/categoryBrowse.server'
import CategoryBrowseFields from './CategoryBrowseFields.vue'
import { eventBus } from '@/main'
const store = appStoreData()
const props = defineProps<{ disabled?: boolean; horizontal?: boolean }>()
const draft = ref(categorySettings(store.currentConfigApp))
const fields = ref<InstanceType<typeof CategoryBrowseFields>>()
// Match the other quick display controls: update the in-memory preview only.
// The enclosing Save button persists the complete currentConfigApp.
function applyPreview() {
  if (props.disabled || !fields.value) return
  if (JSON.stringify(draft.value) === JSON.stringify(categorySettings(store.currentConfigApp))) return
  if (!fields.value.validate()) return
  Object.assign(store.currentConfigApp, draft.value)
  eventBus.emit('category-browse-applied')
}
watch(
  () => [store.currentFilesBases.id, store.currentConfigApp, JSON.stringify(categorySettings(store.currentConfigApp))],
  () => { draft.value = categorySettings(store.currentConfigApp) },
)
function validate() {
  if (!fields.value?.validate()) return false
  applyPreview()
  return true
}
defineExpose({ validate })
watch(draft, applyPreview, { deep: true, flush: 'post' })
watch(() => props.disabled, disabled => { if (!disabled) applyPreview() })
</script>
<style scoped>
.category-quick-settings {
  border-top: 1px solid var(--el-border-color-lighter);
  margin-top: 12px;
  padding-top: 6px;
}
.is-horizontal {
  position: relative;
  border: 0;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
  height: 100%;
  box-sizing: border-box;
}
.is-horizontal :deep(.category-browse-fields) {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px 16px;
}
.is-horizontal :deep(.category-field) {
  margin: 0;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 6px;
}
.is-horizontal :deep(.category-field:first-child) {
  grid-column: 1 / -1;
  flex-direction: row;
  align-items: center;
  justify-content: flex-start;
  gap: 14px;
}
.is-horizontal :deep(.category-field:first-child > span) {
  font-size: 16px;
  font-weight: 650;
  color: var(--el-text-color-primary);
}
.is-horizontal :deep(.category-field .el-select),
.is-horizontal :deep(.category-field .el-input-number) {
  width: 100%;
}
.is-horizontal :deep(p) {
  grid-column: 1 / -1;
  margin: 0;
}
@media (max-width: 1100px) {
  .is-horizontal :deep(.category-browse-fields) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 620px) {
  .is-horizontal :deep(.category-browse-fields) {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
