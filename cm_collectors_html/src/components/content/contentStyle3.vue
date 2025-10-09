<template>
  <contentRightClickMenu :resource="props.resource">
    <div class="content-style3" :style="{ width: coverPosterSize_C.width + 'px' }">
      <div class="content-cover">
        <el-image :src="getResourceCoverPoster(props.resource)" fit="cover" />
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
            <contentTagDisplay :resource="props.resource"></contentTagDisplay>
          </div>
        </div>
        <div class="block-two">
          <div>{{ props.resource.issuingDate }}</div>
          <div>
            <el-rate v-model="localStars" disabled />
          </div>
        </div>
        <div class="title">{{ props.resource.title }}</div>
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
import { getResourceCoverPoster } from '@/common/photo';
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
.content-style3 {
  display: flex;
  flex-direction: column;
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

  .content-cover {
    position: relative;
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
