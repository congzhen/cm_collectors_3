<template>
  <el-select v-model="selectVal" clearable :style="{ width: props.width }">
    <el-option label="中国演员数据集" value="ALL"></el-option>
    <el-option label="日本演员数据集" value="5"></el-option>
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
