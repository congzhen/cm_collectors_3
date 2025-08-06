<template>
  <el-select v-model="selectVal" :clearable="props.clearable" :style="{ width: props.width }" placeholder="标签显示模式"
    @change="changeHandle" @clear="handleClear">
    <el-option v-for="item, index in dataset.tagMode" :key="index" :label="$t(`tagMode.${item}`)"
      :value="item"></el-option>
  </el-select>
</template>
<script setup lang="ts">
import dataset from '@/assets/dataset';
import { appStoreData } from '@/storeData/app.storeData';
const store = {
  appStoreData: appStoreData(),
}
const selectVal = defineModel({ type: String, default: '' })
const props = defineProps({
  width: {
    type: String,
    default: '100%',
  },
  clearable: {
    type: Boolean,
    default: false
  }
})
const emit = defineEmits(['change'])

const changeHandle = () => {
  emit('change', selectVal.value || '')
}
const handleClear = () => {
  selectVal.value = '';
}
</script>
