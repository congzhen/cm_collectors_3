<template>
  <dialogCommon ref="dialogCommonRef" title="刮削演员数据" btnSubmitTitle="刮削" width="860px" @submit="submitHandle"
    @closed="closeHandle">
    <div class="scraper-performer-main" :loading="loading">
      <div class="form">
        <el-form ref="ruleFormRef" :model="formData" label-width="160px" status-icon>
          <div class="form-main">
            <div class="form-item">
              <selectScraperConfig v-model="formData.scraperConfig" width="260px" />
            </div>
            <div class="form-item">
              <el-radio-group v-model="formData.operate" fill="#4C4D4F">
                <el-radio-button label="更新" value="update" />
                <el-radio-button label="覆盖" value="cover" />>
              </el-radio-group>
            </div>
            <div class="form-item">
              <datePicker v-model="formData.lastScraperUpdateTime" placeholder="最后更新时间" width="140px" />
            </div>
            <div class="form-item">
              <el-button type="success" icon="Search" plain @click="searchScraperDataHandle">检索数据</el-button>
            </div>
          </div>
        </el-form>
      </div>
      <div class="data-list">
        <el-table :data="dataList" height="100%" border size="small" style="width: 100%">
          <el-table-column type="index" width="80" />
          <el-table-column width="80">
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
          <el-table-column prop="path" label="姓名">
            <template #default="scope">
              <p>
                <label>{{ scope.row.info.name }}</label>
                <label class="alias-name" v-if="scope.row.info.aliasName != ''">({{ scope.row.info.aliasName }})</label>
              </p>
              <p class="error" v-if="scope.row.msg != ''">{{ scope.row.msg }}</p>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </dialogCommon>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import dialogCommon from '../com/dialog/dialog-common.vue';
import selectScraperConfig from '../com/form/selectScraperConfig.vue';
import datePicker from '../com/form/datePicker.vue';
import { debounceNow } from '@/assets/debounce';
import { scraperDataServer } from '@/server/scraper.server';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { I_performerBasic } from '@/dataType/performer.dataType';
interface I_scraperPerformer {
  info: I_performerBasic;
  status: boolean;
  waiting: boolean;
  msg: string;
}
const emits = defineEmits(['success'])
const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();
const formData = ref({
  scraperConfig: '',
  operate: 'update',
  lastScraperUpdateTime: ''
})
const loading = ref(false);
const dataList = ref<I_scraperPerformer[]>([]);
let performerBasesId = '';
let workStatus = true;

const init = (_performerBasesId: string) => {
  performerBasesId = _performerBasesId;
  workStatus = true;
  dataList.value = [];
  dialogCommonRef.value?.disabledSubmit(true);
}

const searchScraperDataHandle = debounceNow(async () => {
  try {
    loading.value = true;
    const result = await scraperDataServer.searchScraperPerformerData(performerBasesId, formData.value.lastScraperUpdateTime);
    if (result && result.status) {
      dataList.value = [];
      result.data.forEach(item => {
        dataList.value.push({
          info: item,
          status: false,
          waiting: true,
          msg: '',
        });
      })
      dialogCommonRef.value?.disabledSubmit(false);
    } else if (result.msg) {
      ElMessage.error(result.msg);
    }
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }
})

const submitHandle = debounceNow(async () => {
  if (formData.value.scraperConfig == '') {
    ElMessage.error('请先选择刮削器配置文件');
    return
  }
  dialogCommonRef.value?.disabledSubmit(true);
  for (let i = 0; i < dataList.value.length; i++) {
    if (!workStatus) {
      return;
    }
    dataList.value[i].waiting = false;
    const result = await scraperDataServer.scraperPerformerDataProcess(performerBasesId, dataList.value[i].info.id, dataList.value[i].info.name, formData.value.scraperConfig, formData.value.operate);
    if (!result.status) {
      dataList.value[i].msg = result.msg;
    }
    dataList.value[i].status = result.status;
  }
  success();
})
const success = () => {
  ElMessageBox.alert('刮削完成', {
    confirmButtonText: 'OK',
  })
  emits('success');
}
const open = (_performerBasesId: string) => {
  init(_performerBasesId);
  dialogCommonRef.value?.open();
};
const closeHandle = () => {
  workStatus = false;
}

defineExpose({ open })
</script>
<style scoped lang="scss">
.form-main {
  display: flex;
  gap: 10px;
}

.data-list {
  padding-top: 10px;
  width: 100%;
  height: 400px;

  .alias-name {
    padding-left: 10px;
  }
}
</style>
