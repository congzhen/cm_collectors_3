<template>
  <div class="index-view">
    <HeaderView class="header"></HeaderView>
    <div class="main">
      <TagView class="tag"></TagView>
      <ContentView class="content"></ContentView>
      <DetailsView class="details"></DetailsView>
    </div>
  </div>
</template>
<script setup lang="ts">
import HeaderView from './HeaderView.vue'
import TagView from './TagView.vue'
import ContentView from './ContentView.vue'
import DetailsView from './DetailsView.vue'
</script>
<style lang="scss" scoped>
.index-view {
  width: 100%;
  height: 100%;
  overflow: hidden;
  background-color: #1f1f1f;
  display: flex;
  flex-direction: column;

  .header {
    width: 100%;
  }

  .main {
    flex-grow: 1;
    display: flex;
    overflow: hidden;

    .tag {
      flex-shrink: 0;
      height: 100%;
    }

    .content {
      flex-grow: 1;
      overflow: auto;
    }

    .details {
      flex-shrink: 0;
      height: 100%;
    }
  }
}
</style>
