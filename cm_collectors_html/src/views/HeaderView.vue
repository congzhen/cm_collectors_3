<template>
  <div class="header-container">
    <div class="title">
      <img src="/public/icon32.png" />
      <label>CM File Collectors</label>
    </div>
    <div class="right">
      <div class="search">
        <el-input-tag v-model="searchValue" trigger="Space" clearable placeholder="Please input">
          <template #suffix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input-tag>
      </div>
      <div class="setting">
        <el-icon title="添加"><Plus /></el-icon>
        <el-icon title="标签"><PriceTag /></el-icon>
        <el-icon title="演员"><User /></el-icon>
        <el-icon title="设置"><Setting /></el-icon>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'

const searchValue = ref<string[]>([])
</script>
<style lang="scss" scoped>
.header-container {
  display: flex;
  justify-content: space-between;
  .title {
    font-size: 1.7em;
    font-weight: 500;
    flex-shrink: 0;
    padding: 0;
    img {
      width: 1em;
      height: 1em;
      margin-right: 0.3em;
    }
  }

  .right {
    display: flex;
    justify-content: flex-end;
    flex-grow: 1;

    .search {
      margin: 0 1em;
      width: 30%;
      transition: width 0.3s;
    }
    .search:focus-within {
      width: 60%;
    }

    .setting {
      flex-shrink: 0;
      font-size: 2em;
      cursor: pointer;
      .el-icon {
        margin-left: 0.2em;
        &:hover {
          color: var(--el-color-primary);
        }
      }
    }
  }
}
</style>
