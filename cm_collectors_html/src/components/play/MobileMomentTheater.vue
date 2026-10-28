<template>
  <section class="mobile-moment-theater" aria-label="手机版片段放映室">
    <div class="theater-stage">
      <MobileVideoPlayer v-if="selected && active" :key="session" :resource-id="selected.resourceId" :drama-series-id="selected.dramaSeriesId" :title="selected.resourceTitle || selected.title" :poster="momentCoverImage(selected.id, selected.start)" :initial-moment="initialPlayback" @playing="recordPlayback" />
      <div v-else class="theater-placeholder">{{ loading ? '正在加载片段…' : '当前文件库暂无可播放片段' }}</div>
    </div>
    <div class="theater-picker">
      <div class="picker-heading"><span>共 {{ total }} 个片段</span><span>左右滑动选择</span></div>
      <p v-if="error" role="alert">{{ error }} <el-button size="small" @click="load()">重试</el-button></p>
      <div ref="filmstrip" class="moment-filmstrip" aria-label="片段选项">
        <button v-for="item in items" :key="item.id" type="button" class="filmstrip-card" :class="{ 'is-selected': selected?.id === item.id }" :disabled="!item.available" :aria-label="`播放片段：${item.title}`" :aria-pressed="selected?.id === item.id" @click="select(item, true)">
          <div class="filmstrip-cover"><MomentCover :moment-id="item.id" :time="item.start" /><span v-if="selected?.id === item.id" class="selected-label">当前片段</span></div>
          <strong :title="item.title">{{ item.title }}</strong>
          <span class="filmstrip-time">{{ item.available ? momentTime(item.start) + (item.end ? ' — ' + momentTime(item.end) : '') : '源视频不可用' }}</span>
        </button>
        <el-button v-if="hasMore" class="filmstrip-more" :loading="loading" @click="load(true)">加载更多</el-button>
      </div>
    </div>
  </section>
</template>
<script setup lang="ts">
import { computed, nextTick, onActivated, onBeforeUnmount, onDeactivated, ref, watch } from 'vue'
import { momentApi, momentTime, momentCoverImage, subscribeMomentChanges, type VideoMoment } from '@/server/videoMoments.server'
import { playUpdate } from '@/common/play'
import MobileVideoPlayer from './mobileVideoPlayer.vue'
import MomentCover from './MomentCover.vue'

const props = defineProps<{ filesBasesId: string }>()
const items = ref<VideoMoment[]>([])
const selected = ref<VideoMoment>()
const total = ref(0)
const loading = ref(false)
const error = ref('')
const active = ref(true)
const autoplay = ref(false)
const session = ref(0)
const filmstrip = ref<HTMLElement>()
const initialPlayback = computed(() => selected.value ? { start: selected.value.start, end: selected.value.end, title: selected.value.title, autoplay: autoplay.value } : undefined)
const page = ref(1)
const hasMore = computed(() => page.value * 50 < total.value)
let version = 0, recorded = false
function select(item: VideoMoment, play: boolean) {
  if (!item.available) return
  selected.value = item
  autoplay.value = play
  recorded = false
  session.value++
}
function recordPlayback() {
  if (recorded || !selected.value) return
  recorded = true
  void playUpdate(selected.value.resourceId, selected.value.dramaSeriesId)
}
async function load(append = false) {
  if (!props.filesBasesId || !active.value || (append && loading.value)) return
  const request = ++version
  const lastRequestedPage = append ? page.value + 1 : page.value
  loading.value = true
  error.value = ''
  try {
    let nextItems = append ? [...items.value] : []
    let nextTotal = total.value, lastLoadedPage = page.value
    for (let targetPage = append ? lastRequestedPage : 1; targetPage <= lastRequestedPage; targetPage++) {
      const result = await momentApi.libraryList(props.filesBasesId, targetPage)
      if (request !== version) return
      if (!result.status) throw new Error(result.msg || '片段加载失败')
      nextTotal = result.data.total
      lastLoadedPage = Math.min(targetPage, Math.max(1, Math.ceil(nextTotal / 50)))
      nextItems.push(...result.data.items)
      if (targetPage * 50 >= nextTotal) break
    }
    const scrollLeft = filmstrip.value?.scrollLeft || 0
    items.value = [...new Map(nextItems.map(item => [item.id, item])).values()]
    page.value = lastLoadedPage
    total.value = nextTotal
    const current = items.value.find(item => item.id === selected.value?.id)
    if (!selected.value || !current?.available) {
      selected.value = undefined
      const first = items.value.find(item => item.available)
      if (first) select(first, false)
    } else if (current.start !== selected.value.start || current.end !== selected.value.end) {
      select(current, false)
    }
    await nextTick()
    if (request === version && filmstrip.value) filmstrip.value.scrollLeft = scrollLeft
  } catch (cause) {
    if (request === version) error.value = cause instanceof Error ? cause.message : '片段加载失败'
  } finally {
    if (request === version) loading.value = false
  }
}
watch(() => props.filesBasesId, () => {
  version++
  items.value = []
  selected.value = undefined
  total.value = 0
  page.value = 1
  void load()
}, { immediate: true })
const unsubscribe = subscribeMomentChanges(() => { void load() })
onDeactivated(() => { active.value = false; version++; selected.value = undefined; items.value = []; page.value = 1 })
onActivated(() => { if (!active.value) { active.value = true; void load() } })
onBeforeUnmount(() => { active.value = false; version++; unsubscribe() })
</script>
<style scoped>
.mobile-moment-theater { display: flex; flex-direction: column; gap: 18px; height: 100%; min-height: 340px; min-width: 0; }
.theater-stage { flex: 1; min-height: 180px; overflow: hidden; border-radius: 12px; background: #000; }
.theater-stage :deep(.phone-player-stage) { height: 100%; aspect-ratio: auto; }
.theater-placeholder { height: 100%; min-height: 180px; display: grid; place-items: center; color: var(--el-text-color-secondary); font-size: 13px; }
.theater-picker { flex-shrink: 0; min-width: 0; }
.picker-heading { display: flex; justify-content: space-between; margin-bottom: 10px; font-size: 12px; color: var(--el-text-color-secondary); }
.moment-filmstrip { display: flex; gap: 10px; overflow-x: auto; padding: 2px 0 8px; overscroll-behavior-x: contain; -webkit-overflow-scrolling: touch; }
.filmstrip-card { flex: 0 0 156px; width: 156px; padding: 5px; border: 1px solid var(--el-border-color-lighter); border-radius: 10px; background: var(--el-fill-color-extra-light); text-align: left; color: var(--el-text-color-primary); cursor: pointer; }
.filmstrip-card.is-selected { border-color: var(--el-color-primary); background: var(--el-color-primary-light-9); }
.filmstrip-card:disabled { opacity: .5; cursor: default; }
.filmstrip-cover { position: relative; height: 86px; border-radius: 6px; overflow: hidden; }
.filmstrip-cover :deep(img), .filmstrip-cover :deep(.cover-placeholder) { width: 100%; height: 100%; object-fit: cover; border-radius: inherit; }
.selected-label { position: absolute; left: 4px; bottom: 4px; padding: 2px 5px; border-radius: 4px; background: var(--el-color-primary); color: #fff; font-size: 10px; }
.filmstrip-card strong { display: block; margin: 7px 2px 5px; font-size: 12px; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.filmstrip-time { display: block; margin: 0 2px 3px; font-size: 10px; color: var(--el-text-color-secondary); font-variant-numeric: tabular-nums; }
.filmstrip-more { flex: 0 0 100px; align-self: center; margin: 0; }
</style>
