<template>
  <section class="video-timeline">
    <div class="timeline-heading">
      <strong>画面时间轴</strong>
      <el-button size="small" @click="reloadFrames">刷新状态</el-button>
      <select
        v-if="manage"
        v-model="interval"
        aria-label="预览图间隔"
        @change="intervalEdited = true"
      >
        <option :value="10">每10秒</option>
        <option :value="30">每30秒</option>
        <option :value="60">每60秒</option>
        <option :value="120">每120秒</option>
      </select>
      <el-button
        size="small"
        v-if="manage"
        :loading="busy"
        :disabled="pending || preview?.status === 'ready'"
        @click="generate"
        >{{
          preview?.status === 'missing'
            ? '重新生成'
            : preview?.status === 'failed'
              ? '重试生成'
              : '生成预览'
        }}</el-button
      >
      <el-button
        v-if="manage && preview && preview.status !== 'none'"
        size="small"
        type="danger"
        plain
        @click="deletePreview"
        >删除预览</el-button
      >
    </div>
    <p v-if="manage && preview?.status === 'ready'">
      已有预览会保留。调整密度或替换视频后，需要手动删除预览再生成。
    </p>
    <p v-if="error" role="status">{{ error }}</p>
    <p v-else-if="pending">
      {{
        preview?.status === 'queued' ? '等待后台处理（队列暂停时可在片段收藏页恢复）' : '正在生成'
      }}
      · {{ preview?.completed }} / {{ preview?.count || '…' }}
    </p>
    <p v-else-if="preview?.status === 'missing'" role="status">
      {{
        manage
          ? '时间轴缓存已清理或不完整，点击重新生成即可恢复。'
          : '时间轴缓存已清理，可在管理中重新生成。'
      }}
    </p>
    <p v-else-if="preview?.status === 'failed'" role="alert">{{ preview.error }}</p>
    <p v-else-if="preview?.status === 'stale'">视频已改变，请重新生成预览。</p>
    <p v-else-if="!preview?.completed">
      {{
        manage
          ? '按需生成预览图，不会修改源视频。较长视频会适当加大间隔。'
          : '暂无画面预览，可在管理中生成。'
      }}
    </p>
    <div v-if="preview && preview.status !== 'stale'" class="timeline-frames">
      <button
        v-for="i in preview.completed"
        :key="`${preview.sourceKey}-${preview.interval}-${i}-${frameReload}`"
        type="button"
        @click="emit('seek', (i - 1) * preview.interval)"
      >
        <img
          @error="onFrameError"
          @load="onFrameLoad"
          loading="lazy"
          :src="previewImage(preview, (i - 1) * preview.interval)"
          :alt="`跳转到 ${momentTime((i - 1) * preview.interval)}`"
        />
        <span>{{ momentTime((i - 1) * preview.interval) }}</span>
      </button>
    </div>
  </section>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, ref, toRef, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/assets/request'
import {
  momentApi,
  momentTime,
  previewImage,
  notifyPreviewChanged,
} from '@/server/videoMoments.server'
import { useVideoPreview } from './useVideoPreview'
const props = withDefaults(defineProps<{ dramaSeriesId: string; manage?: boolean }>(), {
  manage: true,
})
const emit = defineEmits<{ seek: [time: number] }>()
const { preview, error, refresh } = useVideoPreview(toRef(props, 'dramaSeriesId'))
const interval = ref(30),
  busy = ref(false)
const intervalEdited = ref(false)
let sourceVersion = 0
onBeforeUnmount(() => sourceVersion++)
watch(
  () => props.dramaSeriesId,
  () => {
    sourceVersion++
    busy.value = false
    intervalEdited.value = false
    interval.value = 30
  },
  { flush: 'sync' },
)
watch(preview, (value) => {
  if (!intervalEdited.value && value && [10, 30, 60, 120].includes(value.interval))
    interval.value = value.interval
})
const pending = computed(() => ['queued', 'processing'].includes(preview.value?.status || ''))
const frameReload = ref(0)
function reloadFrames() {
  frameReload.value++
  void refresh()
}
function onFrameLoad(event: Event) {
  ;(event.target as HTMLImageElement).style.visibility = ''
}
let frameRefresh: { version: number; promise: Promise<void> } | undefined
async function onFrameError(event: Event) {
  const img = event.target as HTMLImageElement
  img.style.visibility = 'hidden'
  if (img.dataset.retried) return
  img.dataset.retried = 'true'
  const targetId = props.dramaSeriesId
  const version = sourceVersion
  const source = img.getAttribute('src') || ''
  if (!frameRefresh || frameRefresh.version !== version) {
    const entry = { version, promise: Promise.resolve() }
    entry.promise = refresh().finally(() => {
      if (frameRefresh === entry) frameRefresh = undefined
    })
    frameRefresh = entry
  }
  await frameRefresh.promise
  // A transient image request failure must not leave a valid frame hidden forever.
  // Retry once; physical cache loss is handled by the missing-preview response.
  if (
    targetId === props.dramaSeriesId &&
    version === sourceVersion &&
    preview.value?.status === 'ready' &&
    source &&
    img.getAttribute('src') === source
  )
    img.src = `${source}&retry=1`
}
async function generate() {
  if (busy.value || pending.value || preview.value?.status === 'ready') return
  const version = sourceVersion
  const targetId = props.dramaSeriesId
  busy.value = true
  try {
    const result = await momentApi.queue(targetId, interval.value)
    if (version !== sourceVersion) return
    if (!result.status) ElMessage.error(result.msg)
    else await refresh()
  } finally {
    if (version === sourceVersion) busy.value = false
  }
}
async function deletePreview() {
  const targetId = props.dramaSeriesId
  try {
    await ElMessageBox.confirm(
      '删除这个视频的预览图片？片段收藏、首图和源视频会保留。',
      '删除预览',
      {
        type: 'warning',
      },
    )
  } catch {
    return
  }
  const result = await request({
    url: `/videoPreviews/${encodeURIComponent(targetId)}`,
    method: 'delete',
  })
  if (!result.status) ElMessage.error(result.msg)
  else notifyPreviewChanged(targetId)
}
</script>
<style scoped>
.video-timeline {
  margin: 12px 0;
}
.timeline-heading {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.timeline-heading .el-button,
.timeline-heading select {
  box-sizing: border-box;
  height: 32px;
  min-height: 32px;
  margin-left: 0;
  padding: 0 12px;
  border-radius: 8px;
  font-family: inherit;
  font-size: 12px;
  line-height: 1;
}
.timeline-heading select {
  border: 1px solid var(--el-border-color);
}
.video-timeline p {
  font-size: 12px;
  opacity: 0.7;
  line-height: 1.6;
}
.timeline-frames {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding: 12px 0;
}
.timeline-frames button {
  position: relative;
  flex: 0 0 144px;
  padding: 0;
  border: 0;
  border-radius: 6px;
  overflow: hidden;
  background: #171b23;
  color: white;
  cursor: pointer;
}
.timeline-frames img {
  width: 144px;
  height: 82px;
  object-fit: contain;
  display: block;
}
.timeline-frames span {
  display: block;
  font-size: 11px;
  padding: 4px;
}
</style>
