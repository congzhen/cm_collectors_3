<template>
  <div class="database-setting">
    <el-tabs :tab-position="compact ? 'top' : 'left'" class="setting-tabs" v-model="activeName">
      <el-tab-pane name="general" label="通用设置">
        <generalSet v-if="activeName === 'general'"></generalSet>
      </el-tab-pane>
      <el-tab-pane name="videoPreviewCache" label="视频预览缓存">
        <VideoPreviewSetting v-if="activeName === 'videoPreviewCache'" />
      </el-tab-pane>
      <!--
      <el-tab-pane name="user" label="用户管理">
        <h3>未开放</h3>
      </el-tab-pane>
      -->
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
import { ref, onMounted, onBeforeUnmount } from 'vue';
import VideoPreviewSetting from '../VideoPreviewSetting.vue';
import generalSet from './generalSet.vue';
import updateLog from './updateLog.vue';
import aboutSoftware from './aboutSoftware.vue';
const activeName = ref('general');
const compactQuery = window.matchMedia('(max-width: 700px)');
const compact = ref(compactQuery.matches);
const syncCompact = () => { compact.value = compactQuery.matches; };
onMounted(() => compactQuery.addEventListener('change', syncCompact));
onBeforeUnmount(() => compactQuery.removeEventListener('change', syncCompact));
</script>
<style lang="scss" scoped>
@media (max-width: 700px) {
  .setting-tabs {
    display: flex;
    flex-direction: column;
    :deep(> .el-tabs__header) {
      flex-shrink: 0;
      margin: 0 0 12px;
      .el-tabs__item { padding: 0 12px; font-size: 12px; }
    }
    :deep(> .el-tabs__content) { flex: 1; min-height: 0; }
  }
}
</style>
