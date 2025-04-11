<template>
  <el-drawer
    v-model="drawerVisible"
    :direction="direction"
    :size="props.width"
    :close-on-click-modal="false"
    :append-to-body="true"
  >
    <template v-if="props.title != ''" #header>
      <label style="font-size: 1.5em">{{ props.title }}</label>
    </template>
    <template #default>
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
    </template>
    <template #footer v-if="props.footer">
      <div class="drawer-footer" style="flex: auto">
        <div>
          <slot name="footerBtn"></slot>
        </div>
        <div>
          <el-button @click="drawerVisible = false"> 关闭 </el-button>
          <el-button type="primary" @click="submitHandle(ruleFormTagRef)" :loading="btnLoading">
            {{ btnSubmitTitle_C }}
          </el-button>
        </div>
      </div>
    </template>
  </el-drawer>
</template>
<script setup lang="ts">
import { ref, computed, type PropType, reactive } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { debounceNow } from '@/assets/debounce'
const ruleFormTagRef = ref<FormInstance>()
type directionType = 'rtl' | 'ltr' | 'ttb' | 'btt'

const drawerVisible = ref(false)
const btnLoading = ref(false)

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

const btnSubmitTitle_C = computed(() => {
  return props.btnSubmitTitle == '' ? '提交' : props.btnSubmitTitle
})

const emits = defineEmits(['closed', 'submit'])

const resetLoading = () => {
  setTimeout(() => {
    btnLoading.value = false
  }, 1000)
}

const submitHandle = debounceNow((formEl: FormInstance | undefined) => {
  try {
    if (!formEl) return
    formEl.validate((valid) => {
      if (valid) {
        btnLoading.value = true
        resetLoading()
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
  drawerVisible.value = true
}
const close = () => {
  drawerVisible.value = false
  emits('closed')
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
