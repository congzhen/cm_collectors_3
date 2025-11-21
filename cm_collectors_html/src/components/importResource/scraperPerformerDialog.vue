<template>
  <dialogCommon ref="dialogCommonRef" title="刮削演员数据" btnSubmitTitle="刮削" width="860px" @submit="submitHandle"
    @closed="closeHandle">
    <div class="scraper-performer-main" :loading="loading">
      <div class="form">
        <el-form ref="ruleFormRef" :model="formData" label-width="160px" status-icon>
          <div class="form-main">
            <div class="form-item">
              <selectScraperConfig v-model="formData.scraperConfig" width="220px" />
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
              <el-select v-model="formData.concurrency" style="width: 60px">
                <el-option v-for="item in [1, 2, 3, 4, 5, 6, 7, 8, 9]" :key="item" :label="item" :value="item" />
              </el-select>
            </div>
            <div class="form-item">
              <el-button-group>
                <el-button type="success" plain @click="searchScraperDataHandle">检索数据</el-button>
                <el-button type="success" plain @click="saveConfigHandle">保存配置</el-button>
              </el-button-group>
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
import { appStoreData } from '@/storeData/app.storeData';
import { filesBasesServer } from '@/server/filesBases.server';
import { defualtConfigScraperPerformerData, E_config_type, type I_config_scraperPerformerData } from '@/dataType/config.dataType';
interface I_scraperPerformer {
  info: I_performerBasic;
  status: boolean;
  waiting: boolean;
  msg: string;
}
const store = {
  appStoreData: appStoreData(),
}
const emits = defineEmits(['success'])
const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();
const formData = ref<I_config_scraperPerformerData>({
  scraperConfig: '',
  operate: 'update',
  lastScraperUpdateTime: '',
  concurrency: 3,
  timeout: 30,
})
const loading = ref(false);
const dataList = ref<I_scraperPerformer[]>([]);
let performerBasesId = '';
let workStatus = true;

const init = async (_performerBasesId: string) => {
  await getConfig();
  performerBasesId = _performerBasesId;
  workStatus = true;
  dataList.value = [];
  dialogCommonRef.value?.disabledSubmit(true);
}

const getConfig = async () => {
  try {
    loading.value = true;
    const result = await filesBasesServer.getConfigById(store.appStoreData.currentFilesBases.id, E_config_type.scraperPerformer);
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    const configStr = result.data;
    if (configStr != '') {
      const config = JSON.parse(configStr);
      formData.value = { ...defualtConfigScraperPerformerData, ...config };
    } else {
      formData.value = { ...defualtConfigScraperPerformerData };
    }
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }
}

const searchScraperDataHandle = debounceNow(async () => {
  try {
    loading.value = true;
    const result = await scraperDataServer.searchScraperPerformerData(store.appStoreData.currentFilesBases.id, performerBasesId, formData.value);
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

const saveConfigHandle = debounceNow(async () => {
  try {
    loading.value = true;
    const result = await scraperDataServer.updatePerformerScraperConfig(store.appStoreData.currentFilesBases.id, formData.value);
    if (result && result.status) {
      ElMessage.success('保存成功');
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

  // 使用信号量控制并发数
  const concurrency = formData.value.concurrency;
  const total = dataList.value.length;
  let index = 0;

  const processItem = async () => {
    const currentIndex = index++;
    if (currentIndex >= total || !workStatus) {
      return;
    }

    const item = dataList.value[currentIndex];
    item.waiting = false;

    // 添加随机延迟（0-2秒）
    const delay = Math.random() * 2000;
    await new Promise(resolve => setTimeout(resolve, delay));

    try {
      const result = await scraperDataServer.scraperPerformerDataProcess(
        performerBasesId,
        item.info.id,
        item.info.name,
        formData.value.scraperConfig,
        formData.value.operate
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
  flex-wrap: wrap;
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
