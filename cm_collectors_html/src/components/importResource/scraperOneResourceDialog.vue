<template>
  <dialogCommon ref="dialogCommonRef" title="刮削资源信息" width="560px" btnSubmitTitle="刮削数据" @submit="submitHandle">
    <div class="scraperOneResource" :loading="loading">
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
            <el-form-item>
              <div class="form-item-checkbox-list">
                <el-checkbox class="checkbox-wrap" v-model="formData.saveNfo" :label="'保存NFO到：' + path_C"
                  size="large" />
                <el-checkbox class="checkbox-wrap" v-model="formData.saveNfo" :label="'保存图片到：' + path_C" size="large" />
                <el-checkbox class="checkbox-wrap" v-model="formData.cutPoster" label="是否裁切海报" size="large" />
                <el-checkbox class="checkbox-wrap" v-model="formData.useExistNfo" label="如果NFO已存在,直接调用" size="large" />
              </div>
            </el-form-item>
          </div>
        </el-form>
      </div>
    </div>
  </dialogCommon>
</template>
<script lang="ts" setup>
import { ref, computed } from 'vue';
import dialogCommon from '../com/dialog/dialog-common.vue';
import type { I_scraperOneResource } from '@/dataType/other.dataType';
import { getDirectoryFromPath, getFileNameFromPath } from '@/assets/path';
import selectScraperConfig from '../com/form/selectScraperConfig.vue';
import { debounceNow } from '@/assets/debounce';
import { scraperDataServer } from '@/server/scraper.server';
import { ElMessage } from 'element-plus';
import type { I_resource } from '@/dataType/resource.dataType';
import { appStoreData } from '@/storeData/app.storeData';
const store = {
  app: appStoreData(),
}
const emits = defineEmits(['success'])
const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();

const scraperOneResource = ref<I_scraperOneResource>({
  title: '',
  issueNumber: '',
  dramaSeriesSrc: '',
});
const loading = ref(false);
const formData = ref({
  scraperConfig: '',
  operate: 'update',
  timeout: 30,
  saveNfo: true,
  saveImages: true,
  cutPoster: false,
  useExistNfo: true,
})
let resourceId = '';
const title_C = computed(() => {
  const parts = [];
  if (scraperOneResource.value?.issueNumber) {
    parts.push(scraperOneResource.value.issueNumber);
  }
  if (scraperOneResource.value?.title) {
    parts.push(scraperOneResource.value.title);
  }
  if (scraperOneResource.value?.dramaSeriesSrc) {
    parts.push(getFileNameFromPath(scraperOneResource.value.dramaSeriesSrc));
  }

  return "刮削主题: " + parts.join(" / ");
});
const path_C = computed(() => {
  if (!scraperOneResource.value || scraperOneResource.value.dramaSeriesSrc == '') return '';
  return getDirectoryFromPath(scraperOneResource.value.dramaSeriesSrc);
});

const init = (_resourceId: string, _scraperOneResource: I_scraperOneResource) => {
  dialogCommonRef.value?.disabledSubmit(false);
  resourceId = _resourceId;
  scraperOneResource.value = _scraperOneResource;
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
    const result = await scraperDataServer.scraperOneResourceDataProcess(
      resourceId,
      store.app.currentFilesBases.id,
      formData.value.scraperConfig,
      formData.value.timeout,
      formData.value.operate,
      formData.value.saveNfo,
      formData.value.saveImages,
      formData.value.cutPoster,
      formData.value.useExistNfo,
      scraperOneResource.value.title,
      scraperOneResource.value.issueNumber,
      scraperOneResource.value.dramaSeriesSrc,
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

const success = (data: I_resource) => {
  dialogCommonRef.value?.close();
  emits('success', data);
};

const open = (_resourceId: string, _scraperOneResource: I_scraperOneResource) => {
  init(_resourceId, _scraperOneResource);
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
