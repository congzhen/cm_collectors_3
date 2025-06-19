<template>
  <div class="database-setting">
    <div class="database-setting-btn">
      <el-button icon="Plus" type="success" plain>创建新文件数据库</el-button>
      <el-button type="warning" plain>路径替换</el-button>
      <el-button type="warning" plain>批量删除</el-button>
    </div>
    <el-tabs tab-position="left" class="setting-tabs" v-model="activeName">
      <el-tab-pane name="default" label="默认文件数据库">
        <fileSettingData></fileSettingData>
      </el-tab-pane>
      <el-tab-pane name="default1" label="中国数据库">用户管理</el-tab-pane>
      <el-tab-pane name="default2" label="主播数据库">管理工具</el-tab-pane>
      <el-tab-pane name="default3" label="写真">更新日志</el-tab-pane>
      <el-tab-pane name="default14" label="动漫">关于软件</el-tab-pane>
    </el-tabs>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import fileSettingData from './fileSettingData.vue';
const activeName = ref('default');
</script>
<style lang="scss" scoped></style>
