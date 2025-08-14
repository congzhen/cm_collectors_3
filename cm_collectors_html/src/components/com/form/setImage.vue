<template>
  <div class="set-image" :style="{ width: props.width, height: props.height }">
    <el-upload action="/" :on-change="handleUploadPhotos" :show-file-list="false" :auto-upload="false" drag>
      <div class="photo">
        <el-image :src="image_C">
          <template #error>
            <div class="image-slot">{{ props.placeholder }}</div>
          </template>
        </el-image>
      </div>
    </el-upload>
    <comCropperDialog ref="comCropperDialogRef" @sumbit="cropperSubmit"></comCropperDialog>
  </div>
</template>
<script lang="ts" setup>
import comCropperDialog from '@/components/com/cropper/cropperDialog.vue';
import type { UploadFile } from 'element-plus';
import { ref, computed } from 'vue';

const props = defineProps({
  src: {
    type: String,
    default: '',
  },
  width: {
    type: String,
    default: '100%',
  },
  height: {
    type: String,
    default: '100%',
  },
  placeholder: {
    type: String,
    default: '点击或拖拽上传图片'
  },
  cropperWidth: {
    type: Number,
    default: 270
  },
  cropperHeight: {
    type: Number,
    default: 320
  }
})
const comCropperDialogRef = ref<InstanceType<typeof comCropperDialog>>();

const imageBase64 = ref('');

const image_C = computed(() => {
  return imageBase64.value == '' ? props.src : imageBase64.value;
});

const init = () => {
  imageBase64.value = '';
}

const handleUploadPhotos = (_uploadFile: UploadFile) => {
  if (_uploadFile.raw) {
    openCropper(_uploadFile.raw, '50%', props.cropperWidth, props.cropperHeight, props.cropperWidth, props.cropperHeight)
  }
}

const openCropper = (file: File | undefined = undefined, mode = '100%', _cropWidth: number | undefined = undefined, _cropHeight: number | undefined = undefined, maxCropWidth: number | undefined, maxCropHeight: number | undefined) => {
  comCropperDialogRef.value?.open(file, mode, _cropWidth, _cropHeight, maxCropWidth, maxCropHeight);
}

const cropperSubmit = (fileData: string) => {
  imageBase64.value = fileData;
}

const getImageBase64 = () => {
  return imageBase64.value;
}
const getImageSize = () => {
  return {
    width: props.cropperWidth,
    height: props.cropperHeight,
  }
}

defineExpose({ init, getImageBase64, getImageSize, openCropper })


</script>
<style scoped lang="scss">
.set-image {
  :first-child {
    width: 100%;
    height: 100%;
  }

  :deep(.el-upload-dragger) {
    padding: 0px;
    border: 0px;
    border-radius: 5px;
  }

  .image-slot {
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.2em;
  }

}
</style>
