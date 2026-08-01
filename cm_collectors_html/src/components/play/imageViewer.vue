<template>
  <el-image-viewer
    v-if="showImageViewer && !isModernAppearance"
    :url-list="imageList_C"
    :initial-index="initialIndex"
    @close="closeImageViewer"
    @switch="handleSwitch"
  />

  <Teleport to="body">
    <Transition name="modern-image-viewer-fade">
      <div
        v-if="showImageViewer && isModernAppearance"
        class="modern-image-viewer"
        :class="{ 'modern-image-viewer--thumbs': showThumbnails && props.imageList.length > 1 }"
        role="dialog"
        aria-modal="true"
        aria-label="图片查看器"
      >
        <header class="modern-image-viewer__header">
          <div class="modern-image-viewer__title">
            <el-icon><Picture /></el-icon>
            <strong>图片查看器</strong>
          </div>
          <span class="modern-image-viewer__progress">{{ progressText }}</span>
          <button
            class="modern-image-viewer__close"
            type="button"
            title="关闭"
            aria-label="关闭"
            @click="closeImageViewer"
          >
            <el-icon><Close /></el-icon>
          </button>
        </header>

        <button
          v-if="props.imageList.length > 1"
          class="modern-image-viewer__side-button modern-image-viewer__side-button--prev"
          type="button"
          title="上一张"
          aria-label="上一张"
          @click="changeImage(-1)"
        >
          <el-icon><ArrowLeft /></el-icon>
        </button>
        <button
          v-if="props.imageList.length > 1"
          class="modern-image-viewer__side-button modern-image-viewer__side-button--next"
          type="button"
          title="下一张"
          aria-label="下一张"
          @click="changeImage(1)"
        >
          <el-icon><ArrowRight /></el-icon>
        </button>

        <div class="modern-image-viewer__canvas" @wheel.prevent="handleWheel">
          <img
            v-if="currentImage"
            class="modern-image-viewer__image"
            :class="`is-${imageMode}`"
            :src="currentImage"
            :alt="`图片 ${currentIndex + 1}`"
            :style="imageStyle"
            draggable="false"
            @mousedown.prevent="startDrag"
          />
        </div>

        <div
          v-if="showThumbnails && props.imageList.length > 1"
          class="modern-image-viewer__thumbnail-panel"
        >
          <button
            class="modern-image-viewer__thumbnail-scroll"
            type="button"
            title="向前滚动"
            aria-label="向前滚动缩略图"
            @click="scrollThumbnails(-1)"
          >
            <el-icon><ArrowLeft /></el-icon>
          </button>
          <div ref="thumbnailTrackRef" class="modern-image-viewer__thumbnail-track">
            <button
              v-for="(image, index) in props.imageList"
              :key="`${image}-${index}`"
              ref="thumbnailItemsRef"
              class="modern-image-viewer__thumbnail"
              :class="{ active: index === currentIndex }"
              type="button"
              :title="`第 ${index + 1} 张`"
              :aria-label="`查看第 ${index + 1} 张图片`"
              :aria-current="index === currentIndex ? 'true' : undefined"
              @click="setCurrentIndex(index)"
            >
              <img
                :src="thumbnailSource(index)"
                :alt="`缩略图 ${index + 1}`"
                loading="lazy"
                decoding="async"
                draggable="false"
              />
              <span>{{ index + 1 }}</span>
            </button>
          </div>
          <button
            class="modern-image-viewer__thumbnail-scroll"
            type="button"
            title="向后滚动"
            aria-label="向后滚动缩略图"
            @click="scrollThumbnails(1)"
          >
            <el-icon><ArrowRight /></el-icon>
          </button>
        </div>

        <div class="modern-image-viewer__toolbar" role="toolbar" aria-label="图片操作">
          <button type="button" title="上一张" aria-label="上一张" @click="changeImage(-1)">
            <el-icon><ArrowLeft /></el-icon>
          </button>
          <span class="modern-image-viewer__toolbar-progress">{{ progressText }}</span>
          <button type="button" title="下一张" aria-label="下一张" @click="changeImage(1)">
            <el-icon><ArrowRight /></el-icon>
          </button>
          <i class="modern-image-viewer__divider"></i>
          <button type="button" title="缩小" aria-label="缩小" @click="zoom(-zoomStep)">
            <el-icon><ZoomOut /></el-icon>
          </button>
          <span class="modern-image-viewer__scale">{{ scaleText }}</span>
          <button type="button" title="放大" aria-label="放大" @click="zoom(zoomStep)">
            <el-icon><ZoomIn /></el-icon>
          </button>
          <i class="modern-image-viewer__divider"></i>
          <button
            type="button"
            :class="{ active: imageMode === 'contain' }"
            :title="imageMode === 'contain' ? '显示原始尺寸' : '适应窗口'"
            :aria-label="imageMode === 'contain' ? '显示原始尺寸' : '适应窗口'"
            @click="toggleImageMode"
          >
            <el-icon><ScaleToOriginal /></el-icon>
            <span class="modern-image-viewer__fit-text">适应</span>
          </button>
          <i v-if="props.imageList.length > 1" class="modern-image-viewer__divider"></i>
          <button
            v-if="props.imageList.length > 1"
            type="button"
            :class="{ active: showThumbnails }"
            :title="showThumbnails ? '收起缩略图' : '展开缩略图'"
            :aria-label="showThumbnails ? '收起缩略图' : '展开缩略图'"
            @click="toggleThumbnails"
          >
            <el-icon><Grid /></el-icon>
          </button>
          <i class="modern-image-viewer__divider"></i>
          <button type="button" title="向左旋转" aria-label="向左旋转" @click="rotate(-90)">
            <el-icon><RefreshLeft /></el-icon>
          </button>
          <button type="button" title="向右旋转" aria-label="向右旋转" @click="rotate(90)">
            <el-icon><RefreshRight /></el-icon>
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script lang="ts" setup>
import {
  computed,
  nextTick,
  onBeforeUnmount,
  reactive,
  ref,
  watch,
  type CSSProperties,
  type PropType,
} from 'vue'
import {
  ArrowLeft,
  ArrowRight,
  Close,
  Grid,
  Picture,
  RefreshLeft,
  RefreshRight,
  ScaleToOriginal,
  ZoomIn,
  ZoomOut,
} from '@element-plus/icons-vue'
import { appStoreData } from '@/storeData/app.storeData'

type T_imageMode = 'contain' | 'original'

const loadingPlaceholder =
  'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzAwIiBoZWlnaHQ9IjMwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSJ0cmFuc3BhcmVudCIvPjx0ZXh0IHg9IjUwJSIgeT0iNTAlIiBmb250LWZhbWlseT0iQXJpYWwsIEhlbHZldGljYSwgc2Fucy1zZXJpZiIgZm9udC1zaXplPSIxOCIgZmlsbD0id2hpdGUiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGRvbWluYW50LWJhc2VsaW5lPSJtaWRkbGUiPkxvYWRpbmcuLi48L3RleHQ+PC9zdmc+'
const errorPlaceholder =
  'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzAwIiBoZWlnaHQ9IjMwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSJ0cmFuc3BhcmVudCIvPjx0ZXh0IHg9IjUwJSIgeT0iNTAlIiBmb250LWZhbWlseT0iQXJpYWwsIEhlbHZldGljYSwgc2Fucy1zZXJpZiIgZm9udC1zaXplPSIxOCIgZmlsbD0id2hpdGUiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGRvbWluYW50LWJhc2VsaW5lPSJtaWRkbGUiPkltYWdlIEVycm9yPC90ZXh0Pjwvc3ZnPg=='

const props = defineProps({
  imageList: {
    type: Array as PropType<string[]>,
    required: true,
  },
  thumbnailList: {
    type: Array as PropType<string[]>,
    default: () => [],
  },
})

const store = appStoreData()
const isModernAppearance = computed(
  () => (store.appConfig.appearanceStyle || store.appConfig.headerStyle || 'modern') === 'modern',
)
const showImageViewer = ref(false)
const initialIndex = ref(0)
const currentIndex = ref(0)
const loadedImages = ref<Record<number, string>>({})
const loadingImages = ref<Record<number, boolean>>({})
const showThumbnails = ref(true)
const thumbnailTrackRef = ref<HTMLElement>()
const thumbnailItemsRef = ref<HTMLElement[]>([])
const imageMode = ref<T_imageMode>('contain')
const scale = ref(1)
const rotation = ref(0)
const offset = reactive({ x: 0, y: 0 })
const dragState = reactive({ active: false, startX: 0, startY: 0, originX: 0, originY: 0 })
const zoomStep = 0.2
let previousBodyOverflow = ''

const imageList_C = computed(() => {
  const result = props.imageList.map((_, index) => loadedImages.value[index] || loadingPlaceholder)
  preloadAround(currentIndex.value)
  return result
})
const currentImage = computed(() => loadedImages.value[currentIndex.value] || loadingPlaceholder)
const progressText = computed(() =>
  props.imageList.length > 0 ? `${currentIndex.value + 1} / ${props.imageList.length}` : '0 / 0',
)
const scaleText = computed(() => `${Math.round(scale.value * 100)}%`)
const imageStyle = computed<CSSProperties>(() => ({
  transform: `translate(-50%, -50%) translate(${offset.x}px, ${offset.y}px) rotate(${rotation.value}deg) scale(${scale.value})`,
}))

const thumbnailSource = (index: number) =>
  props.thumbnailList[index] || props.imageList[index] || loadingPlaceholder

watch(
  () => props.imageList.length,
  (length) => {
    if (length === 0) {
      currentIndex.value = 0
      return
    }
    if (currentIndex.value >= length) currentIndex.value = length - 1
  },
)

const preloadImage = (index: number) => {
  if (index < 0 || index >= props.imageList.length) return
  if (loadedImages.value[index] || loadingImages.value[index]) return

  loadingImages.value[index] = true
  const img = new Image()
  img.onload = () => {
    loadedImages.value[index] = props.imageList[index]
    loadingImages.value[index] = false
  }
  img.onerror = () => {
    loadedImages.value[index] = errorPlaceholder
    loadingImages.value[index] = false
  }
  img.src = props.imageList[index]
}

const preloadAround = (index: number) => {
  for (
    let itemIndex = Math.max(0, index - 2);
    itemIndex <= Math.min(props.imageList.length - 1, index + 2);
    itemIndex++
  ) {
    preloadImage(itemIndex)
  }
}

const resetTransform = () => {
  scale.value = 1
  rotation.value = 0
  offset.x = 0
  offset.y = 0
  imageMode.value = 'contain'
}

const scrollCurrentThumbnailIntoView = () => {
  nextTick(() => {
    thumbnailItemsRef.value[currentIndex.value]?.scrollIntoView({
      behavior: 'auto',
      block: 'nearest',
      inline: 'center',
    })
  })
}

const setCurrentIndex = (index: number) => {
  if (props.imageList.length === 0) return
  const length = props.imageList.length
  currentIndex.value = ((index % length) + length) % length
  preloadAround(currentIndex.value)
  resetTransform()
  scrollCurrentThumbnailIntoView()
}

const changeImage = (step: number) => setCurrentIndex(currentIndex.value + step)

const handleSwitch = (index: number) => {
  currentIndex.value = index
  preloadAround(index)
}

const zoom = (amount: number) => {
  scale.value = Math.min(5, Math.max(0.2, Number((scale.value + amount).toFixed(2))))
}

const rotate = (degrees: number) => {
  rotation.value = (rotation.value + degrees) % 360
}

const toggleImageMode = () => {
  imageMode.value = imageMode.value === 'contain' ? 'original' : 'contain'
  scale.value = 1
  offset.x = 0
  offset.y = 0
}

const toggleThumbnails = () => {
  showThumbnails.value = !showThumbnails.value
  if (showThumbnails.value) scrollCurrentThumbnailIntoView()
}

const scrollThumbnails = (direction: number) => {
  thumbnailTrackRef.value?.scrollBy({ left: direction * 480, behavior: 'smooth' })
}

const handleWheel = (event: WheelEvent) => {
  zoom(event.deltaY < 0 ? zoomStep : -zoomStep)
}

const handleMouseMove = (event: MouseEvent) => {
  if (!dragState.active) return
  offset.x = dragState.originX + event.clientX - dragState.startX
  offset.y = dragState.originY + event.clientY - dragState.startY
}

const stopDrag = () => {
  dragState.active = false
  document.removeEventListener('mousemove', handleMouseMove)
  document.removeEventListener('mouseup', stopDrag)
}

const startDrag = (event: MouseEvent) => {
  dragState.active = true
  dragState.startX = event.clientX
  dragState.startY = event.clientY
  dragState.originX = offset.x
  dragState.originY = offset.y
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', stopDrag)
}

const handleKeyDown = (event: KeyboardEvent) => {
  if (!showImageViewer.value || !isModernAppearance.value) return
  if (event.key === 'Escape') closeImageViewer()
  else if (event.key === 'ArrowLeft') changeImage(-1)
  else if (event.key === 'ArrowRight') changeImage(1)
  else if (event.key === 'ArrowUp') zoom(zoomStep)
  else if (event.key === 'ArrowDown') zoom(-zoomStep)
  else return
  event.preventDefault()
}

const openImageViewer = (index = 0) => {
  if (props.imageList.length === 0) return
  initialIndex.value = Math.min(Math.max(index, 0), props.imageList.length - 1)
  currentIndex.value = initialIndex.value
  resetTransform()
  preloadAround(currentIndex.value)
  showImageViewer.value = true

  if (isModernAppearance.value) {
    previousBodyOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    document.addEventListener('keydown', handleKeyDown)
    scrollCurrentThumbnailIntoView()
  }
}

const closeImageViewer = () => {
  showImageViewer.value = false
  stopDrag()
  document.removeEventListener('keydown', handleKeyDown)
  if (isModernAppearance.value) document.body.style.overflow = previousBodyOverflow
}

onBeforeUnmount(() => {
  stopDrag()
  document.removeEventListener('keydown', handleKeyDown)
  if (showImageViewer.value) document.body.style.overflow = previousBodyOverflow
})

defineExpose({
  openImageViewer,
  closeImageViewer,
})
</script>

<style lang="scss" scoped>
.modern-image-viewer {
  --viewer-accent: #25b8b5;
  --viewer-accent-soft: rgba(37, 184, 181, 0.14);
  --viewer-surface: rgba(27, 30, 35, 0.94);
  --viewer-surface-soft: rgba(31, 35, 40, 0.88);
  --viewer-border: rgba(255, 255, 255, 0.14);
  --viewer-text: #e7eaee;
  --viewer-muted: #a2aab3;

  position: fixed;
  inset: 0;
  z-index: 3000;
  overflow: hidden;
  color: var(--viewer-text);
  background: radial-gradient(circle at center, #171a1f 0, #0d0f12 72%);
  user-select: none;
}

.modern-image-viewer__header {
  position: absolute;
  top: 14px;
  left: 4.5%;
  right: 4.5%;
  z-index: 4;
  height: 46px;
  padding: 0 10px 0 14px;
  box-sizing: border-box;
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  border: 1px solid var(--viewer-border);
  border-radius: 10px;
  background: var(--viewer-surface);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.22);
  backdrop-filter: blur(16px);
}

.modern-image-viewer__title {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;

  .el-icon {
    width: 24px;
    height: 24px;
    flex-shrink: 0;
    color: var(--viewer-accent);
    font-size: 19px;
  }

  strong {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 15px;
  }
}

.modern-image-viewer__progress,
.modern-image-viewer__toolbar-progress,
.modern-image-viewer__scale {
  font-variant-numeric: tabular-nums;
}

.modern-image-viewer__progress {
  color: var(--viewer-text);
  font-size: 14px;
}

.modern-image-viewer__close {
  justify-self: end;
}

.modern-image-viewer__close,
.modern-image-viewer__side-button,
.modern-image-viewer__thumbnail-scroll,
.modern-image-viewer__toolbar button {
  display: grid;
  place-items: center;
  padding: 0;
  color: var(--viewer-text);
  background: var(--viewer-surface-soft);
  border: 1px solid var(--viewer-border);
  cursor: pointer;
  transition:
    color 0.16s ease,
    border-color 0.16s ease,
    background-color 0.16s ease;

  &:hover,
  &:focus-visible {
    color: var(--viewer-accent);
    border-color: var(--viewer-accent);
    outline: none;
  }
}

.modern-image-viewer__close {
  width: 32px;
  height: 32px;
  border-radius: 7px;
  font-size: 19px;
}

.modern-image-viewer__side-button {
  position: absolute;
  top: 46%;
  z-index: 3;
  width: 58px;
  height: 58px;
  border-radius: 50%;
  font-size: 25px;
  transform: translateY(-50%);

  &--prev {
    left: 4.5%;
  }

  &--next {
    right: 4.5%;
  }
}

.modern-image-viewer__canvas {
  position: absolute;
  inset: 72px 9% 94px;
  overflow: hidden;
}

.modern-image-viewer--thumbs .modern-image-viewer__canvas {
  bottom: 252px;
}

.modern-image-viewer__image {
  position: absolute;
  top: 50%;
  left: 50%;
  display: block;
  background: #08090b;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 4px;
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.36);
  cursor: grab;
  transform-origin: center center;
  transition: box-shadow 0.16s ease;

  &:active {
    cursor: grabbing;
  }

  &.is-contain {
    max-width: 100%;
    max-height: 100%;
  }

  &.is-original {
    max-width: none;
    max-height: none;
  }
}

.modern-image-viewer__thumbnail-panel {
  position: absolute;
  left: 50%;
  bottom: 84px;
  z-index: 4;
  width: min(78vw, 1180px);
  height: 112px;
  padding: 10px 12px;
  box-sizing: border-box;
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr) 34px;
  align-items: center;
  gap: 10px;
  border: 1px solid var(--viewer-border);
  border-radius: 10px;
  background: var(--viewer-surface);
  box-shadow: 0 14px 34px rgba(0, 0, 0, 0.25);
  backdrop-filter: blur(16px);
  transform: translateX(-50%);
}

.modern-image-viewer__thumbnail-scroll {
  width: 32px;
  height: 42px;
  border-radius: 7px;
  font-size: 16px;
}

.modern-image-viewer__thumbnail-track {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  overflow-x: auto;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.modern-image-viewer__thumbnail {
  position: relative;
  width: 126px;
  height: 82px;
  padding: 3px;
  flex: 0 0 auto;
  overflow: hidden;
  border: 1px solid var(--viewer-border);
  border-radius: 7px;
  background: rgba(255, 255, 255, 0.035);
  cursor: pointer;
  transition:
    border-color 0.16s ease,
    background-color 0.16s ease;

  &:hover {
    border-color: rgba(255, 255, 255, 0.34);
  }

  &.active {
    border: 2px solid var(--viewer-accent);
    background: var(--viewer-accent-soft);
  }

  img {
    width: 100%;
    height: 100%;
    display: block;
    object-fit: cover;
    border-radius: 4px;
  }

  span {
    position: absolute;
    top: 6px;
    left: 7px;
    min-width: 20px;
    height: 20px;
    padding: 0 5px;
    box-sizing: border-box;
    display: grid;
    place-items: center;
    border-radius: 5px;
    color: #ffffff;
    background: rgba(12, 14, 17, 0.72);
    font-size: 11px;
    line-height: 1;
  }

  &.active span {
    background: var(--viewer-accent);
  }
}

.modern-image-viewer__toolbar {
  position: absolute;
  bottom: 22px;
  left: 50%;
  z-index: 5;
  height: 50px;
  padding: 0 10px;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  gap: 8px;
  border: 1px solid var(--viewer-border);
  border-radius: 10px;
  background: var(--viewer-surface);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.26);
  backdrop-filter: blur(16px);
  transform: translateX(-50%);

  button {
    width: 34px;
    height: 34px;
    border-radius: 7px;
    font-size: 17px;

    &.active {
      color: var(--viewer-accent);
      border-color: rgba(37, 184, 181, 0.42);
      background: var(--viewer-accent-soft);
    }
  }
}

.modern-image-viewer__toolbar-progress {
  min-width: 54px;
  text-align: center;
  font-size: 13px;
}

.modern-image-viewer__scale {
  min-width: 45px;
  color: var(--viewer-text);
  text-align: center;
  font-size: 13px;
}

.modern-image-viewer__divider {
  width: 1px;
  height: 24px;
  margin: 0 1px;
  flex-shrink: 0;
  background: var(--viewer-border);
}

.modern-image-viewer__fit-text {
  position: absolute;
  left: 50%;
  bottom: 1px;
  font-size: 9px;
  line-height: 1;
  transform: translateX(-50%);
}

.modern-image-viewer__toolbar button:has(.modern-image-viewer__fit-text) {
  position: relative;

  .el-icon {
    transform: translateY(-4px);
  }
}

.modern-image-viewer-fade-enter-active,
.modern-image-viewer-fade-leave-active {
  transition: opacity 0.18s ease;
}

.modern-image-viewer-fade-enter-from,
.modern-image-viewer-fade-leave-to {
  opacity: 0;
}

@media (max-width: 820px) {
  .modern-image-viewer__header {
    left: 12px;
    right: 12px;
  }

  .modern-image-viewer__side-button {
    width: 42px;
    height: 42px;

    &--prev {
      left: 12px;
    }

    &--next {
      right: 12px;
    }
  }

  .modern-image-viewer__canvas {
    inset-inline: 54px;
  }

  .modern-image-viewer__thumbnail-panel {
    width: calc(100vw - 24px);
    height: 96px;
  }

  .modern-image-viewer__thumbnail {
    width: 104px;
    height: 66px;
  }

  .modern-image-viewer__toolbar {
    max-width: calc(100vw - 20px);
    gap: 4px;
    overflow-x: auto;
  }

  .modern-image-viewer__toolbar button {
    flex: 0 0 32px;
  }
}
</style>
