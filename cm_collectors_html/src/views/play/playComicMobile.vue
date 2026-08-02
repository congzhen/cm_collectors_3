<template>
  <div class="play-comic-mobile-container">
    <MobileHeaderView class="header" :title="resourceInfo?.title || ''" :show-menu-button="true" />
    <div class="main-container" v-loading="loading">
      <div class="main" v-if="resourceInfo">
        <!-- 漫画显示区域 -->
        <div v-if="comicImageList.length > 0" class="comic-image-container" ref="readImageRef" @touchstart="handleTouchStart"
          @touchmove="handleTouchMove" @touchend="handleTouchEnd">
          <el-image class="comic-image"
            :src="getFileImageByDramaSeriesId(selectedDramaSeriesId, comicImageList[nowPage])" fit="contain"
            :style="{ width: '100%' }" />
        </div>
        <div v-else class="comic-empty">暂无可显示的漫画图片</div>

        <!-- 底部控制区域 -->
        <div class="bottom-controls">
          <div class="page-info">
            第 {{ nowPage + 1 }} 页 / 共 {{ comicImageList.length }} 页
          </div>

          <!-- 阅读操作保持单排，避免占用过多漫画显示空间 -->
          <div class="reader-actions">
            <el-button plain size="small" @click="changeNowPage('per')" :disabled="nowPage <= 0">
              上一页
            </el-button>
            <el-button size="small" @click="toggleDramaSeries" :type="showDramaSeries ? 'primary' : 'default'">
              {{ showDramaSeries ? '隐藏选集' : '显示选集' }}
            </el-button>
            <el-button size="small" @click="toggleThumbnail" :type="showThumbnail ? 'primary' : 'default'">
              {{ showThumbnail ? '隐藏缩略图' : '显示缩略图' }}
            </el-button>
            <el-button plain size="small" @click="changeNowPage('next')"
              :disabled="nowPage >= comicImageList.length - 1">
              下一页
            </el-button>
          </div>

          <!-- 选集区域 -->
          <div v-show="showDramaSeries" class="drama-series-selector">
            <resourceDramaSeriesList :drama-series="resourceInfo.dramaSeries" :selected-id="selectedDramaSeriesId"
              :show-mode="store.appStoreData.currentFilesBasesAppConfig.detailsDramaSeriesMode"
              @play-resource-drama-series="playResourceDramaSeriesHandle">
            </resourceDramaSeriesList>
          </div>

          <!-- 缩略图列表 -->
          <div v-show="showThumbnail" class="thumbnail-container">
            <div class="thumbnail-scroll" ref="thumbnailRef">
              <div v-for="filesName, key in comicImageList" :key="key" :class="{ active: key === nowPage }"
                class="thumbnail-item" @click="setPage(key)">
                <el-image :src="getFileImageByDramaSeriesId(selectedDramaSeriesId, filesName)" fit="cover"
                  class="thumbnail-image" />
                <div class="page-number">({{ key + 1 }})</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { nextTick, onMounted, ref, watch } from 'vue';
import MobileHeaderView from '@/views/MobileHeaderView.vue';
import type { I_resource, I_resourceDramaSeries } from '@/dataType/resource.dataType';
import { resourceServer } from '@/server/resource.server';
import { ElMessage } from 'element-plus';
import resourceDramaSeriesList from '@/components/resource/resourceDramaSeriesList.vue'
import { appStoreData } from '@/storeData/app.storeData';
import { filesServer } from '@/server/files.server';
import { getFileImageByDramaSeriesId } from '@/common/photo'

const store = {
  appStoreData: appStoreData(),
}

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

// 触摸事件相关变量
const touchStartY = ref(0);
const touchStartX = ref(0);
const startY = ref(0);
const scrollTop = ref(0);
const swipeHandled = ref(false)

const readImageRef = ref<HTMLDivElement>();
const thumbnailRef = ref<HTMLDivElement>();

const resourceInfo = ref<I_resource>();
const comicImageList = ref<string[]>([]);
const selectedDramaSeriesId = ref<string>('');
const nowPage = ref(0);
const loading = ref(false);
let comicRequestVersion = 0
const showThumbnail = ref(false); // 控制缩略图显示/隐藏
const showDramaSeries = ref(false); // 控制选集显示/隐藏

// 初始化
const init = async () => {
  nowPage.value = 0;
  const loaded = await getResourceInfo();
  if (!loaded) return
  setDramaSeries();
  await getResourceComic();
}

// 设置当前剧集
const setDramaSeries = () => {
  if (props.dramaSeriesId != '') {
    selectedDramaSeriesId.value = props.dramaSeriesId;
  } else if (resourceInfo.value && resourceInfo.value.dramaSeries.length > 0) {
    selectedDramaSeriesId.value = resourceInfo.value.dramaSeries[0].id;
  }
}

// 获取资源信息
const getResourceInfo = async () => {
  loading.value = true;
  try {
    const result = await resourceServer.info(props.resourceId);
    if (!result || !result.status) {
      ElMessage.error(result?.msg || '资源信息加载失败');
      return false;
    }
    resourceInfo.value = result.data;
    return true
  } catch (error) {
    console.error(error)
    ElMessage.error('资源信息加载失败')
    return false
  } finally {
    loading.value = false;
  }
};

// 获取漫画图片列表
const getResourceComic = async () => {
  const requestVersion = ++comicRequestVersion
  if (!selectedDramaSeriesId.value) {
    comicImageList.value = []
    loading.value = false
    return
  }
  loading.value = true;
  try {
    const result = await filesServer.filesDListByDramaSeriesId_Image(selectedDramaSeriesId.value);
    if (requestVersion !== comicRequestVersion) return
    if (!result || !result.status) {
      ElMessage.error(result?.msg || '漫画图片加载失败');
      comicImageList.value = []
      return;
    }
    comicImageList.value = result.data;
  } catch (error) {
    if (requestVersion !== comicRequestVersion) return
    console.error(error)
    comicImageList.value = []
    ElMessage.error('漫画图片加载失败')
  } finally {
    if (requestVersion === comicRequestVersion) loading.value = false;
  }
}

// 切换剧集
const playResourceDramaSeriesHandle = async (ds: I_resourceDramaSeries) => {
  selectedDramaSeriesId.value = ds.id;
  await getResourceComic();
  nowPage.value = 0;
  // 切换剧集后自动隐藏选集面板
  showDramaSeries.value = false;
}

// 切换页面
const changeNowPage = (mode: 'per' | 'next') => {
  if (mode == 'per' && nowPage.value > 0) {
    nowPage.value = nowPage.value - 1;
  } else if (mode == 'next' && nowPage.value < comicImageList.value.length - 1) {
    nowPage.value = nowPage.value + 1;
  }
  nextTick(() => {
    try {
      scrollToThumbnail();
      // 重置滚动位置到顶部
      if (readImageRef.value) {
        readImageRef.value.scrollTop = 0;
      }
    } catch (e) {
      console.log(e);
    }
  })
}

// 设置当前页
const setPage = (page: number) => {
  nowPage.value = page;
  // 点击缩略图后自动隐藏缩略图面板
  showThumbnail.value = false;
  nextTick(() => {
    try {
      scrollToThumbnail();
      // 重置滚动位置到顶部
      if (readImageRef.value) {
        readImageRef.value.scrollTop = 0;
      }
    } catch (e) {
      console.log(e);
    }
  })
}

// 切换缩略图显示/隐藏
const toggleThumbnail = () => {
  showThumbnail.value = !showThumbnail.value;
  // 确保选集面板关闭
  if (showThumbnail.value) {
    showDramaSeries.value = false;
  }
  if (showThumbnail.value) {
    nextTick(() => {
      scrollToThumbnail();
    });
  }
}

// 切换选集显示/隐藏
const toggleDramaSeries = () => {
  showDramaSeries.value = !showDramaSeries.value;
  // 确保缩略图面板关闭
  if (showDramaSeries.value) {
    showThumbnail.value = false;
  }
}

// 滚动到当前缩略图
const scrollToThumbnail = () => {
  nextTick(() => {
    if (thumbnailRef.value) {
      const itemWidth = 90; // 每个缩略图的宽度+间隔
      const container = thumbnailRef.value;
      const scrollPosition = (nowPage.value * itemWidth) - (container.clientWidth / 2) + (itemWidth / 2);
      container.scrollTo({
        left: Math.max(0, scrollPosition),
        behavior: 'smooth'
      });
    }
  });
}

// 触摸事件处理 - 开始
const handleTouchStart = (e: TouchEvent) => {
  swipeHandled.value = false
  touchStartY.value = e.touches[0].clientY;
  touchStartX.value = e.touches[0].clientX;

  // 记录当前滚动位置
  if (readImageRef.value) {
    startY.value = e.touches[0].clientY;
    scrollTop.value = readImageRef.value.scrollTop;
  }
}

// 触摸事件处理 - 移动
const handleTouchMove = (e: TouchEvent) => {
  if (!readImageRef.value) return;

  const touchY = e.touches[0].clientY;
  const touchX = e.touches[0].clientX;
  const deltaY = touchStartY.value - touchY;
  const deltaX = touchStartX.value - touchX;

  // 如果水平滑动距离大于垂直滑动距离，则切换页面
  if (Math.abs(deltaX) > Math.abs(deltaY)) {
    if (swipeHandled.value) return
    if (deltaX > 50) { // 向左滑动
      changeNowPage('next');
      swipeHandled.value = true
      if (e.cancelable) e.preventDefault()
    } else if (deltaX < -50) { // 向右滑动
      changeNowPage('per');
      swipeHandled.value = true
      if (e.cancelable) e.preventDefault()
    }
  } else {
    // 垂直滑动时滚动图片
    const deltaY = startY.value - touchY;
    readImageRef.value.scrollTop = scrollTop.value + deltaY;
  }
}

// 触摸事件处理 - 结束
const handleTouchEnd = () => {
  swipeHandled.value = false
}

// 监听当前页变化，自动滚动缩略图
watch(nowPage, () => {
  scrollToThumbnail();
});

onMounted(async () => {
  nextTick(async () => {
    await init();
  })
});
</script>

<style lang="scss" scoped>
.play-comic-mobile-container {
  --comic-mobile-bg: #121619;
  --comic-mobile-surface: #1b2024;
  --comic-mobile-border: rgba(255, 255, 255, 0.12);
  --comic-mobile-text: #edf1f3;
  --comic-mobile-muted: #a0aab2;

  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: var(--comic-mobile-text);
  background-color: var(--comic-mobile-bg);

  .header {
    flex-shrink: 0;
  }

  .main-container {
    flex: 1;
    overflow: hidden;

    .main {
      width: 100%;
      height: 100%;
      display: flex;
      flex-direction: column;

      .comic-image-container {
        flex: 1;
        overflow-y: auto;
        display: flex;
        align-items: flex-start;
        justify-content: center;
        touch-action: pan-y;
        background: var(--comic-mobile-bg);

        .comic-image {
          max-width: 100%;
          flex-shrink: 0;
        }
      }

      .comic-empty {
        flex: 1;
        display: grid;
        place-items: center;
        color: var(--comic-mobile-muted);
      }

      .bottom-controls {
        background-color: var(--comic-mobile-surface);
        color: var(--comic-mobile-text);
        border-top: 1px solid var(--comic-mobile-border);
        flex-shrink: 0;

        .page-info {
          padding: 5px 10px 3px;
          color: var(--comic-mobile-muted);
          font-size: 11px;
          line-height: 1.2;
          text-align: center;
          white-space: nowrap;
        }

        .reader-actions {
          display: grid;
          grid-template-columns: 0.9fr 1.1fr 1.3fr 0.9fr;
          align-items: center;
          gap: 4px;
          padding: 3px 8px calc(8px + env(safe-area-inset-bottom));

          .el-button {
            width: 100%;
            min-width: 0;
            height: 30px;
            padding: 4px 3px;
            margin: 0 !important;
            font-size: 12px;
          }
        }

        .drama-series-selector {
          padding: 0 10px 10px;
          max-height: 30vh;
          overflow-y: auto;
        }

        .thumbnail-container {
          padding: 0 10px 10px;

          .thumbnail-scroll {
            display: flex;
            overflow-x: auto;
            gap: 10px;
            padding: 5px 0;

            .thumbnail-item {
              flex-shrink: 0;
              width: 80px;
              border-radius: 4px;
              transition: all 0.3s;
              cursor: pointer;
              border: 2px solid transparent;

              &.active {
                border: 2px solid #409eff;
                background-color: rgba(64, 158, 255, 0.1);
              }

              .thumbnail-image {
                width: 100%;
                height: 100px;
              }

              .page-number {
                text-align: center;
                color: var(--comic-mobile-text);
                font-size: 12px;
                padding: 2px 0;
              }
            }
          }
        }
      }
    }
  }
}

:global(html.bright) .play-comic-mobile-container {
  --comic-mobile-bg: #f5f7f8;
  --comic-mobile-surface: #ffffff;
  --comic-mobile-border: #dce3e7;
  --comic-mobile-text: #23323b;
  --comic-mobile-muted: #71808a;
}

@media (max-width: 360px) {
  .play-comic-mobile-container .main-container .main .bottom-controls .reader-actions {
    gap: 3px;
    padding-inline: 5px;

    .el-button {
      font-size: 11px;
    }
  }
}
</style>
