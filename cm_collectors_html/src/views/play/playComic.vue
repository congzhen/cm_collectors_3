<template>
  <div
    class="play-comic-container"
    :class="{
      'play-comic-container--modern': isModernAppearance,
      'play-comic-container--bright': isModernAppearance && isBrightTheme,
    }"
  >
    <HeaderView
      class="header"
      :mode="E_headerMode.GoBack"
      :title="resourceInfo?.title || ''"
      solid
    ></HeaderView>
    <div v-if="isModernAppearance && resourceInfo" class="modern-header-context" aria-hidden="true">
      <strong>{{ resourceInfo.title }}</strong>
      <span v-if="resourceInfo.dramaSeries.length > 0"
        >第 {{ selectedDramaSeriesIndex + 1 }} 话</span
      >
    </div>
    <div class="main-container" v-loading="loading">
      <div v-if="resourceInfo && isModernAppearance" class="modern-reader">
        <aside class="modern-page-panel">
          <div class="modern-panel-header">
            <strong>页面</strong>
            <span
              >{{ comicImageList.length > 0 ? nowPage + 1 : 0 }} / {{ comicImageList.length }}</span
            >
          </div>
          <div
            class="modern-thumbnail-scroll"
            ref="thumbnailRef"
            @keydown="
              (event: KeyboardEvent) => {
                event.preventDefault()
              }
            "
          >
            <ul class="modern-thumbnail-list">
              <li
                v-for="(filesName, key) in comicImageList"
                :key="key"
                :class="{ active: key === nowPage }"
                ref="thumbnailItemsRef"
                @click="setPage(key)"
              >
                <div class="modern-thumbnail-image">
                  <el-image
                    :src="getFileImageByDramaSeriesId(selectedDramaSeriesId, filesName)"
                    fit="contain"
                  />
                  <span v-if="key === nowPage" class="modern-current-badge">当前</span>
                </div>
                <div class="page">{{ key + 1 }}</div>
              </li>
            </ul>
          </div>
        </aside>

        <main class="modern-read-stage">
          <div class="modern-read-image" ref="readImageRef">
            <el-image
              v-if="comicImageList.length > 0"
              class="modern-full-show"
              :src="getFileImageByDramaSeriesId(selectedDramaSeriesId, comicImageList[nowPage])"
              fit="contain"
              :style="{ width: showImageWidth + '%', marginLeft: (100 - showImageWidth) / 2 + '%' }"
            />
          </div>
          <div class="modern-read-toolbar">
            <el-button plain size="small" @click="changeNowPage('per')" :disabled="nowPage <= 0">
              上一页
            </el-button>
            <span class="modern-page-label">第 {{ nowPage + 1 }} 页</span>
            <el-button
              plain
              size="small"
              @click="changeNowPage('next')"
              :disabled="nowPage >= comicImageList.length - 1"
            >
              下一页
            </el-button>
            <span class="modern-toolbar-divider"></span>
            <span class="modern-width-icon" title="阅读宽度" aria-label="阅读宽度"></span>
            <el-slider v-model="showImageWidth" :min="20" :max="100" />
            <span class="modern-width-value">{{ showImageWidth }}%</span>
          </div>
        </main>

        <aside class="modern-series-panel">
          <div class="modern-series-summary">
            <el-image :src="getResourceCoverPoster(resourceInfo)" fit="cover" />
            <div class="title">{{ resourceInfo.title }}</div>
            <div class="modern-series-count">共 {{ resourceInfo.dramaSeries.length }} 话</div>
          </div>
          <div class="modern-series-scroll">
            <resourceDramaSeriesList
              :drama-series="resourceInfo.dramaSeries"
              :selected-id="selectedDramaSeriesId"
              :show-mode="store.appStoreData.currentFilesBasesAppConfig.detailsDramaSeriesMode"
              modern
              @play-resource-drama-series="playResourceDramaSeriesHandle"
            >
            </resourceDramaSeriesList>
          </div>
        </aside>
      </div>

      <div class="main" v-else-if="resourceInfo">
        <div class="main-left">
          <div
            class="thumbnail"
            ref="thumbnailRef"
            @keydown="
              (event: KeyboardEvent) => {
                event.preventDefault()
              }
            "
          >
            <ul>
              <li
                v-for="(filesName, key) in comicImageList"
                :key="key"
                :class="{ active: key === nowPage }"
                ref="thumbnailItemsRef"
                @click="setPage(key)"
              >
                <div>
                  <el-image
                    :src="getFileImageByDramaSeriesId(selectedDramaSeriesId, filesName)"
                    fit="scale-down"
                  />
                </div>
                <div class="page">({{ key + 1 }})</div>
              </li>
            </ul>
          </div>
          <div class="read">
            <div class="read-image" ref="readImageRef">
              <el-image
                class="full-show"
                :src="getFileImageByDramaSeriesId(selectedDramaSeriesId, comicImageList[nowPage])"
                fit="cover"
                :style="{
                  width: showImageWidth + '%',
                  marginLeft: (100 - showImageWidth) / 2 + '%',
                }"
              />
            </div>
            <div class="read-tool-btn">
              <el-button plain size="small" @click="changeNowPage('per')" :disabled="nowPage <= 0">
                上一页
              </el-button>
              <label class="nowPageLabel">第 {{ nowPage + 1 }} 页</label>
              <el-button
                plain
                size="small"
                @click="changeNowPage('next')"
                :disabled="nowPage >= comicImageList.length - 1"
              >
                下一页
              </el-button>
              <el-slider v-model="showImageWidth" :min="20" :max="100" style="width: 200px" />
            </div>
          </div>
        </div>
        <div class="main-right">
          <el-image :src="getResourceCoverPoster(resourceInfo)" fit="cover" />
          <div class="title">{{ resourceInfo.title }}</div>
          <resourceDramaSeriesList
            :drama-series="resourceInfo.dramaSeries"
            :selected-id="selectedDramaSeriesId"
            :show-mode="store.appStoreData.currentFilesBasesAppConfig.detailsDramaSeriesMode"
            @play-resource-drama-series="playResourceDramaSeriesHandle"
          >
          </resourceDramaSeriesList>
          <div class="c-height"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import HeaderView from '../HeaderView.vue'
import { E_headerMode } from '@/dataType/app.dataType'
import type { I_resource, I_resourceDramaSeries } from '@/dataType/resource.dataType'
import { resourceServer } from '@/server/resource.server'
import { ElMessage } from 'element-plus'
import { getResourceCoverPoster } from '@/common/photo'
import resourceDramaSeriesList from '@/components/resource/resourceDramaSeriesList.vue'
import { appStoreData } from '@/storeData/app.storeData'
import { filesServer } from '@/server/files.server'
import { getFileImageByDramaSeriesId } from '@/common/photo'

enum EsetContentScrollbarMode {
  init,
  add,
}

const store = {
  appStoreData: appStoreData(),
}

const isModernAppearance = computed(
  () =>
    (store.appStoreData.appConfig.appearanceStyle ||
      store.appStoreData.appConfig.headerStyle ||
      'modern') === 'modern',
)
const isBrightTheme = computed(() => store.appStoreData.appConfig.theme === 'bright')

const localStorage_showImageWidth_key =
  'play-comic-show-image-width-' + store.appStoreData.currentFilesBases.id

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

const readImageRef = ref<HTMLDivElement>()
const thumbnailRef = ref<HTMLDivElement>()
const thumbnailItemsRef = ref<HTMLDivElement[]>([])

const resourceInfo = ref<I_resource>()
const comicImageList = ref<string[]>([])
const selectedDramaSeriesId = ref<string>('')
const nowPage = ref(0)
const showImageWidth = ref(
  parseInt(localStorage.getItem(localStorage_showImageWidth_key) || '50', 10),
)
const loading = ref(false)
const selectedDramaSeriesIndex = computed(() => {
  const index =
    resourceInfo.value?.dramaSeries.findIndex((item) => item.id === selectedDramaSeriesId.value) ??
    -1
  return index < 0 ? 0 : index
})

// 监听waterfallColumn变化，保存到本地存储
watch(showImageWidth, (newVal) => {
  localStorage.setItem(localStorage_showImageWidth_key, newVal.toString())
})

const init = async () => {
  nowPage.value = 0
  await getResourceInfo()
  setDramaSeries()
  await getResourceComic()
}

const setDramaSeries = () => {
  if (props.dramaSeriesId != '') {
    selectedDramaSeriesId.value = props.dramaSeriesId
  } else if (resourceInfo.value && resourceInfo.value.dramaSeries.length > 0) {
    selectedDramaSeriesId.value = resourceInfo.value.dramaSeries[0].id
  }
}

const getResourceInfo = async () => {
  loading.value = true
  const result = await resourceServer.info(props.resourceId)
  if (!result || !result.status) {
    ElMessage.error(result.msg)
    return
  }
  resourceInfo.value = result.data
  loading.value = false
}

const getResourceComic = async () => {
  loading.value = true
  const result = await filesServer.filesDListByDramaSeriesId_Image(selectedDramaSeriesId.value)
  if (!result || !result.status) {
    ElMessage.error(result.msg)
    return
  }
  comicImageList.value = result.data
  loading.value = false
}

const playResourceDramaSeriesHandle = async (ds: I_resourceDramaSeries) => {
  selectedDramaSeriesId.value = ds.id
  await getResourceComic()
  nowPage.value = 0
}

const changeNowPage = (mode: 'per' | 'next') => {
  if (mode == 'per' && nowPage.value > 0) {
    nowPage.value = nowPage.value - 1
  } else if (mode == 'next' && nowPage.value < comicImageList.value.length - 1) {
    nowPage.value = nowPage.value + 1
  }
  nextTick(() => {
    try {
      setContentScrollbar()
      scrollToThumbnail()
    } catch (e) {
      console.log(e)
    }
  })
}
const setPage = (page: number) => {
  nowPage.value = page
  nextTick(() => {
    try {
      setContentScrollbar()
      scrollToThumbnail()
    } catch (e) {
      console.log(e)
    }
  })
}

const scrollToThumbnail = () => {
  nextTick(() => {
    if (thumbnailItemsRef.value && thumbnailItemsRef.value[nowPage.value] && thumbnailRef.value) {
      const targetElement = thumbnailItemsRef.value[nowPage.value]
      const container = thumbnailRef.value

      // 计算目标元素在容器中的位置
      const elementTop = targetElement.offsetTop
      const containerHeight = container.clientHeight
      const elementHeight = targetElement.clientHeight

      // 滚动到目标元素在容器中间位置
      const scrollToPosition = elementTop - containerHeight / 2 + elementHeight / 2
      container.scrollTo({
        top: scrollToPosition,
        behavior: 'smooth',
      })
    }
  })
}

const setContentScrollbar = (
  num = 0,
  mode: EsetContentScrollbarMode = EsetContentScrollbarMode.init,
) => {
  if (!readImageRef.value) return
  if (EsetContentScrollbarMode.init == mode) {
    readImageRef.value.scrollTo(0, 0)
  } else {
    const offsetTop = readImageRef.value.scrollTop || 0
    const top = offsetTop + num
    readImageRef.value?.scrollTo(0, top)
  }
}

const addEventListeners = () => {
  document.addEventListener('keydown', handleKeyDown)
}

const removeEventListeners = () => {
  document.removeEventListener('keydown', handleKeyDown)
}

const isInputTarget = (target: EventTarget | null) => {
  const el = target as HTMLElement | null
  if (!el) return false
  const tagName = el.tagName?.toLowerCase()
  return (
    tagName === 'input' || tagName === 'textarea' || tagName === 'select' || el.isContentEditable
  )
}

const handleKeyDown = (event: KeyboardEvent) => {
  if (isInputTarget(event.target)) return
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
  scrollToThumbnail()
})

onMounted(async () => {
  nextTick(async () => {
    await init()
    addEventListeners()
  })
})

onUnmounted(() => {
  removeEventListeners()
})
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

        .el-image {
          width: 100%;
        }

        .title {
          font-size: 14px;
        }
      }
    }
  }
}

.play-comic-container--modern {
  --comic-bg: #17191d;
  --comic-panel-bg: #1d2025;
  --comic-stage-bg: #111317;
  --comic-soft-bg: #25292f;
  --comic-border: rgba(255, 255, 255, 0.1);
  --comic-border-strong: rgba(255, 255, 255, 0.16);
  --comic-text: #e7eaee;
  --comic-muted: #9aa3ad;
  --comic-accent: #25b8b5;
  --comic-accent-soft: rgba(37, 184, 181, 0.12);
  --comic-shadow: 0 12px 34px rgba(0, 0, 0, 0.24);

  position: relative;
  color: var(--comic-text);
  background: var(--comic-bg);

  &.play-comic-container--bright {
    --comic-bg: #f4f6f8;
    --comic-panel-bg: #ffffff;
    --comic-stage-bg: #e9edf1;
    --comic-soft-bg: #f4f6f8;
    --comic-border: #dfe4e9;
    --comic-border-strong: #cfd6dd;
    --comic-text: #2d3741;
    --comic-muted: #6d7883;
    --comic-accent: #159fa1;
    --comic-accent-soft: rgba(21, 159, 161, 0.1);
    --comic-shadow: 0 12px 30px rgba(33, 45, 57, 0.12);
  }

  > .header {
    position: relative;
    z-index: 10;
    flex-shrink: 0;
  }

  .modern-header-context {
    position: absolute;
    top: 0;
    left: 50%;
    z-index: 11;
    min-width: 0;
    height: 58px;
    display: flex;
    align-items: center;
    gap: 14px;
    color: var(--comic-text);
    pointer-events: none;
    transform: translateX(-50%);

    strong {
      max-width: 360px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      font-size: 16px;
    }

    span {
      color: var(--comic-muted);
      font-size: 13px;
    }
  }

  > .main-container {
    min-height: 0;
    padding: 10px;
    box-sizing: border-box;
    background: var(--comic-bg);
  }
}

.modern-reader {
  width: 100%;
  height: 100%;
  min-width: 0;
  min-height: 0;
  display: grid;
  grid-template-columns: 168px minmax(360px, 1fr) minmax(250px, 300px);
  gap: 10px;
}

.modern-page-panel,
.modern-series-panel,
.modern-read-stage {
  min-width: 0;
  min-height: 0;
  border: 1px solid var(--comic-border);
  border-radius: 10px;
  background: var(--comic-panel-bg);
}

.modern-page-panel {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modern-panel-header {
  height: 48px;
  padding: 0 14px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--comic-border);

  strong {
    font-size: 14px;
  }

  span {
    color: var(--comic-muted);
    font-size: 12px;
    font-variant-numeric: tabular-nums;
  }
}

.modern-thumbnail-scroll {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--comic-border-strong) transparent;
}

.modern-thumbnail-list {
  margin: 0;
  padding: 10px;
  list-style: none;

  li {
    position: relative;
    margin: 0 0 9px;
    padding: 5px;
    border: 1px solid transparent;
    border-radius: 8px;
    color: var(--comic-muted);
    background: var(--comic-soft-bg);
    cursor: pointer;
    transition:
      border-color 0.16s ease,
      background-color 0.16s ease,
      color 0.16s ease;

    &:hover {
      color: var(--comic-text);
      border-color: var(--comic-border-strong);
    }

    &.active {
      color: var(--comic-accent);
      border-color: var(--comic-accent);
      background: var(--comic-accent-soft);
      box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--comic-accent) 24%, transparent);
    }

    .page {
      padding-top: 5px;
      text-align: center;
      font-size: 12px;
      font-weight: 600;
      line-height: 18px;
    }
  }
}

.modern-thumbnail-image {
  position: relative;
  height: 150px;
  overflow: hidden;
  border-radius: 5px;
  background: var(--comic-stage-bg);

  .el-image {
    width: 100%;
    height: 100%;
  }
}

.modern-current-badge {
  position: absolute;
  top: 6px;
  left: 6px;
  padding: 2px 6px;
  border-radius: 5px;
  color: #ffffff;
  background: var(--comic-accent);
  font-size: 11px;
  line-height: 18px;
}

.modern-read-stage {
  position: relative;
  overflow: hidden;
  background: var(--comic-stage-bg);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.02);
}

.modern-read-image {
  width: 100%;
  height: 100%;
  padding: 14px 14px 86px;
  box-sizing: border-box;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--comic-border-strong) transparent;
}

.modern-full-show {
  display: block;
  overflow: visible;
  border-radius: 4px;
  background: #090a0c;
  box-shadow: var(--comic-shadow);

  :deep(.el-image__inner) {
    position: static;
    display: block;
    width: 100%;
    height: auto;
  }
}

.modern-read-toolbar {
  position: absolute;
  bottom: 16px;
  left: 50%;
  z-index: 2;
  min-width: 480px;
  max-width: calc(100% - 32px);
  height: 52px;
  padding: 0 12px;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border: 1px solid var(--comic-border-strong);
  border-radius: 11px;
  background: color-mix(in srgb, var(--comic-panel-bg) 92%, transparent);
  box-shadow: var(--comic-shadow);
  backdrop-filter: blur(14px);
  transform: translateX(-50%);

  .el-button {
    height: 32px;
    margin: 0;
    color: var(--comic-text);
    background: var(--comic-soft-bg);
    border-color: var(--comic-border);

    &:not(.is-disabled):hover {
      color: var(--comic-accent);
      border-color: var(--comic-accent);
      background: var(--comic-accent-soft);
    }
  }

  .el-slider {
    width: 180px;
    flex-shrink: 1;
    --el-slider-main-bg-color: var(--comic-accent);
    --el-slider-runway-bg-color: var(--comic-border-strong);
  }
}

.modern-page-label,
.modern-width-value {
  flex-shrink: 0;
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}

.modern-page-label {
  min-width: 54px;
  color: var(--comic-text);
  text-align: center;
  font-weight: 600;
}

.modern-width-value {
  min-width: 36px;
  color: var(--comic-muted);
  text-align: right;
}

.modern-toolbar-divider {
  width: 1px;
  height: 24px;
  flex-shrink: 0;
  background: var(--comic-border-strong);
}

.modern-width-icon {
  position: relative;
  width: 18px;
  height: 14px;
  flex-shrink: 0;
  box-sizing: border-box;
  border: 1px solid var(--comic-muted);
  border-radius: 2px;

  &::before,
  &::after {
    content: '';
    position: absolute;
    top: 3px;
    bottom: 3px;
    width: 1px;
    background: var(--comic-muted);
  }

  &::before {
    left: 4px;
  }

  &::after {
    right: 4px;
  }
}

.modern-series-panel {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modern-series-summary {
  padding: 12px;
  flex-shrink: 0;
  border-bottom: 1px solid var(--comic-border);

  > .el-image {
    width: 100%;
    height: 150px;
    overflow: hidden;
    border-radius: 7px;
    background: var(--comic-stage-bg);
  }

  .title {
    margin-top: 10px;
    overflow: hidden;
    color: var(--comic-text);
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 15px;
    font-weight: 700;
  }
}

.modern-series-count {
  margin-top: 3px;
  color: var(--comic-muted);
  font-size: 12px;
}

.modern-series-scroll {
  min-height: 0;
  padding: 10px;
  flex: 1;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--comic-border-strong) transparent;

  :deep(.resourceDramaSeries-list--modern) {
    --modern-details-border: var(--comic-border);
    --modern-details-soft-bg: var(--comic-soft-bg);
    --modern-details-text: var(--comic-text);
    --modern-details-text-muted: var(--comic-muted);
    --series-active-bg: var(--comic-accent-soft);
    --series-active-border: var(--comic-accent);
    --series-active-text: var(--comic-accent);
    --series-surface-hover: var(--comic-accent-soft);
  }
}

.modern-thumbnail-scroll::-webkit-scrollbar,
.modern-read-image::-webkit-scrollbar,
.modern-series-scroll::-webkit-scrollbar {
  width: 7px;
}

.modern-thumbnail-scroll::-webkit-scrollbar-thumb,
.modern-read-image::-webkit-scrollbar-thumb,
.modern-series-scroll::-webkit-scrollbar-thumb {
  border: 2px solid transparent;
  border-radius: 999px;
  background: var(--comic-border-strong);
  background-clip: padding-box;
}

@media (max-width: 1180px) {
  .modern-reader {
    grid-template-columns: 142px minmax(320px, 1fr) 250px;
  }

  .modern-header-context {
    display: none !important;
  }

  .modern-read-toolbar {
    min-width: 430px;

    .el-slider {
      width: 130px;
    }
  }
}
</style>
