<template>
  <dialogCommon ref="dialogCommonRef" title="导入数据列表" btnSubmitTitle="导入数据" @submit="submitHandle" @closed="closeHandle">
    <el-table :data="pathList" height="400px" border size="small" style="width: 100%">
      <el-table-column type="index" width="50" />
      <el-table-column width="32">
        <template #default="scope">
          <div class="table-icon">
            <el-icon v-if="scope.row.status" size="14">
              <Select />
            </el-icon>
            <el-icon v-else-if="waiting">
              <Paperclip />
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
import { debounceNow } from '@/assets/debounce';
import { importDataServer } from '@/server/importData.server';
import { appStoreData } from '@/storeData/app.storeData';
import type { I_config_scanDisk } from '@/dataType/config.dataType';
import { ElMessage } from 'element-plus';
const store = {
  appStoreData: appStoreData(),
}
interface I_pathList {
  path: string;
  status: boolean;
  msg: string;
}

const emits = defineEmits(['success'])

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();
const pathList = ref<I_pathList[]>([]);
const waiting = ref(true);

let config: I_config_scanDisk;
let workStatus = true;


const init = (_pathList: string[], _config: I_config_scanDisk) => {
  workStatus = true;
  waiting.value = true;
  pathList.value = [];
  _pathList.forEach(path => {
    pathList.value.push({
      path,
      status: false,
      msg: '',
    });
  });
  config = _config;
}

const submitHandle = debounceNow(async () => {
  dialogCommonRef.value?.disabledSubmit(true);
  waiting.value = false;
  for (let i = 0; i < pathList.value.length; i++) {
    if (!workStatus) {
      return;
    }
    const result = await importDataServer.scanDiskImportData(store.appStoreData.currentFilesBases.id, pathList.value[i].path, config);
    if (!result.status) {
      pathList.value[i].msg = result.msg;
    }
    pathList.value[i].status = true;
  }
  success();
})

const closeHandle = () => {
  workStatus = false;
  success();
}

const success = () => {
  ElMessage.success('导入成功');
  close();
  emits('success')
}

const open = (_pathList: string[], _config: I_config_scanDisk) => {
  init(_pathList, _config)
  dialogCommonRef.value?.open();
};
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
