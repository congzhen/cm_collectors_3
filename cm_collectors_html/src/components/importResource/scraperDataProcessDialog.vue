<template>
  <dialogCommon ref="dialogCommonRef" title="刮削数据列表" btnSubmitTitle="刮削数据" @submit="submitHandle" @closed="closeHandle">
    <el-table :data="pathList" height="400px" border size="small" style="width: 100%">
      <el-table-column type="index" width="50" />
      <el-table-column width="32">
        <template #default="scope">
          <div class="table-icon">
            <el-icon v-if="scope.row.status" size="14">
              <Select />
            </el-icon>
            <el-icon v-else-if="scope.row.waiting">
              <Paperclip />
            </el-icon>
            <el-icon v-else-if="scope.row.msg != ''">
              <CloseBold />
            </el-icon>
            <el-icon v-else class="element-rotating" size="14">
              <Loading />
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
import { ref } from 'vue';
import dialogCommon from '../com/dialog/dialog-common.vue';
import type { I_config_scraperData } from '@/dataType/config.dataType';
import { debounceNow } from '@/assets/debounce';
import { ElMessageBox } from 'element-plus';
import { scraperDataServer } from '@/server/scraper.server';
import { appStoreData } from '@/storeData/app.storeData';

interface I_pathList {
  path: string;
  status: boolean;
  waiting: boolean;
  msg: string;
}
const store = {
  appStoreData: appStoreData(),
}


const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();
const pathList = ref<I_pathList[]>([]);
const waiting = ref(true);
let config: I_config_scraperData;
let workStatus = true;
const init = (_pathList: string[], _config: I_config_scraperData) => {
  workStatus = true;
  waiting.value = true;
  dialogCommonRef.value?.disabledSubmit(false);
  pathList.value = [];
  _pathList.forEach(path => {
    pathList.value.push({
      path,
      status: false,
      waiting: true,
      msg: '',
    });
  });
  config = _config;
}
const submitHandle = debounceNow(async () => {
  dialogCommonRef.value?.disabledSubmit(true);
  waiting.value = false;

  // 使用信号量控制并发数
  const concurrency = config.concurrency || 3; // 默认并发数为3
  const total = pathList.value.length;
  let index = 0;

  const processItem = async () => {
    const currentIndex = index++;
    if (currentIndex >= total || !workStatus) {
      return;
    }

    const item = pathList.value[currentIndex];
    item.waiting = false;

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
