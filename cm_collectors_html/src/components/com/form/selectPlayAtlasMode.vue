<template>
  <el-select v-model="selectVal" :style="{ width: props.width }" placeholder="图集模式" @change="changeHandle"
    @clear="handleClear">
    <el-option v-for="item, index in dataset.playAtlasMode" :key="index" :label="item" :value="item"></el-option>
  </el-select>
</template>
<script setup lang="ts">
import dataset from '@/assets/dataset';
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
