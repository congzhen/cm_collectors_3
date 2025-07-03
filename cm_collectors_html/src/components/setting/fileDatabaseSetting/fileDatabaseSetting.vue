<template>
  <div class="database-setting">
    <div class="database-setting-btn">
      <el-button icon="Plus" type="success" plain>创建新文件数据库</el-button>
      <el-button type="warning" plain>路径替换</el-button>
      <el-button type="warning" plain>批量删除</el-button>
    </div>
    <el-tabs tab-position="left" class="setting-tabs" v-model="activeName">
      <el-tab-pane v-for="item, key in store.filesBasesStoreData.filesBases" :key="key" :name="item.id"
        :label="item.name">
        <fileSettingData v-if="activeName === item.id" :filesBasesId="item.id" @set-success="setSuccessHandle">
        </fileSettingData>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import fileSettingData from './fileSettingData.vue';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
import { LoadingService } from '@/assets/loading';
import { appStoreData } from '@/storeData/app.storeData';
import { ElMessage } from 'element-plus';

const store = {
  appStoreData: appStoreData(),
  filesBasesStoreData: filesBasesStoreData(),
}
const activeName = ref(store.filesBasesStoreData.filesBasesFirst?.id);


const setSuccessHandle = async (filesBasesId: string) => {
  try {
    LoadingService.show()
    const result = await store.appStoreData.initApp();
    if (result && !result.status) {
      ElMessage.error(result.message);
      return
    }
    if (filesBasesId == store.appStoreData.currentFilesBases.id) {
      const result = await store.appStoreData.initCurrentFilesBases(filesBasesId)
      if (result && !result.status) {
        ElMessage.error(result.message);
        return
      }
    }
  } catch (err) {
    console.log(err)
  } finally {
    LoadingService.hide()
  }
}

</script>
<style lang="scss" scoped></style>
