<template>
  <el-select v-model="selectVal" clearable :style="{ width: props.width }">
    <el-option label="中国" value="ALL"></el-option>
    <el-option label="日本" value="5"></el-option>
    <el-option label="韩国" value="4"></el-option>
    <el-option label="美国" value="3"></el-option>
    <el-option label="英国" value="2"></el-option>
    <el-option label="法国" value="1"></el-option>
    <el-option label="德国" value="1"></el-option>
    <el-option label="其他" value="noStars"></el-option>
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
