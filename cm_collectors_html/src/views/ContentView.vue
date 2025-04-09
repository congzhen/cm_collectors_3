<template>
  <div class="content-view">
    <div class="list">
      <el-scrollbar>
        <ul class="list-ul">
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>

          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>

          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>

          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>

          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
          <li><contentItem></contentItem></li>
        </ul>
      </el-scrollbar>
    </div>
    <div class="paging">
      <el-pagination background layout="total, prev, pager, next, jumper" :total="1000" />
    </div>
  </div>
</template>
<script setup lang="ts">
import contentItem from '@/components/content/contentItem.vue'
</script>
<style lang="scss" scoped>
.content-view {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .list {
    flex-grow: 1;
    overflow: hidden;
    .list-ul {
      list-style-type: none;
      display: flex;
      flex-wrap: wrap;
      gap: 0.4em;
      padding-bottom: 1em;
    }
  }

  .paging {
    width: 100%;
    padding-top: 5px;
  }
}
</style>
