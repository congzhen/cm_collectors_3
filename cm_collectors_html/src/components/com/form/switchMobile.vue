<template>
  <div class="switch-mobile" v-if="isMobileDevice()">
    <el-switch v-model="switchVal" class="ml-2" inline-prompt
      style="--el-switch-on-color: #444444; --el-switch-off-color: #3a3a3a" active-text="移动端" inactive-text="桌面端"
      @change="changeHandle" />
  </div>
</template>
<script setup lang="ts">
import { getMobileShow, goToMobileOrPC, setMobileShow, isMobileDevice } from '@/assets/mobile';
import { onMounted } from 'vue'
const switchVal = defineModel({ type: Boolean, default: true })
const changeHandle = (val: boolean) => {
  setMobileShow(val)
  goToMobileOrPC()
}
onMounted(() => {
  switchVal.value = getMobileShow()
})
</script>
<style scoped lang="scss"></style>
