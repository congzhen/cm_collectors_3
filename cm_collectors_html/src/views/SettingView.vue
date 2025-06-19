<template>
  <div class="setting-view">
    <HeaderView class="header" :mode="E_headerMode.GoBack"></HeaderView>
    <div class="main">
      <el-tabs v-model="activeName" class="setting-tabs" type="border-card">
        <el-tab-pane label="软件设置" name="softSettings">
          <softSettings></softSettings>
        </el-tab-pane>
        <el-tab-pane label="文件数据库" name="fileDatabase">
          <fileDatabaseSetting></fileDatabaseSetting>
        </el-tab-pane>
        <el-tab-pane label="演员数据集" name="performerDatabase">
          <performerSetting></performerSetting>
        </el-tab-pane>
        <el-tab-pane label="标签数据集" name="tagDatabase">
          <tagSetting></tagSetting>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import HeaderView from './HeaderView.vue';
import { E_headerMode } from '@/dataType/app.dataType';
import softSettings from '@/components/setting/softSettings/softSettings.vue';
import fileDatabaseSetting from '@/components/setting/fileDatabaseSetting/fileDatabaseSetting.vue';
import performerSetting from '@/components/setting/performerDatabaseSetting/performerDatabaseSetting.vue';
import tagSetting from '@/components/setting/tagSetting/tagSetting.vue';
const activeName = ref('softSettings');
</script>
<style lang="scss" scoped>
.setting-view {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .main {
    flex-grow: 1;
    display: flex;
    overflow: hidden;
    padding: 10px;

    .setting-tabs {
      width: 100%;
      height: 100%;

      .el-tab-pane {
        height: 100%;
      }

      .el-tab-pane {
        height: 100%;
      }
    }
  }
}
</style>
