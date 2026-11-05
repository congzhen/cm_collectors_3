<template>
  <div class="category-quick-settings" :class="{ 'is-horizontal': horizontal }">
    <el-form :disabled="saving || disabled">
      <CategoryBrowseFields
        ref="fields"
        :config="draft"
        :files-bases-id="store.currentFilesBases.id || ''"
        always-show
      />
    </el-form>
    <el-button
      class="category-apply"
      size="small"
      plain
      title="应用分类浏览"
      aria-label="应用分类浏览"
      :loading="saving"
      :disabled="disabled"
      @click="save"
      >应用</el-button
    >
  </div>
</template>
<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { appStoreData } from '@/storeData/app.storeData'
import { categoryBrowseServer, categorySettings } from '@/server/categoryBrowse.server'
import CategoryBrowseFields from './CategoryBrowseFields.vue'
import { eventBus } from '@/main'
const store = appStoreData()
const props = defineProps<{ disabled?: boolean; horizontal?: boolean }>()
const emit = defineEmits<{ saving: [value: boolean] }>()
const draft = ref(categorySettings(store.currentConfigApp))
const saving = ref(false)
const fields = ref<InstanceType<typeof CategoryBrowseFields>>()
let version = 0
watch(saving, (value) => emit('saving', value), { flush: 'sync' })
watch(
  () => [
    store.currentFilesBases.id,
    store.currentConfigApp,
    JSON.stringify(categorySettings(store.currentConfigApp)),
  ],
  () => {
    version++
    saving.value = false
    draft.value = categorySettings(store.currentConfigApp)
  },
)
async function save() {
  if (saving.value || props.disabled) return
  if (fields.value?.validate() === false) return
  const current = ++version,
    library = store.currentFilesBases.id,
    config = store.currentConfigApp
  const snapshot = { ...draft.value }
  const isCurrent = () =>
    current === version &&
    library === store.currentFilesBases.id &&
    config === store.currentConfigApp
  saving.value = true
  try {
    const result = await categoryBrowseServer.settings(library, snapshot)
    if (!isCurrent()) return
    if (!result.status) throw new Error(result.msg || '分类浏览保存失败')
    Object.assign(config, snapshot)
    eventBus.emit('category-browse-applied')
    ElMessage.success('分类浏览已应用')
  } catch (error) {
    if (isCurrent()) ElMessage.error(error instanceof Error ? error.message : '分类浏览保存失败')
  } finally {
    if (current === version) saving.value = false
  }
}
onBeforeUnmount(() => {
  version++
  emit('saving', false)
})
</script>
<style scoped>
.category-quick-settings {
  border-top: 1px solid var(--el-border-color-lighter);
  margin-top: 12px;
  padding-top: 6px;
}
.category-apply {
  --el-button-text-color: var(--adjuster-accent, #159a9a);
  --el-button-bg-color: var(--adjuster-accent-soft, rgba(21, 154, 154, 0.11));
  --el-button-border-color: transparent;
  --el-button-hover-text-color: var(--adjuster-accent, #159a9a);
  --el-button-hover-bg-color: var(--adjuster-accent-soft, rgba(21, 154, 154, 0.16));
  --el-button-hover-border-color: var(--adjuster-accent, #159a9a);
  --el-button-active-text-color: var(--adjuster-accent, #159a9a);
  --el-button-active-border-color: var(--adjuster-accent, #159a9a);
  height: 28px;
  padding: 0 14px;
  border-radius: 6px;
  font-weight: 500;
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
.is-horizontal > .category-apply {
  position: absolute;
  right: 0;
  top: 2px;
  margin: 0;
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
  padding-right: 90px;
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
