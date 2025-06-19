<template>
  <div class="database-menu">
    <el-menu :default-active="activeDataBase" mode="horizontal" text-color="#fff" active-text-color="#ffd04b">
      <el-menu-item index="1">日本数据库</el-menu-item>
      <el-menu-item index="2">中国数据库</el-menu-item>
      <el-menu-item index="3">主播</el-menu-item>
      <el-menu-item index="4">VR</el-menu-item>
      <el-menu-item index="5">短剧</el-menu-item>
      <el-menu-item index="6">漫画</el-menu-item>
      <el-menu-item index="7">写真</el-menu-item>
      <el-menu-item index="8">动漫</el-menu-item>
      <el-menu-item index="9">电视剧</el-menu-item>
      <el-menu-item index="10">日本数据库</el-menu-item>
      <el-menu-item index="20">中国数据库</el-menu-item>
      <el-menu-item index="30">主播</el-menu-item>
      <el-menu-item index="40">VR</el-menu-item>
      <el-menu-item index="50">短剧</el-menu-item>
      <el-menu-item index="60">漫画</el-menu-item>
      <el-menu-item index="70">写真</el-menu-item>
      <el-menu-item index="80">动漫</el-menu-item>
      <el-menu-item index="90">电视剧</el-menu-item>
      <el-menu-item index="100">日本数据库</el-menu-item>
      <el-menu-item index="200">中国数据库</el-menu-item>
      <el-menu-item index="300">主播</el-menu-item>
      <el-menu-item index="400">VR</el-menu-item>
      <el-menu-item index="500">短剧</el-menu-item>
      <el-menu-item index="600">漫画</el-menu-item>
      <el-menu-item index="700">写真</el-menu-item>
      <el-menu-item index="800">动漫</el-menu-item>
      <el-menu-item index="900">电视剧</el-menu-item>
    </el-menu>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
const activeDataBase = ref('1')
</script>
<style lang="scss" scoped>
.database-menu {
  padding-bottom: 5px;

  :deep(.el-menu--horizontal) {
    height: auto;
  }

  :deep(.el-menu--horizontal>.el-menu-item) {
    height: 14px;
    line-height: 16px;
    padding-top: 12px;
    padding-bottom: 16px;

    &:hover {
      background-color: unset;
      color: #409EFF;
      user-select: none;
    }
  }

}
</style>
