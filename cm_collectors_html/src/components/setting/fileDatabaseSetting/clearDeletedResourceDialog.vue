<template>
  <dialogCommon ref="dialogCommonRef" title="清理已删除资源" btnSubmitTitle="批量删除" @submit="deleteHandle">
    <div class="clear-delete-resource-container" v-loading="loading">
      <div class="condition">
        <div class="condition-item">
          <div class="condition-item-label">清理数据库:</div>
          <div class="condition-item-content">
            <el-select v-model="formData.filesBasesIds" placeholder="全部数据库" multiple>
              <el-option v-for="item, index in store.filesBasesStoreData.filesBases" :key="index" :label="item.name"
                :value="item.id"></el-option>
            </el-select>
            <el-button icon="Search" @click="searchHandle">查询</el-button>
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
        </li>
      </ul>
    </div>
  </dialogCommon>
</template>
<script lang="ts" setup>
import dialogCommon from '@/components/com/dialog/dialog-common.vue';
import { ref } from 'vue';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
import { debounceNow } from '@/assets/debounce';

import type { I_resource } from '@/dataType/resource.dataType';
import { ElMessage, ElMessageBox } from 'element-plus';
import { resourceServer } from '@/server/resource.server';
import { resourceBatchDelete } from '@/common/resource';

const store = {
  filesBasesStoreData: filesBasesStoreData(),
}

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();
const loading = ref(false);
const formData = ref({
  filesBasesIds: [],
})

const dataList = ref<I_resource[]>([])

const init = () => {
  dialogCommonRef.value?.disabledSubmit(true)
  formData.value.filesBasesIds = [];
  dataList.value = [];
}

const searchHandle = debounceNow(async () => {
  try {
    loading.value = true;
    const result = await resourceServer.deletedDataList(formData.value.filesBasesIds)
    if (result.status) {
      dataList.value = result.data;
      if (dataList.value.length > 0) {
        dialogCommonRef.value?.disabledSubmit(false)
      } else {
        dialogCommonRef.value?.disabledSubmit(true)
      }
    } else {
      ElMessage.error(result.msg);
    }
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
  } catch (error) {
    ElMessage.error('检索失败');
  } finally {
    loading.value = false;
  }
})

const deleteHandle = debounceNow(async () => {
  try {
    loading.value = true;
    if (dataList.value.length === 0) {
      ElMessage.error('没有可删除的资源');
      return;
    }
    resourceBatchDelete(dataList.value, () => {
      ElMessageBox.alert('删除成功,刷新资源后生效', {})
      close();
    })
  } catch (error) {
    ElMessage.error('删除失败:' + String(error));
  } finally {
    loading.value = false;
  }
})

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
.clear-delete-resource-container {
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
        display: flex;
        gap: 10px;

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
