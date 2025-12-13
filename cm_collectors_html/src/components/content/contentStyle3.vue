<template>
  <contentRightClickMenu :resource="props.resource">
    <div class="content-style content-style3" :style="{ width: coverPosterSize_C.width + 'px' }">
      <div class="content-cover" :style="{ height: coverPosterSize_C.height + 'px' }">
        <contentCoverImage :resource="props.resource"></contentCoverImage>
        <div class="play-icon" @click.stop="playResource(props.resource)">
          <el-icon>
            <VideoPlay />
          </el-icon>
        </div>
      </div>
      <div class="info">
        <div class="block-two">
          <div class="issueNumber">{{ props.resource.issueNumber }}</div>
          <div class="tags">
            <contentTagDisplay :resource="props.resource" :font-size-disable="true"></contentTagDisplay>
          </div>
        </div>
        <div class="block-two">
          <div>{{ props.resource.issuingDate }}</div>
          <div>
            <el-rate v-model="localStars" disabled />
          </div>
        </div>
        <div class="title" :style="titleStyleObj_C">{{ props.resource.title }}</div>
      </div>
    </div>
  </contentRightClickMenu>
</template>
<script setup lang="ts">
import type { I_resource } from '@/dataType/resource.dataType';
import contentRightClickMenu from './contentRightClickMenu.vue';
import contentTagDisplay from './contentTagDisplay.vue';
import { computed, ref, type PropType } from 'vue';
import { playResource } from '@/common/play';
import { appStoreData } from '@/storeData/app.storeData';
import contentCoverImage from './contentCoverImage.vue';
import dataset from '@/assets/dataset';
import { coverPosterSize } from '@/common/photo';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource>,
    required: true,
  },
})
const localStars = ref(props.resource.stars)
const coverPosterSize_C = computed(() => {
  const { width, height } = coverPosterSize(props.resource.coverPosterWidth, props.resource.coverPosterHeight, store.appStoreData.currentConfigApp.coverPosterWidthStatus, store.appStoreData.currentConfigApp.coverPosterWidthBase, store.appStoreData.currentConfigApp.coverPosterHeightStatus, store.appStoreData.currentConfigApp.coverPosterHeightBase)
  return {
    width,
    height,
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

</script>
<style lang="scss" scoped>
.content-style3 {
  display: flex;
  flex-direction: column;
  cursor: pointer;
  overflow: hidden;

  &:hover {
    .play-icon {
      display: block;
    }

    .content-cover {
      overflow: hidden;

      .el-image {
        scale: 1.05;
      }
    }
  }

  .content-cover {
    position: relative;
    width: 100%;
    height: 100%;

    .el-image {
      width: 100%;
      height: 100%;
    }
  }

  .play-icon {
    position: absolute;
    z-index: 10;
    font-size: 3.8em;
    color: #f3f3f3;
    opacity: 0.75;
    display: none;
    margin-left: 0.05em;
    margin-top: -1.5em;
  }

  .info {
    height: 90px;
    overflow: hidden;


    .block-two {
      display: flex;
      gap: 10px;
      justify-content: space-between;
      align-items: center;
      height: 24px;
      overflow: hidden;
    }

    .el-rate {
      height: auto;
    }

    .issueNumber {
      font-weight: 800;
      flex-shrink: 0;
    }

    .tags {
      padding-top: 2px;
      height: 18px;
      flex: 1;
      overflow: hidden;
      display: flex;
      justify-content: flex-end;

      .content-tag-display {
        justify-content: flex-end;
      }
    }

    .title {
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
}
</style>
