<template>
  <el-dialog v-model="dialogVisible" title="文件管理" width="860" :append-to-body="true" :close-on-click-modal="false">
    <serverFileManagement ref="serverFileManagementRef" class="serverFileManagement"
      :fileOperate="[E_sfm_FileOperate.Rename, E_sfm_FileOperate.Delete]"
      :column="[E_sfm_Column.Name, E_sfm_Column.Size, E_sfm_Column.ModifiedAt, E_sfm_Column.Operate]"
      :show="props.show">
    </serverFileManagement>
    <template #footer>
      <div class="dialog-footer">
        <el-button type="primary" @click="selectedFilesHandle">选择文件</el-button>
      </div>
    </template>
  </el-dialog>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { E_sfm_FileOperate, E_sfm_Column, E_sfm_FileType } from './com/dataType.ts';
import serverFileManagement from './serverFileManagement.vue';
import { ElMessage } from 'element-plus';
const props = defineProps({
  show: { // 显示的文件类型
    type: Array<E_sfm_FileType>,
    default: () => {
      return [E_sfm_FileType.Directory, E_sfm_FileType.File]
    }
  },
})
const emit = defineEmits(['selectedFiles'])

const dialogVisible = ref(false)
const serverFileManagementRef = ref<InstanceType<typeof serverFileManagement>>();
const selectedFilesHandle = () => {
  const slc = serverFileManagementRef.value?.getSelectedFiles() || [];
  if (slc.length == 0) {
    ElMessage.warning('请选择文件');
    return;
  }
  emit('selectedFiles', slc)
  dialogVisible.value = false;
}

const open = () => {
  dialogVisible.value = true
}
defineExpose({ open })
</script>
<style scoped lang="scss">
.serverFileManagement {
  height: 50vh;
}
</style>
