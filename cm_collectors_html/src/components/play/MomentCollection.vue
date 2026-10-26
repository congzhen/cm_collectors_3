<template>
  <section
    ref="collectionRoot"
    class="moment-collection"
    :class="{ 'moment-collection--library': !!filesBasesId, 'moment-collection--list': layout === 'list' }"
  >
    <div v-if="showHeading && !filesBasesId" class="collection-heading">
      <strong>片段收藏</strong>
      <button type="button" class="collection-refresh" :disabled="loading" @click="refresh">
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.8"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"
        >
          <path d="M20 7v5h-5M4 17v-5h5M6 7a7 7 0 0 1 12-1l2 3M4 15l2 3a7 7 0 0 0 12-1" />
        </svg>
        刷新
      </button>
    </div>
    <form v-if="filesBasesId" class="collection-search" @submit.prevent="search">
      <el-input
        v-model="keyword"
        placeholder="搜索片段、备注或资源标题"
        clearable
        maxlength="200"
        @clear="search"
      />
      <el-button native-type="submit" :loading="loading">搜索</el-button>
    </form>
    <p v-if="error" role="alert">{{ error }}</p>
    <p v-else-if="!loading && !items.length" class="collection-empty">
      {{
        filesBasesId && appliedKeyword
          ? '没有找到匹配的片段。'
          : '还没有收藏。在播放器的“管理”面板中记录喜欢的片段。'
      }}
    </p>
    <div class="collection-list">
      <article v-for="item in items" :key="item.id" class="collection-card">
        <button
          type="button"
          class="cover-button"
          :disabled="!item.available || disabled"
          :aria-label="`播放片段：${item.title}`"
          @click="open(item)"
        >
          <MomentCover :moment-id="item.id" :time="item.start" :refresh-token="coverRefresh" />
          <span v-if="item.end > item.start" class="duration-badge">{{
            momentDuration(item.start, item.end).replace(/^共 /, '').replace(' 秒', 's')
          }}</span>
        </button>
        <div class="card-text">
          <strong :title="item.title">{{ item.title }}</strong>
          <div v-if="showResourceTitle ?? !props.resourceId" class="resource-title" :title="item.resourceTitle">
            {{ item.resourceTitle || '原资源已移除' }}
          </div>
          <div class="card-time">
            <span
              >{{ momentTime(item.start) }}{{ item.end ? ' — ' + momentTime(item.end) : '' }}</span
            >
          </div>
          <p v-if="item.note" class="card-note" :title="item.note">{{ item.note }}</p>
          <span v-if="!item.available" class="unavailable">源视频不可用</span>
        </div>
        <button
          type="button"
          class="collection-play"
          :disabled="!item.available || disabled"
          :title="item.end ? '循环播放片段' : '播放片段'"
          :aria-label="`播放片段：${item.title}`"
          @click="open(item)"
        >
          <svg width="13" height="13" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
            <path d="m7 4 14 8-14 8Z" />
          </svg>
          播放
        </button>
      </article>
    </div>
    <div v-if="filesBasesId" class="collection-pagination">
      <span>共 {{ libraryTotal }} 个片段</span>
      <el-pagination
        :current-page="libraryPage"
        :page-size="50"
        :total="libraryTotal"
        layout="prev, pager, next"
        :pager-count="5"
        :disabled="loading"
        @current-change="changeLibraryPage"
      />
    </div>
    <el-button v-else-if="!filesBasesId && hasMore" :loading="loading" @click="more"
      >加载更多</el-button
    >
  </section>
</template>
<script setup lang="ts">
import { onActivated, onDeactivated, onBeforeUnmount, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElInput, ElPagination } from 'element-plus'
import { isMobile } from '@/assets/mobile'
import {
  momentApi,
  momentTime,
  momentDuration,
  subscribeMomentChanges,
  type VideoMoment,
} from '@/server/videoMoments.server'
import MomentCover from './MomentCover.vue'
const props = withDefaults(
  defineProps<{
    resourceId?: string
    filesBasesId?: string
    playInDialog?: boolean
    playMode?: 'page' | 'emit'
    layout?: 'grid' | 'list'
    showHeading?: boolean
    showResourceTitle?: boolean
    disabled?: boolean
  }>(),
  { resourceId: '', filesBasesId: '', playMode: 'page', layout: 'grid', showHeading: true, showResourceTitle: undefined },
)
const emit = defineEmits<{
  play: [moment: VideoMoment]
  'has-items-change': [hasItems: boolean]
  'load-error': [failed: boolean]
}>()
const router = useRouter()
const collectionRoot = ref<HTMLElement>()
const items = ref<VideoMoment[]>([])
watch(() => items.value.length > 0, hasItems => emit('has-items-change', hasItems), { immediate: true })
const coverRefresh = ref(0)
const keyword = ref(''),
  appliedKeyword = ref('')
const libraryPage = ref(1),
  libraryTotal = ref(0)
const loading = ref(false),
  error = ref(''),
  hasMore = ref(false)
watch(error, message => emit('load-error', !!message), { immediate: true })
let page = 1,
  version = 0
async function load(append = false, targetPage = 1) {
  const v = ++version
  const requestedPage = append ? page + 1 : 1
  loading.value = true
  if (props.filesBasesId) {
    const result = await momentApi.libraryList(props.filesBasesId, targetPage, appliedKeyword.value)
    if (v !== version) return
    loading.value = false
    error.value = result.status ? '' : result.msg
    if (result.status) {
      const lastPage = Math.max(1, Math.ceil(result.data.total / 50))
      if (targetPage > lastPage) {
        void load(false, lastPage)
        return
      }
      libraryTotal.value = result.data.total
      libraryPage.value = targetPage
      coverRefresh.value++
      items.value = result.data.items
      collectionRoot.value?.querySelector('.collection-list')?.scrollTo?.({ top: 0 })
    }
    return
  }
  const result = await momentApi.list(props.resourceId, requestedPage)
  if (v !== version) return
  loading.value = false
  error.value = result.status ? '' : result.msg
  if (result.status) {
    page = requestedPage
    if (!append) coverRefresh.value++
    items.value = append ? [...items.value, ...result.data] : result.data
    hasMore.value = result.data.length === 50
  }
}
function refresh() {
  void load()
}
function search() {
  appliedKeyword.value = keyword.value.trim()
  items.value = []
  libraryTotal.value = 0
  refresh()
}
function changeLibraryPage(value: number) {
  void load(false, value)
}
function more() {
  if (loading.value) return
  void load(true)
}
function open(item: VideoMoment) {
  if (!item.available || props.disabled) return
  emit('play', item)
  if (props.playInDialog || props.playMode === 'emit') {
    return
  }
  void router.push({
    path: `/play/${isMobile() ? 'moviesMobile' : 'movies'}/${encodeURIComponent(item.resourceId)}/${encodeURIComponent(item.dramaSeriesId)}`,
    query: { t: String(item.start), end: String(item.end), momentTitle: item.title },
  })
}
watch(
  () => [props.resourceId, props.filesBasesId],
  () => {
    page = 1
    keyword.value = ''
    appliedKeyword.value = ''
    libraryPage.value = 1
    libraryTotal.value = 0
    hasMore.value = false
    error.value = ''
    items.value = []
    refresh()
  },
  { immediate: true },
)
let active = true
const unsubscribe = subscribeMomentChanges(() => {
  if (active) void load(false, props.filesBasesId ? libraryPage.value : 1)
})
onDeactivated(() => {
  active = false
  version++
})
onActivated(() => {
  active = true
  refresh()
})
onBeforeUnmount(() => {
  unsubscribe()
  version++
})
</script>
<style scoped>
.moment-collection--library {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}
.moment-collection--library .collection-search,
.moment-collection--library .collection-pagination {
  flex-shrink: 0;
}
.moment-collection--library .collection-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  align-content: start;
  margin-bottom: 0;
  scrollbar-gutter: stable;
}
.collection-search {
  display: flex;
  gap: 8px;
  margin-bottom: 14px;
}
.collection-pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  background: var(--el-bg-color);
  z-index: 1;
}
.collection-heading {
  box-sizing: border-box;
  width: 100%;
  min-height: 40px;
  padding: 8px 16px;
  border-radius: 4px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}
.collection-heading strong {
  font-size: 14px;
  font-weight: 400;
  color: inherit;
}
.collection-refresh {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  min-height: 24px;
  padding: 0 8px;
  border: 0;
  border-radius: 7px;
  background: transparent;
  color: inherit;
  font: inherit;
  font-size: 12px;
  cursor: pointer;
}
.collection-refresh:hover {
  background: var(--el-color-primary-light-8);
  color: inherit;
}
.collection-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 340px), 1fr));
  gap: 12px;
  margin-bottom: 12px;
}
.moment-collection--list .collection-list {
  grid-template-columns: minmax(0, 1fr);
}
.collection-card {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  padding: 12px;
  background: var(--el-bg-color-overlay, #222831);
  border: 1px solid var(--el-border-color-lighter, #ffffff15);
  border-radius: 12px;
  box-sizing: border-box;
  transition: border-color 0.15s ease;
}
.collection-card:hover {
  border-color: var(--el-border-color, #ffffff30);
}
.cover-button {
  position: relative;
  flex: 0 0 104px;
  width: 104px;
  height: 66px;
  padding: 0;
  border: 0;
  border-radius: 8px;
  overflow: hidden;
  background: var(--el-fill-color-darker);
  cursor: pointer;
}
.duration-badge {
  position: absolute;
  right: 3px;
  bottom: 3px;
  padding: 0 3px;
  border-radius: 3px;
  background: #000b;
  color: #fff;
  font-size: 9px;
  font-weight: 400;
  line-height: 13px;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  pointer-events: none;
}
.cover-button :deep(img),
.cover-button :deep(.cover-placeholder) {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: inherit;
}
.cover-button :deep(.cover-placeholder) {
  display: grid;
  font-size: 11px;
}
.card-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.card-text strong {
  color: var(--el-text-color-primary);
  font-size: 13px;
  font-weight: 500;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.resource-title {
  color: var(--el-text-color-secondary);
  font-size: 11px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.card-time {
  display: flex;
  flex-wrap: wrap;
  gap: 2px 6px;
  color: var(--el-text-color-secondary);
  font-size: 11px;
  line-height: 1.4;
  font-variant-numeric: tabular-nums;
}
.card-time > span {
  white-space: nowrap;
}
.card-note {
  margin: 0;
  color: var(--el-text-color-regular);
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.collection-play {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  min-height: 32px;
  padding: 0 10px;
  border: 1px solid var(--el-color-primary-light-5);
  border-radius: 8px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font: inherit;
  font-size: 12px;
  cursor: pointer;
}
.collection-play:hover {
  background: var(--el-color-primary-light-8);
}
.collection-play:disabled,
.cover-button:disabled,
.collection-refresh:disabled {
  opacity: 0.45;
  cursor: default;
}
.unavailable {
  color: var(--el-color-warning);
  font-size: 11px;
}
.collection-empty {
  margin: 0 0 12px;
  padding: 18px 14px;
  border: 1px dashed var(--el-border-color);
  border-radius: 10px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.7;
}
@media (max-width: 480px) {
  .collection-card {
    gap: 9px;
    padding: 10px;
  }
  .cover-button {
    flex-basis: 84px;
    width: 84px;
    height: 58px;
  }
  .collection-play {
    padding: 0 8px;
  }
}
</style>
