<template>
  <el-select v-model="selectVal" clearable :style="{ width: props.width }">
    <el-option label="8K" value="ALL"></el-option>
    <el-option label="4K" value="5"></el-option>
    <el-option label="2K" value="4"></el-option>
    <el-option label="1080P" value="3"></el-option>
    <el-option label="720P" value="2"></el-option>
    <el-option label="高清" value="1"></el-option>
    <el-option label="标清" value="1"></el-option>
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
