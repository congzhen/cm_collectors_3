<template>
  <dialogCommon ref="dialogCommonRef" title="导入数据列表" btnSubmitTitle="导入数据" @submit="submitHandle" @closed="closeHandle">
    <el-table ref="tableRef" :data="pathList" height="400px" border size="small" style="width: 100%">
      <el-table-column type="selection" width="55" />
      <el-table-column type="index" width="50" />
      <el-table-column width="32">
        <template #default="scope">
          <div class="table-icon">
            <el-icon v-if="scope.row.importing" class="element-rotating" size="14">
              <Loading />
            </el-icon>
            <el-icon v-else-if="scope.row.status" size="14">
              <Select />
            </el-icon>
            <el-icon v-else>
              <Paperclip />
            </el-icon>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="path" label="文件路径">
        <template #default="scope">
          <p>{{ scope.row.path }}</p>
          <p class="error" v-if="scope.row.msg != ''">{{ scope.row.msg }}</p>
        </template>
      </el-table-column>
    </el-table>
  </dialogCommon>
</template>
<script lang="ts" setup>
import { ref, nextTick } from 'vue';
import dialogCommon from '../com/dialog/dialog-common.vue';
import { debounceNow } from '@/assets/debounce';
import { importDataServer } from '@/server/importData.server';
import { appStoreData } from '@/storeData/app.storeData';
import type { I_config_scanDisk } from '@/dataType/config.dataType';
import { ElMessageBox, ElTable } from 'element-plus';
const store = {
  appStoreData: appStoreData(),
}
interface I_pathList {
  path: string;
  status: boolean;
  msg: string;
  importing?: boolean; // 标记是否正在导入
}

const emits = defineEmits(['success'])

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();
// 新增 tableRef 用于获取选中行
const tableRef = ref<InstanceType<typeof ElTable>>();
const pathList = ref<I_pathList[]>([]);
const waiting = ref(true);

let config: I_config_scanDisk;
let workStatus = true;


const init = (_pathList: string[], _config: I_config_scanDisk) => {
  workStatus = true;
  waiting.value = true;
  dialogCommonRef.value?.disabledSubmit(false);
  pathList.value = [];
  _pathList.forEach(path => {
    pathList.value.push({
      path,
      status: false,
      msg: '',
      importing: false, // 初始化 importing 状态
    });
  });
  config = _config;
}

const submitHandle = debounceNow(async () => {
  dialogCommonRef.value?.disabledSubmit(true);
  waiting.value = false;

  // 获取选中的行
  const selectedRows = tableRef.value?.getSelectionRows() || [];

  // 确定要导入的数据列表：如果用户没选，则导入全部；否则导入选中的
  const rowsToImport = selectedRows.length > 0 ? selectedRows : pathList.value;

  // 标记所有待导入行的状态为 importing
  pathList.value.forEach(row => {
    const shouldImport = rowsToImport.some(r => r.path === row.path);
    if (shouldImport) {
      row.importing = true;
    }
  });

  for (let i = 0; i < pathList.value.length; i++) {
    if (!workStatus) {
      return;
    }

    const row = pathList.value[i];

    // 检查当前行是否属于需要导入的列表
    const shouldImport = rowsToImport.some(r => r.path === row.path);

    // 如果不导入该数据，则跳过，不改变其原始状态（图标、消息等）
    if (!shouldImport) {
      continue;
    }

    const result = await importDataServer.scanDiskImportData(store.appStoreData.currentFilesBases.id, row.path, config);
    if (!result.status) {
      row.msg = result.msg;
    }
    row.status = true;
    row.importing = false; // 导入完成，取消 importing 状态
  }
  success();
})

const closeHandle = () => {
  workStatus = false;
}

const success = () => {
  ElMessageBox.alert('导入成功', {
    confirmButtonText: 'OK',
  })
  emits('success');
}

const open = (_pathList: string[], _config: I_config_scanDisk) => {
  init(_pathList, _config)
  dialogCommonRef.value?.open();

  // 默认全选：在对话框打开后，确保表格渲染完成再执行全选
  nextTick(() => {
    nextTick(() => {
      tableRef.value?.toggleAllSelection();
    });
  });
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const close = () => {
  dialogCommonRef.value?.close();
}

defineExpose({ open })

</script>
<style scoped lang="scss">
.table-icon {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
}

.error {
  color: #F56C6C;
}
</style>
