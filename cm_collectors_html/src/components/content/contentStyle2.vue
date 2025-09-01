<template>
  <contentRightClickMenu :resource="props.resource">
    <div class="content-style2">
      <div class="content-cover"
        :style="{ width: coverPosterSize_C.width + 'px', height: coverPosterSize_C.height + 'px' }">
        <el-image :src="getResourceCoverPoster(props.resource)" fit="contain" />
        <div class="play-icon" @click.stop="playResource(props.resource)">
          <el-icon>
            <VideoPlay />
          </el-icon>
        </div>
      </div>
      <div class="content-info" :style="{ width: store.appStoreData.currentConfigApp.coverPosterBoxInfoWidth + 'px' }">
        <h4 class="title">{{ props.resource.title }}</h4>
        <contentTagDisplay :resource="props.resource"></contentTagDisplay>
        <div class="tag-list">
          <el-tag effect="plain" size="small" v-for="item, key in props.resource.performers" :key="key">
            {{ item.name }}
          </el-tag>
        </div>
        <div class="tag-list">
          <el-tag type="info" effect="plain" size="small" v-for="item, key in props.resource.tags" :key="key">
            {{ item.name }}
          </el-tag>
        </div>
      </div>
    </div>
  </contentRightClickMenu>
</template>
<script setup lang="ts">
import type { I_resource } from '@/dataType/resource.dataType';
import { computed, type PropType } from 'vue';
import { getResourceCoverPoster } from '@/common/photo';
import { appStoreData } from '@/storeData/app.storeData';
import { coverPosterSize } from '@/common/photo';
import contentTagDisplay from './contentTagDisplay.vue'
import { playResource } from '@/common/play';
import contentRightClickMenu from './contentRightClickMenu.vue';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource>,
    required: true,
  },
})
const coverPosterSize_C = computed(() => {
  const { width, height } = coverPosterSize(props.resource.coverPosterWidth, props.resource.coverPosterHeight, store.appStoreData.currentConfigApp.coverPosterWidthStatus, store.appStoreData.currentConfigApp.coverPosterWidthBase, store.appStoreData.currentConfigApp.coverPosterHeightStatus, store.appStoreData.currentConfigApp.coverPosterHeightBase)
  return {
    width,
    height,
  }
})
</script>
<style lang="scss" scoped>
.content-style2 {
  display: flex;
  gap: 12px;
  position: relative;

  &:hover {
    .play-icon {
      display: block;
    }
  }

  .play-icon {
    position: absolute;
    z-index: 10;
    margin-left: 0.05em;
    margin-top: -1.5em;
    font-size: 3.8em;
    color: #f3f3f3;
    opacity: 0.75;
    display: none;
    cursor: pointer;
  }

  .content-cover {
    flex-shrink: 0;
    padding: 2px;
    overflow: hidden;

    .el-image {
      width: 100%;
      height: 100%;
      border-radius: 5px;
    }

  }

  .content-info {
    width: 200px;
    flex: 1;

    .title {
      display: -webkit-box;
      -webkit-line-clamp: 3;
      -webkit-box-orient: vertical;
      overflow: hidden;
      text-overflow: ellipsis;
      word-break: break-word;
      margin-bottom: 5px;
    }

    .tag-list {
      display: flex;
      flex-wrap: wrap;
      gap: 2px;
      padding: 5px 0;
    }
  }
}
</style>
