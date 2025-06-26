<template>
  <div class="tag-block-stars">
    <div class="tag-content">
      <tagSpan title="全部"></tagSpan>
      <tagSpan title="未评分"></tagSpan>
    </div>
    <ul class="tag-stars">
      <li v-for="(star, index) in starData" :key="index">
        <el-rate v-model="starData[index]" disabled />
      </li>
    </ul>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import tagSpan from './tagSpan.vue'
const starData = ref([1, 2, 3, 4, 5])
</script>
<style lang="scss" scoped>
.tag-block-stars {
  display: flex;

  .tag-content {
    display: flex;
    flex-wrap: wrap;
    flex-direction: column;
    gap: 0.3em;
  }

  .tag-stars {
    padding-left: 0.3em;
    list-style-type: none;
    cursor: pointer;

    li {
      background-color: #303131;
      padding: 0.4em 0.6em;
      margin-bottom: 0.3em;
      border-radius: 3px;
      height: 1em;
      line-height: 1em;
      overflow: hidden;

      &:hover {
        background-color: var(--el-color-primary-light-9);
      }

      .el-rate {
        height: 1em;
        cursor: pointer;

        :deep(.el-rate__item) {
          cursor: pointer;
        }
      }
    }
  }
}
</style>
