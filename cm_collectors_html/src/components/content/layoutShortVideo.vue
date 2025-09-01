<template>
  <div class="layout-short-video">
    <div class="shortVideoList">
      <div class="shortVideoListWaterfall">
        <el-scrollbar ref="scrollbarRef">
          <Waterfall ref="waterfallRef" :list="dataList_C" :gutter="10" :breakpoints="waterfallBreakpoints_C"
            :img-selector="'src'" class="waterfall-list">
            <template #default="{ item, index }">
              <contentRightClickMenu :resource="props.dataList[index]">
                <div class="waterfall-item"
                  :class="{ active: index === currentPlayIndex, ['waterfall-item-' + index]: true }"
                  @click.stop="clickResourceHandle(index)">

                  <el-image :src="item.src" :title="item.title" @load="onImageLoad" />

                  <div v-if="!isMobile()" class="play-icon" @click.stop="selectResourcesHandle(item)">
                    <el-icon>
                      <Setting />
                    </el-icon>
                  </div>
                </div>
              </contentRightClickMenu>
            </template>
          </Waterfall>
        </el-scrollbar>
      </div>
      <div class="tool" v-if="!isMobile()">
        <el-slider v-model="waterfallColumn" :min="2" :max="8" style="width: 200px;" />
      </div>
    </div>
    <div class="shortVideoPlay" :class="{ 'mobile-layout': isMobile() }">
      <div ref="shortVideoPlayContainerRef" class="shortVideoPlayContainer">
        <videoPlay ref="videoPlayRef" />
      </div>
      <div class="tool">
        <el-button-group :size="isMobile() ? 'small' : 'default'">
          <el-button icon="ArrowLeft" :disabled="currentPlayIndex == 0" @click="prevHandle()">上一个</el-button>
          <el-button @click="nextHandle()" :disabled="currentPlayIndex >= (dataListWrapper.length - 1)">
            下一个
            <el-icon class="el-icon--right">
              <ArrowRight />
            </el-icon>
          </el-button>
        </el-button-group>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { Waterfall } from 'vue-waterfall-plugin-next'
import { isMobile } from '@/assets/mobile';
import { getResourceCoverPoster } from '@/common/photo';
import type { I_resource } from '@/dataType/resource.dataType';
import { computed, nextTick, ref, type PropType, onMounted, watch, onBeforeUnmount } from 'vue';
import type { ElScrollbar } from 'element-plus';
import { debounce } from '@/assets/debounce';
import videoPlay from '@/components/play/videoPlay.vue';
import { getPlayVideoURL } from '@/common/play';
import contentRightClickMenu from './contentRightClickMenu.vue';
import { appStoreData } from '@/storeData/app.storeData';

const store = {
  appStoreData: appStoreData(),
}

const localStorage_WaterfallColumn_key = 'layout-short-video-waterfall-column-' + store.appStoreData.currentFilesBases.id;

const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
})
const emits = defineEmits(['selectResources']);

const scrollbarRef = ref<InstanceType<typeof ElScrollbar>>();
const waterfallRef = ref<InstanceType<typeof Waterfall>>();

// 从本地存储获取保存的值，如果没有则使用默认值3
const waterfallColumn = ref(parseInt(localStorage.getItem(localStorage_WaterfallColumn_key) || '3', 10));


const shortVideoPlayContainerRef = ref<HTMLDivElement>();
const videoPlayRef = ref<InstanceType<typeof videoPlay>>();

const currentPlayIndex = ref(-1);
const dataListWrapper = computed(() => props.dataList);

watch(dataListWrapper, () => {
  init()
}, { deep: true })


watch(currentPlayIndex, (newVal) => {
  if (newVal < 0) return;
  scrollToCurrentItem();
  setVideoSource(currentPlayDramaSeriesId_C.value);
})

// 监听waterfallColumn变化，保存到本地存储
watch(waterfallColumn, (newVal) => {
  localStorage.setItem(localStorage_WaterfallColumn_key, newVal.toString());
})

const dataList_C = computed(() => {
  return props.dataList.map((item: I_resource) => {
    const src = getResourceCoverPoster(item);
    return {
      ...item,
      src,
    }
  })
});

const currentPlayDramaSeriesId_C = computed(() => {
  return getResourceDramaSeriesId(currentPlayIndex.value);
});

// 计算动态 breakpoints
const waterfallBreakpoints_C = computed(() => {
  if (isMobile()) {
    return {
      1200: { rowPerView: 3 },
      800: { rowPerView: 2 },
      500: { rowPerView: 1 }
    }
  } else {
    // 可以根据当前列数设置不同的断点
    return {
      9999: { rowPerView: waterfallColumn.value },
    }
  }
});

const init = () => {
  currentPlayIndex.value = 0;
  setVideoSource(currentPlayDramaSeriesId_C.value);
}

const selectResourcesHandle = (item: I_resource) => {
  emits('selectResources', item)
}

const clickResourceHandle = (index: number) => {
  currentPlayIndex.value = index;
}

// 图片加载完成事件处理
const onImageLoad = debounce(() => {
  nextTick(() => {
    waterfallRef.value?.renderer();
  });
}, 100);

const setVideoSource = (dramaSeriesId: string) => {
  const vp = videoPlayRef.value;
  if (!vp || dramaSeriesId == '') return;
  const isPlaying = vp.isPlaying() || false;
  vp.setVideoSource(getPlayVideoURL(dramaSeriesId, 'mp4'), 'mp4', () => {
    vp.addTextTrack(
      `/api/video/subtitle/${dramaSeriesId}`,
      '默认字幕',
      'zh',
      true // 设为默认字幕
    )
    if (isPlaying) {
      vp.play();
    }
  });
}



const setVideoPlaySize = () => {
  const width = shortVideoPlayContainerRef.value?.clientWidth;
  const height = shortVideoPlayContainerRef.value?.clientHeight;
  const controllerHeight = videoPlayRef.value?.getControllerHeight() || 0;
  if (width && height) {
    videoPlayRef.value?.setAspectRatio(width + ':' + (height - controllerHeight));
  }
}

const getResourceDramaSeriesId = (index: number) => {
  if (props.dataList[index] && props.dataList[index].dramaSeries.length > 0) {
    return props.dataList[index].dramaSeries[0].id;
  }
  return '';
}

const prevHandle = () => {
  if (currentPlayIndex.value <= 0) return;
  currentPlayIndex.value = currentPlayIndex.value - 1;
}
const nextHandle = () => {
  if (currentPlayIndex.value >= (dataListWrapper.value.length - 1)) return;
  currentPlayIndex.value = currentPlayIndex.value + 1;
}

// 滚动到当前选中的项目
const scrollToCurrentItem = () => {
  if (currentPlayIndex.value < 0 || !waterfallRef.value || !scrollbarRef.value) return;

  nextTick(() => {
    const container = scrollbarRef.value?.wrapRef;
    if (!container) return;

    // 通过 Waterfall 组件的 $el 获取所有 item 元素
    //const items = waterfallRef.value?.$el.querySelectorAll('.waterfall-item');

    const items = waterfallRef.value?.$el.getElementsByClassName('waterfall-item-' + currentPlayIndex.value);
    if (!items || items.length === 0) return;
    // 获取第一个匹配的元素（应该只有一个）
    const targetElement = items[0];
    // 获取容器和目标元素的位置信息
    const containerRect = container.getBoundingClientRect();
    const targetRect = targetElement.getBoundingClientRect();
    // 计算目标元素相对于容器的位置
    const containerScrollTop = container.scrollTop;
    const relativeTop = targetRect.top - containerRect.top;

    // 计算需要滚动到的位置，使目标元素居中
    const scrollToPosition = containerScrollTop + relativeTop - (containerRect.height / 2) + (targetRect.height / 2);

    // 平滑滚动到目标位置
    container.scrollTo({
      top: Math.max(0, scrollToPosition),
      behavior: 'smooth'
    });
  });
}

// 键盘事件处理函数
const handleKeyDown = (event: KeyboardEvent) => {
  // 上/左箭头键触发上一个
  if (event.key === 'ArrowUp' || event.key === 'ArrowLeft') {
    event.preventDefault();
    prevHandle();
  }
  // 下/右箭头键触发下一个
  else if (event.key === 'ArrowDown' || event.key === 'ArrowRight') {
    event.preventDefault();
    nextHandle();
  }
}

onMounted(() => {
  nextTick(() => {
    setVideoPlaySize();
  });
  // 添加键盘事件监听
  window.addEventListener('keydown', handleKeyDown);
})
onBeforeUnmount(() => {
  // 移除键盘事件监听
  window.removeEventListener('keydown', handleKeyDown);
})

const change = () => {
  init();
  nextTick(() => {
    scrollbarRef.value?.setScrollTop(0);
  });

};


defineExpose({ change, });
</script>
<style lang="scss" scoped>
.layout-short-video {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;

  .shortVideoList {
    flex: 1;
    display: flex;
    flex-direction: column;

    .shortVideoListWaterfall {
      flex: 1;
      overflow: hidden;
    }

    .tool {
      flex-shrink: 0;
      display: flex;
      justify-content: center;
    }

    .waterfall-list {
      background-color: unset;
    }

    .waterfall-item {
      position: relative;
      cursor: pointer;
      border: 2px solid transparent;
      border-radius: 4px;
      transition: border-color 0.3s ease;

      &.active {
        border-color: #409eff;
        box-shadow: 0 0 8px rgba(64, 158, 255, 0.5);
      }

      &:hover {
        .play-icon {
          display: block;
        }
      }

      .el-image:hover {
        transform: scale(1.01);
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
        cursor: pointer;
      }
    }
  }

  .shortVideoPlay {
    width: 50%;
    padding: 10px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;

    &.mobile-layout {
      width: 70%;
    }

    .shortVideoPlayContainer {
      flex: 1;
    }

    .tool {
      flex-shrink: 0;
      padding-top: 10px;
      display: flex;
      justify-content: center;
    }
  }
}
</style>
