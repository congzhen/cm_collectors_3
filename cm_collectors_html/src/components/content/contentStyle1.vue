<template>
  <div class="content-style1">
    <div class="top-bar">
      <contentTag title="1080P"></contentTag>
      <contentTag title="1080P"></contentTag>
      <contentTag title="1080P"></contentTag>
      <contentTag title="1080P"></contentTag>
    </div>
    <div class="content-cover">
      <el-image
        src="https://picx.zhimg.com/v2-d6f44389971daab7e688e5b37046e4e4_720w.jpg"
        fit="cover"
      />
    </div>
    <div class="play-icon">
      <el-icon>
        <VideoPlay />
      </el-icon>
    </div>
    <div class="title-bg"></div>
    <div class="title">11122312</div>
  </div>
</template>
<script setup lang="ts">
import contentTag from './contentTag.vue'
</script>
<style lang="scss" scoped>
.content-style1 {
  width: 158px;
  height: 214px;
  background-color: aqua;
  overflow: hidden;
  position: relative;
  cursor: pointer;

  &:hover {
    .play-icon {
      display: block;
    }
  }

  .top-bar {
    position: absolute;
    z-index: 10;
    display: flex;
    flex-wrap: wrap;
    gap: 2px;
    padding: 0.05em 0.1em;
  }

  .content-cover {
    width: 100%;
    height: 100%;
    overflow: hidden;
    .el-image {
      width: 100%;
      height: 100%;
    }
  }

  .play-icon {
    position: absolute;
    z-index: 10;
    margin-left: 0.05em;
    margin-top: -1.5em;
    font-size: 3.8em;
    color: #f3f3f3;
    opacity: 0.75;
    display: none;
  }

  .title-bg {
    position: absolute;
    margin-top: -1.7em;
    width: 100%;
    height: 1.7em;
    background-color: #282923;
    opacity: 0.5;
    z-index: 5;
  }

  .title {
    position: absolute;
    margin-top: -1.7em;
    width: 100%;
    line-height: 1.7em;
    z-index: 6;
    padding: 0 0.2em;
  }
}
</style>
