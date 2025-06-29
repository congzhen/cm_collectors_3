<template>
  <el-select v-model="selectVal" clearable :style="{ width: props.width }" placeholder="'国家" @change="changeHandle"
    @clear="handleClear" :multiple="props.multiple">
    <el-option-group v-for="(countryList, key) in dataset.country" :key="key" :label="key">
      <el-option v-for="item, index in countryList" :key="index" :label="item" :value="item"></el-option>
    </el-option-group>
  </el-select>
</template>
<script setup lang="ts">
import dataset from '@/assets/dataset';
const selectVal = defineModel<string | string[]>({ type: [String, Array], default: "" as string | string[] });
const props = defineProps({
  width: {
    type: String,
    default: '100%',
  },
  multiple: {
    type: Boolean,
    default: false
  }
})
const emit = defineEmits(['change'])

const changeHandle = () => {
  emit('change', selectVal.value || '')
}
const handleClear = () => {
  if (props.multiple) {
    selectVal.value = [];
  } else {
    selectVal.value = '';
  }
}
</script>
