<template>
  <div v-if="visible && src" class="mobile-seek-preview">
    <img :src="src" alt="跳转位置预览" /><span>{{
      momentTime(time - (getRange?.()?.start || 0))
    }}</span>
  </div>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, toRef, ref, watch } from 'vue'
import { momentTime, previewImage, type MomentPlayback } from '@/server/videoMoments.server'
import { useVideoPreview } from './useVideoPreview'
const props = defineProps<{
  dramaSeriesId: string
  root: HTMLElement
  getDuration: () => number
  getRotation: () => number
  getRange?: () => MomentPlayback | undefined
}>()
const { preview, refresh } = useVideoPreview(toRef(props, 'dramaSeriesId'))
const visible = ref(false),
  time = ref(0)
const src = computed(() => (preview.value ? previewImage(preview.value, time.value) : ''))
let progress: HTMLElement | undefined
function move(e: TouchEvent | MouseEvent) {
  if (!progress) return
  const point = 'touches' in e ? e.touches[0] : e
  if (!point) return
  const rect = progress.getBoundingClientRect()
  const rotated = props.getRotation() === 90
  const fraction = rotated
    ? (point.clientY - rect.top) / rect.height
    : (point.clientX - rect.left) / rect.width
  const range = props.getRange?.()
  time.value =
    (range?.start ?? 0) +
    Math.max(0, Math.min(1, fraction)) * (range ? range.end - range.start : props.getDuration())
  visible.value = true
}
function start(e: TouchEvent | MouseEvent) {
  progress = (e.target as HTMLElement)?.closest<HTMLElement>('.xgplayer-progress') || undefined
  if (progress) {
    void refresh()
    move(e)
  }
}
function end() {
  progress = undefined
  visible.value = false
}
watch(
  () => props.root,
  (root, _, cleanup) => {
    root.addEventListener('touchstart', start, { capture: true, passive: true })
    root.addEventListener('mousedown', start, true)
    document.addEventListener('touchmove', move, { capture: true, passive: true })
    document.addEventListener('mousemove', move, true)
    document.addEventListener('touchend', end, true)
    document.addEventListener('touchcancel', end, true)
    document.addEventListener('mouseup', end, true)
    cleanup(() => {
      root.removeEventListener('touchstart', start, true)
      root.removeEventListener('mousedown', start, true)
      document.removeEventListener('touchmove', move, true)
      document.removeEventListener('mousemove', move, true)
      document.removeEventListener('touchend', end, true)
      document.removeEventListener('touchcancel', end, true)
      document.removeEventListener('mouseup', end, true)
      end()
    })
  },
  { immediate: true },
)
onBeforeUnmount(end)
</script>
<style scoped>
.mobile-seek-preview {
  position: absolute;
  bottom: 70px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 30;
  pointer-events: none;
  background: #111d;
  border-radius: 8px;
  overflow: hidden;
  color: white;
  text-align: center;
  font-size: 12px;
}
.mobile-seek-preview img {
  display: block;
  width: 160px;
  height: 90px;
  object-fit: contain;
}
.mobile-seek-preview span {
  display: block;
  padding: 4px;
}
</style>
