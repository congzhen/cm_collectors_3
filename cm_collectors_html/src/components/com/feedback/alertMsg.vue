<template>
  <p class="alert-msg" :style="{ color: color_C }">
    <slot />
  </p>
</template>
<script lang="ts" setup>
import { computed, type PropType } from 'vue'
const props = defineProps({
  color: {
    type: String as PropType<'primary' | 'success' | 'info' | 'warning' | 'error'>,
    default: 'info',
  },
})
const color_C = computed(() => {
  switch (props.color) {
    case 'primary':
      return '#409eff';
    case 'success':
      return '#67c23a';
    case 'info':
      return '#909399';
    case 'warning':
      return '#e6a23c';
    case 'error':
      return '#f56c6c';
  }
})
</script>
<style lang="scss" scoped>
.alert-msg {
  font-size: 0.9em;
  line-height: 1.3em;
  padding: 0.1em 0;
}
</style>
