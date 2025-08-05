<template>
  <div class="database-setting">
    <div class="database-setting-btn">
      <el-button icon="Plus" type="success" @click="createNewFilesBases()" plain>创建新文件数据库</el-button>
      <el-button type="warning" @click="sortFilesBasesHandle()" plain>排序</el-button>
      <el-button type="warning" @click="replacePathHandle()" plain>路径替换</el-button>
      <el-button type="warning" plain>批量删除</el-button>
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
</template>
<script setup lang="ts">
import { ref } from 'vue';
import fileSettingData from './fileSettingData.vue';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
import { LoadingService } from '@/assets/loading';
import { appStoreData } from '@/storeData/app.storeData';
import { ElMessage } from 'element-plus';
import fileDatabaseFormDialog from './fileDatabaseFormDialog.vue';
import fileDatabaseSortDialog from './fileDatabaseSortDialog.vue';
import pathReplaceDialog from './pathReplaceDialog.vue';

const store = {
  appStoreData: appStoreData(),
  filesBasesStoreData: filesBasesStoreData(),
}

const fileDatabaseFormDialogRef = ref<InstanceType<typeof fileDatabaseFormDialog>>();
const fileDatabaseSortDialogRef = ref<InstanceType<typeof fileDatabaseSortDialog>>();
const pathReplaceDialogRef = ref<InstanceType<typeof pathReplaceDialog>>();

const activeName = ref(store.filesBasesStoreData.filesBasesFirst?.id);


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

const createNewFilesBases = () => {
  fileDatabaseFormDialogRef.value?.open();
}

const sortFilesBasesHandle = () => {
  fileDatabaseSortDialogRef.value?.open();
}

const replacePathHandle = () => {
  pathReplaceDialogRef.value?.open();
}

</script>
<style lang="scss" scoped>
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
