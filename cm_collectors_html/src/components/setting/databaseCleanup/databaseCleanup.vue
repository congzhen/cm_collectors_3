<template>
  <div class="database-cleanup" v-loading="loading">
    <section class="maintenance-group"><div class="maintenance-heading"><h3>清理范围与内容</h3><p>选择文件库及数据类型，确认提示后执行清理。</p></div>
    <el-form v-model="formData" label-width="200px">
      <el-form-item label="文件数据库集">
        <selectFilesBases v-model="formData.filesBases_ids" multiple></selectFilesBases>
        <el-text class="warning-text" type="warning" size="small">
          注意：若未指定文件数据库集，将默认清理全部
        </el-text>
      </el-form-item>
      <el-form-item label="清理数据">
        <div style="margin-bottom: 10px;">
          <el-button type="primary" link @click="selectMultipleItem('')">全选</el-button>
          <el-button type="primary" link @click="deselectAll">取消</el-button>
          <el-button type="primary" link @click="selectMultipleItem('data')">仅清除数据保留配置</el-button>
        </div>
        <el-checkbox-group v-model="formData.clear_items" @change="handleClearItemChange">
          <el-checkbox label="资源" value="resource"></el-checkbox>
          <el-checkbox label="演员" value="performer"></el-checkbox>
          <el-checkbox label="标签" value="tags"></el-checkbox>
          <el-checkbox label="标签分类" value="tagClass"></el-checkbox>
          <el-checkbox label="文件数据库配置" value="fileDatabaseConfig"></el-checkbox>
          <el-checkbox label="导入配置" value="importConfig"></el-checkbox>
          <el-checkbox label="资源刮削配置" value="resourceScraperConfig"></el-checkbox>
          <el-checkbox label="演员刮削配置" value="performerScraperConfig"></el-checkbox>
          <el-checkbox label="通用设置" value="generalConfig"></el-checkbox>
          <el-checkbox label="计划任务" value="cronJobs"></el-checkbox>
        </el-checkbox-group>
        <div>
          <el-text class="warning-text" type="warning" size="small">
            注意：选择了标签分类时必须同时选择标签
          </el-text>
          <el-text class="warning-text" type="warning" size="small">
            注意：清理演员数据将删除文件数据库中所有关联的演员数据集
          </el-text>
        </div>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" icon="delete" @click="clearHandle">清理</el-button>
      </el-form-item>
    </el-form>
    </section>

    <!-- 数据库备份列表移到表单外部 -->
    <section class="maintenance-group"><div class="maintenance-heading"><h3>数据库备份</h3><p>查看和管理已有数据库备份文件。</p></div>
    <div class="cleanup-backups">
      <div class="backup-list">
        <div v-if="dbBackupList.length === 0" class="empty-backup">
          暂无数据库备份
        </div>
        <div v-for="(backup, index) in dbBackupList" :key="index" class="cleanup-backup-item">
          <span class="backup-name">{{ backup }}</span>
          <el-button type="danger" size="small" @click="deleteBackup(backup, index)" plain>
            删除
          </el-button>
        </div>
      </div>
    </div>
    </section>
  </div>
</template>
<script lang="ts" setup>
import { debounceNow } from '@/assets/debounce';
import { getFileNameFromPath } from '@/assets/path';
import selectFilesBases from '@/components/com/form/selectFilesBases.vue';
import type { I_databaseCleanupForm, I_databaseCleanupFormClearItem } from '@/dataType/other.dataType';
import { appDataServer } from '@/server/app.server';
import { ElMessage, ElMessageBox } from 'element-plus';
import { h, onMounted, ref } from 'vue';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';

const formData = ref<I_databaseCleanupForm>({
  filesBases_ids: [],
  clear_items: [],
})
const filesBasesStore = filesBasesStoreData();
const cleanupLabels: Record<I_databaseCleanupFormClearItem, string> = {
  resource: '资源', performer: '演员', tags: '标签', tagClass: '标签分类',
  fileDatabaseConfig: '文件数据库配置', importConfig: '导入配置',
  resourceScraperConfig: '资源刮削配置', performerScraperConfig: '演员刮削配置',
  generalConfig: '通用设置', cronJobs: '计划任务',
};
const dbBackupList = ref<string[]>([]);
const loading = ref(false);

const init = async () => {
  await getDbBackupList();
}

const getDbBackupList = async () => {
  try {
    loading.value = true;
    const result = await appDataServer.dbBackupList();
    if (!result.status) {
      //ElMessage.error(result.msg);
      return
    }
    dbBackupList.value = result.data;
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }
}

// 添加删除备份功能
const deleteBackup = (backupPath: string, index: number) => {
  const backupName = getFileNameFromPath(backupPath, false)
  ElMessageBox.confirm(
    `确定要删除备份 "${backupName}" 吗？此操作不可恢复。`,
    '确认删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      loading.value = true;
      const result = await appDataServer.deleteDbBackup(backupName);
      if (!result.status) {
        ElMessage.error(result.msg);
        return
      }
      dbBackupList.value.splice(index, 1);
      ElMessage.success('删除成功');
    } catch (error) {
      console.log(error);
      ElMessage.error('删除失败');
    } finally {
      loading.value = false;
    }
  }).catch(() => {
    // 用户取消删除
  });
}

const handleClearItemChange = (value: string[]) => {
  // 如果选中了tagClass但没有选中tags，则自动选中tags
  if (value.includes('tagClass') && !value.includes('tags')) {
    formData.value.clear_items = [...value, 'tags'] as I_databaseCleanupFormClearItem[];
  }
}

// 全选函数
const selectMultipleItem = (m: string) => {
  switch (m) {
    case 'data':
      formData.value.clear_items = ['resource', 'performer', 'tags', 'tagClass'];
      break;
    default:
      formData.value.clear_items = ['resource', 'performer', 'tags', 'tagClass', 'fileDatabaseConfig', 'importConfig', 'resourceScraperConfig', 'performerScraperConfig', 'generalConfig', 'cronJobs'];
      break;
  }

}

// 取消全选函数
const deselectAll = () => {
  formData.value.clear_items = [];
}

const clearHandle = debounceNow(async () => {
  if (loading.value) return;
  if (formData.value.clear_items.length === 0) {
    ElMessage.warning('请选择要清理的数据');
    return;
  }
  // 确认内容和请求共用快照，避免弹窗期间的变动改变实际清理范围。
  const request: I_databaseCleanupForm = {
    filesBases_ids: [...formData.value.filesBases_ids],
    clear_items: [...formData.value.clear_items],
  };
  const libraryNames = request.filesBases_ids.map(id => filesBasesStore.getFilesBasesNameById(id) || `未知文件库（${id}）`);
  loading.value = true;
  try {
    try {
      await ElMessageBox.confirm(
        h('div', { class: 'cleanup-confirm-content' }, [
          h('p', { style: 'margin:0 0 8px;font-weight:600' }, '请核对本次清理范围和内容：'),
          h('p', { style: 'margin:8px 0;overflow-wrap:anywhere' }, [
            h('strong', '文件库范围：'),
            h('span', { style: libraryNames.length ? '' : 'color:var(--el-color-danger);font-weight:600' },
              libraryNames.length ? `指定 ${libraryNames.length} 个文件库：${libraryNames.join('、')}` : '全部文件库（未指定文件库）'),
          ]),
          h('p', { style: 'margin:8px 0' }, [h('strong', '清理内容：'), request.clear_items.map(item => cleanupLabels[item]).join('、')]),
          request.clear_items.includes('performer')
            ? h('p', { style: 'color:var(--el-color-warning)' }, '演员清理会删除所选文件库关联的演员数据集，可能影响共用这些演员集的其他文件库。') : null,
          request.clear_items.includes('generalConfig')
            ? h('p', { style: 'color:var(--el-color-danger)' }, '通用设置为全局配置，将重置整个软件的通用设置，不受所选文件库范围限制。') : null,
          h('p', { style: 'margin:12px 0 0;color:var(--el-text-color-secondary)' }, '确认后将执行清理；取消不会修改任何数据。'),
        ]),
        '确认清理数据库',
        { type: 'warning', confirmButtonText: '确认清理', cancelButtonText: '取消', closeOnClickModal: false },
      );
    } catch (action) {
      if (action === 'cancel' || action === 'close') return;
      throw action;
    }
    const result = await appDataServer.databaseCleanup(request);
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    await getDbBackupList();
    ElMessage.success('清理成功');
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }
});

onMounted(() => {
  init();
})

</script>
<style lang="scss" scoped>
@use '../settingsModules' as modules;

.database-cleanup {
  max-width: 960px;
  padding: 20px 0;

  .el-text {
    display: block;
  }

  .cleanup-backups {
    margin-top: 30px;
    padding: 20px;
    background-color: #2d2d2d;
    border-radius: 8px;
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);

    h3 {
      margin-top: 0;
      margin-bottom: 15px;
      color: #ffffff;
      font-size: 18px;
      font-weight: 600;
    }

    .backup-list {
      max-height: 300px;
      overflow-y: auto;

      &::-webkit-scrollbar {
        width: 6px;
      }

      &::-webkit-scrollbar-thumb {
        background-color: #666;
        border-radius: 3px;
      }
    }

    .cleanup-backup-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 12px 15px;
      margin-bottom: 10px;
      background-color: #3a3a3a;
      border-radius: 6px;
      transition: all 0.3s ease;

      &:hover {
        background-color: #444;
        transform: translateY(-2px);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
      }

      &:last-child {
        margin-bottom: 0;
      }

      .backup-name {
        flex: 1;
        color: #ffffff;
        font-size: 14px;
        word-break: break-all;
        margin-right: 15px;
      }
    }

    .empty-backup {
      text-align: center;
      color: #aaa;
      padding: 30px 0;
      font-size: 14px;
    }
  }
}

.database-cleanup {
  max-width: 1320px; margin: 0; padding: 0 0 16px; overflow: auto; box-sizing: border-box;
  .maintenance-group > .el-form { @include modules.panel; @include modules.form-rows; @include modules.mobile-form; }
  .maintenance-group > .el-form :deep(.el-form-item__content) { align-items: flex-start; flex-direction: column; }
  .maintenance-group > .el-form :deep(.el-select) { width: 100%; }
  .maintenance-group > .el-form :deep(.el-checkbox-group) { display: flex; flex-wrap: wrap; gap: 6px 20px; }
  .maintenance-group > .el-form :deep(.el-checkbox) { margin-right: 0; }
  .warning-text { line-height: 1.7; margin-top: 6px; }
  .cleanup-backups { @include modules.panel; margin-top: 12px; box-shadow: none; h3 { color: var(--el-text-color-primary); font-size: 16px; } }
  .cleanup-backups .cleanup-backup-item { background: var(--el-fill-color-light); border: 1px solid var(--el-border-color-lighter); box-shadow: none; &:hover { background: var(--el-fill-color); transform: none; box-shadow: none; } .backup-name { color: var(--el-text-color-regular); } }
}


.database-cleanup { max-width: none; @include modules.maintenance-groups;
  .maintenance-group .cleanup-backups { margin: 0; min-width: 0; }
  .maintenance-group > .el-form :deep(.el-form-item__label) { justify-content: flex-start; }
  .maintenance-group > .el-form :deep(.el-checkbox-group) { display: grid; width: 100%; max-width: 960px; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 4px 12px; }
  .maintenance-group > .el-form :deep(.el-select) { max-width: 760px; }
  .warning-text { text-align: left; align-self: flex-start; }
}
</style>
