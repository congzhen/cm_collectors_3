<template>
  <div class="performer-info">
    <div class="performer-cover">
      <el-image
        src="https://picx.zhimg.com/v2-d6f44389971daab7e688e5b37046e4e4_720w.jpg"
        fit="cover"
      />
    </div>
    <div class="performer-info-data">
      <div class="performer-name">櫻茉日</div>
      <div><el-rate disabled /></div>
      <ul class="performer-info-ul">
        <li>别名：堀北実来、堀北实来</li>
        <li>国籍：日本</li>
        <li>罩杯：E-Cup</li>
        <li>
          <el-breadcrumb>
            <el-breadcrumb-item> <span>胸围：</span><label>101 </label> </el-breadcrumb-item>
            <el-breadcrumb-item> <span>腰围：</span><label>58 </label> </el-breadcrumb-item>
            <el-breadcrumb-item> <span>臀围：</span><label>90 </label> </el-breadcrumb-item>
          </el-breadcrumb>
        </li>
        <li>出生日期：2000-03-03</li>
        <li>身高：160cm</li>
        <li>
          简介：在给文档站做升级之前，每一次文档改动都需要改五种语言，大部分的贡献者都只会自己的母语和一门外语，大部分的贡献者也会寻求翻译器来翻译其它自己不熟悉的外语，而翻译器翻译的内容对母语者可能造成误解，并且对贡献者来说也是一件麻烦的事情。
        </li>
      </ul>
    </div>
  </div>
</template>
<script lang="ts" setup></script>
<style lang="scss" scoped>
.performer-info {
  .performer-cover {
    width: 100%;
    flex-shrink: 0;
    overflow: hidden;
    .el-image {
      width: 100%;
    }
  }
  .performer-info-data {
    flex-grow: 1;
    padding-left: 0.5em;
    .performer-name {
      font-family: 500;
      font-size: 1.5em;
      color: #ffaa47;
    }
    .performer-info-ul {
      list-style-type: none;
      font-size: 1em;
      line-height: 1.5em;
      color: #cfd3dc;
      :deep(.el-breadcrumb) {
        font-size: 1em;
        line-height: 1.5em;
      }
    }
  }
}
</style>
