<template>
  <el-collapse-item title="Consistency">
    <template #title>
      <tagLogic text="国家"></tagLogic>
    </template>
    <tagBlock></tagBlock>
  </el-collapse-item>
</template>
<script setup lang="ts">
import tagBlock from './tagBlock.vue'
import tagLogic from './tagLogic.vue'
</script>
<style lang="scss" scoped>
.title-label {
  padding-left: 0.5em;
  font-size: 1em;
  font-weight: 600;
}
</style>
