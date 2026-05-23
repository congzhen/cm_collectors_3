<template>
  <dialogTable ref="dialogTableRef" :dataList="dataList" :loading="loading" title="演员回收站"
    @selection-change="selectionChangeHandle">
    <template #toolbar>
      <div class="recycle-bin-toolbar">
        <el-button type="danger" icon="Delete" size="small" plain :disabled="selectedRows.length === 0"
          @click="batchDeleteHandle">批量删除</el-button>
        <span class="selected-count" v-if="selectedRows.length > 0">已选择 {{ selectedRows.length }} 个演员</span>
      </div>
    </template>
    <el-table-column type="selection" width="55" />
    <el-table-column prop="name" label="姓名" width="200" show-overflow-tooltip></el-table-column>
    <el-table-column prop="aliasName" label="别名" show-overflow-tooltip></el-table-column>
    <el-table-column label="操作" width="160">
      <template #default="scope">
        <el-button type="warning" icon="RefreshLeft" size="small" plain @click="restoreHandle(scope.row)">恢复</el-button>
        <el-button type="danger" icon="Delete" size="small" plain
          @click="deleteHandle(scope.row, scope.$index)">删除</el-button>
      </template>
    </el-table-column>
  </dialogTable>
</template>
<script setup lang="ts">
import { debounceNow } from '@/assets/debounce';
import { LoadingService } from '@/assets/loading';
import { messageBoxConfirm } from '@/common/messageBox';
import dialogTable from '@/components/com/dialog/dialog-table.vue';
import type { I_performer } from '@/dataType/performer.dataType';
import { performerServer } from '@/server/performer.server';
import { ElMessage } from 'element-plus';
import { ref } from 'vue';

const props = defineProps({
  performerBasesId: {
    type: String,
    required: true,
  },
})
const emits = defineEmits(['success'])

const dialogTableRef = ref<InstanceType<typeof dialogTable>>();

const loading = ref(false);
const dataList = ref<I_performer[]>([]);
const selectedRows = ref<I_performer[]>([]);

const init = async () => {
  await getDataList();
}

const getDataList = async () => {
  loading.value = true;
  const result = await performerServer.recycleBin(props.performerBasesId);
  if (result && result.status) {
    dataList.value = result.data;
    selectedRows.value = [];
  } else {
    ElMessage.error(result.msg);
  }
  loading.value = false;
}

const selectionChangeHandle = (selection: I_performer[]) => {
  selectedRows.value = selection;
}

const restoreHandle = debounceNow(async (data: I_performer) => {
  LoadingService.show();
  try {
    const result = await performerServer.updateStatus(data.id, true);
    if (result.status) {
      ElMessage.success('恢复成功');
      emits('success', true);
      getDataList();
    } else {
      ElMessage.error(result.msg);
    }
  } catch (error) {
    ElMessage.error('提交失败，请稍后再试');
    console.log(error);
  } finally {
    LoadingService.hide();
  }
})

const deleteHandle = (data: I_performer, index: number) => {
  messageBoxConfirm({
    text: `确定要彻底删除( ${data.name} )吗？`,
    type: 'error',
    successCallBack: debounceNow(async () => {
      LoadingService.show();
      try {
        const result = await performerServer.delete(data.id);
        if (result.status) {
          ElMessage.success('删除成功');
          dataList.value.splice(index, 1);
        } else {
          ElMessage.error(result.msg);
        }
      } catch (error) {
        ElMessage.error('提交失败，请稍后再试');
        console.log(error);
      } finally {
        LoadingService.hide();
      }
    }),
  })
}

const batchDeleteHandle = () => {
  const rows = [...selectedRows.value];
  if (rows.length === 0) {
    ElMessage.warning('请先选择要删除的演员');
    return;
  }

  messageBoxConfirm({
    text: `确定要彻底删除选中的 ${rows.length} 个演员吗？`,
    type: 'error',
    successCallBack: debounceNow(async () => {
      LoadingService.show();
      try {
        const deletedIds: string[] = [];
        const failedNames: string[] = [];

        // 批量删除仍然逐条调用已有的单个删除接口，避免新增后端批量接口和重复业务逻辑。
        for (const performer of rows) {
          const result = await performerServer.delete(performer.id);
          if (result.status) {
            deletedIds.push(performer.id);
          } else {
            failedNames.push(performer.name);
          }
        }

        if (deletedIds.length > 0) {
          dataList.value = dataList.value.filter(item => !deletedIds.includes(item.id));
          dialogTableRef.value?.clearSelection();
          selectedRows.value = [];
        }

        if (failedNames.length > 0) {
          ElMessage.error(`删除完成，${failedNames.length} 个失败：${failedNames.join('、')}`);
        } else {
          ElMessage.success(`已删除 ${deletedIds.length} 个演员`);
        }
      } catch (error) {
        ElMessage.error('提交失败，请稍后再试');
        console.log(error);
      } finally {
        LoadingService.hide();
      }
    }),
  })
}

const open = async () => {
  await init();
  dialogTableRef.value?.open();
}


defineExpose({ open })
</script>
<style lang="scss" scoped>
.recycle-bin-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;

  .selected-count {
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }
}
</style>
