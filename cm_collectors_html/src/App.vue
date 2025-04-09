<template>
  <div class="app-container">
    <RouterView />
  </div>
</template>
<script setup lang="ts"></script>
<style lang="scss" scoped>
.app-container {
  width: calc(100vw - 10px);
  height: calc(100vh - 10px);
  padding: 5px;
  overflow: hidden;
  background-color: #1f1f1f;
  display: flex;
  flex-direction: column;
}
</style>
