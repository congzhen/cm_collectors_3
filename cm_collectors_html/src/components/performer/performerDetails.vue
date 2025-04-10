<template>
  <div class="performer-details">
    <div class="performer-photo">
      <el-image
        src="https://picx.zhimg.com/v2-d6f44389971daab7e688e5b37046e4e4_720w.jpg"
        fit="cover"
      />
    </div>
    <div class="performer-info">
      <div class="performer-name">櫻茉日</div>
      <ul class="performer-info-ul">
        <li>别名: 堀北実来、堀北实来</li>
        <li>
          <el-breadcrumb>
            <el-breadcrumb-item> <span>胸围: </span><label>101 </label> </el-breadcrumb-item>
            <el-breadcrumb-item> <span>腰围: </span><label>58 </label> </el-breadcrumb-item>
            <el-breadcrumb-item> <span>臀围: </span><label>90 </label> </el-breadcrumb-item>
          </el-breadcrumb>
        </li>
        <li>出生日期: 2000-03-03</li>
        <li>拍摄年龄: 23岁</li>
        <li>身高：160cm</li>
      </ul>
      <div class="performer-btn">
        <el-button icon="Search" size="small" round> 查看【櫻茉日】所有资源 </el-button>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup></script>
<style lang="scss" scoped>
.performer-details {
  display: flex;
  .performer-photo {
    flex-shrink: 0;
    width: 110px;
    padding: 1px;
    border-radius: 5px;
    aspect-ratio: 1/1.3;
    box-shadow: 0 0 0 1px #cdd0d6;
    overflow: hidden;
    .el-image {
      width: 100%;
      height: 100%;
      border-radius: 5px;
    }
  }
  .performer-info {
    flex-grow: 1;
    padding-left: 15px;
    display: flex;
    flex-direction: column;
    .performer-name {
      flex-shrink: 0;
      font-family: 300;
      font-size: 1.2em;
      color: #ffaa47;
    }
    .performer-info-ul {
      flex-grow: 1;
      list-style-type: none;
      font-size: 0.8em;
      :deep(.el-breadcrumb) {
        .el-breadcrumb__inner {
          font-size: 0.8em;
        }
      }
    }
    .performer-btn {
      flex-shrink: 0;
      transform: scale(0.85);
    }
  }
}
</style>
