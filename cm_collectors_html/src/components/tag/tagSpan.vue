<template>
  <span class="tag-span" :style="{ width: width_C }">{{ props.title }}</span>
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
  }
})

const width_C = computed(() => {
  if (props.tagModeFixed) {
    return '4.8em';
  }
  return store.appStoreData.currentConfigApp.tagMode === 'fixed' ? 'calc(25% - 1.8em)' : 'auto';
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
  overflow: hidden;
  /*文本不换行*/
  white-space: nowrap;
  /*ellipsis:文本溢出显示省略号（...）*/
  text-overflow: ellipsis;

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
