<template>
  <div class="BWH">
    <div class="BWH_k">
      <el-input v-model="bust" @input="emit('update:waist', waist)" :maxlength="3"
        :formatter="(value: string) => value.replace(/[^\d]/g, '')">
        <template #prepend>Bust</template>
      </el-input>
    </div>
    <div class="BWH_k">
      <el-input v-model="waist" @input="emit('update:hip', hip)" :maxlength="3"
        :formatter="(value: string) => value.replace(/[^\d]/g, '')">
        <template #prepend>Waist</template>
      </el-input>
    </div>
    <div>
      <el-input v-model="hip" @input="emit('update:bust', bust)" :maxlength="3"
        :formatter="(value: string) => value.replace(/[^\d]/g, '')">
        <template #prepend>Hip</template>
      </el-input>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  waist: string
  hip: string
  bust: string
}>()

const emit = defineEmits<{
  (e: 'update:waist', val: string): void
  (e: 'update:hip', val: string): void
  (e: 'update:bust', val: string): void
}>()

// 使用 ref 创建可变的本地变量
const bust = ref(props.bust)
const waist = ref(props.waist)
const hip = ref(props.hip)

// 当 props 更新时同步本地变量
watch(() => props.bust, (newVal) => {
  bust.value = newVal
})
watch(() => props.waist, (newVal) => {
  waist.value = newVal
})
watch(() => props.hip, (newVal) => {
  hip.value = newVal
})

// 当本地变量变化时触发 emit
watch(bust, (newVal) => {
  emit('update:bust', newVal)
})
watch(waist, (newVal) => {
  emit('update:waist', newVal)
})
watch(hip, (newVal) => {
  emit('update:hip', newVal)
})
</script>
<style lang="scss" scoped>
.BWH {
  display: flex;

  .BWH_k {
    margin-right: 5px;
  }

  :deep(.el-input-group__prepend) {
    padding: 0 7px;
  }
}
</style>
