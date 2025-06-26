<template>
  <div class="tag-block-performer">
    <div class="tag-content">
      <tagSpan title="全部"></tagSpan>
      <tagSpan title="无演员"></tagSpan>
    </div>
    <div class="tag-performer">
      <performerBlock class="tag-performer-item"
        v-for="performer, key in store.appStoreData.currentTopPreferredPerformers" :key="key" :performer="performer">
      </performerBlock>
    </div>
  </div>
</template>
<script setup lang="ts">
import tagSpan from './tagSpan.vue'
import performerBlock from '../performer/performerBlock.vue'
import { appStoreData } from '@/storeData/app.storeData';
const store = {
  appStoreData: appStoreData(),
}
</script>
<style lang="scss" scoped>
.tag-block-performer {
  .tag-content {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3em;
  }

  .tag-performer {
    padding: 0.5em 0;
    display: flex;
    flex-wrap: wrap;
    gap: 0.3em;

    .tag-performer-item {
      width: calc(25% - 0.9em);

      &:hover {
        background-color: var(--el-color-primary-light-9);
      }
    }
  }
}
</style>
