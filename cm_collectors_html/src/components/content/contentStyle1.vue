<template>
  <contentRightClickMenu :resource="props.resource">
    <div class="content-style1"
      :style="{ width: coverPosterSize_C.width + 'px', height: coverPosterSize_C.height + 'px' }">
      <div class="top-bar">
        <contentTagDisplay :resource="props.resource"></contentTagDisplay>
      </div>

      <div class="content-cover">
        <el-image :src="getResourceCoverPoster(props.resource)" fit="cover" />
      </div>


      <div class="play-icon" @click.stop="playResource(props.resource)">
        <el-icon>
          <VideoPlay />
        </el-icon>
      </div>
      <div class="title-bg"></div>
      <div class="title">{{ props.resource.title }}</div>
    </div>
  </contentRightClickMenu>
</template>
<script setup lang="ts">
import type { I_resource } from '@/dataType/resource.dataType';
import contentTagDisplay from './contentTagDisplay.vue';
import { computed, type PropType } from 'vue';
import { appStoreData } from '@/storeData/app.storeData';
import { playResource } from '@/common/play';
import { getResourceCoverPoster } from '@/common/photo';
import { isMobile } from '@/assets/mobile';
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
  let width = props.resource.coverPosterWidth;
  let height = props.resource.coverPosterHeight;
  if (store.appStoreData.currentConfigApp.coverPosterWidthStatus) {
    width = store.appStoreData.currentConfigApp.coverPosterWidthBase;
  }
  if (store.appStoreData.currentConfigApp.coverPosterHeightStatus) {
    width = store.appStoreData.currentConfigApp.coverPosterHeightBase / height * width;
    height = store.appStoreData.currentConfigApp.coverPosterHeightBase;
  }
  return {
    width,
    height,
  }
})

</script>
<style lang="scss" scoped>
.content-style1 {
  width: 158px;
  height: 214px;
  overflow: hidden;
  position: relative;
  cursor: pointer;

  &:hover {
    .play-icon {
      display: block;
    }

    .content-cover {
      .el-image {
        scale: 1.05;
      }
    }
  }

  .top-bar {
    position: absolute;
    z-index: 10;
    padding: 2px;
  }

  .content-cover {
    width: 100%;
    height: 100%;
    overflow: hidden;

    .el-image {
      width: 100%;
      height: 100%;
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

  }

  .title-bg {
    position: absolute;
    margin-top: -1.7em;
    width: 100%;
    height: 1.7em;
    background-color: #282923;
    opacity: 0.5;
    z-index: 5;
    border-radius: 0px 0px 4px 4px;
  }

  .title {
    position: absolute;
    margin-top: -1.7em;
    width: 100%;
    line-height: 1.7em;
    z-index: 6;
    padding: 0 0.2em;
    /* 禁止换行 */
    white-space: nowrap;
    /* 隐藏溢出内容 */
    overflow: hidden;
    /* 添加省略号 */
    text-overflow: ellipsis;
  }
}

@media (max-width: 768px) {
  .content-style1 {
    width: v-bind('isMobile() ? "100%" : coverPosterSize_C.width + "px"');
    height: v-bind('isMobile() ? "auto" : coverPosterSize_C.height + "px"');
  }

  .content-style1:where(.mobile-layout *) {
    width: 100%;
    height: auto;
    aspect-ratio: v-bind('coverPosterSize_C.width') / v-bind('coverPosterSize_C.height');
  }
}
</style>
