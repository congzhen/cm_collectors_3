<template>
  <div>
    <el-button icon="Memo" size="small" @click="openPlayListHandle"></el-button>
    <playListDrawer ref="playListDrawerRef"></playListDrawer>
  </div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import playListDrawer from './playListDrawer.vue';
const playListDrawerRef = ref<InstanceType<typeof playListDrawer>>();


const openPlayListHandle = async () => {
  playListDrawerRef.value?.open();
};
</script>
