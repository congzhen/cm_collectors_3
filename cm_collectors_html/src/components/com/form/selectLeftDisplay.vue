<template>
  <el-select v-model="selectVal" clearable :style="{ width: props.width }" placeholder="'左侧边栏显示项" @change="changeHandle"
    @clear="handleClear" :multiple="props.multiple">
    <el-option v-for="item, index in dataset.leftDisplay" :key="index" :label="appLang.attributeTags(item)"
      :value="item"></el-option>
  </el-select>
</template>
<script setup lang="ts">
import dataset from '@/assets/dataset';
import { appLang } from '@/language/app.lang'
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
