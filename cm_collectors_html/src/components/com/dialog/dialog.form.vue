<template>
  <dialogCommon
    ref="dialogCommonRef"
    :width="props.width"
    :title="props.title"
    :btnSubmitTitle="props.btnSubmitTitle"
    :footer="props.footer"
    @submit="submitHandle(ruleFormTagRef)"
    @closed="close"
  >
    <el-form
      ref="ruleFormTagRef"
      :model="props.modelValue"
      :rules="props.rules"
      :label-width="props.labelWidth"
      :label-position="props.labelPosition"
      status-icon
    >
      <slot></slot>
    </el-form>
  </dialogCommon>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue'
import dialogCommon from './dialog.common.vue'
import type { FormInstance, FormRules } from 'element-plus'
import { debounceNow } from '@/assets/debounce'
const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>()
const ruleFormTagRef = ref<FormInstance>()
const props = defineProps({
  width: {
    type: String,
    default: '800px',
  },
  title: {
    type: String,
    default: '',
  },
  footer: {
    type: Boolean,
    default: true,
  },
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
  btnSubmitTitle: {
    type: String,
    default: '',
  },
})
const emits = defineEmits(['closed', 'submit'])

const submitHandle = debounceNow((formEl: FormInstance | undefined) => {
  try {
    if (!formEl) return
    formEl.validate((valid) => {
      if (valid) {
        emits('submit')
        return Promise.resolve()
      } else {
        return Promise.reject()
      }
    })
  } finally {
  }
})

const open = () => {
  dialogCommonRef.value?.open()
}
const close = () => {
  dialogCommonRef.value?.close()
}
// eslint-disable-next-line no-undef
defineExpose({ open, close })
</script>
