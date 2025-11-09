<template>
  <el-dialog v-model="dialogVisible" :title="props.title" :width="width_C" :append-to-body="true" :top="props.top"
    :z-index="props.zIndex" :close-on-click-modal="false" @closed="closed">
    <slot></slot>
    <template v-if="props.footer" #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false"> {{ btnCloseTitle_C }} </el-button>
        <el-button type="primary" @click="submitHandle" :disabled="submitBtnDisabled">
          {{ btnSubmitTitle_C }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
import { ref, computed } from 'vue'
import { AppLang } from '@/language/app.lang'
const appLang = AppLang()
const dialogVisible = ref(false)

const props = defineProps({
  width: {
    type: String,
    default: '800px',
  },
  title: {
    type: String,
    default: '',
  },
  top: {
    type: String,
    default: '15vh',
  },
  footer: {
    type: Boolean,
    default: true,
  },
  btnSubmitTitle: {
    type: String,
    default: '',
  },
  btnCloseTitle: {
    type: String,
    default: '',
  },
  zIndex: {
    type: Number,
    default: 999,
  },
})

// eslint-disable-next-line no-undef
const emits = defineEmits(['closed', 'submit'])

const submitBtnDisabled = ref(false)

const btnSubmitTitle_C = computed(() => {
  return props.btnSubmitTitle == '' ? appLang.btn('submit') : props.btnSubmitTitle
})
const btnCloseTitle_C = computed(() => {
  return props.btnCloseTitle == '' ? appLang.btn('close') : props.btnCloseTitle
})

const width_C = computed(() => {
  const screenWidth = window.innerWidth
  const targetWidth = parseInt(props.width)
  if (targetWidth > screenWidth * 0.9) {
    return '90%'
  }
  return props.width
})

const disabledSubmit = (b: boolean) => {
  submitBtnDisabled.value = b
}

const closed = () => {
  emits('closed')
}

const submitHandle = () => {
  emits('submit')
}

const open = () => {
  dialogVisible.value = true
}
const close = () => {
  dialogVisible.value = false
}
// eslint-disable-next-line no-undef
defineExpose({ open, close, disabledSubmit })
</script>
