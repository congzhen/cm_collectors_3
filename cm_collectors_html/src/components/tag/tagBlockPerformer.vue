<template>
  <div class="tag-block-performer">
    <div class="tag-content">
      <tagSpan></tagSpan>
      <tagSpan></tagSpan>
    </div>
    <div class="tag-performer">
      <performerBlock class="tag-performer-item"></performerBlock>
      <performerBlock class="tag-performer-item"></performerBlock>
      <performerBlock class="tag-performer-item"></performerBlock>
      <performerBlock class="tag-performer-item"></performerBlock>
      <performerBlock class="tag-performer-item"></performerBlock>
      <performerBlock class="tag-performer-item"></performerBlock>
    </div>
  </div>
</template>
<script setup lang="ts">
import tagSpan from './tagSpan.vue'
import performerBlock from '../performer/performerBlock.vue'
</script>
<style lang="scss" scoped>
.tag-block-performer {
  .tag-content {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3em;
  }
  .tag-performer {
    padding: 0.5em 0;
    display: flex;
    flex-wrap: wrap;
    gap: 0.3em;
    .tag-performer-item {
      width: calc(25% - 0.9em);
      &:hover {
        background-color: var(--el-color-primary-light-9);
      }
    }
  }
}
</style>
