<template>
  <div class="play-comic-container">
    <HeaderView class="header" :mode="E_headerMode.GoBack" :title="resourceInfo?.title || ''"></HeaderView>
    <div class="main-container" v-loading="loading">
      <div class="main" v-if="resourceInfo">
        <div class="main-left">
          <div class="thumbnail" ref="thumbnailRef" @keydown="(event: KeyboardEvent) => { event.preventDefault(); }">
            <ul>
              <li v-for="filesName, key in comicImageList" :key="key" :class="{ active: key === nowPage }"
                ref="thumbnailItemsRef" @click="setPage(key)">
                <div>
                  <el-image :src="getFileImageByDramaSeriesId(selectedDramaSeriesId, filesName)" fit="scale-down" />
                </div>
                <div class="page">({{ key + 1 }})</div>
              </li>
            </ul>
          </div>
          <div class="read">
            <div class="read-image" ref="readImageRef">
              <el-image class="full-show"
                :src="getFileImageByDramaSeriesId(selectedDramaSeriesId, comicImageList[nowPage])" fit="cover"
                :style="{ width: showImageWidth + '%', marginLeft: ((100 - showImageWidth) / 2) + '%' }" />
            </div>
            <div class="read-tool-btn">
              <el-button plain size="small" @click="changeNowPage('per')" :disabled="nowPage <= 0">
                上一页
              </el-button>
              <label class="nowPageLabel">第 {{ nowPage + 1 }} 页</label>
              <el-button plain size="small" @click="changeNowPage('next')"
                :disabled="nowPage >= comicImageList.length - 1">
                下一页
              </el-button>
              <el-slider v-model="showImageWidth" :min="20" :max="100" style="width: 200px;" />
            </div>
          </div>
        </div>
        <div class="main-right">
          <el-image :src="getResourceCoverPoster(resourceInfo)" fit="cover" />
          <div class="title">{{ resourceInfo.title }}</div>
          <resourceDramaSeriesList :drama-series="resourceInfo.dramaSeries" :selected-id="selectedDramaSeriesId"
            :show-mode="store.appStoreData.currentFilesBasesAppConfig.detailsDramaSeriesMode"
            @play-resource-drama-series="playResourceDramaSeriesHandle">
          </resourceDramaSeriesList>
          <div class="c-height"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import HeaderView from '../HeaderView.vue'
import { E_headerMode } from '@/dataType/app.dataType'
import type { I_resource, I_resourceDramaSeries } from '@/dataType/resource.dataType';
import { resourceServer } from '@/server/resource.server';
import { ElMessage } from 'element-plus';
import { getResourceCoverPoster } from '@/common/photo';
import resourceDramaSeriesList from '@/components/resource/resourceDramaSeriesList.vue'
import { appStoreData } from '@/storeData/app.storeData';
import { filesServer } from '@/server/files.server';
import { getFileImageByDramaSeriesId } from '@/common/photo'

enum EsetContentScrollbarMode {
  init,
  add,
}

const store = {
  appStoreData: appStoreData(),
}

const localStorage_showImageWidth_key = 'play-comic-show-image-width-' + store.appStoreData.currentFilesBases.id;

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

const readImageRef = ref<HTMLDivElement>();
const thumbnailRef = ref<HTMLDivElement>();
const thumbnailItemsRef = ref<HTMLDivElement[]>([]);

const resourceInfo = ref<I_resource>();
const comicImageList = ref<string[]>([]);
const selectedDramaSeriesId = ref<string>('');
const nowPage = ref(0);
const showImageWidth = ref(parseInt(localStorage.getItem(localStorage_showImageWidth_key) || '50', 10));
const loading = ref(false);

// 监听waterfallColumn变化，保存到本地存储
watch(showImageWidth, (newVal) => {
  localStorage.setItem(localStorage_showImageWidth_key, newVal.toString());
})


const init = async () => {
  nowPage.value = 0;
  await getResourceInfo();
  setDramaSeries();
  await getResourceComic();
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

const getResourceComic = async () => {
  loading.value = true;
  const result = await filesServer.filesDListByDramaSeriesId_Image(selectedDramaSeriesId.value);
  if (!result || !result.status) {
    ElMessage.error(result.msg);
    return;
  }
  comicImageList.value = result.data;
  loading.value = false;
}

const playResourceDramaSeriesHandle = async (ds: I_resourceDramaSeries) => {
  selectedDramaSeriesId.value = ds.id;
  await getResourceComic();
  nowPage.value = 0;
}

const changeNowPage = (mode: 'per' | 'next') => {
  if (mode == 'per' && nowPage.value > 0) {
    nowPage.value = nowPage.value - 1;
  } else if (mode == 'next' && nowPage.value < comicImageList.value.length - 1) {
    nowPage.value = nowPage.value + 1;
  }
  nextTick(() => {
    try {
      setContentScrollbar();
      scrollToThumbnail();
    } catch (e) {
      console.log(e);
    }
  })
}
const setPage = (page: number) => {
  nowPage.value = page;
  nextTick(() => {
    try {
      setContentScrollbar();
      scrollToThumbnail();
    } catch (e) {
      console.log(e);
    }
  })
}

const scrollToThumbnail = () => {
  nextTick(() => {
    if (thumbnailItemsRef.value && thumbnailItemsRef.value[nowPage.value] && thumbnailRef.value) {
      const targetElement = thumbnailItemsRef.value[nowPage.value];
      const container = thumbnailRef.value;

      // 计算目标元素在容器中的位置
      const elementTop = targetElement.offsetTop;
      const containerHeight = container.clientHeight;
      const elementHeight = targetElement.clientHeight;

      // 滚动到目标元素在容器中间位置
      const scrollToPosition = elementTop - (containerHeight / 2) + (elementHeight / 2);
      container.scrollTo({
        top: scrollToPosition,
        behavior: 'smooth'
      });
    }
  });
}

const setContentScrollbar = (num = 0, mode: EsetContentScrollbarMode = EsetContentScrollbarMode.init) => {
  if (!readImageRef.value) return;
  if (EsetContentScrollbarMode.init == mode) {
    readImageRef.value.scrollTo(0, 0);
  } else {
    const offsetTop = readImageRef.value.scrollTop || 0;
    const top = offsetTop + num;
    readImageRef.value?.scrollTo(0, top);
  }
}

const addEventListeners = () => {
  document.addEventListener('keydown', handleKeyDown);
}

const removeEventListeners = () => {
  document.removeEventListener('keydown', handleKeyDown);
}

const handleKeyDown = (event: KeyboardEvent) => {
  if (event.key === 'ArrowUp') {
    setContentScrollbar(-200, EsetContentScrollbarMode.add)
  } else if (event.key === 'ArrowDown') {
    setContentScrollbar(200, EsetContentScrollbarMode.add)
  } else if (event.key === 'ArrowLeft') {
    changeNowPage('per')
  } else if (event.key === 'ArrowRight') {
    changeNowPage('next')
  }
}

// 监听当前页变化，自动滚动缩略图
watch(nowPage, () => {
  scrollToThumbnail();
});

onMounted(async () => {
  nextTick(async () => {
    await init();
    addEventListeners();
  })

});

onUnmounted(() => {
  removeEventListeners();
});
</script>

<style lang="scss" scoped>
.play-comic-container {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;

  .main-container {
    flex: 1;
    overflow: hidden;

    .main {
      width: 100%;
      height: 100%;
      display: flex;
      justify-content: space-between;
      gap: 10px;

      .main-left {
        flex: 1;
        display: flex;
        gap: 10px;

        .thumbnail {
          width: 160px;
          height: 100%;
          flex-shrink: 0;
          overflow-y: auto;

          ul {
            list-style: none;
            height: 100%;
            padding: 0 2px;

            li {
              padding: 2px;
              border-radius: 4px;
              transition: all 0.3s;
              margin-bottom: 5px;
              cursor: pointer;

              &.active {
                border: 2px solid #409eff;
                border-radius: 4px;
                background-color: rgba(64, 158, 255, 0.1);
              }

              .page {
                text-align: center;
                line-height: 18px;
                padding-bottom: 5px;
              }
            }
          }
        }

        .read {
          flex: 1;
          display: flex;
          flex-direction: column;

          .read-image {
            flex: 1;
            overflow-y: auto;
          }

          .read-tool-btn {
            padding: 10px 0;
            flex-shrink: 0;
            display: flex;
            justify-content: center;
            gap: 10px;
            line-height: 24px;
          }
        }
      }

      .main-right {
        width: 280px;
        flex-shrink: 0;
        display: flex;
        flex-direction: column;
        gap: 10px;

        .title {
          font-size: 14px;
        }
      }
    }
  }
}
</style>
