<template>
  <div>
    <el-button icon="Memo" size="small" title="播放列表" aria-label="播放列表" @click="openPlayListHandle">
      <span class="toolbar-button-label">待播</span>
    </el-button>
    <playListDrawer ref="playListDrawerRef"></playListDrawer>
  </div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import playListDrawer from './playListDrawer.vue';
defineProps({
  modern: {
    type: Boolean,
    default: false,
  },
});
const playListDrawerRef = ref<InstanceType<typeof playListDrawer>>();


const openPlayListHandle = async () => {
  playListDrawerRef.value?.open();
};
</script>
