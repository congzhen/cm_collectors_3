<template>
  <el-select v-model="selectVal" clearable :style="{ width: props.width }"
    :placeholder="'请选择 ' + store.appStoreData.currentCupText" @change="changeHandle" @clear="handleClear">
    <el-option label="全部" value="ALL"></el-option>
    <el-option v-for="cup, key in dataset.cup" :key="key" :label="cup + ' - ' + store.appStoreData.currentCupText"
      :value="cup"></el-option>
    <el-option label="未定义" value="noCup"></el-option>
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
})
const emit = defineEmits(['change'])

const changeHandle = () => {
  emit('change', selectVal.value || '')
}
const handleClear = () => {
  selectVal.value = '';
}
</script>
