<template>
  <div class="database-setting">
    <div class="database-setting-btn">
      <div class="database-primary-actions">
        <el-button icon="Plus" type="success" @click="createNewPerformerBases" plain>创建新演员集</el-button>
        <el-button type="warning" plain @click="importHandle">导入</el-button>
        <el-button type="warning" plain @click="exportHandle">导出</el-button>
        <el-button type="danger" plain icon="Delete" :disabled="!activePerformerBases"
          @click="deleteHandle">真实删除当前演员集</el-button>
      </div>

    </div>
    <el-tabs tab-position="left" class="setting-tabs" v-model="activeName">
      <el-tab-pane v-for="item, key in store.performerBasesStoreData.performerBases" :key="key" :name="item.id"
        :label="item.name">
        <performerSettingData v-if="activeName === item.id" :performer-bases="item"></performerSettingData>
      </el-tab-pane>
    </el-tabs>
  </div>
  <performerDatabaseFormDialog ref="performerDatabaseFormDialogRef"></performerDatabaseFormDialog>
  <exportPerformerDatabaseDialog ref="exportPerformerDatabaseDialogRef"></exportPerformerDatabaseDialog>
  <importPerformerDatabaseDialog ref="importPerformerDatabaseDialogRef"></importPerformerDatabaseDialog>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue';
import performerSettingData from './performerSettingData.vue';
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';
import performerDatabaseFormDialog from './performerDatabaseFormDialog.vue';
import exportPerformerDatabaseDialog from './exportPerformerDatabaseDialog.vue';
import importPerformerDatabaseDialog from './importPerformerDatabaseDialog.vue';
import { appStoreData } from '@/storeData/app.storeData';
import { LoadingService } from '@/assets/loading';
import { ElMessage, ElMessageBox } from 'element-plus';
import { performerBasesServer } from '@/server/performerBases.server';
const store = {
  appStoreData: appStoreData(),
  performerBasesStoreData: performerBasesStoreData(),
}
const performerDatabaseFormDialogRef = ref<InstanceType<typeof performerDatabaseFormDialog>>();
const exportPerformerDatabaseDialogRef = ref<InstanceType<typeof exportPerformerDatabaseDialog>>();
const importPerformerDatabaseDialogRef = ref<InstanceType<typeof importPerformerDatabaseDialog>>();
// 当前左侧 tab 选中的演员库 ID。
// 顶部工具栏的删除按钮针对“当前演员库”，因此删除目标统一从这里推导。
const activeName = ref(store.performerBasesStoreData.activeFirstPerformerBases?.id);

// 当前选中的演员库完整对象。
// computed 会随 activeName 更新，用于确认弹窗展示名称和提交删除 ID。
const activePerformerBases = computed(() => store.performerBasesStoreData.getPerformerBasesById(activeName.value || ''));


const createNewPerformerBases = () => {
  performerDatabaseFormDialogRef.value?.open();
}

const exportHandle = () => {
  exportPerformerDatabaseDialogRef.value?.open();
}

const importHandle = () => {
  importPerformerDatabaseDialogRef.value?.open();
}

// 真实删除当前演员库。
// 前端只做确认和请求；后端会校验该演员库是否还有演员、是否仍被文件库关联。
// 这两个条件任何一个不满足，都不会执行物理删除。
const deleteHandle = async () => {
  if (!activePerformerBases.value) return;
  try {
    // 演员库可能被文件库配置引用，弹窗里把限制条件写清楚，方便用户知道失败原因。
    await ElMessageBox.confirm(
      `确定要真实删除演员库 [ ${activePerformerBases.value.name} ] 吗？删除后无法恢复。仅当库内没有演员记录、且未被文件库关联时才允许删除。`,
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

  // 只提交当前演员库 ID，避免子组件名称编辑状态影响删除目标。
  const result = await performerBasesServer.delete(activePerformerBases.value.id);
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }
  ElMessage.success('删除成功');
  await deleteSuccessHandle();
}

// 删除成功后刷新应用数据和演员库列表。
// 如果当前库被删除，需要切换到第一个启用的演员库，避免 tab 指向已经不存在的数据。
const deleteSuccessHandle = async () => {
  try {
    LoadingService.show()
    const result = await store.appStoreData.initApp();
    if (result && !result.status) {
      ElMessage.error(result.message);
      return
    }
    activeName.value = store.performerBasesStoreData.activeFirstPerformerBases?.id;
  } catch (err) {
    console.log(err)
  } finally {
    LoadingService.hide()
  }
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
</style>
