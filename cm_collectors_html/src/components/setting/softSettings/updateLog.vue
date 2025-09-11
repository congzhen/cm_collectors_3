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
      <el-timeline-item center timestamp="2025/9/10" placement="top">
        <h4 class="version-title">版本: v 3.0.1</h4>
        <div class="rawContent">
          <h5 class="section-title">新特性</h5>
          <p class="logContent">新增全新的 Web 端播放控制器</p>
          <p class="logContent">播放控制器支持直接调用本地播放器播放视频</p>
          <p class="logContent">播放控制器新增最大化播放功能</p>
          <p class="logContent">新增短视频形式的资源展示模式</p>
          <p class="logContent">新增资源右键菜单功能</p>
          <p class="logContent">新增瀑布流布局、短视频布局、图集的列数以及漫画图片宽度的本地记忆功能</p>
          <p class="logContent">优化移动端播放布局与交互体验</p>
          <p class="logContent">文件管理器新增按文件名和修改时间排序功能</p>
          <p class="logContent">新增挂载指定磁盘和网络地址的使用说明</p>
          <p class="logContent">新增右键资源快捷打标签功能</p>
          <p class="logContent">新增资源导入命名模式与封面海报的命名匹配功能</p>
          <p class="logContent">新增同文件夹下视频导入为剧集功能</p>
          <p class="logContent">新增服务器打开文件与文件夹的权限设置</p>
          <p class="logContent">优化视频播放热度统计，并新增资源显示属性 hot</p>
          <p class="logContent">新增 NFO 格式文件导入资源信息功能</p>
          <p class="logContent">资源编辑页面的资源列表允许通过拖拽方式进行重新排序</p>
        </div>
        <div class="rawContent">
          <h5 class="section-title bug-fix">BUG修复</h5>
          <p class="logContent">修复文件管理器在 Windows 上无法正确处理 Junction 的问题</p>
          <p class="logContent">修复移动端播放视频时地址获取错误的问题</p>
          <p class="logContent">修复封面瀑布流多次加载后出现重叠的问题</p>
          <p class="logContent">修复视频播放音量状态无法保存和获取的问题</p>
          <p class="logContent">修复TVBOX资源推荐失败的问题</p>
        </div>
      </el-timeline-item>
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
