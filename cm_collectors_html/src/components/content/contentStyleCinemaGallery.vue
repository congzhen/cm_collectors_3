<template>
  <contentRightClickMenu :resource="props.resource">
    <article class="cinema-gallery-card" :style="cardStyle_C">
      <div class="cinema-gallery-cover" :style="{ height: coverPosterSize_C.height + 'px' }">
        <contentCoverImage :resource="props.resource" />
        <div class="cinema-gallery-config-tags" @click.stop>
          <contentTagDisplay :resource="props.resource" />
        </div>
        <contentVideoDurationBadge :resource="props.resource" offset-bottom="7px" offset-right="7px" />
        <button type="button" class="cinema-gallery-play" aria-label="播放" @click.stop="playResource(props.resource)">
          <el-icon><VideoPlay /></el-icon>
        </button>
      </div>

      <div class="cinema-gallery-info" :style="infoStyle_C">
        <h4 class="cinema-gallery-title" :title="props.resource.title" :style="titleStyleObj_C">
          {{ props.resource.title }}
        </h4>
        <div class="cinema-gallery-meta" :class="{ 'without-performers': !props.resource.performers.length }">
          <div v-if="props.resource.performers.length" class="cinema-gallery-performers" :title="performerNames_C">
            <span class="performer-names">{{ visiblePerformerNames_C }}</span>
            <span v-if="hiddenPerformerCount_C" class="performer-more">+{{ hiddenPerformerCount_C }}</span>
          </div>
          <div v-if="props.resource.tags.length" class="cinema-gallery-tags"
            :class="{ 'has-more': hiddenTagCount_C, 'two-rows': !props.resource.performers.length }">
            <el-tag v-for="item in visibleTags_C" :key="item.id" type="info" effect="plain" :title="item.name">
              {{ item.name }}
            </el-tag>
            <el-tag v-if="hiddenTagCount_C" class="more-tag" type="info" effect="plain"
              :title="hiddenTagsTitle_C">
              +{{ hiddenTagCount_C }}
            </el-tag>
          </div>
        </div>
      </div>
    </article>
  </contentRightClickMenu>
</template>

<script setup lang="ts">
import { computed, type PropType } from 'vue';

import dataset from '@/assets/dataset';
import { playResource } from '@/common/play';
import { coverPosterSize } from '@/common/photo';
import type { I_resource } from '@/dataType/resource.dataType';
import { appStoreData } from '@/storeData/app.storeData';

import contentCoverImage from './contentCoverImage.vue';
import contentRightClickMenu from './contentRightClickMenu.vue';
import contentTagDisplay from './contentTagDisplay.vue';
import contentVideoDurationBadge from './contentVideoDurationBadge.vue';

const store = {
  appStoreData: appStoreData(),
};

const props = defineProps({
  resource: {
    type: Object as PropType<I_resource>,
    required: true,
  },
});

const coverPosterSize_C = computed(() => {
  return coverPosterSize(
    props.resource.coverPosterWidth,
    props.resource.coverPosterHeight,
    store.appStoreData.currentConfigApp.coverPosterWidthStatus,
    store.appStoreData.currentConfigApp.coverPosterWidthBase,
    store.appStoreData.currentConfigApp.coverPosterHeightStatus,
    store.appStoreData.currentConfigApp.coverPosterHeightBase,
  );
});

const titleStyleObj_C = computed<Record<string, string>>(() => ({
  textAlign: dataset.coverTitleAlign.includes(store.appStoreData.currentConfigApp.coverTitleAlign)
    ? store.appStoreData.currentConfigApp.coverTitleAlign
    : 'left',
}));

const performerNames_C = computed(() => {
  return props.resource.performers.map(item => item.name).filter(Boolean).join(' · ');
});

const tagFontSize_C = computed(() => store.appStoreData.currentConfigApp.coverDisplayTagFontSize || 12);
const tagRowHeight_C = computed(() => tagFontSize_C.value + 8);
const infoHeight_C = computed(() => 67 + tagFontSize_C.value * 2);
const cardStyle_C = computed<Record<string, string>>(() => ({
  width: `${coverPosterSize_C.value.width}px`,
  height: `${coverPosterSize_C.value.height + infoHeight_C.value}px`,
}));
const infoStyle_C = computed<Record<string, string>>(() => ({
  height: `${infoHeight_C.value}px`,
  '--cinema-gallery-row-height': `${tagRowHeight_C.value}px`,
  fontSize: `${tagFontSize_C.value}px`,
}));
const estimateTextWidth = (text: string, fontSize: number, horizontalPadding = 0) => {
  const textUnits = Array.from(text).reduce((width, character) => {
    return width + (character.charCodeAt(0) <= 0xff ? 0.62 : 1);
  }, 0);
  return Math.ceil(textUnits * fontSize + horizontalPadding);
};
const visiblePerformers_C = computed(() => {
  const performers = props.resource.performers;
  if (performers.length === 0) return [];

  const availableWidth = Math.max(0, coverPosterSize_C.value.width - 20);
  const fontSize = 12;
  const separatorWidth = estimateTextWidth(' · ', fontSize);
  const allWidth = performers.reduce((width, performer, index) => {
    return width + estimateTextWidth(performer.name, fontSize) + (index > 0 ? separatorWidth : 0);
  }, 0);
  if (allWidth <= availableWidth) return performers;

  const moreWidth = estimateTextWidth(`+${performers.length}`, fontSize, 10);
  const visibleWidth = Math.max(0, availableWidth - moreWidth - 5);
  let usedWidth = 0;
  let visibleCount = 0;
  for (const performer of performers) {
    const nextWidth = usedWidth + estimateTextWidth(performer.name, fontSize)
      + (visibleCount > 0 ? separatorWidth : 0);
    if (nextWidth > visibleWidth) break;
    usedWidth = nextWidth;
    visibleCount++;
  }
  return performers.slice(0, Math.max(1, visibleCount));
});
const visiblePerformerNames_C = computed(() => visiblePerformers_C.value.map(item => item.name).join(' · '));
const hiddenPerformerCount_C = computed(() => {
  return Math.max(0, props.resource.performers.length - visiblePerformers_C.value.length);
});
const visibleTags_C = computed(() => {
  const tags = props.resource.tags;
  if (tags.length === 0) return [];

  const availableWidth = Math.max(0, coverPosterSize_C.value.width - 20);
  const fontSize = tagFontSize_C.value;
  const gap = 5;
  const rowCount = props.resource.performers.length ? 1 : 2;
  const tagWidths = tags.map(tag => estimateTextWidth(tag.name, fontSize, 18));
  let row = 0;
  let rowWidth = 0;
  let allFit = true;
  for (const tagWidth of tagWidths) {
    const nextWidth = rowWidth + tagWidth + (rowWidth > 0 ? gap : 0);
    if (nextWidth <= availableWidth) {
      rowWidth = nextWidth;
      continue;
    }
    row++;
    rowWidth = tagWidth;
    if (row >= rowCount || tagWidth > availableWidth) {
      allFit = false;
      break;
    }
  }
  if (allFit) return tags;

  const moreTagWidth = estimateTextWidth(`+${tags.length}`, fontSize, 18);
  row = 0;
  rowWidth = 0;
  let visibleCount = 0;
  for (const tagWidth of tagWidths) {
    const rowLimit = row === rowCount - 1 ? Math.max(0, availableWidth - moreTagWidth - gap) : availableWidth;
    const nextWidth = rowWidth + tagWidth + (rowWidth > 0 ? gap : 0);
    if (nextWidth > rowLimit) {
      row++;
      rowWidth = 0;
      if (row >= rowCount) break;
    }
    const currentRowLimit = row === rowCount - 1 ? Math.max(0, availableWidth - moreTagWidth - gap) : availableWidth;
    if (tagWidth > currentRowLimit) break;
    rowWidth += tagWidth + (rowWidth > 0 ? gap : 0);
    visibleCount++;
  }

  return tags.slice(0, visibleCount);
});
const hiddenTagCount_C = computed(() => Math.max(0, props.resource.tags.length - visibleTags_C.value.length));
const hiddenTagsTitle_C = computed(() => {
  return props.resource.tags.slice(visibleTags_C.value.length).map(item => item.name).join('、');
});
</script>

<style lang="scss" scoped>
.cinema-gallery-card {
  position: relative;
  overflow: hidden;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  color: var(--el-text-color-primary);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 9px;
  background: var(--el-bg-color-overlay);
  box-shadow: var(--el-box-shadow-light);
  cursor: pointer;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;

  &:hover {
    border-color: var(--el-color-primary-light-5);
    box-shadow: var(--el-box-shadow);
    transform: translateY(-1px);

    .cinema-gallery-cover :deep(.el-image) {
      transform: scale(1.035);
    }

    .cinema-gallery-play {
      opacity: 1;
      transform: translate(-50%, -50%) scale(1);
    }
  }
}

.cinema-gallery-cover {
  position: relative;
  width: 100%;
  flex: 0 0 auto;
  overflow: hidden;
  background: var(--el-fill-color-dark);

  :deep(.el-image) {
    width: 100%;
    height: 100%;
    display: block;
    transition: transform 0.22s ease;
  }
}

.cinema-gallery-config-tags {
  position: absolute;
  top: 7px;
  left: 7px;
  z-index: 12;
  max-width: calc(100% - 14px);
  max-height: 48px;
  overflow: hidden;

  :deep(.content-tag-display) {
    gap: 4px !important;
  }

  :deep(.content-tag) {
    box-shadow: 0 2px 7px rgba(0, 0, 0, 0.22);
    backdrop-filter: blur(5px);
  }
}

.cinema-gallery-play {
  position: absolute;
  top: 50%;
  left: 50%;
  z-index: 13;
  width: 44px;
  height: 44px;
  padding: 0;
  display: grid;
  place-items: center;
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.52);
  border-radius: 50%;
  background: rgba(9, 15, 18, 0.58);
  box-shadow: 0 8px 22px rgba(0, 0, 0, 0.28);
  opacity: 0;
  cursor: pointer;
  backdrop-filter: blur(6px);
  transform: translate(-50%, -50%) scale(0.92);
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;

  .el-icon {
    font-size: 22px;
  }
}

.cinema-gallery-info {
  min-width: 0;
  padding: 9px 10px 10px;
  box-sizing: border-box;
  flex: 0 0 auto;
}

.cinema-gallery-title {
  margin: 0;
  overflow: hidden;
  color: var(--el-text-color-primary);
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
  line-height: 20px;
}

.cinema-gallery-meta {
  height: calc(var(--cinema-gallery-row-height) * 2 + 5px);
  margin-top: 7px;
  overflow: hidden;
}

.cinema-gallery-performers {
  height: var(--cinema-gallery-row-height);
  display: flex;
  align-items: center;
  gap: 5px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: var(--cinema-gallery-row-height);

  .performer-names {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .performer-more {
    padding: 0 4px;
    flex: 0 0 auto;
    border: 1px dashed var(--el-border-color);
    border-radius: 4px;
    line-height: calc(var(--cinema-gallery-row-height) - 4px);
  }
}

.cinema-gallery-tags {
  height: var(--cinema-gallery-row-height);
  margin-top: 5px;
  display: flex;
  align-items: center;
  gap: 5px;
  overflow: hidden;

  &.two-rows {
    height: calc(var(--cinema-gallery-row-height) * 2 + 5px);
    margin-top: 0;
    align-content: flex-start;
    align-items: flex-start;
    flex-wrap: wrap;
  }

  .el-tag {
    max-width: 100%;
    height: var(--cinema-gallery-row-height);
    padding: 0 7px;
    flex: 0 0 auto;
    color: var(--el-text-color-regular);
    border-color: var(--el-border-color);
    border-radius: 5px;
    background: var(--el-fill-color-light);
    font-size: inherit;
    line-height: calc(var(--cinema-gallery-row-height) - 2px);

    :deep(.el-tag__content) {
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .more-tag {
    flex: 0 0 auto;
    color: var(--el-text-color-secondary);
    border-style: dashed;
  }
}
</style>
