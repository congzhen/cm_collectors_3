<template>
  <div class="database-setting">
    <div class="database-setting-btn">
      <el-button icon="Plus" type="success" plain>创建新演员集</el-button>
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
</template>
<script setup lang="ts">
import { ref } from 'vue';
import performerSettingData from './performerSettingData.vue';
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';

const store = {
  performerBasesStoreData: performerBasesStoreData(),
}

const activeName = ref(store.performerBasesStoreData.activeFirstPerformerBases?.id);
</script>
<style lang="scss" scoped></style>
