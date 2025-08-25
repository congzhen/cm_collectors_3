<template>
  <el-timeline class="update-log-timeline">
    <el-scrollbar height="100%">
      <!--
      <el-timeline-item timestamp="2025/8/20" placement="top">
        <h4 class="version-title">版本: v 2.0.7</h4>
        <div class="rawContent">
          <h5 class="section-title">新特性</h5>
          <p class="logContent">示例1</p>
          <p class="logContent">示例2</p>
        </div>
        <div class="rawContent">
          <h5 class="section-title bug-fix">BUG修复</h5>
          <p class="logContent">示例3</p>
          <p class="logContent">示例4</p>
        </div>
      </el-timeline-item>
      -->
      <el-timeline-item center timestamp="2025/9/1" placement="top">
        <h4 class="version-title">版本: v 3.0.0</h4>
        <div class="rawContent">
          <h5 class="section-title">新特性</h5>
          <p class="logContent">v3版本发布</p>
        </div>
      </el-timeline-item>
    </el-scrollbar>
  </el-timeline>
</template>
<script setup lang="ts">
</script>
<style scoped lang="scss">
.update-log-timeline {
  width: 100%;
  height: 100%;
  padding: 20px;

  :deep(.el-timeline-item) {
    .el-timeline-item__timestamp {
      color: #409eff;
      font-weight: bold;
      font-size: 14px;
    }

    .el-timeline-item__wrapper {
      padding-bottom: 20px;
    }

    .el-timeline-item__tail {
      border-color: rgba(64, 158, 255, 0.3);
    }

    .el-timeline-item__node--normal {
      background-color: #409eff;
      border-color: #409eff;
    }
  }

  .version-title {
    color: #f2f2f2;
    font-size: 16px;
    margin-bottom: 10px;
    padding: 4px 8px;
    background: rgba(64, 158, 255, 0.15);
    border-radius: 6px;
    display: inline-block;
  }

  .rawContent {
    background: rgba(0, 0, 0, 0.1);
    border-radius: 8px;
    padding: 5px 15px;
    margin-bottom: 15px;

    .section-title {
      color: #409eff;
      font-size: 14px;
      margin: 5px 0;
      padding-bottom: 5px;
      border-bottom: 1px solid rgba(64, 158, 255, 0.3);

      &.optimization {
        color: #67c23a;
      }

      &.bug-fix {
        color: #f56c6c;
      }
    }

    .logContent {
      color: #f2f2f2;
      font-size: 14px;
      line-height: 1.8;
      margin-bottom: 10px;
      padding-left: 15px;
      position: relative;

      &::before {
        content: "•";
        color: #409eff;
        position: absolute;
        left: 0;
        top: 0;
      }
    }
  }
}
</style>
