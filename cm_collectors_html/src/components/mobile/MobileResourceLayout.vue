<template>
  <div class="mobile-resource-layout"
    :class="[`mobile-resource-layout--${layoutFamily}`, modeClass, { 'mobile-resource-layout--bright': isBrightTheme }]">
    <div v-if="props.dataList.length === 0" class="mobile-empty">
      <el-icon><Picture /></el-icon>
      <span>暂无资源</span>
    </div>

    <div v-else-if="layoutFamily === 'poster'" class="mobile-poster-grid">
      <button v-for="resource in props.dataList" :key="resource.id" class="mobile-poster-card"
        :class="{ 'mobile-poster-card--wide': isWideCover(resource) }" type="button"
        @click="selectResource(resource)">
        <span class="mobile-cover-shell mobile-poster-cover" :style="wideCoverStyle(resource)">
          <img v-if="coverUrl(resource)" :src="coverUrl(resource)" :alt="resource.title" loading="lazy">
          <span v-else class="mobile-cover-placeholder"><el-icon><Picture /></el-icon></span>
          <span v-if="duration(resource)" class="mobile-duration">{{ duration(resource) }}</span>
        </span>
        <span class="mobile-card-body">
          <strong class="mobile-card-title">{{ resource.title }}</strong>
          <span class="mobile-card-meta">
            <span v-if="year(resource)">{{ year(resource) }}</span>
            <span v-if="resource.definition" class="mobile-quality">{{ definition(resource) }}</span>
            <span v-if="showExpandedPosterInfo && resource.dramaSeries.length > 1">
              {{ resource.dramaSeries.length }} 集
            </span>
          </span>
        </span>
      </button>
    </div>

    <div v-else-if="layoutFamily === 'list'" class="mobile-list">
      <button v-for="resource in props.dataList" :key="resource.id" class="mobile-list-card" type="button"
        @click="selectResource(resource)">
        <span class="mobile-cover-shell mobile-list-cover">
          <img v-if="coverUrl(resource)" :src="coverUrl(resource)" :alt="resource.title" loading="lazy">
          <span v-else class="mobile-cover-placeholder"><el-icon><Picture /></el-icon></span>
        </span>
        <span class="mobile-list-info">
          <strong class="mobile-card-title">{{ resource.title }}</strong>
          <span class="mobile-card-meta mobile-list-meta">
            <span v-if="year(resource)">{{ year(resource) }}</span>
            <span v-if="resource.definition" class="mobile-quality">{{ definition(resource) }}</span>
            <span v-if="duration(resource)">{{ duration(resource) }}</span>
            <span v-if="resource.dramaSeries.length > 1">共 {{ resource.dramaSeries.length }} 集</span>
          </span>
          <span v-if="resource.tags.length > 0" class="mobile-list-tags">
            <span v-for="tag in resource.tags.slice(0, 2)" :key="tag.id">{{ tag.name }}</span>
          </span>
        </span>
        <span class="mobile-play-button" aria-hidden="true"><el-icon><VideoPlay /></el-icon></span>
      </button>
    </div>

    <div v-else-if="layoutFamily === 'wall'" class="mobile-wall">
      <button v-for="resource in props.dataList" :key="resource.id" class="mobile-wall-card"
        :class="{ 'mobile-wall-card--wide': shouldSpanWideWall(resource) }"
        type="button"
        @click="selectResource(resource)">
        <span class="mobile-cover-shell mobile-wall-cover" :style="wallCoverStyle(resource)">
          <img v-if="coverUrl(resource)" :src="coverUrl(resource)" :alt="resource.title" loading="lazy">
          <span v-else class="mobile-cover-placeholder"><el-icon><Picture /></el-icon></span>
          <span class="mobile-wall-gradient"></span>
          <strong class="mobile-wall-title">{{ resource.title }}</strong>
          <span v-if="resource.definition" class="mobile-wall-quality">{{ definition(resource) }}</span>
          <span v-if="duration(resource)" class="mobile-wall-duration">{{ duration(resource) }}</span>
        </span>
      </button>
    </div>

    <div v-else class="mobile-short-video">
      <section class="mobile-short-player-shell">
        <videoPlayComponent ref="videoPlayRef" class="mobile-short-player" :use-video-play-controls="false" />
        <div v-if="shortVideoLoading" class="mobile-short-player-state">
          <span class="mobile-short-loading"></span>
          <span>正在加载</span>
        </div>
        <div v-else-if="!currentDramaSeriesId" class="mobile-short-player-state">
          <el-icon><VideoPlay /></el-icon>
          <span>暂无可播放视频</span>
        </div>
      </section>

      <section v-if="currentResource" class="mobile-short-info">
        <div class="mobile-short-heading">
          <div>
            <strong>{{ currentResource.title }}</strong>
            <span class="mobile-feature-meta">
              <span v-if="currentResource.definition" class="mobile-quality">
                {{ definition(currentResource) }}
              </span>
              <span v-if="duration(currentResource)">{{ duration(currentResource) }}</span>
              <span v-if="year(currentResource)">{{ year(currentResource) }}</span>
            </span>
          </div>
          <span class="mobile-short-position">{{ currentPlayIndex + 1 }} / {{ props.dataList.length }}</span>
        </div>
        <div class="mobile-short-navigation">
          <button type="button" :disabled="currentPlayIndex <= 0" @click="playPrevious">
            上一个
          </button>
          <button type="button" :disabled="currentPlayIndex >= props.dataList.length - 1" @click="playNext">
            下一个
          </button>
        </div>
      </section>

      <div class="mobile-short-queue">
        <button v-for="(resource, index) in props.dataList" :key="resource.id" type="button"
          class="mobile-short-card" :class="{ active: index === currentPlayIndex }"
          @click="switchShortVideo(index)">
          <span class="mobile-cover-shell mobile-short-cover">
            <img v-if="coverUrl(resource)" :src="coverUrl(resource)" :alt="resource.title" loading="lazy">
            <span v-else class="mobile-cover-placeholder"><el-icon><Picture /></el-icon></span>
            <span class="mobile-wall-gradient"></span>
            <span class="mobile-short-card-play"><el-icon><VideoPlay /></el-icon></span>
            <strong>{{ resource.title }}</strong>
          </span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch, type CSSProperties, type PropType } from 'vue'
import { Picture, VideoPlay } from '@element-plus/icons-vue'
import type { I_resource } from '@/dataType/resource.dataType'
import type { T_resourcesShowMode } from '@/dataType/app.dataType'
import { getResourceCoverPoster } from '@/common/photo'
import { getResourceDurationText } from '@/common/videoDuration'
import { appStoreData } from '@/storeData/app.storeData'
import { AppLang } from '@/language/app.lang'
import videoPlayComponent from '@/components/play/videoPlay.vue'
import { getPlayVideoURLAndType, playUpdate } from '@/common/play'

type MobileLayoutFamily = 'poster' | 'list' | 'wall' | 'immersive'

const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
  mode: {
    type: String as PropType<T_resourcesShowMode>,
    required: true,
  },
})

const emit = defineEmits<{
  (event: 'select-resource', resource: I_resource): void
}>()

const store = appStoreData()
const appLang = AppLang()
const videoPlayRef = ref<InstanceType<typeof videoPlayComponent>>()
const currentPlayIndex = ref(0)
const currentPlayingDramaSeriesId = ref('')
const shortVideoLoading = ref(false)
let shortVideoRequestVersion = 0

const mobileLayoutMap: Record<T_resourcesShowMode, MobileLayoutFamily> = {
  coverPoster: 'poster',
  coverPosterSimple: 'poster',
  coverPosterSimpleExpand: 'poster',
  coverPosterBox: 'list',
  coverPosterBoxWideSeparate: 'list',
  table: 'list',
  coverPosterWaterfall: 'wall',
  coverPosterMosaic: 'wall',
  coverPosterCompactWall: 'wall',
  coverPosterMosaicShortVideo: 'immersive',
  shortVideo: 'immersive',
  shortVideoTopBottom: 'immersive',
}

const layoutFamily = computed<MobileLayoutFamily>(() => mobileLayoutMap[props.mode] || 'poster')
const modeClass = computed(() => `mobile-resource-mode--${props.mode}`)
const isBrightTheme = computed(() => store.appConfig.theme === 'bright')
const showExpandedPosterInfo = computed(() => props.mode === 'coverPosterSimpleExpand')
const currentResource = computed(() => props.dataList[currentPlayIndex.value])
const currentDramaSeriesId = computed(() => currentResource.value?.dramaSeries[0]?.id || '')
const resourceListKey = computed(() => props.dataList.map(resource => resource.id).join('|'))

const coverUrl = (resource: I_resource) => getResourceCoverPoster(resource)
const duration = (resource: I_resource) => {
  if (!store.currentConfigApp.showVideoDuration) return ''
  return getResourceDurationText(resource, { compact: true })
}
const year = (resource: I_resource) => resource.issuingDate ? resource.issuingDate.slice(0, 4) : ''
const definition = (resource: I_resource) => appLang.definition(resource.definition)

const coverAspectRatio = (resource: I_resource) => {
  if (resource.coverPosterWidth <= 0 || resource.coverPosterHeight <= 0) return 0
  return resource.coverPosterWidth / resource.coverPosterHeight
}

const isWideCover = (resource: I_resource) => coverAspectRatio(resource) >= 1.2
const shouldSpanWideWall = (resource: I_resource) =>
  props.mode === 'coverPosterMosaic' && isWideCover(resource)

const wideCoverStyle = (resource: I_resource): CSSProperties => {
  if (!isWideCover(resource)) return {}
  return { aspectRatio: Math.min(2.2, coverAspectRatio(resource)).toString() }
}

const wallCoverStyle = (resource: I_resource): CSSProperties => {
  const coverRatio = coverAspectRatio(resource)
  if (coverRatio <= 0) return {}
  if (props.mode !== 'coverPosterWaterfall' && !shouldSpanWideWall(resource)) return {}
  const ratio = Math.min(2.2, Math.max(0.58, coverRatio))
  return { aspectRatio: ratio.toString() }
}

const selectResource = (resource: I_resource) => emit('select-resource', resource)

const resetShortVideoPlayer = () => {
  shortVideoRequestVersion++
  shortVideoLoading.value = false
  currentPlayingDramaSeriesId.value = ''
  videoPlayRef.value?.pause()
  videoPlayRef.value?.resetPlayer()
}

const playShortVideo = async (index: number, shouldPlay: boolean) => {
  if (layoutFamily.value !== 'immersive') return
  const resource = props.dataList[index]
  const dramaSeries = resource?.dramaSeries[0]
  currentPlayIndex.value = index
  if (!resource || !dramaSeries) {
    resetShortVideoPlayer()
    return
  }

  const player = videoPlayRef.value
  if (!player) return
  if (dramaSeries.id === currentPlayingDramaSeriesId.value && !shortVideoLoading.value) {
    if (shouldPlay) player.play()
    if (shouldPlay) void playUpdate(resource.id, dramaSeries.id)
    return
  }

  const requestVersion = ++shortVideoRequestVersion
  shortVideoLoading.value = true
  currentPlayingDramaSeriesId.value = dramaSeries.id
  player.pause()
  try {
    const { playUrl, playType } = await getPlayVideoURLAndType(dramaSeries.id)
    if (requestVersion !== shortVideoRequestVersion) return
    if (!playUrl) throw new Error('播放地址为空')
    player.setVideoSource(playUrl, playType, () => {
      if (requestVersion !== shortVideoRequestVersion) return
      shortVideoLoading.value = false
      player.addTextTrack(`/api/video/subtitle/${dramaSeries.id}`, '默认字幕', 'zh', true)
      if (shouldPlay) {
        void player.play()
        void playUpdate(resource.id, dramaSeries.id)
      } else {
        player.pause()
      }
    }, dramaSeries.src || resource.title, 0, () => {
      if (requestVersion !== shortVideoRequestVersion) return
      shortVideoLoading.value = false
      currentPlayingDramaSeriesId.value = ''
    })
  } catch (error) {
    if (requestVersion !== shortVideoRequestVersion) return
    console.error(error)
    shortVideoLoading.value = false
    currentPlayingDramaSeriesId.value = ''
  }
}

const playPrevious = () => {
  if (currentPlayIndex.value <= 0) return
  void switchShortVideo(currentPlayIndex.value - 1)
}

const playNext = () => {
  if (currentPlayIndex.value >= props.dataList.length - 1) return
  void switchShortVideo(currentPlayIndex.value + 1)
}

const switchShortVideo = (index: number) => {
  const shouldResume = videoPlayRef.value?.isPlaying() || false
  return playShortVideo(index, shouldResume)
}

watch(
  [layoutFamily, resourceListKey],
  ([family, listKey], [previousFamily, previousListKey]) => {
    if (family !== 'immersive') {
      if (previousFamily === 'immersive') resetShortVideoPlayer()
      return
    }
    if (family !== previousFamily || listKey !== previousListKey) {
      resetShortVideoPlayer()
      currentPlayIndex.value = 0
      nextTick(() => void playShortVideo(0, false))
    }
  },
  { immediate: true },
)

onBeforeUnmount(resetShortVideoPlayer)
</script>

<style scoped lang="scss">
.mobile-resource-layout {
  --mobile-card-bg: #20252a;
  --mobile-card-border: rgba(255, 255, 255, 0.11);
  --mobile-text: #eef1f4;
  --mobile-muted: #9ca6ad;
  --mobile-accent: #2bc2c4;
  --mobile-accent-soft: rgba(43, 194, 196, 0.14);
  width: 100%;
  color: var(--mobile-text);
}

.mobile-resource-layout--immersive {
  height: 100%;
  min-height: 0;
}

button {
  min-width: 0;
  padding: 0;
  color: inherit;
  font: inherit;
  text-align: left;
  background: none;
  border: 0;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}

.mobile-cover-shell {
  position: relative;
  display: block;
  min-width: 0;
  overflow: hidden;
  background: #11161a;

  img {
    width: 100%;
    height: 100%;
    display: block;
    object-fit: cover;
  }
}

.mobile-cover-placeholder {
  width: 100%;
  height: 100%;
  display: grid;
  place-items: center;
  color: var(--mobile-muted);
  font-size: 28px;
}

.mobile-card-title {
  min-width: 0;
  display: -webkit-box;
  overflow: hidden;
  color: var(--mobile-text);
  font-size: 14px;
  font-weight: 650;
  line-height: 1.35;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.mobile-card-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 5px 8px;
  color: var(--mobile-muted);
  font-size: 11px;
}

.mobile-quality {
  color: var(--mobile-accent);
}

.mobile-duration {
  position: absolute;
  right: 6px;
  bottom: 6px;
  padding: 2px 5px;
  color: #fff;
  font-size: 10px;
  line-height: 1.25;
  background: rgba(0, 0, 0, 0.72);
  border-radius: 4px;
}

.mobile-empty {
  min-height: 240px;
  display: grid;
  place-content: center;
  justify-items: center;
  gap: 10px;
  color: var(--mobile-muted);

  .el-icon {
    font-size: 36px;
  }
}

.mobile-poster-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.mobile-poster-card {
  overflow: hidden;
  background: var(--mobile-card-bg);
  border: 1px solid var(--mobile-card-border);
  border-radius: 9px;

  &:active {
    border-color: var(--mobile-accent);
    transform: scale(0.985);
  }
}

.mobile-poster-card--wide {
  grid-column: 1 / -1;
}

.mobile-poster-cover {
  aspect-ratio: 4 / 5;
}

.mobile-card-body {
  min-height: 56px;
  padding: 8px 9px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  box-sizing: border-box;
}

.mobile-resource-mode--coverPosterSimple .mobile-card-body {
  min-height: 42px;
}

.mobile-resource-mode--coverPosterSimple .mobile-card-meta {
  display: none;
}

.mobile-resource-mode--coverPosterSimpleExpand .mobile-card-body {
  min-height: 68px;
}

.mobile-list {
  display: flex;
  flex-direction: column;
  gap: 9px;
}

.mobile-list-card {
  min-height: 112px;
  padding: 9px;
  display: grid;
  grid-template-columns: minmax(104px, 34%) minmax(0, 1fr) 34px;
  align-items: center;
  gap: 11px;
  background: var(--mobile-card-bg);
  border: 1px solid var(--mobile-card-border);
  border-radius: 10px;

  &:active {
    border-color: var(--mobile-accent);
    background: var(--mobile-accent-soft);
  }
}

.mobile-list-cover {
  width: 100%;
  height: 92px;
  border-radius: 7px;
}

.mobile-resource-mode--coverPosterBoxWideSeparate .mobile-list-card {
  grid-template-columns: minmax(132px, 42%) minmax(0, 1fr) 32px;
}

.mobile-list-info {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.mobile-list-meta {
  line-height: 1.45;
}

.mobile-list-tags {
  min-width: 0;
  display: flex;
  gap: 5px;
  overflow: hidden;

  span {
    max-width: 92px;
    padding: 2px 5px;
    overflow: hidden;
    color: var(--mobile-muted);
    font-size: 10px;
    text-overflow: ellipsis;
    white-space: nowrap;
    background: rgba(255, 255, 255, 0.045);
    border-radius: 4px;
  }
}

.mobile-play-button {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  color: var(--mobile-text);
  border: 1px solid var(--mobile-card-border);
  border-radius: 50%;
}

.mobile-wall {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 6px;
}

.mobile-wall-card {
  min-width: 0;
  overflow: hidden;
  border-radius: 7px;

  &:active {
    outline: 2px solid var(--mobile-accent);
    outline-offset: -2px;
  }
}

.mobile-wall-card--wide {
  grid-column: 1 / -1;
}

.mobile-wall-cover {
  aspect-ratio: 3 / 4;
}

.mobile-wall-gradient,
.mobile-feature-gradient {
  position: absolute;
  inset: 35% 0 0;
  background: linear-gradient(180deg, transparent, rgba(0, 0, 0, 0.92));
  pointer-events: none;
}

.mobile-wall-title {
  position: absolute;
  right: 6px;
  bottom: 24px;
  left: 6px;
  overflow: hidden;
  color: #fff;
  font-size: 12px;
  line-height: 1.3;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mobile-wall-quality,
.mobile-wall-duration {
  position: absolute;
  bottom: 6px;
  color: #fff;
  font-size: 9px;
}

.mobile-wall-quality {
  left: 6px;
  color: var(--mobile-accent);
}

.mobile-wall-duration {
  right: 6px;
}

.mobile-resource-mode--coverPosterCompactWall .mobile-wall {
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 5px;
}

.mobile-resource-mode--coverPosterCompactWall .mobile-wall-cover {
  aspect-ratio: 3 / 4;
}

.mobile-resource-mode--coverPosterCompactWall .mobile-wall-title {
  right: 4px;
  bottom: 18px;
  left: 4px;
  font-size: 10px;
}

.mobile-resource-mode--coverPosterCompactWall .mobile-wall-quality,
.mobile-resource-mode--coverPosterCompactWall .mobile-wall-duration {
  bottom: 4px;
  font-size: 8px;
}

.mobile-resource-mode--coverPosterCompactWall .mobile-wall-quality {
  left: 4px;
}

.mobile-resource-mode--coverPosterCompactWall .mobile-wall-duration {
  right: 4px;
}

.mobile-resource-mode--coverPosterMosaic .mobile-wall-card:nth-child(5n + 1) {
  grid-column: span 2;

  .mobile-wall-cover {
    aspect-ratio: 16 / 9;
  }
}

.mobile-resource-mode--coverPosterWaterfall .mobile-wall {
  display: block;
  column-count: 2;
  column-gap: 7px;
}

.mobile-resource-mode--coverPosterWaterfall .mobile-wall-card {
  width: 100%;
  margin-bottom: 7px;
  display: inline-block;
  break-inside: avoid;
}

.mobile-resource-mode--coverPosterWaterfall .mobile-wall-cover {
  min-height: 130px;
}

.mobile-short-video {
  --mobile-short-gap: clamp(5px, 1.1dvh, 9px);
  --mobile-short-queue-height: clamp(84px, 14dvh, 116px);
  --mobile-short-card-width: calc(var(--mobile-short-queue-height) * 0.82);
  --mobile-short-button-height: clamp(30px, 4.6dvh, 36px);
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: var(--mobile-short-gap);
  overflow: hidden;
}

.mobile-short-player-shell {
  position: relative;
  width: 100%;
  min-height: 180px;
  flex: 1 1 auto;
  overflow: hidden;
  background: #050606;
  border: 1px solid var(--mobile-card-border);
  border-radius: 12px;
}

.mobile-short-player,
.mobile-short-player :deep(.video-player-container),
.mobile-short-player :deep(.video-player-windows) {
  width: 100%;
  height: 100%;
  min-height: 0;
}

.mobile-short-player :deep(.video-player-container) {
  background: #050606;
}

.mobile-short-player :deep(.native-video-player),
.mobile-short-player :deep(.video-js) {
  width: 100% !important;
  height: 100% !important;
  object-fit: contain;
  background: #050606;
}

.mobile-short-player :deep(.video-js.vjs-fluid) {
  height: 100% !important;
  padding-top: 0 !important;
}

.mobile-short-player-state {
  position: absolute;
  top: 50%;
  left: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: rgba(255, 255, 255, 0.72);
  font-size: 12px;
  pointer-events: none;
  transform: translate(-50%, -50%);

  .el-icon {
    font-size: 34px;
  }
}

.mobile-short-loading {
  width: 28px;
  height: 28px;
  border: 2px solid rgba(255, 255, 255, 0.22);
  border-top-color: var(--mobile-accent);
  border-radius: 50%;
  animation: mobile-short-spin 0.8s linear infinite;
}

@keyframes mobile-short-spin {
  to { transform: rotate(360deg); }
}

.mobile-short-info {
  flex: 0 0 auto;
  padding: clamp(1px, 0.4dvh, 3px) 2px 0;
  display: flex;
  flex-direction: column;
  gap: var(--mobile-short-gap);
}

.mobile-short-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;

  > div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: clamp(3px, 0.7dvh, 6px);
  }

  strong {
    display: block;
    overflow: hidden;
    color: var(--mobile-text) !important;
    font-size: clamp(13px, 3.8vw, 16px);
    line-height: 1.35;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.mobile-short-position {
  flex: 0 0 auto;
  padding-top: 2px;
  color: var(--mobile-muted) !important;
  font-size: clamp(10px, 2.8vw, 12px);
}

.mobile-feature-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: clamp(3px, 0.6dvh, 6px) clamp(5px, 1.8vw, 9px);
  color: var(--mobile-muted) !important;
  font-size: clamp(10px, 2.8vw, 12px);
}

.mobile-short-navigation {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: clamp(5px, 1.8vw, 9px);

  button {
    height: var(--mobile-short-button-height);
    min-height: var(--mobile-short-button-height);
    color: var(--mobile-text);
    font-size: 12px;
    text-align: center;
    background: var(--mobile-card-bg);
    border: 1px solid var(--mobile-card-border);
    border-radius: 7px;

    &:disabled {
      cursor: default;
      opacity: 0.38;
    }
  }
}

.mobile-short-queue {
  height: var(--mobile-short-queue-height);
  min-height: var(--mobile-short-queue-height);
  flex: 0 0 var(--mobile-short-queue-height);
  padding-bottom: clamp(2px, 0.5dvh, 4px);
  display: grid;
  grid-auto-columns: var(--mobile-short-card-width);
  grid-auto-flow: column;
  align-items: stretch;
  gap: clamp(5px, 1.8vw, 8px);
  overflow-x: auto;
  overflow-y: hidden;
  overscroll-behavior-inline: contain;
  scroll-snap-type: inline proximity;
  scrollbar-width: none;
  touch-action: pan-x;

  &::-webkit-scrollbar {
    display: none;
  }
}

.mobile-short-card {
  min-width: 0;
  height: 100%;
  overflow: hidden;
  scroll-snap-align: start;

  &.active .mobile-short-cover {
    box-shadow: 0 0 0 2px var(--mobile-accent) inset;
  }
}

.mobile-short-cover {
  width: 100%;
  height: 100%;
  border-radius: 8px;

  strong {
    position: absolute;
    right: 6px;
    bottom: 7px;
    left: 6px;
    overflow: hidden;
    color: #fff;
    font-size: 11px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.mobile-short-card-play {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 28px;
  height: 28px;
  display: grid;
  place-items: center;
  color: #fff;
  background: rgba(0, 0, 0, 0.48);
  border: 1px solid rgba(255, 255, 255, 0.28);
  border-radius: 50%;
  transform: translate(-50%, -50%);
}

.mobile-short-card.active .mobile-short-card-play {
  color: #083638;
  background: var(--mobile-accent);
  border-color: var(--mobile-accent);
}

.mobile-resource-layout--bright {
  --mobile-card-bg: #ffffff;
  --mobile-card-border: #dde4e8;
  --mobile-text: #26343d;
  --mobile-muted: #71808a;
  --mobile-accent: #159fa3;
  --mobile-accent-soft: rgba(21, 159, 163, 0.09);
}

@media (max-width: 360px) {
  .mobile-poster-grid {
    gap: 8px;
  }

  .mobile-list-card {
    grid-template-columns: 96px minmax(0, 1fr) 30px;
    gap: 8px;
  }

  .mobile-resource-mode--coverPosterBoxWideSeparate .mobile-list-card {
    grid-template-columns: 112px minmax(0, 1fr) 28px;
  }
}

@media (orientation: landscape) and (min-aspect-ratio: 3 / 2) and (max-height: 600px) {
  .mobile-poster-grid,
  .mobile-wall,
  .mobile-resource-mode--coverPosterCompactWall .mobile-wall {
    grid-template-columns: minmax(0, 1fr);
  }

  .mobile-resource-mode--coverPosterWaterfall .mobile-wall {
    column-count: 1;
  }

  .mobile-poster-card--wide,
  .mobile-wall-card--wide,
  .mobile-resource-mode--coverPosterMosaic .mobile-wall-card:nth-child(5n + 1) {
    grid-column: 1;
  }

  .mobile-poster-cover,
  .mobile-wall-cover,
  .mobile-resource-mode--coverPosterCompactWall .mobile-wall-cover,
  .mobile-resource-mode--coverPosterMosaic .mobile-wall-card:nth-child(5n + 1) .mobile-wall-cover {
    aspect-ratio: 1.2 / 1 !important;
  }
}
</style>
