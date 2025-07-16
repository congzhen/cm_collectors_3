<template>
  <el-dialog v-model="dialogVisible" :style="{ width: dialogWidth, maxWidth: '80vw' }" :append-to-body="true"
    :show-close="false" :close-on-click-modal="false" top="6vh">
    <template #header="{ close, titleId, titleClass }">
      <div class="dialog-image-header">
        <div :id="titleId" :class="titleClass">{{ filePath }}</div>
        <el-icon @click="close">
          <Close />
        </el-icon>
      </div>
    </template>
    <div class="sfm-image" :style="{ height: dialogHeight }">
      <el-image :src="imageBase64" fit="contain"></el-image>
    </div>
  </el-dialog>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { Close } from '@element-plus/icons-vue'
const dialogVisible = ref(false)

const filePath = ref('')
const imageBase64 = ref('')
const dialogWidth = ref('') // 动态宽度
const dialogHeight = ref('') // 动态高度

const open = (_filePath: string, _base64: string) => {
  filePath.value = _filePath
  imageBase64.value = _base64
  // 动态获取图片尺寸
  const img = new Image()
  img.src = _base64
  img.onload = () => {
    const maxWidth = window.innerWidth * 0.8 // 80vw
    const width = img.width > maxWidth ? maxWidth : img.width
    const height = (img.height * width) / img.width

    dialogWidth.value = `${width}px`
    dialogHeight.value = `${height}px`
    dialogVisible.value = true
  }
}
const close = () => {
  dialogVisible.value = false
}

defineExpose({ open, close })
</script>
<style lang="scss" scoped>
.dialog-image-header {
  display: flex;
  justify-content: space-between;

  .el-dialog__title {
    font-size: 14px;
    line-height: 16px;
  }

  .el-icon {
    cursor: pointer;
  }
}

.sfm-image {
  padding: 5px;
  border: 2px solid #eee;
  border-radius: 5px;
  display: flex;

  img {
    width: 100%;
    height: auto;
  }
}
</style>
