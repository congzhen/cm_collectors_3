<template>
  <div class="database-setting">
    <div class="database-setting-btn">
      <el-button icon="Plus" type="success" @click="createNewPerformerBases" plain>创建新演员集</el-button>
      <el-button type="warning" plain>导入</el-button>
      <el-button type="warning" plain>导出</el-button>
    </div>
    <el-tabs tab-position="left" class="setting-tabs" v-model="activeName">
      <el-tab-pane v-for="item, key in store.performerBasesStoreData.performerBases" :key="key" :name="item.id"
        :label="item.name">
        <performerSettingData v-if="activeName === item.id" :performer-bases="item"></performerSettingData>
      </el-tab-pane>
    </el-tabs>
  </div>
  <performerDatabaseFormDialog ref="performerDatabaseFormDialogRef"></performerDatabaseFormDialog>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import performerSettingData from './performerSettingData.vue';
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';
import performerDatabaseFormDialog from './performerDatabaseFormDialog.vue';
const store = {
  performerBasesStoreData: performerBasesStoreData(),
}
const performerDatabaseFormDialogRef = ref<InstanceType<typeof performerDatabaseFormDialog>>();
const activeName = ref(store.performerBasesStoreData.activeFirstPerformerBases?.id);


const createNewPerformerBases = () => {
  performerDatabaseFormDialogRef.value?.open();
}
</script>
<style lang="scss" scoped></style>
