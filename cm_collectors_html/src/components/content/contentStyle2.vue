<template>
  <contentRightClickMenu :resource="props.resource">
    <div class="content-style content-style2">
      <div class="content-cover"
        :style="{ width: coverPosterSize_C.width + 'px', height: coverPosterSize_C.height + 'px' }">
        <contentCoverImage :resource="props.resource"></contentCoverImage>
        <contentVideoDurationBadge :resource="props.resource" offset-bottom="52px"></contentVideoDurationBadge>
        <div class="cover-display-tags" @click.stop>
          <contentTagDisplay :resource="props.resource"></contentTagDisplay>
        </div>
        <div class="cover-title-overlay">
          <h4 class="title" :title="props.resource.title" :style="titleStyleObj_C">{{ props.resource.title }}</h4>
        </div>
        <div class="play-icon" @click.stop="playResource(props.resource)">
          <span class="play-icon-surface">
            <el-icon><VideoPlay /></el-icon>
          </span>
        </div>
      </div>
      <div class="content-info" :style="{
        width: infoWidth_C + 'px',
        height: coverPosterSize_C.height + 'px',
      }">
        <div v-if="props.resource.performers.length" class="info-section performer-section">
          <div class="section-label">
            <el-icon><User /></el-icon>
            <span>演员</span>
          </div>
          <div class="performer-list" :title="performerNames_C">{{ performerNames_C }}</div>
        </div>

        <div v-if="props.resource.tags.length" class="info-section resource-tag-section">
          <div class="section-label">
            <el-icon><PriceTag /></el-icon>
            <span>标签</span>
          </div>
          <div class="tag-list">
            <el-tag type="info" effect="plain" size="small" v-for="item in visibleTags_C" :key="item.id"
              :title="item.name">
              {{ item.name }}
            </el-tag>
            <el-tag v-if="hiddenTagCount_C > 0" class="more-tag" type="info" effect="plain" size="small"
              :title="props.resource.tags.slice(visibleTags_C.length).map(item => item.name).join('、')">
              +{{ hiddenTagCount_C }}
            </el-tag>
          </div>
        </div>
      </div>
    </div>
  </contentRightClickMenu>
</template>
<script setup lang="ts">
import type { I_resource } from '@/dataType/resource.dataType';
import { computed, type PropType } from 'vue';
import contentCoverImage from './contentCoverImage.vue';
import { appStoreData } from '@/storeData/app.storeData';
import { coverPosterSize } from '@/common/photo';
import contentTagDisplay from './contentTagDisplay.vue'
import contentVideoDurationBadge from './contentVideoDurationBadge.vue';
import { playResource } from '@/common/play';
import contentRightClickMenu from './contentRightClickMenu.vue';
import dataset from '@/assets/dataset';
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
const infoWidth_C = computed(() => {
  return store.appStoreData.currentConfigApp.coverPosterBoxInfoWidth || 200;
})
const performerNames_C = computed(() => {
  return props.resource.performers.map(item => item.name).filter(Boolean).join(' · ');
})
const visibleTagLimit_C = computed(() => {
  if (infoWidth_C.value < 220) return 6;
  if (infoWidth_C.value < 300) return 8;
  return 10;
})
const visibleTags_C = computed(() => props.resource.tags.slice(0, visibleTagLimit_C.value))
const hiddenTagCount_C = computed(() => Math.max(0, props.resource.tags.length - visibleTags_C.value.length))
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
.content-style2 {
  display: flex;
  position: relative;
  overflow: hidden;
  color: var(--el-text-color-primary);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-bg-color-overlay);
  box-shadow: var(--el-box-shadow-light);
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;

  &:hover {
    border-color: var(--el-color-primary-light-5);
    box-shadow: var(--el-box-shadow);
    transform: translateY(-1px);

    .play-icon {
      opacity: 1;
    }

    .content-cover {
      .el-image {
        transform: scale(1.035);
      }
    }
  }

  .play-icon {
    position: absolute;
    inset: 0;
    z-index: 13;
    display: grid;
    place-items: center;
    color: #ffffff;
    opacity: 0;
    pointer-events: none;
    cursor: pointer;
    transition: opacity 0.18s ease;

    .play-icon-surface {
      width: 46px;
      height: 46px;
      display: grid;
      place-items: center;
      border: 1px solid rgba(255, 255, 255, 0.5);
      border-radius: 50%;
      background: rgba(10, 16, 19, 0.58);
      box-shadow: 0 8px 22px rgba(0, 0, 0, 0.28);
      backdrop-filter: blur(6px);
      pointer-events: auto;

      .el-icon {
        font-size: 24px;
      }
    }
  }

  .content-cover {
    flex-shrink: 0;
    overflow: hidden;
    position: relative;
    background: var(--el-fill-color-dark);

    .el-image {
      width: 100%;
      height: 100%;
      display: block;
      transition: transform 0.22s ease;
    }

    .cover-display-tags {
      position: absolute;
      top: 7px;
      left: 7px;
      z-index: 12;
      max-width: calc(100% - 14px);
      max-height: 46px;
      overflow: hidden;

      :deep(.content-tag-display) {
        gap: 4px !important;
      }

      :deep(.content-tag) {
        box-shadow: 0 2px 7px rgba(0, 0, 0, 0.2);
        backdrop-filter: blur(5px);
      }
    }

    .cover-title-overlay {
      position: absolute;
      right: 0;
      bottom: 0;
      left: 0;
      z-index: 11;
      padding: 32px 12px 10px;
      color: #ffffff;
      background: linear-gradient(180deg, transparent, rgba(8, 13, 16, 0.86));
      pointer-events: none;

      .title {
        margin: 0;
        display: -webkit-box;
        -webkit-box-orient: vertical;
        -webkit-line-clamp: 2;
        overflow: hidden;
        text-overflow: ellipsis;
        word-break: break-word;
        color: #ffffff;
        font-size: 14px;
        font-weight: 600;
        line-height: 20px;
        text-shadow: 0 1px 4px rgba(0, 0, 0, 0.55);
      }
    }
  }

  .content-info {
    width: 200px;
    min-width: 0;
    padding: 14px;
    box-sizing: border-box;
    flex: 0 0 auto;
    display: flex;
    flex-direction: column;
    overflow: hidden;

    .info-section {
      min-width: 0;
      margin-top: 0;
    }

    .section-label {
      display: flex;
      align-items: center;
      gap: 5px;
      color: var(--el-text-color-secondary);
      font-size: 12px;

      .el-icon {
        font-size: 13px;
      }
    }

    .performer-list {
      min-width: 0;
      overflow: hidden;
      display: -webkit-box;
      -webkit-box-orient: vertical;
      -webkit-line-clamp: 2;
      text-overflow: ellipsis;
      word-break: break-word;
      color: var(--el-text-color-regular);
      font-size: 13px;
      line-height: 20px;
    }

    .performer-section {
      padding-bottom: 7px;
      border-bottom: 1px solid var(--el-border-color-lighter);

      .section-label {
        margin-bottom: 6px;
      }
    }

    .resource-tag-section {
      margin-top: 7px;

      &:first-child {
        margin-top: 0;
      }

      .section-label {
        margin-bottom: 6px;
      }
    }

    .tag-list {
      display: flex;
      flex-wrap: wrap;
      gap: 5px;
      max-height: 53px;
      overflow: hidden;

      .el-tag {
        height: 24px;
        max-width: 100%;
        padding: 0 7px;
        box-sizing: border-box;
        border-radius: 5px;
        line-height: 22px;
      }

      .el-tag {
        color: var(--el-text-color-regular);
        border-color: var(--el-border-color);
        background: var(--el-fill-color-light);

        :deep(.el-tag__content) {
          min-width: 0;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
      }

      .more-tag {
        color: var(--el-text-color-secondary);
        border-style: dashed;
      }
    }
  }
}
</style>
