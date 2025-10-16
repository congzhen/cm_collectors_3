<template>
  <div class="scraper-data" v-loading="loading">
    <div class="block">
      <el-alert title="监控磁盘" type="success" :closable="false" />
      <ul class="scan-list">
        <li v-for="(item, index) in formData.scanDiskPaths" :key="index">
          <el-input v-model="formData.scanDiskPaths[index]" :disabled="true">
            <template #append>
              <el-button icon="Delete" @click="deleteDiskLocationHandle(index)" />
            </template>
          </el-input>
        </li>
      </ul>
      <div class="tool">
        <el-button type="primary" plain @click="addDiskLocationHandle">添加文件夹位置</el-button>
      </div>
    </div>
    <el-form ref="ruleFormRef" :model="formData" label-width="160px" status-icon>
      <div class="block">
        <el-alert title="刮削配置" type="success" :closable="false" />
        <el-form-item label="监控文件后缀名">
          <selectVideoSuffixName v-model="formData.videoSuffixName" multiple filterable allow-create
            default-first-option />
        </el-form-item>
        <el-form-item label="应用刮削器配置文件">
          <selectScraperConfig v-model="formData.scraperConfigs" multiple />
        </el-form-item>
        <el-form-item label="并发处理数量">
          <el-input-number v-model="formData.concurrency" :min="1" :max="10" />
        </el-form-item>
        <el-form-item label="重试次数">
          <el-input-number v-model="formData.retryCount" :min="0" :max="10" />
        </el-form-item>
        <el-form-item label="超时时间">
          <el-input-number v-model="formData.timeout" :min="1" :max="300" />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="formData.skipIfNfoExists" label="已存在nfo文件时跳过" />
          <el-checkbox v-model="formData.saveNfo" label="保存元数据为nfo文件" />
        </el-form-item>
        <el-form-item>
          <div>
            <el-checkbox v-model="formData.enableDownloadImages" label="下载元数据中的图片链接" />
            <el-checkbox v-model="formData.useTagAsImageName" label="使用标签名作为图片名" />
            <el-checkbox v-model="formData.enableUserSimulation" label="开启用户模拟操作" />
          </div>
          <div class="warning-list">
            <el-text type="warning">下载图片会耗费更多时间进行反反爬</el-text>
            <el-text type="warning">开启模拟操作会降低刮削速度，增加反反爬能力</el-text>
          </div>
        </el-form-item>
      </div>
    </el-form>
  </div>
  <serverFileManagementDialog ref="serverFileManagementDialogRef" @selectedFiles="selectedFilesHandle"
    :show="[E_sfm_FileType.Directory]">
  </serverFileManagementDialog>
  <scraperDataProcessDialog ref="scraperDataProcessDialogRef" @success="successHandle">
  </scraperDataProcessDialog>
</template>
<script setup lang="ts">
import serverFileManagementDialog from '@/components/serverFileManagement/serverFileManagementDialog.vue';
import scraperDataProcessDialog from './scraperDataProcessDialog.vue';
import { E_sfm_FileType, type I_sfm_FileEntry } from '@/components/serverFileManagement/com/dataType';
import { E_config_type, type I_config_scraperData, defualtConfigScraperData } from '@/dataType/config.dataType';
import selectVideoSuffixName from '../com/form/selectVideoSuffixName.vue';
import selectScraperConfig from '../com/form/selectScraperConfig.vue';
import { ref } from 'vue';
import { filesBasesServer } from '@/server/filesBases.server';
import { appStoreData } from '@/storeData/app.storeData';
import { ElMessage } from 'element-plus';
import { debounceNow } from '@/assets/debounce';
import { scraperDataServer } from '@/server/scraper.server';
const store = {
  appStoreData: appStoreData(),
}
const emits = defineEmits(['success'])
const serverFileManagementDialogRef = ref<InstanceType<typeof serverFileManagementDialog>>();
const scraperDataProcessDialogRef = ref<InstanceType<typeof scraperDataProcessDialog>>();
const loading = ref(false)
const formData = ref<I_config_scraperData>({ ...defualtConfigScraperData })

const init = async () => {
  await getConfig();
}
const getConfig = async () => {
  try {
    loading.value = true;
    const result = await filesBasesServer.getConfigById(store.appStoreData.currentFilesBases.id, E_config_type.scraper);
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    const configStr = result.data;
    if (configStr != '') {
      const config = JSON.parse(configStr);
      formData.value = { ...defualtConfigScraperData, ...config };
    } else {
      formData.value = { ...defualtConfigScraperData };
    }
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }
}

const addDiskLocationHandle = () => {
  serverFileManagementDialogRef.value?.open();
}
const selectedFilesHandle = (slc: I_sfm_FileEntry[]) => {
  if (slc.length == 0) {
    return;
  }
  slc.forEach(item => {
    if (item.is_dir && !formData.value.scanDiskPaths.includes(item.path)) {
      formData.value.scanDiskPaths.push(item.path);
    }
  });
}
const deleteDiskLocationHandle = (index: number) => {
  formData.value.scanDiskPaths.splice(index, 1);
}

const submit = debounceNow(async () => {
  if (formData.value.scanDiskPaths.length == 0) {
    ElMessage.error('请先设置监控路径');
    return;
  }
  if (formData.value.scraperConfigs.length == 0) {
    ElMessage.error('请先选择刮削器配置文件');
    return;
  }
  try {
    const configData = formData.value;
    loading.value = true;
    const result = await scraperDataServer.pretreatment(store.appStoreData.currentFilesBases.id, configData);
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    } else if (result.data.length == 0) {
      ElMessage.error('没有可刮削的数据');
      return;
    } else {
      scraperDataProcessDialogRef.value?.open(result.data, configData);
    }
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }
});
const successHandle = () => {
  emits('success')
}


defineExpose({ init, submit })
</script>
<style lang="scss" scoped>
.scraper-data {
  width: 100%;
  height: 100%;
  overflow: hidden;

  .block {

    .el-alert {
      margin: 0 0 10px 0;
    }

    .scan-list {
      list-style-type: none;
      display: flex;
      flex-direction: column;
      gap: 5px;
    }

    .tool {
      padding: 10px 0;
    }

    .el-form {
      width: 90%;
    }

    .warning-list {
      display: contents;
    }
  }
}
</style>
