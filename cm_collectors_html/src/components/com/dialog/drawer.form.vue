<template>
  <drawerCommon ref="drawerCommonRef" :width="props.width" :direction="props.direction" :title="props.title"
    :btnSubmitTitle="props.btnSubmitTitle" :btnSubmit="props.btnSubmit" :footer="props.footer"
    @submit="submitHandle(ruleFormTagRef)">
    <el-form ref="ruleFormTagRef" :model="props.modelValue" :rules="props.rules" :label-width="props.labelWidth"
      :label-position="props.labelPosition" status-icon>
      <slot></slot>
    </el-form>
  </drawerCommon>
</template>
<script setup lang="ts">
import { ref, type PropType, reactive } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import drawerCommon from './drawer.common.vue'
import { debounceNow } from '@/assets/debounce'

const drawerCommonRef = ref<InstanceType<typeof drawerCommon>>()
const ruleFormTagRef = ref<FormInstance>()

type directionType = 'rtl' | 'ltr' | 'ttb' | 'btt'

// eslint-disable-next-line no-undef
const props = defineProps({
  width: {
    type: String,
    default: '480px',
  },
  direction: {
    type: String as PropType<directionType>,
    default: 'rtl',
  },
  title: {
    type: String,
    default: '',
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
  footer: {
    type: Boolean,
    default: true,
  },
  btnSubmitTitle: {
    type: String,
    default: '',
  },
  btnSubmit: {
    type: Boolean,
    default: true,
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
  drawerCommonRef.value?.open()
}
const close = () => {
  drawerCommonRef.value?.close()
}
// eslint-disable-next-line no-undef
defineExpose({ open, close })
</script>
<style lang="scss" scoped>
.drawer-footer {
  display: flex;
  justify-content: space-between;
}
</style>
