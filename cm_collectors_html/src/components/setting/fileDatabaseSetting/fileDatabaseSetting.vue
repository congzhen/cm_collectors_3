<template>
  <div class="database-setting">
    <div class="database-setting-btn">
      <div class="database-primary-actions">
        <el-button icon="Plus" type="success" @click="createNewFilesBases()" plain>创建新文件数据库</el-button>
        <el-button type="warning" @click="sortFilesBasesHandle()" plain>排序</el-button>
        <el-button type="warning" @click="replacePathHandle()" plain>路径替换</el-button>
        <el-button type="warning" @click="clearDeletedResourceHandle()" plain>清除已删除的资源</el-button>
        <el-button type="danger" plain icon="Delete" :disabled="!activeFilesBases"
          @click="deleteHandle">真实删除当前文件库</el-button>
      </div>

    </div>
    <el-tabs tab-position="left" class="setting-tabs" v-model="activeName">
      <el-tab-pane v-for="item, key in store.filesBasesStoreData.filesBases" :class="[item.status ? '' : 'disabled']"
        :key="key" :name="item.id" :label="item.name">
        <template #label>
          <span :class="[item.status ? '' : 'disabled']">{{ item.name }}</span>
        </template>
        <fileSettingData v-if="activeName === item.id" :filesBasesId="item.id" @set-success="setSuccessHandle">
        </fileSettingData>
      </el-tab-pane>
    </el-tabs>
  </div>
  <fileDatabaseFormDialog ref="fileDatabaseFormDialogRef"></fileDatabaseFormDialog>
  <fileDatabaseSortDialog ref="fileDatabaseSortDialogRef"></fileDatabaseSortDialog>
  <pathReplaceDialog ref="pathReplaceDialogRef"></pathReplaceDialog>
  <clearDeletedResourceDialog ref="clearDeletedResourceDialogRef"></clearDeletedResourceDialog>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue';
import fileSettingData from './fileSettingData.vue';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
import { LoadingService } from '@/assets/loading';
import { appStoreData } from '@/storeData/app.storeData';
import { ElMessage, ElMessageBox } from 'element-plus';
import fileDatabaseFormDialog from './fileDatabaseFormDialog.vue';
import fileDatabaseSortDialog from './fileDatabaseSortDialog.vue';
import pathReplaceDialog from './pathReplaceDialog.vue';
import clearDeletedResourceDialog from './clearDeletedResourceDialog.vue';
import { filesBasesServer } from '@/server/filesBases.server';

const store = {
  appStoreData: appStoreData(),
  filesBasesStoreData: filesBasesStoreData(),
}

const fileDatabaseFormDialogRef = ref<InstanceType<typeof fileDatabaseFormDialog>>();
const fileDatabaseSortDialogRef = ref<InstanceType<typeof fileDatabaseSortDialog>>();
const pathReplaceDialogRef = ref<InstanceType<typeof pathReplaceDialog>>();
const clearDeletedResourceDialogRef = ref<InstanceType<typeof clearDeletedResourceDialog>>();

// 当前左侧 tab 选中的文件库 ID。
// 删除按钮操作的不是子组件里的临时表单数据，而是这里的当前库，避免按钮位置调整后出现目标不明确。
const activeName = ref(store.filesBasesStoreData.filesBasesFirst?.id);

// 根据当前 tab 实时拿到完整文件库对象，用于确认弹窗显示名称和提交删除 ID。
// 使用 computed 可以跟随 activeName 自动变化，切换 tab 后删除按钮始终指向当前选中的文件库。
const activeFilesBases = computed(() => store.filesBasesStoreData.getFilesBasesById(activeName.value || ''));


const setSuccessHandle = async (filesBasesId: string) => {
  try {
    LoadingService.show()
    const result = await store.appStoreData.initApp();
    if (result && !result.status) {
      ElMessage.error(result.message);
      return
    }
    if (filesBasesId == store.appStoreData.currentFilesBases.id) {
      const result = await store.appStoreData.initCurrentFilesBases(filesBasesId)
      if (result && !result.status) {
        ElMessage.error(result.message);
        return
      }
    }
  } catch (err) {
    console.log(err)
  } finally {
    LoadingService.hide()
  }
}

// 删除成功后刷新全局应用数据。
// 文件库列表、当前文件库、左侧 tab 都依赖 appStoreData.initApp 的结果；
// 删除后重新选中第一个可用文件库，避免界面继续停留在已经不存在的库上。
const deleteSuccessHandle = async () => {
  try {
    LoadingService.show()
    const result = await store.appStoreData.initApp();
    if (result && !result.status) {
      ElMessage.error(result.message);
      return
    }
    activeName.value = store.filesBasesStoreData.filesBasesFirst?.id;
    if (activeName.value) {
      await store.appStoreData.initCurrentFilesBases(activeName.value);
    }
  } catch (err) {
    console.log(err)
  } finally {
    LoadingService.hide()
  }
}

const createNewFilesBases = () => {
  fileDatabaseFormDialogRef.value?.open();
}

const sortFilesBasesHandle = () => {
  fileDatabaseSortDialogRef.value?.open();
}

const replacePathHandle = () => {
  pathReplaceDialogRef.value?.open();
}

const clearDeletedResourceHandle = () => {
  clearDeletedResourceDialogRef.value?.open();
}

// 真实删除当前文件库。
// 前端负责二次确认和发起请求，后端会再次校验库内是否还有资源记录；
// 因此即使用户绕过前端确认，也不会删除仍有资源的文件库。
const deleteHandle = async () => {
  if (!activeFilesBases.value) return;
  try {
    // 使用明确的“真实删除”和库名提示，降低误删风险。
    await ElMessageBox.confirm(
      `确定要真实删除文件库 [ ${activeFilesBases.value.name} ] 吗？删除后无法恢复。仅当库内没有资源记录时才允许删除。`,
      '确认删除',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
  } catch {
    return
  }

  // 这里只传当前选中文件库 ID，业务校验和附属配置清理由后端事务完成。
  const result = await filesBasesServer.delete(activeFilesBases.value.id);
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }
  ElMessage.success('删除成功');
  await deleteSuccessHandle();
}

</script>
<style lang="scss" scoped>
.database-setting-btn {
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.database-primary-actions {
  display: flex;
}

:deep(.el-tabs__item) {
  .disabled {
    color: #909399 !important;
    text-decoration: line-through;
    font-style: italic;
    opacity: 0.7;
  }

  // 选中状态覆盖 disabled 样式
  &.is-active .disabled {
    color: var(--el-color-primary) !important;
    opacity: 1;
  }
}
</style>
