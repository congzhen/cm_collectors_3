<template>
  <span class="tag-span" :style="{ width: width_C }" :title="tooltip_C">
    <span class="tag-span-title">{{ props.title }}</span>
    <span v-if="showResourceCount_C" class="tag-resource-count">{{ displayResourceCount_C }}</span>
  </span>
</template>
<script setup lang="ts">
import { appStoreData } from '@/storeData/app.storeData';
import { computed } from 'vue';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  tagModeFixed: {
    type: Boolean,
    default: false,
  },
  resourceCount: {
    type: Number,
    default: undefined,
  }
})

const showResourceCount_C = computed(() => {
  return props.resourceCount !== undefined
    && store.appStoreData.currentConfigApp.showCustomTagResourceCount !== false;
})

const displayResourceCount_C = computed(() => {
  if (props.resourceCount === undefined) return '';
  if (props.resourceCount > 999) return '999+';
  return props.resourceCount.toString();
})

const tooltip_C = computed(() => {
  if (!showResourceCount_C.value || props.resourceCount === undefined) return props.title;
  return `${props.title}：当前库共有 ${props.resourceCount} 个资源使用此标签`;
})

const width_C = computed(() => {
  if (props.tagModeFixed) {
    return '4.8em';
  }
  let rowNumRatio = 25;
  try {
    rowNumRatio = 100 / store.appStoreData.currentConfigApp.tagFixedModeRowShowNum;
  } catch {
    rowNumRatio = 25;
  }
  return store.appStoreData.currentConfigApp.tagMode === 'fixed' ? `calc(${rowNumRatio}% - 1.8em)` : 'auto';
})


</script>
<style lang="scss" scoped>
.tag-span {
  /*width: 4.8em;*/
  padding: 0.6em 0.7em;
  text-align: center;
  border-radius: 2px;
  cursor: pointer;
  user-select: none;
  color: #bdbcbc;
  background-color: #303131;
  font-size: 0.85em;
  line-height: 0.9em;
  /*溢出的部分隐藏*/
  overflow: visible;
  /*文本不换行*/
  white-space: nowrap;
  /*ellipsis:文本溢出显示省略号（...）*/
  text-overflow: ellipsis;
  position: relative;

  .tag-span-title {
    display: block;
    min-width: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .tag-resource-count {
    position: absolute;
    top: -5px;
    right: -4px;
    z-index: 3;
    min-width: 13px;
    height: 12px;
    padding: 0 2px;
    box-sizing: border-box;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 3px;
    color: var(--tag-count-text, #67d4d7);
    background-color: var(--tag-count-bg, rgba(55, 198, 202, 0.14));
    font-size: 8px;
    font-weight: 650;
    line-height: 12px;
    pointer-events: none;
  }

  /* 新增过渡动画 */
  transition:
    background-color 0.2s ease,
    box-shadow 0.2s ease;

  /* 鼠标悬停时的高亮效果 */
  &:hover {
    color: var(--el-color-primary);
    background-color: var(--el-color-primary-light-9);
  }
}
</style>
