<template>
  <el-date-picker v-model="selectVal" type="date" format="YYYY-MM-DD" value-format="YYYY-MM-DD"
    :placeholder="props.placeholder" :style="{ width: props.width }" @change="changeHandle"
    :disabled="props.disabled" />
</template>
<script setup lang="ts">
const selectVal = defineModel({ type: String, default: '' })

const props = defineProps({
  placeholder: {
    type: String,
    default: '',
  },
  width: {
    type: String,
    default: '100%',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['change'])

const changeHandle = (val: string) => {
  if (val == null) {
    selectVal.value = ''
  }
  emit('change', selectVal.value)
}
</script>
