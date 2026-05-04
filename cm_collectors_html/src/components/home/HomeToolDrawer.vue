<template>
  <aside class="home-side-panel">
    <div class="drawer-head">
      <div>
        <strong>{{ title }}</strong>
        <span>{{ subtitle }}</span>
      </div>
      <el-button text :icon="Close" @click="emits('close')" />
    </div>
    <slot></slot>
  </aside>
</template>

<script setup lang="ts">
import { Close } from '@element-plus/icons-vue'

defineProps({
  title: {
    type: String,
    default: '',
  },
  subtitle: {
    type: String,
    default: '',
  },
})

const emits = defineEmits<{
  close: []
}>()

</script>

<style lang="scss" scoped>
.home-side-panel {
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--home-border);
  border-radius: 8px;
  background: var(--home-panel-bg);
}

.drawer-head {
  min-height: 54px;
  padding: 9px 12px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  border-bottom: 1px solid var(--home-border);
  background: var(--home-panel-soft-bg);
  box-sizing: border-box;

  strong {
    display: block;
    color: var(--home-text);
    font-size: 15px;
    line-height: 20px;
  }

  span {
    display: block;
    max-width: 240px;
    color: var(--home-text-muted);
    line-height: 16px;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  :deep(.el-button) {
    color: var(--home-text-muted);
  }
}
</style>
