<template>
  <el-input v-model="searchVal" clearable :style="{ width: props.width }">
    <template #append>
      <el-button icon="Search" />
    </template>
  </el-input>
</template>
<script setup lang="ts">
const searchVal = defineModel({ type: String, default: '' })
const props = defineProps({
  width: {
    type: String,
    default: '100%',
  },
})
</script>
