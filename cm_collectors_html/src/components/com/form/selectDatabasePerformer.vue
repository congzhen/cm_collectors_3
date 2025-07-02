<template>
  <el-select v-model="selectVal" clearable :style="{ width: props.width }" placeholder="演员集" @change="changeHandle"
    @clear="handleClear" :multiple="props.multiple">
    <el-option v-for="item, index in dataList_C" :key="index" :label="item.name" :value="item.id"></el-option>
  </el-select>
</template>
<script setup lang="ts">
import { computed } from 'vue';
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';
const store = {
  performerBasesStoreData: performerBasesStoreData(),
}
const selectVal = defineModel<string | string[]>({ type: [String, Array], default: "" as string | string[] });
const props = defineProps({
  width: {
    type: String,
    default: '100%',
  },
  multiple: {
    type: Boolean,
    default: false
  },
  filesBasesId: {
    type: String,
    default: '',
  },
})
const emit = defineEmits(['change'])

const dataList_C = computed(() => {
  if (props.filesBasesId == '') {
    return store.performerBasesStoreData.performerBases;
  } else {
    return store.performerBasesStoreData.listByFilesBasesId(props.filesBasesId)
  }
})

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
