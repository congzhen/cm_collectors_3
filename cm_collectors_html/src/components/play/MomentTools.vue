<template>
  <div class="moment-tools" @pointerdown.stop @click.stop @touchstart.stop @keydown.stop>
    <LoopPlaybackStatus
      v-if="loop"
      :root="activityRoot"
      :start="loop.start"
      :end="loop.end"
      :title="loop.title"
      :visible="entryVisible && !opened"
      :get-time="getTime"
      :get-duration="getDuration"
      :seek="seek"
    />
    <button
      ref="entryButton"
      class="moment-open"
      :class="{ 'moment-open--hidden': !entryVisible && !opened }"
      :tabindex="entryVisible || opened ? 0 : -1"
      type="button"
      @focus="showEntry"
      @blur="showEntry"
      @click="toggle"
    >
      {{ loop ? '片段循环中 · ' : '' }}片段 / 预览
    </button>
    <button
      v-if="loop"
      ref="exitButton"
      class="moment-open moment-exit"
      :class="{ 'moment-open--hidden': !entryVisible && !opened }"
      :tabindex="entryVisible || opened ? 0 : -1"
      type="button"
      @focus="showEntry"
      @blur="showEntry"
      title="退出片段播放"
      aria-label="退出片段播放"
      @click="exitLoop"
    >
      <svg
        width="18"
        height="18"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.8"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="M10 7V3H3v18h7v-4M8 12h13m-4-4 4 4-4 4" />
      </svg>
    </button>
    <Teleport to="body" :disabled="!props.pageOverlay">
      <div v-if="opened && props.pageOverlay" class="moment-backdrop" />
      <section
        v-if="opened"
        ref="panel"
        class="moment-panel"
        :class="{ 'moment-panel--page': props.pageOverlay }"
        aria-label="片段收藏与画面预览"
        @keydown.stop
      >
        <header class="moment-panel-header">
          <div><strong>精彩片段</strong><span class="moment-subtitle">重看喜欢的时刻</span></div>
          <div class="moment-header-actions">
            <button
              type="button"
              class="moment-manage"
              :aria-pressed="managing"
              @click="managing = !managing"
            >
              {{ managing ? '完成管理' : '管理' }}
            </button>
            <button
              type="button"
              class="moment-close"
              aria-label="关闭片段面板"
              title="关闭"
              @click="opened = false"
            >
              关闭
            </button>
          </div>
        </header>
        <div v-if="managing" class="moment-actions">
          <el-button size="small" @click="capture(false)">收藏当前时刻</el-button>
          <el-button size="small" @click="markStart">标记起点</el-button>
          <el-button size="small" :disabled="markedStart === undefined" @click="capture(true)"
            >标记终点</el-button
          >
          <el-button v-if="loop" size="small" @click="exitLoop">退出循环</el-button>
        </div>
        <p v-if="managing && markedStart !== undefined" class="moment-marked">
          已标记起点 {{ momentTime(markedStart) }}，播放到终点后再标记。
        </p>
        <el-form
          v-if="managing && draft"
          ref="momentForm"
          class="moment-form"
          :model="draft"
          :rules="momentRules"
          label-position="top"
          novalidate
          @submit.prevent="save"
        >
          <el-form-item label="名称" prop="title">
            <el-input v-model="draft.title" maxlength="200" placeholder="给片段起个名字" />
          </el-form-item>
          <el-form-item label="备注" prop="note">
            <el-input
              v-model="draft.note"
              type="textarea"
              maxlength="4000"
              :rows="2"
              placeholder="可选，记录喜欢这个片段的原因"
            />
          </el-form-item>
          <div class="moment-times">
            <el-form-item label="开始（秒）" prop="start">
              <el-input-number
                v-model="draft.start"
                :controls="false"
                :min="0"
                :step="0.1"
                :precision="1"
              />
            </el-form-item>
            <el-form-item label="结束（0为时间点）" prop="end">
              <el-input-number
                v-model="draft.end"
                :controls="false"
                :min="0"
                :step="0.1"
                :precision="1"
              />
            </el-form-item>
          </div>
          <div class="moment-form-actions">
            <el-button native-type="submit" type="primary" :loading="saving">保存收藏</el-button>
            <el-button @click="draft = undefined">取消</el-button>
          </div>
        </el-form>
        <p v-if="listError" class="moment-status" role="alert">{{ listError }}</p>
        <div v-if="!moments.length && !listError" class="moment-empty">
          <span class="moment-empty-icon" aria-hidden="true">▷</span>
          <strong>{{ loading ? '正在加载片段…' : '还没有收藏片段' }}</strong>
          <span v-if="!loading">点击右上角“管理”，收藏当前时刻或标记一段视频。</span>
        </div>
        <div class="moment-list">
          <article v-for="item in moments" :key="item.id" class="moment-item">
            <button
              class="moment-cover"
              type="button"
              :disabled="!item.available"
              :aria-label="`播放片段：${item.title}`"
              @click="jump(item)"
            >
              <MomentCover :moment-id="item.id" :time="item.start" :refresh-token="coverRefresh" />
              <span v-if="item.end > item.start" class="duration-badge">{{
                momentDuration(item.start, item.end).replace(/^共 /, '').replace(' 秒', 's')
              }}</span>
            </button>
            <div class="moment-item-info">
              <strong :title="item.title">{{ item.title }}</strong>
              <span class="moment-item-time"
                ><span
                  >{{ momentTime(item.start)
                  }}{{ item.end > 0 ? ' — ' + momentTime(item.end) : '' }}</span
                ></span
              >
              <span v-if="!item.available" class="moment-unavailable">源视频不可用</span>
            </div>
            <div class="moment-item-actions">
              <button
                class="moment-play"
                type="button"
                :disabled="!item.available"
                @click="jump(item)"
                :aria-label="`播放片段：${item.title}`"
              >
                <svg
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="currentColor"
                  aria-hidden="true"
                >
                  <path d="m7 4 14 8-14 8Z" /></svg
                ><span>播放</span>
              </button>
              <template v-if="managing">
                <button class="moment-edit" type="button" @click="draft = { ...item }">编辑</button>
                <button class="moment-delete" type="button" @click="remove(item)">删除</button>
              </template>
            </div>
          </article>
        </div>
        <el-button v-if="hasMore" :loading="loading" size="small" @click="loadMore"
          >加载更多</el-button
        >
        <VideoTimeline :manage="managing" :drama-series-id="dramaSeriesId" @seek="seekTo" />
        <p class="moment-hint">片段记录播放位置，源视频删除后无法播放。</p>
      </section>
    </Teleport>
  </div>
</template>
<script setup lang="ts">
import { onActivated, onDeactivated, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  ElForm,
  ElFormItem,
  ElInput,
  ElInputNumber,
  ElMessage,
  ElMessageBox,
  type FormInstance,
  type FormRules,
} from 'element-plus'
import {
  momentApi,
  momentTime,
  momentDuration,
  type VideoMoment,
  type MomentPlayback,
} from '@/server/videoMoments.server'
import VideoTimeline from './VideoTimeline.vue'
import MomentCover from './MomentCover.vue'
import LoopPlaybackStatus from './LoopPlaybackStatus.vue'
const props = defineProps<{
  initialMoment?: MomentPlayback
  dramaSeriesId: string
  pageOverlay?: boolean
  activityRoot?: HTMLElement
  getTime: () => number
  getDuration: () => number
  seek: (time: number) => void
  isPlaying?: () => boolean
  isEnded?: () => boolean
  play: () => unknown
}>()
const emit = defineEmits<{ 'loop-change': [range: MomentPlayback | undefined] }>()
const route = useRoute()
const managing = ref(false)
const entryVisible = ref(true)
const entryButton = ref<HTMLButtonElement>()
const exitButton = ref<HTMLButtonElement>()
const panel = ref<HTMLElement>()
function closeOnOutsideClick(event: MouseEvent) {
  if (!opened.value || inactive) return
  const path = event.composedPath()
  if ([props.activityRoot, panel.value, entryButton.value, exitButton.value].some(el => el && path.includes(el))) return
  // Confirmation dialogs are teleported outside the panel by Element Plus.
  if (path.some(el => el instanceof Element && el.matches('.el-message-box__wrapper, .el-message-box, .el-popper'))) return
  opened.value = false
}
let entryTimer: ReturnType<typeof setTimeout> | undefined
function showEntry() {
  clearTimeout(entryTimer)
  entryVisible.value = true
  if (inactive || opened.value) return
  entryTimer = setTimeout(() => {
    if (document.activeElement !== entryButton.value && document.activeElement !== exitButton.value)
      entryVisible.value = false
  }, 3000)
}
watch(
  () => props.activityRoot,
  (root, _, cleanup) => {
    if (!root) return
    const options = { capture: true, passive: true }
    const events = ['mousemove', 'pointerdown', 'touchstart', 'keydown'] as const
    events.forEach((event) => root.addEventListener(event, showEntry, options))
    cleanup(() => events.forEach((event) => root.removeEventListener(event, showEntry, true)))
  },
  { immediate: true },
)
const opened = ref(false),
  saving = ref(false),
  listError = ref('')
const markedStart = ref<number>()
const draft = ref<Partial<VideoMoment>>()
const momentForm = ref<FormInstance>()
const momentRules: FormRules = {
  title: [
    { required: true, whitespace: true, message: '请输入片段名称', trigger: 'blur' },
    { max: 200, message: '名称最多200字', trigger: 'blur' },
  ],
  note: [{ max: 4000, message: '备注最多4000字', trigger: 'blur' }],
  start: [
    {
      validator: (_rule, value, callback) => {
        const duration = props.getDuration()
        if (typeof value !== 'number' || !Number.isFinite(value) || value < 0)
          callback(new Error('请输入有效的开始时间'))
        else if (duration > 0 && value >= duration) callback(new Error('开始时间应小于视频时长'))
        else callback()
      },
      trigger: 'blur',
    },
  ],
  end: [
    {
      validator: (_rule, value, callback) => {
        const duration = props.getDuration()
        if (typeof value !== 'number' || !Number.isFinite(value) || value < 0)
          callback(new Error('请输入有效的结束时间'))
        else if (value > 0 && value <= (draft.value?.start ?? 0))
          callback(new Error('结束时间必须大于开始时间'))
        else if (duration > 0 && value > duration) callback(new Error('结束时间超出视频时长'))
        else callback()
      },
      trigger: 'blur',
    },
  ],
}
const moments = ref<VideoMoment[]>([])
const coverRefresh = ref(0)
const loading = ref(false),
  hasMore = ref(false)
let awaitingLoopSeek = false
const loop = ref<{ start: number; end: number; title?: string }>()
let loopNotice: ReturnType<typeof ElMessage> | undefined
function showLoopNotice() {
  loopNotice?.close()
  loopNotice = ElMessage.success({
    message: '已进入片段循环',
    duration: 2000,
    appendTo: props.activityRoot || document.body,
    customClass: props.activityRoot ? 'moment-loop-message' : '',
  })
}
watch(loop, (range) => {
  emit('loop-change', range ? { ...range, title: range.title || '' } : undefined)
  if (range) showLoopNotice()
  else {
    loopNotice?.close()
    loopNotice = undefined
  }
})
let page = 1,
  version = 0,
  contextVersion = 0,
  initialApplied = false
let timer: ReturnType<typeof setInterval>
async function load(append = false) {
  const v = ++version
  const requestedPage = append ? page + 1 : 1
  loading.value = true
  const result = await momentApi.list('', requestedPage, props.dramaSeriesId)
  if (v !== version) return
  loading.value = false
  listError.value = result.status ? '' : result.msg
  if (result.status) {
    page = requestedPage
    if (!append) coverRefresh.value++
    hasMore.value = result.data.length === 50
    moments.value = append ? [...moments.value, ...result.data] : result.data
  }
}
function loadMore() {
  if (loading.value) return
  void load(true)
}
watch(opened, showEntry)
function toggle() {
  opened.value = !opened.value
  if (opened.value) {
    managing.value = false
    void load()
  }
}
function markStart() {
  if (props.getDuration() > 0) markedStart.value = props.getTime()
  else ElMessage.warning('请等待视频加载')
}
function capture(range: boolean) {
  if (!(props.getDuration() > 0)) {
    ElMessage.warning('请等待视频加载')
    return
  }
  const start = range ? markedStart.value! : props.getTime(),
    end = range ? props.getTime() : 0
  if (range && end <= start) {
    ElMessage.warning('终点必须晚于起点')
    return
  }
  draft.value = {
    dramaSeriesId: props.dramaSeriesId,
    start: Math.floor(start * 10) / 10,
    end: Math.floor(end * 10) / 10,
    title: '',
    note: '',
  }
}
async function save() {
  if (!draft.value || saving.value) return
  const id = props.dramaSeriesId
  const context = contextVersion
  const submittedDraft = draft.value
  if (!momentForm.value) return
  saving.value = true
  const valid = await momentForm.value.validate().catch(() => false)
  if (context !== contextVersion) return
  if (!valid || draft.value !== submittedDraft || id !== props.dramaSeriesId) {
    saving.value = false
    return
  }
  const result = await momentApi.save({ ...submittedDraft })
  if (context !== contextVersion) return
  saving.value = false
  if (id !== props.dramaSeriesId) return
  if (!result.status) {
    ElMessage.error(result.msg)
    return
  }
  if (draft.value === submittedDraft) {
    draft.value = undefined
    markedStart.value = undefined
  }
  await load()
  if (context === contextVersion) ElMessage.success('片段已保存')
}
async function remove(item: VideoMoment) {
  const context = contextVersion
  try {
    await ElMessageBox.confirm('删除这条片段收藏？源视频不会被删除。', '删除收藏', {
      type: 'warning',
    })
  } catch {
    return
  }
  const result = await momentApi.remove(item.id)
  if (context !== contextVersion) return
  if (!result.status) ElMessage.error(result.msg)
  else {
    await load()
  }
}
function exitLoop() {
  loop.value = undefined
  awaitingLoopSeek = false
  showEntry()
}
function seekTo(time: number) {
  loop.value = undefined
  props.seek(time)
}
function canPlayMoment(start: number, end: number, duration: number) {
  if (!(duration > 0) || !Number.isFinite(duration)) {
    ElMessage.warning('请等待视频加载')
    return false
  }
  if (
    !Number.isFinite(start) ||
    !Number.isFinite(end) ||
    start < 0 ||
    end < 0 ||
    start >= duration ||
    (end > 0 && (end <= start || end > duration))
  ) {
    ElMessage.warning('收藏时间超出当前视频范围，请在管理中调整片段')
    return false
  }
  return true
}
function jump(item: VideoMoment) {
  if (!canPlayMoment(item.start, item.end, props.getDuration())) return
  awaitingLoopSeek = item.end > item.start
  loop.value =
    item.end > item.start ? { start: item.start, end: item.end, title: item.title } : undefined
  props.seek(item.start)
  void props.play()
  opened.value = false
}
watch(
  () => props.dramaSeriesId,
  () => {
    version++
    contextVersion++
    saving.value = false
    page = 1
    hasMore.value = false
    loading.value = false
    listError.value = ''
    opened.value = false
    managing.value = false
    draft.value = undefined
    moments.value = []
    loop.value = undefined
    markedStart.value = undefined
    initialApplied = false
    awaitingLoopSeek = false
  },
)
watch(
  () =>
    props.initialMoment
      ? [props.initialMoment]
      : [route?.params.dramaSeriesId, route?.query.t, route?.query.end, route?.query.momentTitle],
  () => {
    initialApplied = false
    awaitingLoopSeek = false
    loop.value = undefined
  },
)
let inactive = false
onDeactivated(() => {
  contextVersion++
  version++
  saving.value = false
  inactive = true
  loopNotice?.close()
  clearTimeout(entryTimer)
  opened.value = false
})
onActivated(() => {
  inactive = false
  showEntry()
})
onMounted(() => {
  document.addEventListener('click', closeOnOutsideClick, true)
  showEntry()
  timer = setInterval(() => {
    if (inactive) return
    const duration = props.getDuration()
    if (!initialApplied && duration > 0 && Number.isFinite(duration)) {
      initialApplied = true
      const fromRoute =
        route?.params.dramaSeriesId === props.dramaSeriesId && route.query.t !== undefined
      if (props.initialMoment || fromRoute) {
        const start = props.initialMoment?.start ?? Number(route.query.t),
          end =
            props.initialMoment?.end ??
            (route.query.end === undefined ? 0 : Number(route.query.end))
        if (canPlayMoment(start, end, duration)) {
          awaitingLoopSeek = true
          props.seek(start)
          if (Number.isFinite(end) && end > start && end <= duration)
            loop.value = {
              start,
              end,
              title:
                props.initialMoment?.title ??
                (typeof route.query.momentTitle === 'string' ? route.query.momentTitle : undefined),
            }
          if (props.initialMoment && props.initialMoment.autoplay !== false) void props.play()
        }
      }
    }
    const range = loop.value
    if (!range) {
      awaitingLoopSeek = false
      return
    }
    const time = props.getTime()
    if (time < range.start) {
      // Keyboard and mobile gestures can seek outside the custom progress bar.
      // Keep the loop bounded without resuming a video the user paused.
      if (!awaitingLoopSeek) {
        awaitingLoopSeek = true
        props.seek(range.start)
      }
      return
    }
    if (time < range.end) awaitingLoopSeek = false
    const playing = props.isPlaying?.() ?? true
    const endedAtRangeEnd = props.isEnded?.() && range.end >= duration - 0.1
    if (duration > 0 && time >= range.end && !awaitingLoopSeek && (playing || endedAtRangeEnd)) {
      awaitingLoopSeek = true
      props.seek(range.start)
      void props.play()
      showLoopNotice()
    }
  }, 150)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', closeOnOutsideClick, true)
  emit('loop-change', undefined)
  version++
  contextVersion++
  loopNotice?.close()
  clearInterval(timer)
  clearTimeout(entryTimer)
})
</script>
<style scoped>
:global(.moment-loop-message) {
  position: absolute;
  max-width: calc(100% - 24px);
  pointer-events: none;
}
.moment-tools {
  position: absolute;
  inset: 10px;
  pointer-events: none;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  align-items: flex-start;
  z-index: 24;
  font: 13px/1.5 sans-serif;
  color: #e9edf5;
}
.moment-open {
  opacity: 1;
  transition: opacity 0.2s ease;
  pointer-events: auto;
  border: 1px solid #ffffff40;
  background: #17202dcc;
  color: white;
  border-radius: 18px;
  padding: 7px 12px;
  cursor: pointer;
}
.moment-exit {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 7px;
  border-color: #ffffff60;
}
.moment-open--hidden {
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
}
.moment-panel {
  color: var(--el-text-color-primary);
  font: 13px/1.5 sans-serif;
  pointer-events: auto;
  position: absolute;
  right: 0;
  top: 40px;
  width: min(530px, calc(100vw - 40px));
  max-width: 100%;
  max-height: calc(100% - 40px);
  overflow: auto;
  background: var(--el-bg-color-overlay);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 18px;
  padding: 20px;
  box-shadow: var(--el-box-shadow-dark);
  box-sizing: border-box;
}
.moment-panel-header,
.moment-header-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}
.moment-panel-header {
  margin-bottom: 20px;
}
.moment-panel-header strong {
  display: block;
  font-size: 17px;
  font-weight: 600;
  letter-spacing: 0.02em;
}
.moment-subtitle {
  display: block;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  margin-top: 3px;
}
.moment-panel button {
  font-family: inherit;
  cursor: pointer;
}
.moment-manage,
.moment-close {
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  padding: 7px 11px;
  font-size: 12px;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-regular);
}
.moment-manage[aria-pressed='true'] {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  border-color: var(--el-color-primary-light-5);
}
.moment-close {
  border-color: transparent;
  color: var(--el-text-color-secondary);
}
.moment-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  padding: 12px;
  margin-bottom: 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-fill-color-light);
}
.moment-actions .el-button {
  margin: 0;
  min-height: 32px;
  border-radius: 6px;
}
.moment-form {
  margin: 14px 0;
  padding: 14px;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
}
.moment-form :deep(.el-form-item) {
  margin-bottom: 18px;
}
.moment-form :deep(.el-form-item__label) {
  color: var(--el-text-color-regular);
  font-size: 12px;
  line-height: 20px;
}
.moment-form :deep(.el-input-number) {
  width: 100%;
}
.moment-form :deep(.el-input-number .el-input__inner) {
  text-align: left;
}
.moment-form-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.moment-panel .el-button {
  height: 32px;
  min-height: 32px;
  padding: 0 12px;
  margin-left: 0;
  border-radius: 8px;
  font-size: 12px;
}
.moment-times {
  display: flex;
  gap: 12px;
}
.moment-times :deep(.el-form-item) {
  flex: 1;
  min-width: 0;
}
.moment-list {
  display: grid;
  gap: 9px;
}
.moment-item {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  padding: 11px;
  background: var(--el-fill-color-blank);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
}
.moment-item:hover {
  background: var(--el-fill-color-light);
  border-color: var(--el-border-color);
}
.moment-cover {
  position: relative;
  flex: 0 0 84px;
  width: 84px;
  height: 52px;
  display: grid;
  place-items: center;
  padding: 0;
  border: 0;
  border-radius: 7px;
  overflow: hidden;
  background: var(--el-fill-color-darker);
  color: var(--el-text-color-secondary);
}
.moment-cover :deep(img),
.moment-cover :deep(.cover-placeholder) {
  width: 100%;
  height: 100%;
  object-fit: cover;
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
.moment-item-info {
  flex: 1;
  min-width: 80px;
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.moment-item-info strong {
  font-size: 14px;
  line-height: 1.4;
  overflow-wrap: anywhere;
  color: var(--el-text-color-primary);
}
.moment-item-time {
  display: flex;
  flex-wrap: wrap;
  gap: 2px 6px;
  color: var(--el-text-color-secondary);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}
.moment-item-time > span {
  white-space: nowrap;
}
.moment-item-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-left: auto;
}
.moment-play {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  padding: 8px 11px;
  border: 1px solid var(--el-color-primary-light-5);
  border-radius: 8px;
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  white-space: nowrap;
  font-size: 12px;
}
.moment-play:hover {
  background: var(--el-color-primary-light-8);
}
.moment-edit,
.moment-delete {
  padding: 8px;
  border: 0;
  border-radius: 7px;
  color: var(--el-text-color-regular);
  background: transparent;
  font-size: 12px;
}
.moment-edit:hover {
  background: var(--el-fill-color-light);
  color: var(--el-text-color-primary);
}
.moment-delete {
  color: var(--el-color-danger);
}
.moment-delete:hover {
  background: var(--el-color-danger-light-9);
}
.moment-unavailable {
  color: var(--el-color-warning);
  font-size: 11px;
}
.moment-panel button:disabled {
  cursor: default;
  opacity: 0.4;
}
.moment-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 26px 12px;
  text-align: center;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
  border: 1px dashed var(--el-border-color-lighter);
  border-radius: 12px;
}
.moment-empty strong {
  color: var(--el-text-color-regular);
  font-weight: 500;
}
.moment-empty span:last-child {
  font-size: 12px;
}
.moment-empty-icon {
  font-size: 28px;
  line-height: 1.2;
  color: var(--el-color-primary);
}
.moment-status,
.moment-marked {
  padding: 10px 12px;
  border-radius: 8px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-size: 12px;
}
.moment-hint {
  margin: 14px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 11px;
}
.moment-panel :deep(.video-timeline) {
  border-top: 1px solid var(--el-border-color-lighter);
  margin-top: 18px;
  padding-top: 16px;
}
.moment-panel :deep(.timeline-heading) {
  gap: 8px;
}
.moment-panel :deep(.timeline-heading strong) {
  font-size: 13px;
  margin-right: auto;
}
.moment-panel :deep(.timeline-heading select) {
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-blank);
  border: 1px solid var(--el-border-color);
}
.moment-panel :deep(.timeline-frames) {
  scrollbar-width: thin;
  scrollbar-color: var(--el-border-color-darker) transparent;
  gap: 10px;
}
.moment-panel :deep(.timeline-frames button) {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 9px;
  background: var(--el-bg-color);
}
.moment-panel :deep(.timeline-frames span) {
  padding: 6px;
  color: var(--el-text-color-regular);
  font-variant-numeric: tabular-nums;
}
@media (max-width: 540px) {
  .moment-panel {
    padding: 14px;
  }
  .moment-cover {
    flex-basis: 64px;
    width: 64px;
    height: 44px;
  }
  .moment-item {
    gap: 8px;
    padding: 9px;
  }
  .moment-item-actions {
    flex-wrap: wrap;
  }
}
.moment-backdrop {
  position: fixed;
  inset: 0;
  background: #0008;
  z-index: 3100;
  pointer-events: none;
}
.moment-panel--page {
  position: fixed;
  left: 12px;
  right: 12px;
  top: 12vh;
  width: auto;
  max-width: none;
  max-height: 76vh;
  z-index: 3101;
}
</style>
