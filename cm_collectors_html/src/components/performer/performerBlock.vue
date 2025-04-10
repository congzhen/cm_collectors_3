<template>
  <div class="performer-block">
    <performerPhoto></performerPhoto>
    <div class="performer-block-name">小栗みゆ</div>
  </div>
</template>
<script lang="ts" setup>
import performerPhoto from './performerPhoto.vue'
</script>
<style lang="scss" scoped>
.performer-block {
  max-width: 95px;
  border-radius: 5px;
  cursor: pointer;
  padding: 3px;
  overflow: hidden;
  background-color: #262727;
  .performer-block-name {
    text-align: center;
    padding: 0.8em 0.5em;
    font-size: 0.8em;
    line-height: 1em;
  }
}
</style>
