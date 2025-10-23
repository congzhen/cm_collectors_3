<template>
  <dialogCommon ref="dialogCommonRef" title="刮削演员信息" width="560px" btnSubmitTitle="刮削数据" @submit="submitHandle">
    <div class="scraperOneResource" v-loading="loading">
      <div class="form">
        <el-form ref="ruleFormRef" :model="formData" label-width="100px" status-icon>
          <div class="form-main">
            <div class="form-item">
              <el-text type="primary">{{ title_C }}</el-text>
            </div>
            <div class="form-item">
              <el-form-item label="刮削器">
                <selectScraperConfig v-model="formData.scraperConfig" width="360px" />
              </el-form-item>
            </div>
            <div class="form-item">
              <el-form-item label="超时时间">
                <el-input-number v-model="formData.timeout" :min="1" :max="300" />
              </el-form-item>
            </div>
            <div class="form-item">
              <el-form-item>
                <el-radio-group v-model="formData.operate" fill="#4C4D4F">
                  <el-radio-button label="更新" value="update" />
                  <el-radio-button label="覆盖" value="cover" />>
                </el-radio-group>
              </el-form-item>
            </div>
          </div>
        </el-form>
      </div>
    </div>
  </dialogCommon>
</template>
<script lang="ts" setup>
import { ref, computed } from 'vue';
import dialogCommon from '../com/dialog/dialog-common.vue';
import selectScraperConfig from '../com/form/selectScraperConfig.vue';
import { debounceNow } from '@/assets/debounce';
import { scraperDataServer } from '@/server/scraper.server';
import { ElMessage, } from 'element-plus';
import type { I_performerBasic } from '@/dataType/performer.dataType';


const emits = defineEmits(['success'])
const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();

const loading = ref(false);
const formData = ref({
  scraperConfig: '',
  operate: 'update',
  timeout: 30,
})
let performerId = '';
let performerBases_id = '';
const performerName = ref('');

const title_C = computed(() => {
  return "刮削演员: " + performerName.value;
});


const init = (_performerId: string, _performerBases_id: string, _name: string) => {
  dialogCommonRef.value?.disabledSubmit(false);
  performerId = _performerId;
  performerBases_id = _performerBases_id;
  performerName.value = _name;
};

const submitHandle = debounceNow(async () => {
  if (formData.value.scraperConfig == '') {
    ElMessage.error('请先选择刮削器配置文件');
    return
  }
  loading.value = true;
  try {
    dialogCommonRef.value?.disabledSubmit(true);
    if (!formData.value) {
      return;
    }
    const result = await scraperDataServer.scraperOnePerformerDataProcess(
      performerId,
      performerName.value,
      performerBases_id,
      formData.value.scraperConfig,
      formData.value.timeout,
      formData.value.operate
    );
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    } else {
      ElMessage.success('刮削成功');
      success(result.data);
    }
  } catch (error) {
    ElMessage.success(String(error));
  } finally {
    loading.value = false;
    dialogCommonRef.value?.disabledSubmit(false);
  }
})


const success = (data: I_performerBasic) => {
  dialogCommonRef.value?.close();
  emits('success', data);
};

const open = (_performerId: string, _performerBases_id: string, _name: string) => {
  if (_performerBases_id == '') {
    ElMessage.error('演员所属库不能为空');
    return;
  }
  if (_name == '') {
    ElMessage.error('请先输入演员名称');
    return;
  }
  init(_performerId, _performerBases_id, _name);
  dialogCommonRef.value?.open();
};


defineExpose({ open })
</script>
<style scoped lang="scss">
.scraperOneResource {
  display: flex;
  flex-direction: column;
  gap: 10px;

  .form-main {
    display: flex;
    gap: 10px;
    flex-direction: column;

    .form-item-checkbox-list {
      display: flex;
      flex-direction: column;
    }
  }
}

.checkbox-wrap {
  white-space: normal !important;
  height: auto;
  line-height: 1.5;
  margin-bottom: 8px;
}

.checkbox-wrap :deep(.el-checkbox__label) {
  white-space: normal !important;
  word-break: break-all;
}
</style>
