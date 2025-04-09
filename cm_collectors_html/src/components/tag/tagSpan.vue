<template>
  <span class="tag-span">美国美国</span>
</template>
<script setup lang="ts"></script>
<style lang="scss" scoped>
.tag-span {
  width: 4em;
  padding: 0.6em 0.9em;
  text-align: center;
  border-radius: 2px;
  cursor: pointer;
  user-select: none;
  color: #bdbcbc;
  background-color: #303131;
  font-size: 0.9em;
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
    box-shadow: 0 0 4px rgba(0, 141, 255, 0.3);
  }
}
</style>
