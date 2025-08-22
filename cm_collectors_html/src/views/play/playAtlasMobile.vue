<template>
  <div class="play-atlas-mobile-container">
    <MobileHeader :title="resourceInfo?.title || ''" :show-menu-button="true" @menu-action="handleMenuAction" />
    <div class="main-container" v-loading="loading">
      <div class="main" ref="mainRef" v-if="resourceInfo">
        <!-- 移动端顶部控制栏 -->
        <div class="mobile-controls">
          <div class="info-text">Total: {{ atlasImageList.length }}</div>
          <el-select v-model="selectedDramaSeriesId" @change="getResourceAtlas" class="episode-select">
            <el-option v-for="item in resourceInfo.dramaSeries" :key="item.id" :label="getFinalPathSegment(item.src)"
              :value="item.id" />
          </el-select>
        </div>

        <!-- 为移动端优化的瀑布流 -->
        <Waterfall ref="waterfallRef" :list="waterfallList" :gutter="8" :breakpoints="waterfallBreakpoints"
          :img-selector="'src'" class="atlas-list" @scroll="handleScroll">
          <template #default="{ item }">
            <el-image class="atlas-image" :src="item.src" @click="openImageViewer(item.id)" @load="onImageLoad" lazy
              fit="cover" />
          </template>
        </Waterfall>

        <!-- 底部加载更多指示器 -->
        <div v-if="displayedCount < atlasImageList.length" class="load-more-container">
          <el-button @click="loadMoreImages" :loading="isHandlingScroll" type="primary" plain size="small">
            Load More ({{ atlasImageList.length - displayedCount }} remaining)
          </el-button>
        </div>
      </div>
    </div>

    <imageViewer ref="imageViewerRef" :imageList="atlasImageListSrc_C"></imageViewer>
  </div>
</template>

<script lang="ts" setup>
import type { I_resource } from '@/dataType/resource.dataType';
import imageViewer from '@/components/play/imageViewer.vue';
import MobileHeader from '../MobileHeaderView.vue'
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import { resourceServer } from '@/server/resource.server';
import { ElMessage } from 'element-plus';
import { filesServer } from '@/server/files.server';
import { getFinalPathSegment } from '@/assets/tool';
import { getFileImageByDramaSeriesId } from '@/common/photo';
import { Waterfall } from 'vue-waterfall-plugin-next';
import 'vue-waterfall-plugin-next/dist/style.css';
import { debounce } from '@/assets/debounce';
import { useRouter } from 'vue-router';

const router = useRouter();

const props = defineProps({
  resourceId: {
    type: String,
    required: true,
  },
  dramaSeriesId: {
    type: String,
    default: '',
  },
});

const mainRef = ref<HTMLDivElement>();
const waterfallRef = ref<InstanceType<typeof Waterfall>>();
const imageViewerRef = ref<InstanceType<typeof imageViewer>>();

const loading = ref(false);
const resourceInfo = ref<I_resource>();
const atlasImageList = ref<string[]>([]);
const selectedDramaSeriesId = ref<string>('');
const waterfallColumn = ref(2); // 移动端默认2列

const isFirstLoadCompleted = ref(false);
const loadedImageCount = ref(0);
const displayedCount = ref(20); // 移动端初始显示数量更少
const incrementCount = 10; // 移动端每次加载更少图片
const isHandlingScroll = ref(false);

// 监听列数变化
watch(waterfallColumn, () => {
  nextTick(() => {
    setTimeout(() => {
      checkAndLoadMore();
    }, 100);
  });
});

// 计算瀑布流列表数据
const waterfallList = computed(() => {
  const mainWidth = mainRef.value?.clientWidth || 0;
  const thumbWidth = Math.floor(mainWidth / waterfallColumn.value);
  return atlasImageList.value
    .slice(0, displayedCount.value)
    .map((fileName, index) => ({
      src: getFileImageByDramaSeriesId(selectedDramaSeriesId.value, fileName, thumbWidth),
      id: index
    }));
});

// 完整图片地址列表
const atlasImageListSrc_C = computed(() => {
  return atlasImageList.value.map(fileName =>
    getFileImageByDramaSeriesId(selectedDramaSeriesId.value, fileName)
  );
});

// 计算动态 breakpoints
const waterfallBreakpoints = computed(() => {
  return {
    9999: { rowPerView: waterfallColumn.value },
  };
});

const init = async () => {
  await getResourceInfo();
  setDramaSeries();
  await getResourceAtlas();
};

const setDramaSeries = () => {
  if (props.dramaSeriesId != '') {
    selectedDramaSeriesId.value = props.dramaSeriesId;
  } else if (resourceInfo.value && resourceInfo.value.dramaSeries.length > 0) {
    selectedDramaSeriesId.value = resourceInfo.value.dramaSeries[0].id;
  }
};

const getResourceInfo = async () => {
  loading.value = true;
  const result = await resourceServer.info(props.resourceId);
  if (!result || !result.status) {
    ElMessage.error(result.msg);
    return;
  }
  resourceInfo.value = result.data;
  loading.value = false;
};

const getResourceAtlas = async () => {
  loading.value = true;
  const result = await filesServer.filesDListByDramaSeriesId_Image(selectedDramaSeriesId.value);
  if (!result || !result.status) {
    ElMessage.error(result.msg);
    return;
  }
  atlasImageList.value = result.data;
  // 重置显示数量
  displayedCount.value = Math.min(20, atlasImageList.value.length);
  loading.value = false;
  // 在下一个 tick 检查是否需要加载更多
  nextTick(() => {
    checkAndLoadMore();
  });
};

// 检查并加载更多图片
const checkAndLoadMore = () => {
  if (!isFirstLoadCompleted.value) return;

  const container = document.querySelector('.atlas-list');
  if (!container) return;

  const { scrollHeight, clientHeight } = container;

  if (scrollHeight <= clientHeight && displayedCount.value < atlasImageList.value.length) {
    loadMoreImages();

    setTimeout(() => {
      const container = document.querySelector('.atlas-list');
      if (container) {
        const { scrollHeight, clientHeight } = container;
        if (scrollHeight <= clientHeight && displayedCount.value < atlasImageList.value.length) {
          if (displayedCount.value < Math.min(100, atlasImageList.value.length)) {
            checkAndLoadMore();
          }
        }
      }
    }, 100);
  }
};

// 滚动处理函数
const handleScroll = debounce((event: Event) => {
  const target = event.target as HTMLElement;
  const { scrollTop, scrollHeight, clientHeight } = target;

  // 当滚动接近底部时增加显示数量
  if (scrollHeight - scrollTop - clientHeight < 100) {
    loadMoreImages();
  }
}, 300);

// 加载更多图片函数
const loadMoreImages = () => {
  // 防止重复触发
  if (isHandlingScroll.value || displayedCount.value >= atlasImageList.value.length) {
    return;
  }

  isHandlingScroll.value = true;

  // 增加显示数量
  displayedCount.value = Math.min(
    displayedCount.value + incrementCount,
    atlasImageList.value.length
  );

  // 确保 DOM 更新完成后再继续
  nextTick(() => {
    setTimeout(() => {
      isHandlingScroll.value = false;
    }, 50);
  });
};

// 图片加载完成事件处理
const onImageLoad = debounce(() => {
  nextTick(() => {
    waterfallRef.value?.renderer();
  });
  loadedImageCount.value++;

  if (!isFirstLoadCompleted.value && loadedImageCount.value >= displayedCount.value) {
    isFirstLoadCompleted.value = true;

    setTimeout(() => {
      checkAndLoadMore();
    }, 200);
  }
}, 500);

const openImageViewer = (index: number) => {
  imageViewerRef.value?.openImageViewer(index);
};

// 处理菜单操作
const handleMenuAction = (action: string) => {
  switch (action) {
    case 'goBack':
      router.go(-1);
      break;
    case 'goHome':
      router.push('/');
      break;
  }
};

onMounted(async () => {
  nextTick(async () => {
    await init();
  });
});
</script>

<style lang="scss" scoped>
.play-atlas-mobile-container {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;

  :deep(.el-image-viewer__wrapper) {
    background-color: rgba($color: #000000, $alpha: 0.9);
  }

  :deep(.atlas-image) {
    cursor: pointer;
    display: block;
  }

  .main-container {
    flex: 1;
    overflow: hidden;

    .main {
      width: 100%;
      height: 100%;
      display: flex;
      flex-direction: column;

      .mobile-controls {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 5px 10px;

        .info-text {
          font-size: 14px;
          color: #666;
        }

        .episode-select {
          width: 60%;
        }
      }

      .atlas-list {
        flex: 1;
        width: 100%;
        height: 100%;
        overflow-y: auto;
        padding: 5px;
        box-sizing: border-box;
        background-color: unset;
      }

      :deep(.waterfall-item) {
        margin-bottom: 8px;
        border-radius: 6px;
        overflow: hidden;
        box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
      }

      .load-more-container {
        display: flex;
        justify-content: center;
        padding: 15px;
      }
    }
  }
}
</style>
