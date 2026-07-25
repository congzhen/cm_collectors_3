<template>
  <el-dialog class="mainDialog" :top="props.top" v-model="dialogVisible" title="图片编辑" :close-on-click-modal="false"
    :width="props.width" :fullscreen="fullscreen()" append-to-body>
    <div :style="{ height: 'calc(' + props.height + ' - 80px)' }">
      <comCropper v-if="dialogVisible" ref="comCropperRef" :cropWidth="cropWidth" :cropHeight="cropHeight" mode="50%"
        :fitImageToCrop="fitImageToCrop" @sumbit="sumbitCropper">
      </comCropper>
    </div>
  </el-dialog>
</template>
<script setup lang="ts">
import comCropper from "./cropperImage.vue"
import { nextTick, ref } from 'vue'

const props = defineProps({
  width: {
    type: String,
    default: '960px'
  },
  height: {
    type: String,
    default: '720px'
  },
  top: {
    type: String,
    default: '10vh'
  }
});
const emits = defineEmits(['sumbit']);

const comCropperRef = ref<InstanceType<typeof comCropper>>();
const dialogVisible = ref(false)

const cropWidth = ref(270);
const cropHeight = ref(320);
const fitImageToCrop = ref(false);

const open = (image: File | string | undefined = undefined, mode = '100%', _cropWidth: number | undefined = undefined, _cropHeight: number | undefined = undefined, maxCropWidth = 900, maxCropHeight = 580, _fitImageToCrop = false) => {
  if (_cropWidth) cropWidth.value = _cropWidth > maxCropWidth ? maxCropWidth : _cropWidth;
  if (_cropHeight) cropHeight.value = _cropHeight > maxCropHeight ? maxCropHeight : _cropHeight;
  fitImageToCrop.value = _fitImageToCrop;
  if (dialogVisible.value == false) {
    dialogVisible.value = true;
  }
  nextTick(() => {
    comCropperRef.value?.setMode(mode);
    if (image instanceof File) {
      comCropperRef.value?.setRawFile(image);
    } else if (image) {
      comCropperRef.value?.setImage(image);
    }
  })

}
const close = () => {
  if (dialogVisible.value == true) {
    dialogVisible.value = false;
  }
}

const sumbitCropper = (fileData: string) => {
  emits('sumbit', fileData)
  close();
}

const fullscreen = () => {
  if (window.innerWidth < 960 || window.innerHeight < 660) {
    return true;
  }
  return false;
}

defineExpose({
  open
})
</script>
