<template>
  <div class="performer-photo">
    <el-image
      src="https://picx.zhimg.com/v2-d6f44389971daab7e688e5b37046e4e4_720w.jpg"
      fit="cover"
    />
  </div>
</template>
<script lang="ts" setup></script>
<style lang="scss" scoped>
.performer-photo {
  width: 100%;
  aspect-ratio: 1/1;
  position: relative;
  border-radius: 50%;
  box-shadow: 0 0 0 1px #ffffff;
  overflow: hidden;
  cursor: pointer;

  .el-image {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    border-radius: inherit;
    img {
      border-radius: inherit;
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }
  &:hover {
    .el-image {
      transform: scale(1.1);
    }
  }
}
</style>
