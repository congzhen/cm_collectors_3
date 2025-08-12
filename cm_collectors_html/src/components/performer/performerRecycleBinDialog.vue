<template>
  <dialogTable ref="dialogTableRef" :dataList="dataList" :loading="loading" title="演员回收站">
    <el-table-column prop="name" label="姓名" width="200" show-overflow-tooltip></el-table-column>
    <el-table-column prop="aliasName" label="别名" width="360" show-overflow-tooltip></el-table-column>
    <el-table-column label="操作">
      <template #default="scope">
        <el-button type="warning" icon="RefreshLeft" size="small" @click="restoreHandle(scope.row)">恢复</el-button>
      </template>
    </el-table-column>
  </dialogTable>
</template>
<script setup lang="ts">
import { debounceNow } from '@/assets/debounce';
import { LoadingService } from '@/assets/loading';
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

const init = async () => {
  await getDataList();
}

const getDataList = async () => {
  loading.value = true;
  const result = await performerServer.recycleBin(props.performerBasesId);
  if (result && result.status) {
    dataList.value = result.data;
  } else {
    ElMessage.error(result.msg);
  }
  loading.value = false;
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

const open = async () => {
  await init();
  dialogTableRef.value?.open();
}


defineExpose({ open })
</script>
