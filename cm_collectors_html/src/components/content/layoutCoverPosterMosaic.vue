<template>
  <div class="layout-cover-poster-mosaic">
    <div class="main">
      <el-scrollbar ref="scrollbarRef">
        <div ref="mosaicContainerRef" class="mosaic-list">
          <div
            v-for="(row, rowIndex) in mosaicRows"
            :key="`${layoutVersion}-${rowIndex}`"
            class="mosaic-row"
            :style="{
              height: `${row.height}px`,
              gap: `${gap}px`,
              marginBottom: `${gap}px`,
            }"
          >
            <div
              v-for="item in row.items"
              :key="item.data.resource.id"
              class="mosaic-item"
              :data-resource-id="item.data.resource.id"
              :style="{ width: `${item.width}px` }"
            >
              <contentRightClickMenu :resource="item.data.resource">
                <div
                  class="mosaic-card"
                  :class="{
                    active:
                      props.playerMode &&
                      item.data.resource.id === props.activeResourceId,
                  }"
                  @click="cardClickHandle(item.data.resource)"
                >
                  <el-image
                    :src="item.data.src"
                    :title="item.data.resource.title"
                    fit="cover"
                  />
                  <contentVideoDurationBadge
                    :resource="item.data.resource"
                    :offset-bottom="props.playerMode ? '8px' : '30px'"
                    offset-right="4px"
                  />
                  <template v-if="!props.playerMode">
                    <div class="title-bg"></div>
                    <div
                      class="resource-title"
                      :style="{ textAlign: coverTitleAlign }"
                    >
                      {{ item.data.resource.title }}
                    </div>
                  </template>
                  <div
                    class="play-icon"
                    @click.stop="cardActionHandle(item.data.resource)"
                  >
                    <el-icon>
                      <Setting v-if="props.playerMode" />
                      <VideoPlay v-else />
                    </el-icon>
                  </div>
                </div>
              </contentRightClickMenu>
            </div>
          </div>
        </div>
        <el-backtop
          class="custom-backtop"
          target=".layout-cover-poster-mosaic .el-scrollbar__wrap"
          :right="20"
          :bottom="20"
        />
      </el-scrollbar>
    </div>
    <div v-if="!isMobile()" class="tool">
      <span class="tool-label">基础列数</span>
      <el-slider
        v-model="mosaicColumn"
        :min="2"
        :max="props.maxColumn"
        style="width: 200px"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
import type { ElScrollbar } from 'element-plus';
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
  type PropType,
} from 'vue';

import { isMobile } from '@/assets/mobile';
import { playResource } from '@/common/play';
import { getResourceCoverPoster } from '@/common/photo';
import type { I_resource } from '@/dataType/resource.dataType';
import { appStoreData } from '@/storeData/app.storeData';

import contentRightClickMenu from './contentRightClickMenu.vue';
import contentVideoDurationBadge from './contentVideoDurationBadge.vue';
import { createMosaicRows, type I_mosaicSource } from './mosaicLayout';

interface I_mosaicResource {
  resource: I_resource;
  src: string;
}

const store = {
  appStoreData: appStoreData(),
};

const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
  playerMode: {
    type: Boolean,
    default: false,
  },
  activeResourceId: {
    type: String,
    default: '',
  },
  storageKeyPrefix: {
    type: String,
    default: 'layout-cover-poster-mosaic-column',
  },
  defaultColumn: {
    type: Number,
    default: 6,
  },
  maxColumn: {
    type: Number,
    default: 20,
  },
  limitColumnsByWidth: {
    type: Boolean,
    default: true,
  },
});

const emits = defineEmits(['selectResources', 'playResource']);

const scrollbarRef = ref<InstanceType<typeof ElScrollbar>>();
const mosaicContainerRef = ref<HTMLElement>();
const containerWidth = ref(0);
const layoutVersion = ref(0);
const defaultColumn = computed(
  () =>
    props.defaultColumn ||
    6,
);
const storageKey = computed(
  () =>
    `${props.storageKeyPrefix}-${store.appStoreData.currentFilesBases.id}`,
);
const getStoredColumn = () => {
  const value = Number.parseInt(
    localStorage.getItem(storageKey.value) || defaultColumn.value.toString(),
    10,
  );
  return Number.isFinite(value)
    ? Math.min(props.maxColumn, Math.max(2, value))
    : defaultColumn.value;
};
const mosaicColumn = ref(getStoredColumn());
const gap = computed(
  () => store.appStoreData.currentConfigApp.coverPosterGap ?? 10,
);
const coverTitleAlign = computed(() => {
  const align = store.appStoreData.currentConfigApp.coverTitleAlign;
  return align === 'center' || align === 'right' ? align : 'left';
});

const effectiveColumn = computed(() => {
  if (!props.limitColumnsByWidth) return mosaicColumn.value;
  if (containerWidth.value <= 520) return 2;
  if (containerWidth.value <= 800) return Math.min(mosaicColumn.value, 4);
  return mosaicColumn.value;
});

const mosaicSources = computed<I_mosaicSource<I_mosaicResource>[]>(() => {
  return props.dataList.map((resource) => {
    const width = resource.coverPosterWidth;
    const height = resource.coverPosterHeight;
    const aspectRatio =
      Number.isFinite(width) && Number.isFinite(height) && width > 0 && height > 0
        ? width / height
        : 2 / 3;

    return {
      data: {
        resource,
        src: getResourceCoverPoster(resource),
      },
      aspectRatio,
      preferredSpan: aspectRatio >= 1.15 ? 2 : 1,
    };
  });
});

const mosaicRows = computed(() => {
  return createMosaicRows(
    mosaicSources.value,
    effectiveColumn.value,
    containerWidth.value,
    gap.value,
  );
});

let resizeObserver: ResizeObserver | undefined;

const updateContainerWidth = () => {
  const width = mosaicContainerRef.value?.clientWidth || 0;
  if (Math.abs(containerWidth.value - width) < 1) return;
  containerWidth.value = width;
  layoutVersion.value++;
};

watch(mosaicColumn, (column) => {
  localStorage.setItem(storageKey.value, column.toString());
  layoutVersion.value++;
});

watch(
  () => store.appStoreData.currentFilesBases.id,
  () => {
    mosaicColumn.value = getStoredColumn();
    nextTick(updateContainerWidth);
  },
);

const selectResourcesHandle = (resource: I_resource) => {
  emits('selectResources', resource);
};

const cardClickHandle = (resource: I_resource) => {
  if (props.playerMode) {
    emits('playResource', resource);
    return;
  }
  selectResourcesHandle(resource);
};

const cardActionHandle = (resource: I_resource) => {
  if (props.playerMode) {
    selectResourcesHandle(resource);
    return;
  }
  playResource(resource);
};

const change = () => {
  scrollbarRef.value?.setScrollTop(0);
};

const scrollToResource = (resourceId: string) => {
  if (!resourceId) return;
  nextTick(() => {
    const container = scrollbarRef.value?.wrapRef;
    const target = Array.from(
      mosaicContainerRef.value?.querySelectorAll<HTMLElement>('.mosaic-item') ||
        [],
    ).find((item) => item.dataset.resourceId === resourceId);
    if (!container || !target) return;

    const containerRect = container.getBoundingClientRect();
    const targetRect = target.getBoundingClientRect();
    const relativeTop = targetRect.top - containerRect.top;
    const scrollTop =
      container.scrollTop +
      relativeTop -
      containerRect.height / 2 +
      targetRect.height / 2;

    container.scrollTo({
      top: Math.max(0, scrollTop),
      behavior: 'smooth',
    });
  });
};

onMounted(() => {
  nextTick(() => {
    updateContainerWidth();
    if (!mosaicContainerRef.value) return;
    resizeObserver = new ResizeObserver(updateContainerWidth);
    resizeObserver.observe(mosaicContainerRef.value);
  });
});

onBeforeUnmount(() => {
  resizeObserver?.disconnect();
});

defineExpose({ change, scrollToResource });
</script>

<style lang="scss" scoped>
.layout-cover-poster-mosaic {
  display: flex;
  width: 100%;
  height: 100%;
  overflow: hidden;
  flex-direction: column;

  .main {
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  .mosaic-list {
    width: calc(100% - 14px);
    padding-bottom: 1em;
  }

  .main :deep(.el-scrollbar__bar.is-vertical) {
    right: 3px;
  }

  .mosaic-row {
    display: flex;
    width: 100%;
    overflow: hidden;
  }

  .mosaic-item {
    height: 100%;
    min-width: 0;
    flex-shrink: 0;
    overflow: hidden;
  }

  .mosaic-item :deep(.content-right-click-menu) {
    width: 100%;
    height: 100%;
  }

  .mosaic-card {
    position: relative;
    width: 100%;
    height: 100%;
    overflow: hidden;
    cursor: pointer;
    background: #181818;
    border-radius: 3px;
    transition:
      box-shadow 0.2s ease,
      transform 0.2s ease;

    &.active {
      box-shadow:
        inset 0 0 0 2px #409eff,
        0 0 8px rgba(64, 158, 255, 0.5);
    }

    &:hover {
      .play-icon {
        display: block;
      }

      .el-image {
        transform: scale(1.015);
      }
    }

    .el-image {
      width: 100%;
      height: 100%;
      transition: transform 0.2s ease;
    }
  }

  .play-icon {
    position: absolute;
    bottom: 0.12em;
    left: 0.05em;
    z-index: 10;
    display: none;
    color: #f3f3f3;
    font-size: 3.8em;
    cursor: pointer;
    opacity: 0.75;
  }

  .title-bg,
  .resource-title {
    position: absolute;
    right: 0;
    bottom: 0;
    left: 0;
    height: 26px;
  }

  .title-bg {
    z-index: 5;
    background: linear-gradient(
      to top,
      rgba(18, 18, 18, 0.88),
      rgba(18, 18, 18, 0.48)
    );
  }

  .resource-title {
    z-index: 6;
    box-sizing: border-box;
    padding: 0 7px;
    overflow: hidden;
    color: #f3f3f3;
    font-size: 13px;
    line-height: 26px;
    text-overflow: ellipsis;
    white-space: nowrap;
    pointer-events: none;
  }

  .tool {
    display: flex;
    height: 32px;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    gap: 12px;
  }

  .tool-label {
    color: var(--el-text-color-regular);
    font-size: 12px;
    white-space: nowrap;
  }
}
</style>
