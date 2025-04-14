<template>
  <div class="tag-item">
    <label>角色扮演</label>
    <div class="tag-item-tool">
      <div class="tag-item-tool-block">
        <el-icon :size="12" color="#bdbcbc">
          <Edit />
        </el-icon>
        <el-icon :size="12" color="#bdbcbc">
          <Delete />
        </el-icon>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup></script>
<style lang="scss" scoped>
.tag-item {
  font-size: 12px;
  height: 30px;
  line-height: 30px;
  padding: 0 20px;
  border-radius: 5px;
  border: 1px solid #4c4d4f;
  cursor: pointer;
  -moz-user-select: none;
  -webkit-user-select: none;
  -ms-user-select: none;
  user-select: none;
  position: relative;
  &:hover {
    border: 1px solid #616264;
    label {
      color: #babdc5;
    }
    .tag-item-tool {
      display: block;
    }
  }
  label {
    cursor: pointer;
  }

  .tag-item-tool {
    display: none;
    position: absolute;
    margin-top: -28px;
    right: 2px;
    .tag-item-tool-block {
      display: flex;
      flex-direction: column;
      gap: 2px;
      .el-icon {
        cursor: pointer;
        &:hover {
          color: #fff;
        }
      }
    }
  }
}
</style>
