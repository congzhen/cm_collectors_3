<template>
  <el-select v-model="selectVal" clearable :style="{ width: props.width }">
    <el-option label="全部" value="ALL"></el-option>
    <el-option label="未定义" value="noCup"></el-option>
  </el-select>
</template>
<script setup lang="ts">
const selectVal = defineModel({ type: String, default: '' })
const props = defineProps({
  width: {
    type: String,
    default: '100%',
  },
})
</script>
