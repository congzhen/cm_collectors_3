<template>
  <div class="database-setting">
    <el-tabs tab-position="left" class="setting-tabs" v-model="activeName">
      <el-tab-pane name="general" label="通用设置">
        <general class="tab-content-main"></general>
      </el-tab-pane>
      <el-tab-pane name="user" label="用户管理">用户管理</el-tab-pane>
      <el-tab-pane name="managementTool" label="管理工具">管理工具</el-tab-pane>
      <el-tab-pane name="updateLog" label="更新日志">更新日志</el-tab-pane>
      <el-tab-pane name="about" label="关于软件">关于软件</el-tab-pane>
    </el-tabs>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import general from './general.vue';
const activeName = ref('general');
</script>
<style lang="scss" scoped></style>
