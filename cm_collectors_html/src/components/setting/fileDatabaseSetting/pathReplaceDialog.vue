<template>
  <dialogCommon ref="dialogCommonRef" title="资源路径替换器" :footer="false">
    <div class="path-replace-container">
      <div class="condition">
        <div class="condition-item">
          <div class="condition-item-label">操作数据库:</div>
          <div class="condition-item-content">
            <el-select v-model="formData.filesBasesIds" placeholder="全部数据库" multiple
              :disabled="!formData.filesBasesIdsStatus">
              <el-option v-for="item, index in store.filesBasesStoreData.filesBases" :key="index" :label="item.name"
                :value="item.id"></el-option>
            </el-select>
          </div>
        </div>
        <div class="condition-item">
          <div class="condition-item-label">检索路径:</div>
          <div class="condition-item-content">
            <el-input v-model="formData.searchPath" :disabled="!formData.searchPathStatus">
              <template #prepend>
                <el-button icon="FolderOpened" @click="openServerFileManagement('searchPath')" />
              </template>
              <template #append> <el-button icon="Search" @click="searchHandle">检索</el-button></template>
            </el-input>
          </div>
        </div>
        <div class="condition-item">
          <div class="condition-item-label">替换路径:</div>
          <div class="condition-item-content">

            <el-input v-model="formData.replacePath" :disabled="!formData.replacePathStatus">
              <template #prepend>
                <el-button icon="FolderOpened" @click="openServerFileManagement('replacePath')" />
              </template>
              <template #append> <el-button icon="Refresh" @click="replacePathHandle()">替换</el-button></template>
            </el-input>
          </div>
        </div>
        <div class="condition-item">
          <div class="condition-item-label"></div>
          <div class="condition-item-content">
            <span class="result">检索到数据 {{ dataList.length }} 条.</span>
          </div>
        </div>
      </div>
      <ul class="result-list">
        <li v-for="item, index in dataList" :key="index">
          <div class="title">{{ item.title }}</div>
          <div class="path" v-html="searchContentShowRed(item.src)"></div>
        </li>
      </ul>
    </div>
  </dialogCommon>
  <serverFileManagementDialog ref="serverFileManagementDialogRef" @selectedFiles="selectedFilesHandle"
    :show="[E_sfm_FileType.Directory]">
  </serverFileManagementDialog>
</template>
<script lang="ts" setup>
import dialogCommon from '@/components/com/dialog/dialog-common.vue';
import serverFileManagementDialog from '@/components/serverFileManagement/serverFileManagementDialog.vue';
import { E_sfm_FileType, type I_sfm_FileEntry } from '@/components/serverFileManagement/com/dataType';
import { ref } from 'vue';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
import { debounceNow } from '@/assets/debounce';
import { LoadingService } from '@/assets/loading';
import { ElMessage } from 'element-plus';
import { resourcesDramaSeriesServer } from '@/server/resource.server';
import type { I_dramaSeriesWithResource } from '@/dataType/resource.dataType';

const store = {
  filesBasesStoreData: filesBasesStoreData(),
}

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();
const serverFileManagementDialogRef = ref<InstanceType<typeof serverFileManagementDialog>>();
const formData = ref({
  filesBasesIds: [],
  filesBasesIdsStatus: true,
  searchPath: '',
  searchPathStatus: true,
  replacePath: '',
  replacePathStatus: true,
})

const dataList = ref<I_dramaSeriesWithResource[]>([])

const init = () => {
  formData.value.filesBasesIds = [];
  formData.value.filesBasesIdsStatus = true;
  formData.value.searchPath = '';
  formData.value.searchPathStatus = true;
  formData.value.replacePath = '';
  formData.value.replacePathStatus = true;
  dataList.value = [];
}

const searchHandle = debounceNow(async () => {
  LoadingService.show();
  try {
    if (formData.value.searchPath == '') {
      ElMessage.error('请输入检索路径');
      return;
    }
    const result = await resourcesDramaSeriesServer.searchPath(formData.value.filesBasesIds, formData.value.searchPath)
    if (result.status) {
      dataList.value = result.data;
      //formData.value.filesBasesIdsStatus = false;
      //formData.value.searchPathStatus = false;
    } else {
      ElMessage.error(result.msg);
    }
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
  } catch (error) {
    ElMessage.error('提交失败，请稍后再试');
  } finally {
    LoadingService.hide();
  }
})

const replacePathHandle = debounceNow(async () => {
  LoadingService.show();
  try {
    if (dataList.value.length == 0) {
      ElMessage.error('没有可替换数据');
      return;
    }
    if (formData.value.replacePath == '') {
      ElMessage.error('请输入替换路径');
      return;
    }
    const result = await resourcesDramaSeriesServer.replacePath(formData.value.filesBasesIds, formData.value.searchPath, formData.value.replacePath)
    if (result.status) {
      ElMessage.success('替换成功');
      dataList.value = result.data;
    } else {
      ElMessage.error(result.msg);
    }
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
  } catch (error) {
    ElMessage.error('提交失败，请稍后再试');
  } finally {
    LoadingService.hide();
  }
})

const searchContentShowRed = (contentText: string) => {
  // 转义正则表达式特殊字符
  const escapeRegExp = (string: string) => {
    return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  };

  const escapedSearchPath = escapeRegExp(formData.value.searchPath);
  return contentText.replace(
    new RegExp(escapedSearchPath, "gi"),
    `<span style="color: red">${formData.value.searchPath.replace(/[&<>"']/g, (char) => {
      // 转义HTML特殊字符以防止XSS
      const escapeMap: { [key: string]: string } = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#39;'
      };
      return escapeMap[char] || char;
    })}</span>`
  );
}

type openServerFileManagementFieldType = 'searchPath' | 'replacePath';
let openServerFileManagementField: openServerFileManagementFieldType = 'searchPath';
const openServerFileManagement = (opt: openServerFileManagementFieldType) => {
  openServerFileManagementField = opt;
  serverFileManagementDialogRef.value?.open();
}
const selectedFilesHandle = (slc: I_sfm_FileEntry[]) => {
  if (slc.length == 0) {
    return;
  }
  if (openServerFileManagementField == 'replacePath') {
    formData.value.replacePath = slc[slc.length - 1].path;
  } else {
    formData.value.searchPath = slc[slc.length - 1].path;
  }
}

const open = () => {
  init();
  dialogCommonRef.value?.open();
}
const close = () => {
  dialogCommonRef.value?.close()
}
// eslint-disable-next-line no-undef
defineExpose({ open, close })
</script>
<style lang="scss" scoped>
.path-replace-container {
  width: 100%;
  height: 460px;
  display: flex;
  flex-direction: column;

  .condition {
    flex-shrink: 0;

    .condition-item {
      display: flex;
      gap: 10px;
      padding-bottom: 10px;

      .condition-item-label {
        width: 160px;
        flex-shrink: 0;
        line-height: 32px;
        text-align: right;
      }

      .condition-item-content {
        flex-grow: 1;

        .result {
          color: #E6A23C;
        }
      }
    }
  }

  .result-list {
    flex-grow: 1;
    overflow-y: auto;
    list-style-type: none;

    li {
      border-radius: 5px;
      margin-bottom: 10px;
      padding: 5px;
      background-color: #404040;
      font-size: 12px;

      .title {
        font-size: 14px;
      }

      .path {
        color: #409EFF;
      }
    }
  }
}
</style>
