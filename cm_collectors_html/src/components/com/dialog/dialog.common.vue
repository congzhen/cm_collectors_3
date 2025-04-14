<template>
  <el-dialog
    v-model="dialogVisible"
    :title="props.title"
    :width="width_C"
    :append-to-body="true"
    :close-on-click-modal="false"
    @closed="closed"
  >
    <slot></slot>
    <template v-if="props.footer" #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false"> Close </el-button>
        <el-button type="primary" @click="submitHandle">
          {{ btnSubmitTitle_C }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
import { ref, computed, type PropType } from 'vue'
type btnType = 'primary' | 'success' | 'info' | 'warning' | 'danger' | ''

const dialogVisible = ref(false)

// eslint-disable-next-line no-undef
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
  btnSubmitTitle: {
    type: String,
    default: '',
  },
})

// eslint-disable-next-line no-undef
const emits = defineEmits(['closed', 'submit'])

const btnSubmitTitle_C = computed(() => {
  return props.btnSubmitTitle == '' ? 'Submit' : props.btnSubmitTitle
})

const width_C = computed(() => {
  const screenWidth = window.innerWidth
  const targetWidth = parseInt(props.width)
  if (targetWidth > screenWidth * 0.9) {
    return `${screenWidth * 0.9}px` // 保留10%边距
  }
  return props.width
})

const closed = () => {
  console.log('closed')
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
defineExpose({ open, close })
</script>
