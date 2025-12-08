<template>
  <el-dialog v-model="dialogVisible" title="文件管理" width="860" :append-to-body="true" :close-on-click-modal="false"
    body-class="dialog-sfm-body" footer-class="dialog-sfm-footer">
    <serverFileManagement ref="serverFileManagementRef" class="serverFileManagement"
      :fileOperate="[E_sfm_FileOperate.Rename, E_sfm_FileOperate.Delete]"
      :column="[E_sfm_Column.Name, E_sfm_Column.Size, E_sfm_Column.ModifiedAt, E_sfm_Column.Operate]"
      :show="props.show">
    </serverFileManagement>
    <template #footer>
      <div class="dialog-footer">
        <el-checkbox label="选择后保持窗口打开" v-model="noClose" size="small" v-if="noCloseOption" />
        <el-button type="primary" @click="selectedFilesHandle">选择文件</el-button>
      </div>
    </template>
  </el-dialog>
</template>
<script lang="ts" setup>
import { ref, watch } from 'vue';
import { E_sfm_FileOperate, E_sfm_Column, E_sfm_FileType } from './com/dataType.ts';
import serverFileManagement from './serverFileManagement.vue';
import { ElMessage, ElNotification } from 'element-plus';
const props = defineProps({
  show: { // 显示的文件类型
    type: Array<E_sfm_FileType>,
    default: () => {
      return [E_sfm_FileType.Directory, E_sfm_FileType.File]
    }
  },
  noCloseOption: {
    type: Boolean,
    default: false
  }
})
const emit = defineEmits(['selectedFiles'])


const dialogVisible = ref(false)
const serverFileManagementRef = ref<InstanceType<typeof serverFileManagement>>();
const noClose = ref(false)

watch(noClose, (newValue) => {
  if (props.noCloseOption) {
    localStorage.setItem('serverFileManagement_noCloseOption', newValue ? 'true' : 'false');
  }
});

const selectedFilesHandle = () => {
  const slc = serverFileManagementRef.value?.getSelectedFiles() || [];
  if (slc.length == 0) {
    ElMessage.warning('请选择文件');
    return;
  }
  emit('selectedFiles', slc)
  if (!noClose.value) {
    dialogVisible.value = false;
  } else {
    const message = '<div>' + slc.map(item => `<div style="margin: 2px 0;">${item.name}</div>`).join('') + '</div>';
    ElNotification({
      title: '已选择文件',
      message: message,
      type: 'success',
      dangerouslyUseHTMLString: true,
    })
    serverFileManagementRef.value?.clearSelectedFiles();
  }
}

const open = () => {
  serverFileManagementRef.value?.clearSelectedFiles();
  if (props.noCloseOption) {
    const noCloseOptionStatus = localStorage.getItem('serverFileManagement_noCloseOption');
    if (noCloseOptionStatus && noCloseOptionStatus == 'true') {
      noClose.value = true;
    } else {
      noClose.value = false;
    }
  }
  dialogVisible.value = true
}
defineExpose({ open })
</script>
<style lang="scss">
.dialog-sfm-body {
  .serverFileManagement {
    height: 50vh;
  }
}


.dialog-sfm-footer {
  padding-top: 5px;

  .dialog-footer {
    display: flex;
    gap: 10px;
    align-items: center;
    width: 100%;
  }

  /* 当只有一个子元素时，靠右对齐 */
  .dialog-footer> :only-child {
    margin-left: auto;
  }

  /* 当有两个子元素时，一个在最左，一个在最右 */
  .dialog-footer> :first-child:not(:only-child) {
    margin-right: auto;
  }
}
</style>
