<template>
  <div class="database-setting">
    <el-tabs tab-position="left" class="setting-tabs" v-model="activeName">
      <el-tab-pane name="general" label="通用设置">
        <generalSet v-if="activeName === 'general'"></generalSet>
      </el-tab-pane>
      <el-tab-pane name="user" label="用户管理">
        <h3>未开放</h3>
      </el-tab-pane>
      <el-tab-pane name="updateLog" label="更新日志">
        <updateLog v-if="activeName === 'updateLog'" />
      </el-tab-pane>
      <el-tab-pane name="about" label="关于软件">
        <aboutSoftware v-if="activeName === 'about'" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import generalSet from './generalSet.vue';
import updateLog from './updateLog.vue';
import aboutSoftware from './aboutSoftware.vue';
const activeName = ref('general');
</script>
<style lang="scss" scoped></style>
