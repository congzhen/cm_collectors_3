<template>
  <div class="database-setting">
    <div class="database-setting-btn">
      <el-button icon="Plus" type="success" @click="createNewPerformerBases" plain>创建新演员集</el-button>
      <el-button type="warning" plain @click="importHandle">导入</el-button>
      <el-button type="warning" plain @click="exportHandle">导出</el-button>
    </div>
    <el-tabs tab-position="left" class="setting-tabs" v-model="activeName">
      <el-tab-pane v-for="item, key in store.performerBasesStoreData.performerBases" :key="key" :name="item.id"
        :label="item.name">
        <performerSettingData v-if="activeName === item.id" :performer-bases="item"></performerSettingData>
      </el-tab-pane>
    </el-tabs>
  </div>
  <performerDatabaseFormDialog ref="performerDatabaseFormDialogRef"></performerDatabaseFormDialog>
  <exportPerformerDatabaseDialog ref="exportPerformerDatabaseDialogRef"></exportPerformerDatabaseDialog>
  <importPerformerDatabaseDialog ref="importPerformerDatabaseDialogRef"></importPerformerDatabaseDialog>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import performerSettingData from './performerSettingData.vue';
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';
import performerDatabaseFormDialog from './performerDatabaseFormDialog.vue';
import exportPerformerDatabaseDialog from './exportPerformerDatabaseDialog.vue';
import importPerformerDatabaseDialog from './importPerformerDatabaseDialog.vue';
const store = {
  performerBasesStoreData: performerBasesStoreData(),
}
const performerDatabaseFormDialogRef = ref<InstanceType<typeof performerDatabaseFormDialog>>();
const exportPerformerDatabaseDialogRef = ref<InstanceType<typeof exportPerformerDatabaseDialog>>();
const importPerformerDatabaseDialogRef = ref<InstanceType<typeof importPerformerDatabaseDialog>>();
const activeName = ref(store.performerBasesStoreData.activeFirstPerformerBases?.id);


const createNewPerformerBases = () => {
  performerDatabaseFormDialogRef.value?.open();
}

const exportHandle = () => {
  exportPerformerDatabaseDialogRef.value?.open();
}

const importHandle = () => {
  importPerformerDatabaseDialogRef.value?.open();
}
</script>
<style lang="scss" scoped></style>
