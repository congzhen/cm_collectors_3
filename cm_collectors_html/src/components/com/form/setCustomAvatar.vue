<template>
  <div class="set-custom-avatar">
    <el-upload action="/" :on-change="handleUploadPhotos" :show-file-list="false" :auto-upload="false" drag>
      <el-avatar class="customAvatar" :size="80" :src="avatar_C" />
    </el-upload>
    <el-icon class="customAvatarDelete" :size="20" @click="customAvatarDelete">
      <Delete />
    </el-icon>
    <comCropperDialog ref="comCropperDialogRef" @sumbit="cropperSubmit"></comCropperDialog>
  </div>
</template>
<script lang="ts" setup>
import comCropperDialog from '@/components/com/cropper/cropperDialog.vue';
import type { UploadFile } from 'element-plus';
import { ref, computed } from 'vue';

const avatarVal = defineModel({ type: String, default: '' })

const comCropperDialogRef = ref<InstanceType<typeof comCropperDialog>>();

const avatar_C = computed(() => {
  if (avatarVal.value !== '') {
    return avatarVal.value;
  }
  return '/emptyPhoto.jpg';
});

const handleUploadPhotos = (_uploadFile: UploadFile) => {
  comCropperDialogRef.value?.open(_uploadFile.raw, '50%', 280, 280);
}
const cropperSubmit = (fileData: string) => {
  avatarVal.value = fileData;
}
const customAvatarDelete = () => {
  avatarVal.value = '';
}



</script>
<style scoped lang="scss">
.set-custom-avatar {
  display: flex;

  :deep(.el-upload-dragger) {
    padding: 10px;
  }

  .customAvatarDelete {
    cursor: pointer;
    margin: 86px 0px 0px 20px;

    :hover {
      color: #F56C6C;
    }
  }

}
</style>
