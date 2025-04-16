<template>
  <el-rate v-model="selectSetVal" clearable :style="{ width: props.width }" />
</template>
<script setup lang="ts">
const selectSetVal = defineModel({ type: Number, default: 0 })
const props = defineProps({
  width: {
    type: String,
    default: '100%',
  },
})
</script>
