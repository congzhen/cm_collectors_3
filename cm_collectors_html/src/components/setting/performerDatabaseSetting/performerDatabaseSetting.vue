<template>
  <div class="database-setting">
    <div class="database-setting-btn">
      <el-button icon="Plus" type="success" plain>创建新演员集</el-button>
      <el-button type="warning" plain>导入</el-button>
      <el-button type="warning" plain>导出</el-button>
    </div>
    <el-tabs tab-position="left" class="setting-tabs" v-model="activeName">
      <el-tab-pane name="default" label="默认演员集">
        <performerSettingData></performerSettingData>
      </el-tab-pane>
      <el-tab-pane name="default1" label="日本AV女优演员集">用户管理</el-tab-pane>
      <el-tab-pane name="default2" label="中国AV女优演员集">管理工具</el-tab-pane>
      <el-tab-pane name="default3" label="主播演员集">更新日志</el-tab-pane>
      <el-tab-pane name="default14" label="写真演员集">关于软件</el-tab-pane>
    </el-tabs>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import performerSettingData from './performerSettingData.vue';
const activeName = ref('default');
</script>
<style lang="scss" scoped></style>
