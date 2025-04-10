<template>
  <div class="tag-block">
    <div class="tag-content">
      <tagSpan></tagSpan>
      <tagSpan></tagSpan>
      <tagSpan></tagSpan>
      <tagSpan></tagSpan>
      <tagSpan></tagSpan>
      <tagSpan></tagSpan>
      <tagSpan></tagSpan>
      <tagSpan></tagSpan>
      <tagSpan></tagSpan>
      <tagSpan></tagSpan>
      <tagSpan></tagSpan>
      <tagSpan></tagSpan>
      <tagSpan></tagSpan>
    </div>
  </div>
</template>
<script setup lang="ts">
import tagSpan from './tagSpan.vue'
</script>
<style lang="scss" scoped>
.tag-block {
  .tag-title {
    font-size: 1.2em;
    font-weight: 500;
    line-height: 2.5em;
  }
  .tag-content {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3em;
  }
}
</style>
