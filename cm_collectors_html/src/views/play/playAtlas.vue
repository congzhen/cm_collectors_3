<template>
  <div class="play-atlas-container">
    <HeaderView class="header" :mode="E_headerMode.GoBack" :title="resourceInfo?.title || ''"></HeaderView>
    <div class="main-container" v-loading="loading">
      <div class="main" ref="mainRef" v-if="resourceInfo">
        <Waterfall ref="waterfallRef" :list="waterfallList" :gutter="10" :breakpoints="waterfallBreakpoints"
          :img-selector="'src'" class="atlas-list" @scroll="handleScroll">
          <template #default="{ item }">
            <el-image class="atlas-image" :src="item.src" @click="openImageViewer(item.id)" @load="onImageLoad" />
          </template>
        </Waterfall>
        <div class="atlas-tool-btn">
          <el-slider v-model="waterfallColumn" :min="1" :max="20" style="width: 200px;" />
          <div>Total: {{ atlasImageList.length }}</div>
          <!-- 添加"加载更多"按钮 -->
          <el-button v-if="displayedCount < atlasImageList.length" @click="loadMoreImages" :loading="isHandlingScroll"
            style="width: 230px;">
            Load More ({{ atlasImageList.length - displayedCount }} remaining)
          </el-button>
          <el-select v-model="selectedDramaSeriesId" @change="getResourceAtlas" style="width: 180px">
            <el-option v-for="item, key in resourceInfo.dramaSeries" :key="key" :label="getFinalPathSegment(item.src)"
              :value="item.id" />
          </el-select>
        </div>
      </div>
    </div>
    <imageViewer ref="imageViewerRef" :imageList="atlasImageListSrc_C"></imageViewer>
  </div>
</template>
<script lang="ts" setup>
import type { I_resource } from '@/dataType/resource.dataType';
import HeaderView from '../HeaderView.vue'
import imageViewer from '@/components/play/imageViewer.vue';
import { E_headerMode } from '@/dataType/app.dataType'
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import { resourceServer } from '@/server/resource.server';
import { ElMessage } from 'element-plus';
import { filesServer } from '@/server/files.server';
import { getFinalPathSegment } from '@/assets/tool'
import { getFileImageByDramaSeriesId } from '@/common/photo'
import { Waterfall } from 'vue-waterfall-plugin-next'
import 'vue-waterfall-plugin-next/dist/style.css'
import { debounce } from '@/assets/debounce';
const props = defineProps({
  resourceId: {
    type: String,
    required: true,
  },
  dramaSeriesId: {
    type: String,
    default: '',
  },
})

const mainRef = ref<HTMLDivElement>();
const waterfallRef = ref<InstanceType<typeof Waterfall>>();
const imageViewerRef = ref<InstanceType<typeof imageViewer>>();

const loading = ref(false);
const resourceInfo = ref<I_resource>();
const atlasImageList = ref<string[]>([]);
const selectedDramaSeriesId = ref<string>('');
const waterfallColumn = ref(6);

const isFirstLoadCompleted = ref(false); // 是否首次加载完成
const loadedImageCount = ref(0); // 已加载的图片数量
const displayedCount = ref(60);//控制当前显示的图片数量
const incrementCount = 20; // 每次加载的图片数量

watch(waterfallColumn, () => {
  // 列数改变时，重新检查是否需要加载更多
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
    }))
})
const atlasImageListSrc_C = computed(() => {
  return atlasImageList.value.map(fileName => getFileImageByDramaSeriesId(selectedDramaSeriesId.value, fileName))
})



// 计算动态 breakpoints
const waterfallBreakpoints = computed(() => {
  // 可以根据当前列数设置不同的断点
  return {
    9999: { rowPerView: waterfallColumn.value },
  }
});


const init = async () => {
  await getResourceInfo();
  setDramaSeries();
  await getResourceAtlas();
}
const setDramaSeries = () => {
  if (props.dramaSeriesId != '') {
    selectedDramaSeriesId.value = props.dramaSeriesId;
  } else if (resourceInfo.value && resourceInfo.value.dramaSeries.length > 0) {
    selectedDramaSeriesId.value = resourceInfo.value.dramaSeries[0].id;
  }
}

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
  displayedCount.value = Math.min(60, atlasImageList.value.length);
  loading.value = false;
  // 在下一个 tick 检查是否需要加载更多
  nextTick(() => {
    checkAndLoadMore();
  });
}

// 添加防抖和加载锁
const isHandlingScroll = ref(false);
const scrollTimer = ref<number | null>(null);


// 检查并加载更多图片
const checkAndLoadMore = () => {
  // 只有在首次加载完成后才执行自动加载更多
  if (!isFirstLoadCompleted.value) return;

  const container = document.querySelector('.atlas-list');
  if (!container) return;

  const { scrollHeight, clientHeight } = container;

  // 如果内容高度小于等于容器高度，且还有更多图片未加载
  if (scrollHeight <= clientHeight && displayedCount.value < atlasImageList.value.length) {
    // 加载下一批图片
    loadMoreImages();

    // 使用 setTimeout 给 DOM 更新和渲染一些时间，然后再检查
    setTimeout(() => {
      // 再次检查容器状态，决定是否继续加载
      const container = document.querySelector('.atlas-list');
      if (container) {
        const { scrollHeight, clientHeight } = container;
        // 只有在确实需要更多内容填满容器时才继续
        if (scrollHeight <= clientHeight && displayedCount.value < atlasImageList.value.length) {
          // 添加一个安全限制，避免无限递归
          if (displayedCount.value < Math.min(200, atlasImageList.value.length)) {
            checkAndLoadMore();
          }
        }
      }
    }, 100); // 给 100ms 时间让 DOM 更新
  }
}


// 优化滚动处理函数
const handleScroll = (event: Event) => {
  const target = event.target as HTMLElement;
  const { scrollTop, scrollHeight, clientHeight } = target;

  // 防抖处理
  if (scrollTimer.value) {
    window.clearTimeout(scrollTimer.value);
  }

  scrollTimer.value = window.setTimeout(() => {
    // 当滚动接近底部时增加显示数量
    if (scrollHeight - scrollTop - clientHeight < 200) {
      loadMoreImages();
    }
  }, 100);
}

// 优化加载更多图片函数
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

  // 使用 nextTick 确保 DOM 更新完成后再继续
  nextTick(() => {
    setTimeout(() => {
      isHandlingScroll.value = false;
    }, 50);
  });
}

// 图片加载完成事件处理
const onImageLoad = debounce(() => {
  nextTick(() => {
    waterfallRef.value?.renderer();
  });
  // 增加已加载图片计数
  loadedImageCount.value++;

  // 当加载的图片数量达到当前显示的数量时，标记首次加载完成
  if (!isFirstLoadCompleted.value && loadedImageCount.value >= displayedCount.value) {
    isFirstLoadCompleted.value = true;

    // 延迟一点时间再检查是否需要加载更多，确保布局已经稳定
    setTimeout(() => {
      checkAndLoadMore();
    }, 200);
  }
}, 1000);


const openImageViewer = (index: number) => {
  imageViewerRef.value?.openImageViewer(index);
}

onMounted(async () => {
  await init();
});

</script>
<style lang="scss" scoped>
.play-atlas-container {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;

  :deep(.el-image-viewer__wrapper) {
    background-color: rgba($color: #000000, $alpha: 0.7);
  }

  :deep(.atlas-image) {
    cursor: pointer;

    &:hover {
      transform: scale(1.1);
    }
  }

  .main-container {
    flex: 1;
    overflow: hidden;

    .main {
      width: 100%;
      height: 100%;
      display: flex;
      flex-direction: column;

      .atlas-list {
        flex: 1;
        width: 100%;
        height: 100%;
        overflow-y: auto;
        padding: 10px;
        box-sizing: border-box;
        background-color: unset;
      }

      :deep(.waterfall-item) {
        margin-bottom: 10px;
        border-radius: 4px;
        overflow: hidden;
      }

      .atlas-tool-btn {
        flex-shrink: 0;
        height: 36px;
        padding: 10px 10px 0 0;
        display: flex;
        justify-content: flex-end;
        gap: 20px;
        line-height: 32px;
        font-size: 14px;
      }
    }
  }
}
</style>
