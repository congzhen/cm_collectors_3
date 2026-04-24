<template>
  <dialogCommon ref="dialogCommonRef" title="刮削数据列表" btnSubmitTitle="刮削数据" @submit="submitHandle" @closed="closeHandle">
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
            <el-icon v-else-if="scope.row.msg != ''">
              <CloseBold />
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
import type { I_config_scraperData } from '@/dataType/config.dataType';
import { debounceNow } from '@/assets/debounce';
import { ElMessageBox, ElTable } from 'element-plus';
import { scraperDataServer } from '@/server/scraper.server';
import { appStoreData } from '@/storeData/app.storeData';

interface I_pathList {
  path: string;
  status: boolean;
  msg: string;
  importing?: boolean; //  importing 标记
}
const store = {
  appStoreData: appStoreData(),
}


const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();

const tableRef = ref<InstanceType<typeof ElTable>>();
const pathList = ref<I_pathList[]>([]);
let config: I_config_scraperData;
let workStatus = true;
const init = (_pathList: string[], _config: I_config_scraperData) => {
  workStatus = true;
  dialogCommonRef.value?.disabledSubmit(false);
  pathList.value = [];
  _pathList.forEach(path => {
    pathList.value.push({
      path,
      status: false,
      msg: '',
      importing: false, // 初始化 importing
    });
  });
  config = _config;
}
const submitHandle = debounceNow(async () => {
  dialogCommonRef.value?.disabledSubmit(true);

  // 获取选中的行
  const selectedRows = tableRef.value?.getSelectionRows() || [];

  // 确定要刮削的数据列表：如果用户没选，则刮削全部；否则刮削选中的
  const rowsToProcess = selectedRows.length > 0 ? selectedRows : pathList.value;

  // 标记所有待处理行的状态为 importing
  pathList.value.forEach(row => {
    const shouldProcess = rowsToProcess.some(r => r.path === row.path);
    if (shouldProcess) {
      row.importing = true;
    }
  });

  // 使用信号量控制并发数
  const concurrency = config.concurrency || 3; // 默认并发数为3
  let index = 0;


  const total = rowsToProcess.length;
  if (total === 0) {
    success();
    return;
  }

  const processItem = async () => {
    const currentIndex = index++;
    if (currentIndex >= total || !workStatus) {
      return;
    }

    const item = rowsToProcess[currentIndex];
    // item.waiting = false; // 删除

    // 添加随机延迟（0-2秒）
    const delay = Math.random() * 2000;
    await new Promise(resolve => setTimeout(resolve, delay));

    try {
      const result = await scraperDataServer.scraperDataProcess(
        store.appStoreData.currentFilesBases.id,
        item.path,
        config
      );

      if (!result.status) {
        item.msg = result.msg;
      }
      item.status = result.status;
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
      item.msg = error.message || '请求失败';
      item.status = false;
    } finally {
      // 处理完成后取消 importing 状态
      item.importing = false;
    }

    // 继续处理下一个
    await processItem();
  };

  // 启动并发任务
  const promises = Array(Math.min(concurrency, total)).fill(null).map(() => processItem());
  await Promise.all(promises);

  success();
})
const closeHandle = () => {
  workStatus = false;
}
const success = () => {
  ElMessageBox.alert('刮削完成', {
    confirmButtonText: 'OK',
  })
}

const open = (_pathList: string[], _config: I_config_scraperData) => {
  init(_pathList, _config)
  dialogCommonRef.value?.open();

  // 默认全选：在对话框打开后，确保表格渲染完成再执行全选
  nextTick(() => {
    nextTick(() => {
      tableRef.value?.toggleAllSelection();
    });
  });
};


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
