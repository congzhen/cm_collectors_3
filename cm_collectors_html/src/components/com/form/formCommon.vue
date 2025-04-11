<template>
  <el-form
    ref="ruleFormTagRef"
    :model="props.modelValue"
    :rules="props.rules"
    :label-width="props.labelWidth"
    :label-position="props.labelPosition"
    status-icon
  >
    <div class="form-container">
      <slot></slot>
    </div>
  </el-form>
</template>
<script setup lang="ts">
import type { FormRules } from 'element-plus'
import { reactive } from 'vue'

const props = defineProps({
  modelValue: {
    required: true,
    type: Object,
  },
  rules: {
    type: Object,
    default: () => {
      return reactive<FormRules>({})
    },
  },
  labelWidth: {
    type: String,
    default: '120px',
  },
  labelPosition: {
    type: String,
    default: 'right',
  },
})
</script>
