<template>
  <div class="database-setting">
    <el-tabs tab-position="left" class="setting-tabs" v-model="activeName">
      <el-tab-pane name="default" label="中国数据库">
        <tagList></tagList>
      </el-tab-pane>
      <el-tab-pane name="default1" label="日本数据库">用户管理</el-tab-pane>
      <el-tab-pane name="default2" label="主播数据库">管理工具</el-tab-pane>
      <el-tab-pane name="default3" label="写真">更新日志</el-tab-pane>
      <el-tab-pane name="default14" label="动漫">关于软件</el-tab-pane>
    </el-tabs>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import tagList from '@/components/tag/tagList.vue';
const activeName = ref('default');
</script>
<style lang="scss" scoped></style>
