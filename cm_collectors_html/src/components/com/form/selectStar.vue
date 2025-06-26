<template>
  <el-select v-model="selectVal" clearable :style="{ width: props.width }" placeholder="请选择星级" @change="changeHandle"
    @clear="handleClear">
    <el-option label="全部" value=""></el-option>
    <el-option label="★★★★★" value="5"></el-option>
    <el-option label="★★★★" value="4"></el-option>
    <el-option label="★★★" value="3"></el-option>
    <el-option label="★★" value="2"></el-option>
    <el-option label="★" value="1"></el-option>
    <el-option label="未评星" value="0"></el-option>
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
const emit = defineEmits(['change'])

const changeHandle = () => {
  emit('change', selectVal.value || '')
}
const handleClear = () => {
  selectVal.value = '';
}
</script>
