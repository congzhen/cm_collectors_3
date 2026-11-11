<template>
  <div class="setting-view">
    <HeaderView class="header" :mode="E_headerMode.GoBack"></HeaderView>
    <div class="main">
      <el-tabs v-model="activeName" class="setting-tabs settings-primary-tabs">
        <el-tab-pane label="软件设置" name="softSettings">
          <softSettings v-if="activeName === 'softSettings'"></softSettings>
        </el-tab-pane>
        <el-tab-pane label="文件数据库" name="fileDatabase">
          <fileDatabaseSetting v-if="activeName === 'fileDatabase'"></fileDatabaseSetting>
        </el-tab-pane>
        <el-tab-pane label="演员数据库" name="performerDatabase">
          <performerSetting v-if="activeName === 'performerDatabase'"></performerSetting>
        </el-tab-pane>
        <el-tab-pane label="演员头像库" name="performerAvatarLibrary">
          <performerAvatarLibrarySetting v-if="activeName === 'performerAvatarLibrary'" />
        </el-tab-pane>
        <el-tab-pane label="标签数据库" name="tagDatabase">
          <tagSetting v-if="activeName === 'tagDatabase'"></tagSetting>
        </el-tab-pane>
        <el-tab-pane label="AI 自动标签" name="aiTag">
          <aiTagSetting v-if="activeName === 'aiTag'"></aiTagSetting>
        </el-tab-pane>
        <el-tab-pane label="视频信息采集" name="videoMetadata">
          <videoMetadataSetting v-if="activeName === 'videoMetadata'" />
        </el-tab-pane>
        <el-tab-pane label="视频转码与剪辑" name="videoTranscode">
          <videoTranscodeSetting v-if="activeName === 'videoTranscode'" />
        </el-tab-pane>
        <el-tab-pane label="计划任务" name="cronJobs">
          <cronJobs v-if="activeName === 'cronJobs'"></cronJobs>
        </el-tab-pane>
        <el-tab-pane label="视频重复检测" name="duplicateDetection">
          <duplicateDetection v-if="activeName === 'duplicateDetection'" @switch-tab="switchTab"></duplicateDetection>
        </el-tab-pane>
        <el-tab-pane label="TVBox推荐" name="tvboxRecommend">
          <tvboxRecommendSetting v-if="activeName === 'tvboxRecommend'"></tvboxRecommendSetting>
        </el-tab-pane>
        <el-tab-pane label="数据库清理" name="databaseCleanup">
          <databaseCleanup v-if="activeName === 'databaseCleanup'"></databaseCleanup>
        </el-tab-pane>
        <el-tab-pane label="自动备份" name="autoBackup">
          <autoBackup v-if="activeName === 'autoBackup'"></autoBackup>
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
import performerAvatarLibrarySetting from '@/components/setting/performerAvatarLibrary/performerAvatarLibrarySetting.vue';
import tagSetting from '@/components/setting/tagSetting/tagSetting.vue';
import aiTagSetting from '@/components/setting/aiTag/aiTagSetting.vue';
import videoMetadataSetting from '@/components/setting/videoMetadata/videoMetadataSetting.vue';
import videoTranscodeSetting from '@/components/setting/videoTranscode/videoTranscodeSetting.vue';
import cronJobs from '@/components/setting/cronJobs/cronJobs.vue';
import duplicateDetection from '@/components/setting/duplicateDetection/duplicateDetection.vue';
import databaseCleanup from '@/components/setting/databaseCleanup/databaseCleanup.vue';
import autoBackup from '@/components/setting/autoBackup/autoBackup.vue';
import tvboxRecommendSetting from '@/components/setting/tvboxRecommendSetting/tvboxRecommendSetting.vue';
const activeName = ref('softSettings');
const switchTab = (tabName: string) => {
  activeName.value = tabName;
}
</script>
<style lang="scss" scoped>
.setting-view {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color-page);
  color: var(--el-text-color-primary);
  .header { flex-shrink: 0; }
}
.main {
  flex: 1;
  min-height: 0;
  display: flex;
  overflow: hidden;
  padding: 8px 8px 10px;
}
.settings-primary-tabs {
  width: 100%;
  height: 100%;
  min-width: 0;
  display: flex;
  flex-direction: column;
  :deep(> .el-tabs__header) {
    flex-shrink: 0;
    margin: 0;
    padding: 4px 12px;
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 12px;
    .el-tabs__nav-wrap::after { display: none; }
    .el-tabs__active-bar { display: none; }
    .el-tabs__item { height: 38px; padding: 0 16px !important; margin: 2px 2px 2px 0; border-radius: 8px; color: var(--el-text-color-regular); }
    .el-tabs__item:hover { color: var(--el-color-primary); background: var(--el-fill-color-light); }
    .el-tabs__item.is-active { color: var(--el-color-primary); background: var(--el-color-primary-light-9); font-weight: 600; }
  }
  :deep(> .el-tabs__content) { flex: 1; min-height: 0; padding: 8px 0 0; overflow: hidden; }
  :deep(> .el-tabs__content > .el-tab-pane) { height: 100%; }
  :deep(.database-setting > .setting-tabs) {
    min-height: 0;
    > .el-tabs__header.is-left {
      margin-right: 8px;
      padding: 8px;
      border-radius: 12px;
      background: var(--el-bg-color);
      .el-tabs__nav-wrap::after, .el-tabs__active-bar { display: none; }
      .el-tabs__item { text-align: left; justify-content: flex-start; padding: 0 16px; margin-bottom: 4px; border-radius: 8px; color: var(--el-text-color-regular); }
      .el-tabs__item.is-active { color: var(--el-color-primary); background: var(--el-color-primary-light-9); }
    }
    > .el-tabs__content { height: 100%; min-width: 0; }
  }
}
@media (max-width: 700px) {
  .main { padding: 12px 10px 10px; }
  .settings-primary-tabs :deep(> .el-tabs__content) { padding-top: 12px; }
  .settings-primary-tabs :deep(.database-setting > .setting-tabs > .el-tabs__header.is-left) { margin-right: 10px; padding: 4px; .el-tabs__item { padding: 0 8px; font-size: 12px; } }
}
// File-library navigation and editor share one panel; other settings keep their own layouts.
.settings-primary-tabs :deep(.database-setting.file-library-setting > .setting-tabs) {
  overflow: hidden;
  box-sizing: border-box;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  background: var(--el-bg-color);
  > .el-tabs__header.is-left {
    margin: 0;
    padding: 8px;
    min-width: 128px;
    max-width: 200px;
    box-sizing: border-box;
    border-right: 1px solid var(--el-border-color-lighter);
    border-radius: 0;
    background: var(--el-fill-color-extra-light);
  }
  > .el-tabs__content { padding: 0; }
}
@media (max-width: 700px) {
  .settings-primary-tabs :deep(.database-setting.file-library-setting > .setting-tabs > .el-tabs__header.is-left) {
    min-width: 88px;
    max-width: 110px;
    padding: 8px 4px;
  }
}
</style>
