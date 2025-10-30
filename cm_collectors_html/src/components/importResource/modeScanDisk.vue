<template>
  <div class="mode-scan-disk" v-loading="loading">
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
        <el-alert title="导入配置" type="success" :closable="false" />

        <el-form-item label="监控文件后缀名">
          <selectVideoSuffixName v-model="formData.videoSuffixName" multiple filterable allow-create
            default-first-option />
          <div><el-checkbox v-model="formData.autoGetVideoDefinition" label="自动获取视频清晰度" /></div>
        </el-form-item>
        <el-form-item label="封面海报类型">
          <el-select v-model="formData.coverPosterType">
            <el-option label="自适应尺寸" :value="-1" />
            <el-option v-for="item, index in store.appStoreData.currentConfigApp.coverPosterData" :key="index"
              :label="item.name" :value="index" />
          </el-select>
        </el-form-item>
        <el-form-item label="资源命名方式">
          <el-radio-group v-model="formData.resourceNamingMode" size="small">
            <el-radio-button label="文件名" value="fileName" />
            <el-radio-button label="目录名" value="dirName" />
            <el-radio-button label="目录名+文件名" value="dirFileName" />
            <el-radio-button label="全路径名" value="fullPathName" />
          </el-radio-group>
        </el-form-item>
        <el-form-item label="导入方式">
          <div class="form-column-list">
            <el-radio-group v-model="formData.importMode" size="small">
              <el-radio-button label="追加导入" value="append" />
              <el-radio-button label="覆盖导入" value="cover" />
            </el-radio-group>
            <div><el-text type="warning">覆盖导入会更新已存在的数据并导入新资源</el-text></div>
            <div><el-text type="warning">覆盖导入当多个资源指向同一视频地址时，仅更新最后的资源记录</el-text></div>
          </div>
        </el-form-item>

        <el-form-item label="封面海报匹配名">
          <el-select v-model="formData.coverPosterMatchName" multiple filterable allow-create default-first-option>
            <el-option v-for="item, key in dataset.coverPosterMatchName" :key="key" :label="item" :value="item" />
          </el-select>
          <el-switch v-model="formData.coverPosterFuzzyMatch" active-text="模糊匹配" inactive-text="严格匹配" />
          <el-checkbox v-model="formData.coverPosterUseRandomImageIfNoMatch" label="匹配的封面失败时，使用目录下随机图片" />
        </el-form-item>
        <el-form-item label="封面海报后缀名">
          <selectImageSuffixName v-model="formData.coverPosterSuffixName" multiple filterable allow-create
            default-first-option />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="formData.autoCreatePoster" label="(未找到封面海报) 自动截取视频内容作封面海报" />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="formData.folderToSeries" label="将同一文件夹下的多个视频文件合并为剧集" />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="formData.folderToSeriesSort" label="合并剧集时，是否按名称重新排序" />
        </el-form-item>
      </div>
      <div class="block">
        <el-alert title="nfo配置" type="success" :closable="false" />
        <el-form-item>
          <div>
            <div><el-checkbox v-model="formData.nfo.nfoStatus" label="导入nfo文件" /></div>
            <div><el-text type="warning">次级节点标签请使用 . 链接</el-text></div>
          </div>

        </el-form-item>
        <el-form-item label="根节点">
          <el-select v-model="formData.nfo.roots" multiple filterable allow-create default-first-option>
            <el-option v-for="item, index in dataset.nfo.roots" :key="index" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题">
          <el-select v-model="formData.nfo.titles" multiple filterable allow-create default-first-option>
            <el-option v-for="item, index in dataset.nfo.titles" :key="index" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="版号番号">
          <el-select v-model="formData.nfo.issueNumbers" multiple filterable allow-create default-first-option>
            <el-option v-for="item, index in dataset.nfo.issueNumbers" :key="index" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="发行日期">
          <el-select v-model="formData.nfo.issuingDates" multiple filterable allow-create default-first-option>
            <el-option v-for="item, index in dataset.nfo.issuingDates" :key="index" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="摘要简介">
          <el-select v-model="formData.nfo.abstracts" multiple filterable allow-create default-first-option>
            <el-option v-for="item, index in dataset.nfo.abstracts" :key="index" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-select v-model="formData.nfo.tags" multiple filterable allow-create default-first-option>
            <el-option v-for="item, index in dataset.nfo.tags" :key="index" :label="item" :value="item" />
          </el-select>
          <el-checkbox v-model="formData.nfo.tagAutoCreate" label="自动添加标签" />
        </el-form-item>
        <el-form-item label="演员姓名">
          <el-select v-model="formData.nfo.performerNames" multiple filterable allow-create default-first-option>
            <el-option v-for="item, index in dataset.nfo.performerNames" :key="index" :label="item" :value="item" />
          </el-select>
          <el-checkbox v-model="formData.nfo.performerMatchAliasName" label="同时匹配别名" />
          <el-checkbox v-model="formData.nfo.performerAutoCreate" label="自动添加演员" />
        </el-form-item>
        <el-form-item label="演员头像">
          <el-select v-model="formData.nfo.performerThumbs" multiple filterable allow-create default-first-option>
            <el-option v-for="item, index in dataset.nfo.performerThumbs" :key="index" :label="item" :value="item" />
          </el-select>
        </el-form-item>
      </div>
    </el-form>
  </div>
  <serverFileManagementDialog ref="serverFileManagementDialogRef" @selectedFiles="selectedFilesHandle"
    :show="[E_sfm_FileType.Directory]">
  </serverFileManagementDialog>
  <modeScanDiskImportDataDialog ref="modeScanDiskImportDataDialogRef" @success="successHandle">
  </modeScanDiskImportDataDialog>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import type { I_sfm_FileEntry } from '@/components/serverFileManagement/com/dataType';
import { E_sfm_FileType } from '@/components/serverFileManagement/com/dataType';
import serverFileManagementDialog from '@/components/serverFileManagement/serverFileManagementDialog.vue';
import modeScanDiskImportDataDialog from './modeScanDiskImportDataDialog.vue';
import selectVideoSuffixName from '../com/form/selectVideoSuffixName.vue';
import selectImageSuffixName from '../com/form/selectImageSuffixName.vue';
import { appStoreData } from '@/storeData/app.storeData';
import { defualtConfigScanDisk, E_config_type, type I_config_scanDisk } from '@/dataType/config.dataType';
import { filesBasesServer } from '@/server/filesBases.server';
import { ElMessage } from 'element-plus';
import { debounceNow } from '@/assets/debounce';
import { importDataServer } from '@/server/importData.server';
import dataset from '@/assets/dataset';
const store = {
  appStoreData: appStoreData(),
}
const emits = defineEmits(['success'])

const serverFileManagementDialogRef = ref<InstanceType<typeof serverFileManagementDialog>>();
const modeScanDiskImportDataDialogRef = ref<InstanceType<typeof modeScanDiskImportDataDialog>>();

const loading = ref(false)
const formData = ref<I_config_scanDisk>({ ...defualtConfigScanDisk })

const init = async () => {
  await getConfig();
}

const getConfig = async () => {
  try {
    loading.value = true;
    const result = await filesBasesServer.getConfigById(store.appStoreData.currentFilesBases.id, E_config_type.importScanDisk);
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    const configStr = result.data;
    if (configStr != '') {
      const config = JSON.parse(configStr);
      formData.value = { ...defualtConfigScanDisk, ...config };
    } else {
      formData.value = { ...defualtConfigScanDisk };
    }
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }
}

const getConfigScanDisk = (): I_config_scanDisk => {
  const configData = formData.value;
  if (configData.coverPosterType == -1) {
    configData.coverPosterWidth = 0;
    configData.coverPosterHeight = 0;
  } else {
    const coverPosterData = store.appStoreData.currentConfigApp.coverPosterData[configData.coverPosterType];
    if (coverPosterData) {
      configData.coverPosterWidth = coverPosterData.width;
      configData.coverPosterHeight = coverPosterData.height;
    }
  }
  return configData
}

const submit = debounceNow(async () => {
  if (formData.value.scanDiskPaths.length == 0) {
    ElMessage.error('请先设置监控路径');
    return;
  }
  try {
    const configData = getConfigScanDisk();

    loading.value = true;
    const result = await importDataServer.scanDiskImportPaths(store.appStoreData.currentFilesBases.id, configData);
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    } else if (result.data.length == 0) {
      ElMessage.error('没有可导入的数据');
      return;
    } else {
      modeScanDiskImportDataDialogRef.value?.open(result.data, configData);
    }
  } catch (error) {
    ElMessage.error(String(error));
  } finally {
    loading.value = false;
  }
});

const saveConfig = debounceNow(async () => {
  try {
    loading.value = true;
    const configData = getConfigScanDisk();
    const result = await importDataServer.updateScanDiskConfig(store.appStoreData.currentFilesBases.id, configData);
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    } else {
      ElMessage.success('保存成功');
    }
  } catch (error) {
    ElMessage.error(String(error));
  } finally {
    loading.value = false;
  }
})

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

const successHandle = () => {
  emits('success')
}


defineExpose({ init, submit, saveConfig })

</script>
<style lang="scss" scoped>
.mode-scan-disk {
  width: 100%;
  height: 100%;
  overflow: hidden;

  .block {
    .form-column-list {
      line-height: normal;
    }

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
  }
}
</style>
