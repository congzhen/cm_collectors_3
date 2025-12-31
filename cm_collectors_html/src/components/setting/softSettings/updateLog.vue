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
      <el-timeline-item center timestamp="2026/1/1" placement="top">
        <h4 class="version-title">版本: v 3.0.8</h4>
        <div class="rawContent">
          <h5 class="section-title">新特性</h5>
          <p class="logContent">新增应用界面缩放功能</p>
          <p class="logContent">优化 Windows 资源管理器打开功能</p>
          <p class="logContent">新增 Windows 端启动配置项，支持设置启动默认是否打开应用界面</p>
          <p class="logContent">新增资源批量关联演员功能</p>
          <p class="logContent">新增演员数据集导入导出功能</p>
          <p class="logContent">新增文件管理器搜索结果中文件选择功能，增强操作便捷性</p>
          <p class="logContent">新增资源随机排序展示选项</p>
          <p class="logContent">优化演员管理功能，将编辑、删除等操作迁移至右键菜单</p>
          <p class="logContent">新增演员数据集迁移功能，支持数据在不同集合间转移</p>
        </div>
        <div class="rawContent">
          <h5 class="section-title bug-fix">BUG修复</h5>
          <p class="logContent">修复应用端管理员登录功能异常问题</p>
          <p class="logContent">修复云播设置中应用参数配置错误</p>
        </div>
      </el-timeline-item>
      <el-timeline-item center timestamp="2025/12/15" placement="top">
        <h4 class="version-title">版本: v 3.0.7</h4>
        <div class="rawContent">
          <h5 class="section-title">新特性</h5>
          <p class="logContent">云播功能优化：默认使用文件名作为播放标题</p>
          <p class="logContent">定时任务增强：支持立即执行已设定的定时任务</p>
          <p class="logContent">刮削功能改进：增加未搜索到ID时的错误处理机制</p>
          <p class="logContent">资源管理界面升级：新增文件图标及优化对话框交互体验</p>
          <p class="logContent">用户体验优化：新增一键返回页面顶部功能</p>
          <p class="logContent">演员模块增强：演员卡片显示该演员关联的资源数量</p>
          <p class="logContent">演员管理完善：新增彻底删除演员数据的功能</p>
          <p class="logContent">NFO导入智能匹配：支持模糊匹配和随机匹配模式</p>
          <p class="logContent">界面个性化：新增明亮主题模式，支持深色/明亮主题自由切换</p>
          <p class="logContent">标题提取强化：字母提取支持日文假名</p>
          <p class="logContent">资源排序自定义：支持配置资源排序</p>
          <p class="logContent">资源推荐功能："随便看看"、"历史记录"、"热门资源" 模块</p>
          <p class="logContent">云播配置细化：新增云播功能相关配置选项</p>
        </div>
        <div class="rawContent">
          <h5 class="section-title bug-fix">BUG修复</h5>
          <p class="logContent">修复导入导出功能在软件端无法正常打开文件的问题</p>
          <p class="logContent">修复资源排序逻辑错误导致排序结果不正确的问题</p>
          <p class="logContent">修复移动端部分页面元素显示样式异常的问题</p>
        </div>
      </el-timeline-item>
      <el-timeline-item center timestamp="2025/12/01" placement="top">
        <h4 class="version-title">版本: v 3.0.6</h4>
        <div class="rawContent">
          <h5 class="section-title">新特性</h5>
          <p class="logContent">新增计划任务功能</p>
          <p class="logContent">新增数据库清理功能</p>
          <p class="logContent">新增标签导入导出功能</p>
          <p class="logContent">新增资源右键添加演员功能</p>
          <p class="logContent">新增文件数据库配置导入导出功能</p>
          <p class="logContent">新增移动端与PC端切换控制</p>
          <p class="logContent">新增导演信息展示</p>
          <p class="logContent">新增标签右键以及真实删除功能</p>
          <p class="logContent">新增封面标签字体大小与资源对齐方式控制</p>
          <p class="logContent">新增资源与演员刮削配置保存功能</p>
        </div>
        <div class="rawContent">
          <h5 class="section-title bug-fix">BUG修复</h5>
          <p class="logContent">修复文件管理器路径拼接错误</p>
          <p class="logContent">修复云播插件指定播放器路径错误</p>
          <p class="logContent">修复资源清理资源路径问题以确保跨平台兼容性</p>
          <p class="logContent">修复导入nfo值获取逻辑以支持更复杂的嵌套结构</p>
        </div>
      </el-timeline-item>
      <el-timeline-item center timestamp="2025/11/11" placement="top">
        <h4 class="version-title">版本: v 3.0.5</h4>
        <div class="rawContent">
          <h5 class="section-title">新特性</h5>
          <p class="logContent">优化视频播放页面背景图片加载及模糊遮罩效果</p>
          <p class="logContent">新增资源评分功能</p>
          <p class="logContent">新增 HLS 视频播放支持，包括 m3u8 视频播放、HLS 视频转码、自动切换播放类型及云播类型选择功能</p>
          <p class="logContent">新增服务器文件管理挂载路径配置功能</p>
          <p class="logContent">新增 Linux 视频调用器支持</p>
          <p class="logContent">新增封面海报精简和精简扩展展示模式</p>
          <p class="logContent">新增内容区域左右内边距设置功能</p>
          <p class="logContent">新增 Windows 托盘菜单自定义功能，支持设置自定义浏览器打开</p>
          <p class="logContent">新增清理已删除资源的功能</p>
          <p class="logContent">优化配置合并逻辑，支持配置优先级控制</p>
          <p class="logContent">新增刮削器自定义浏览器路径功能，支持其他 Chromium 内核浏览器</p>
          <p class="logContent">新增导入方式"追加导入"和"覆盖导入"，并优化剧集导入逻辑</p>
          <p class="logContent">新增粘贴图片自动将剪切板图片打开到裁剪器功能</p>
          <p class="logContent">优化资源文件夹打开逻辑</p>
          <p class="logContent">优化资源搜索逻辑</p>
          <p class="logContent">支持点击详情中的标签设置搜索条件</p>
          <p class="logContent">新增按发行日期排序功能</p>
          <p class="logContent">重构标签列表为独立组件并优化标签排序逻辑</p>
        </div>
        <div class="rawContent">
          <h5 class="section-title bug-fix">BUG修复</h5>
          <p class="logContent">修复演员默认图片显示错误的问题</p>
          <p class="logContent">修复对话框宽度计算逻辑</p>
          <p class="logContent">修复资源文件夹打开逻辑问题</p>
          <p class="logContent">修复封面标签排序功能异常的问题</p>
          <p class="logContent">修复添加资源封面海报模式默认值设置问题</p>
        </div>
      </el-timeline-item>
      <el-timeline-item center timestamp="2025/10/27" placement="top">
        <h4 class="version-title">版本: v 3.0.4</h4>
        <div class="rawContent">
          <h5 class="section-title">新特性</h5>
          <p class="logContent"></p>
          <p class="logContent">新增播放列表功能</p>
          <p class="logContent">新增演员多库数据显示支持</p>
          <p class="logContent">新增按字母筛选演员功能</p>
          <p class="logContent">新增单个演员信息刮削功能</p>
          <p class="logContent">添加封面标签自定义排序功能</p>
          <p class="logContent">新增批量增删标签功能</p>
          <p class="logContent">优化封面图像尺寸计算</p>
          <p class="logContent">添加封面图片填充方式设置</p>
          <p class="logContent">添加封面标题对齐方式设置</p>
        </div>
        <div class="rawContent">
          <h5 class="section-title bug-fix">BUG修复</h5>
          <p class="logContent">修复右键菜单位置问题</p>
          <p class="logContent">修复表单label触发表单组件的问题</p>
        </div>
      </el-timeline-item>
      <el-timeline-item center timestamp="2025/10/17" placement="top">
        <h4 class="version-title">版本: v 3.0.3</h4>
        <div class="rawContent">
          <h5 class="section-title">新特性</h5>
          <p class="logContent">优化视频流传输性能及缓存机制</p>
          <p class="logContent">新增视频限流器，提升带宽控制能力</p>
          <p class="logContent">新增视频与演员信息刮削器</p>
          <p class="logContent">新增刮削器配置调试工具</p>
          <p class="logContent">新增路径替换功能，支持通过文件管理器选择路径</p>
          <p class="logContent">新增剧照展示功能</p>
          <p class="logContent">新增宽图盒子资源显示模式</p>
          <p class="logContent">新增快捷调整资源显示方式的功能</p>
          <p class="logContent">更新Dockerfile以适配刮削器功能</p>
        </div>
        <div class="rawContent">
          <h5 class="section-title bug-fix">BUG修复</h5>
          <p class="logContent">修复部分视频因特殊编码格式导致误转码的问题</p>
          <p class="logContent">修复移动端页面样式异常问题</p>
        </div>
      </el-timeline-item>
      <el-timeline-item center timestamp="2025/9/23" placement="top">
        <h4 class="version-title">版本: v 3.0.2</h4>
        <div class="rawContent">
          <h5 class="section-title">新特性</h5>
          <p class="logContent">新增云播放功能（推荐使用 PotPlayer 播放器以获得最佳体验）</p>
          <p class="logContent">新增云播放协议脚本</p>
          <p class="logContent">新增 Docker 容器化部署</p>
          <p class="logContent">优化 NFO 文件导入性能，移除链接尝试下载功能</p>
          <p class="logContent">新增数据库迁移过程的错误日志记录</p>
        </div>
      </el-timeline-item>
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
