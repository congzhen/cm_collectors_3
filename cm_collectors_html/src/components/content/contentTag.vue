<template>
  <div class="content-tag" :style="{
    '--current-bg': props.bgColor,
    '--solid-bg': solidBgColor,
    color: props.color,
    border: '1px solid ' + props.bgColor,
    borderRadius: props.borderRadius,
    fontSize: props.fontSize,
  }">
    {{ props.title }}
  </div>
</template>
<script lang="ts" setup>
import { computed } from 'vue';

const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  bgColor: {
    type: String,
    default: 'rgba(244, 54, 16, 0.4)',
  },
  color: {
    type: String,
    default: '#FFFFFF',
  },
  borderRadius: {
    type: String,
    default: '0.3em',
  },
  fontSize: {
    type: String,
    default: '12px',
  },
})

// 计算不带透明度的背景色 (将 rgba 转为 rgb)
const solidBgColor = computed(() => {
  const match = props.bgColor.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*[\d.]+)?\)/);
  if (match) {
    return `rgb(${match[1]}, ${match[2]}, ${match[3]})`;
  }
  // 如果不是 rgba 格式，直接返回原色（假设已是不透明或无需处理）
  return props.bgColor;
});
</script>
<style lang="scss" scoped>
.content-tag {
  padding: 0 5px;
  font-size: 12px;
  line-height: 1.3em;
  // 使用 CSS 变量作为背景色，以便 hover 时覆盖
  background-color: var(--current-bg); 
  transition: background-color 0.3s ease; /* 添加过渡效果使变化更平滑 */

  &:hover {
    // 鼠标经过时，使用不透明的背景色
    background-color: var(--solid-bg) !important;
    // 同时也更新边框颜色为不透明，保持视觉一致（可选）
    border-color: var(--solid-bg) !important;
  }
}
</style>