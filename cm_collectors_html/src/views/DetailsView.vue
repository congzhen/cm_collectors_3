<template>
  <div class="details-view">
    <div class="content-cover">
      <el-image
        src="https://picx.zhimg.com/v2-d6f44389971daab7e688e5b37046e4e4_720w.jpg"
        fit="cover"
      />
    </div>
    <div class="tool">
      <el-button-group>
        <el-button icon="VideoPlay" />
        <el-button icon="Folder" />
        <el-button icon="Edit" />
        <el-button icon="Delete" />
      </el-button-group>
    </div>
    <div class="title">
      こんなに甘える本郷愛、見たことない！
      プライベート丸出しで2人きりで一晩中ヤリまくった生々ハメ撮りFUCK
    </div>
    <div class="info-base">
      <div class="info-base-item">版号、番号、刊号: SAME-101</div>
      <div class="info-base-item">
        <el-breadcrumb separator="|">
          <el-breadcrumb-item>年份: 2024-03-01 </el-breadcrumb-item>
          <el-breadcrumb-item>国家:日本 </el-breadcrumb-item>
          <el-breadcrumb-item>清晰度:1080P</el-breadcrumb-item>
        </el-breadcrumb>
      </div>
      <div class="info-base-item">收录时间: 2024-06-16 10:53:42</div>
    </div>
    <div class="info-block">
      <el-alert class="tagAlert" title="演员" type="success" :closable="false" />
    </div>
    <div class="info-block">
      <el-alert class="tagAlert" title="标签" type="warning" :closable="false" />
      <div class="tag-list">
        <el-tag type="info" effect="plain" size="large"> 有码 </el-tag>
        <el-tag type="info" effect="plain" size="large"> 高帧率 </el-tag>
      </div>
    </div>
    <div class="info-block">
      <el-alert class="tagAlert" title="摘要" type="info" :closable="false" />
      <div class="abstract">
        在给文档站做升级之前，每一次文档改动都需要改五种语言，大部分的贡献者都只会自己的母语和一门外语，大部分的贡献者也会寻求翻译器来翻译其它自己不熟悉的外语，而翻译器翻译的内容对母语者可能造成误解，并且对贡献者来说也是一件麻烦的事情。
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup></script>
<style lang="scss" scoped>
.details-view {
  width: 278px;
  height: 100%;
  padding-left: 2px;

  .content-cover {
    width: 100%;
    overflow: hidden;
    .el-image {
      width: 100%;
    }
  }

  .tool {
    width: 100%;

    padding-bottom: 5px;
    .el-button-group {
      width: 100%;
      display: flex;
      justify-content: center;
      .el-button {
        width: 25%;
      }
    }
  }

  .title {
    font-size: 1.1em;
    font-weight: 500;
    line-height: 1.2em;
  }

  .info-base {
    font-size: 12px;
    line-height: 1.5em;
    padding: 5px 0;
    color: #a8abb2;
    :deep(.el-breadcrumb) {
      .el-breadcrumb__inner {
        font-size: 12px;
        color: #a8abb2;
      }
    }
  }

  .info-block {
    padding: 5px 0;
    .el-alert {
      padding: 4px 8px;
    }
    .tag-list {
      padding: 5px;
      display: flex;
      flex-wrap: wrap;
      gap: 5px;
    }
    .abstract {
      text-indent: 2em;
      padding: 10px;
      color: #a8abb2;
    }
  }
}
</style>
