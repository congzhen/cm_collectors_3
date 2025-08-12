<template>
  <el-drawer v-model="drawerVisible" :direction="direction" :size="props.width" :close-on-click-modal="false"
    :append-to-body="true">
    <template v-if="props.title != ''" #header>
      <label style="font-size: 1.5em">{{ props.title }}</label>
    </template>
    <template #default>
      <slot></slot>
    </template>
    <template #footer v-if="props.footer">
      <div class="drawer-footer" style="flex: auto">
        <div>
          <slot name="footerBtn"></slot>
        </div>
        <div v-if="props.btnSubmit">
          <el-button @click="drawerVisible = false"> Close </el-button>
          <el-button type="primary" @click="emits('submit')">
            {{ btnSubmitTitle_C }}
          </el-button>
        </div>
      </div>
    </template>
  </el-drawer>
</template>
<script setup lang="ts">
import { ref, computed, type PropType } from 'vue'
type directionType = 'rtl' | 'ltr' | 'ttb' | 'btt'

const drawerVisible = ref(false)

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
  return props.btnSubmitTitle == '' ? 'Submit' : props.btnSubmitTitle
})

const emits = defineEmits(['closed', 'submit'])

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
