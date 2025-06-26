<template>
  <el-input v-model="searchVal" clearable :placeholder="props.placeholder" :style="{ width: props.width }"
    @input="changeHandle">
    <template #append>
      <el-button icon="Search" />
    </template>
  </el-input>
</template>
<script setup lang="ts">
import { debounce } from '@/assets/debounce'

const searchVal = defineModel({ type: String, default: '' })
const props = defineProps({
  width: {
    type: String,
    default: '100%',
  },
  placeholder: {
    type: String,
    default: '',
  }
})
const emit = defineEmits(['change'])

const changeHandle = debounce(() => {
  emit('change', searchVal.value)
}, 200)
</script>
