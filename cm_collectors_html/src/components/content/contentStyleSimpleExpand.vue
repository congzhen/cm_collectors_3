<template>
  <contentRightClickMenu :resource="props.resource">
    <div class="content-style content-style-simple" :style="simpleStyle_C">
      <div class="top-bar">
        <contentTagDisplay :resource="props.resource"></contentTagDisplay>
      </div>

      <div class="content-cover"
        :style="{ width: coverPosterSize_C.width + 'px', height: coverPosterSize_C.height + 'px' }">
        <contentCoverImage :resource="props.resource"></contentCoverImage>
      </div>
      <div class="play-icon" @click.stop="playResource(props.resource)">
        <el-icon>
          <VideoPlay />
        </el-icon>
      </div>
      <div class="title" :style="titleStyleObj_C">{{ props.resource.title }}</div>
      <div class="performer" :style="performerStyleObj_C">
        <div v-for="(item, index) in props.resource.performers" :key="index">
          <div class="performer-item">
            <el-icon>
              <Avatar />
            </el-icon>
            <label>{{ item.name }}</label>
          </div>
        </div>
      </div>
    </div>
  </contentRightClickMenu>
</template>
<script setup lang="ts">
import contentCoverImage from './contentCoverImage.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import contentTagDisplay from './contentTagDisplay.vue';
import { computed, type PropType } from 'vue';
import { appStoreData } from '@/storeData/app.storeData';
import { playResource } from '@/common/play';
import { isMobile } from '@/assets/mobile';
import contentRightClickMenu from './contentRightClickMenu.vue';
import dataset from '@/assets/dataset';
import { coverPosterSize } from '@/common/photo';
import { alignConvertJustifyContent } from '@/assets/css';
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

const simpleStyle_C = computed(() => {
  return {
    width: coverPosterSize_C.value.width + 'px',
    height: (coverPosterSize_C.value.height + 22 + 16) + 'px',
  }
})
const titleStyleObj_C = computed(() => {
  const obj: Record<string, string> = {};
  if (dataset.coverTitleAlign.indexOf(store.appStoreData.currentConfigApp.coverTitleAlign) > -1) {
    obj['text-align'] = store.appStoreData.currentConfigApp.coverTitleAlign;
  } else {
    obj['text-align'] = 'left'
  }
  return obj
})
const performerStyleObj_C = computed(() => {
  const obj: Record<string, string> = {};
  if (dataset.coverTitleAlign.indexOf(store.appStoreData.currentConfigApp.coverTitleAlign) > -1) {
    obj['justify-content'] = alignConvertJustifyContent(store.appStoreData.currentConfigApp.coverTitleAlign);
  } else {
    obj['justify-content'] = 'flex-start'
  }
  return obj
})



</script>
<style lang="scss" scoped>
.content-style-simple {
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

  .title {
    width: calc(100% - 0.4em);
    line-height: 22px;
    padding: 0 0.2em;
    /* 禁止换行 */
    white-space: nowrap;
    /* 隐藏溢出内容 */
    overflow: hidden;
    /* 添加省略号 */
    text-overflow: ellipsis;
  }

  .performer {
    width: calc(100% - 0.4em);
    line-height: 16px;
    height: 16px;
    padding: 0 0.2em;
    font-size: 0.8em;
    color: #919191;
    display: flex;
    gap: 10px;
    /* 禁止换行 */
    white-space: nowrap;
    /* 隐藏溢出内容 */
    overflow: hidden;
    /* 添加省略号 */
    text-overflow: ellipsis;


    .performer-item {
      display: flex;
      align-items: center;
      gap: 1px;
    }
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
