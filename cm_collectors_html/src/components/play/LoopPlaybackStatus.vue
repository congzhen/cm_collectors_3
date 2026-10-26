<template>
  <Teleport v-if="progressTarget && rangeStyle" :to="progressTarget">
    <div v-if="!seek" class="loop-range" :style="rangeStyle" aria-hidden="true"><i /><i /></div>
    <div
      v-else-if="!customControls"
      class="loop-progress"
      @pointerdown.stop
      @mousedown.stop
      @touchstart.stop
      @touchmove.stop
      @pointerup.stop
      @mouseup.stop
      @touchend.stop
      @click.stop
      @keydown.stop
    >
      <div class="loop-progress-fill" :style="{ width: progressPercent + '%' }" />
      <input
        class="loop-seek"
        type="range"
        min="0"
        :max="end - start"
        step="0.1"
        :value="relativeTime"
        aria-label="片段播放进度"
        @input="seekWithinClip"
      />
    </div>
  </Teleport>
  <Teleport v-if="timeTarget" :to="timeTarget">
    <span
      class="loop-status"
      :class="{ 'loop-status--overlay': timeTarget?.matches('.xgplayer-controls') }"
      :title="description"
      ><span class="loop-name">↻ {{ title || '片段循环' }}</span
      ><span v-if="!customControls || !seek" class="loop-time">{{ elapsed }} / {{ length }}</span></span
    >
  </Teleport>
  <span v-else-if="visible" class="loop-status loop-status--fallback" :title="description"
    >↻ {{ title || '片段循环' }} · {{ elapsed }} / {{ length }}</span
  >
</template>
<script setup lang="ts">
import {
  computed,
  onActivated,
  onDeactivated,
  onBeforeUnmount,
  onMounted,
  shallowRef,
  ref,
} from 'vue'
import { momentTime } from '@/server/videoMoments.server'
const props = defineProps<{
  root?: HTMLElement
  start: number
  end: number
  title?: string
  visible: boolean
  getTime: () => number
  getDuration: () => number
  seek?: (time: number) => void
}>()
const progressTarget = shallowRef<HTMLElement | null>(null)
const timeTarget = shallowRef<HTMLElement | null>(null)
const customControls = ref(false)
const time = ref(0),
  duration = ref(0)
function update() {
  time.value = props.getTime()
  duration.value = props.getDuration()
  // Teleport into each player's controls so their own idle/fullscreen behavior applies.
  // video.js keeps its hidden controls before our custom controls in DOM order.
  // Query each player in priority order, rather than using a selector list.
  const root = props.root
  const custom = root?.querySelector<HTMLElement>('.video-controller')
  customControls.value = !!custom
  const mobile = root?.querySelector<HTMLElement>('.xgplayer-controls')
  progressTarget.value = custom
    ? custom.querySelector<HTMLElement>('.progress-slider-container')
    : mobile
      ? (root?.querySelector<HTMLElement>('.xgplayer-progress-outer') ?? null)
      : (root?.querySelector<HTMLElement>(
          '.video-js:not(.vjs-controls-disabled) .vjs-progress-holder',
        ) ?? null)
  timeTarget.value = custom
    ? custom.querySelector<HTMLElement>('.controls-left')
    : (mobile ??
      root?.querySelector<HTMLElement>('.video-js:not(.vjs-controls-disabled) .vjs-control-bar') ??
      null)
}
const rangeStyle = computed(() => {
  if (!Number.isFinite(duration.value) || duration.value <= 0) return undefined
  const start = Math.max(0, Math.min(1, props.start / duration.value))
  const end = Math.max(start, Math.min(1, props.end / duration.value))
  return { left: `${start * 100}%`, width: `${(end - start) * 100}%` }
})
const compactTime = (seconds: number) => momentTime(seconds).replace(/^00:/, '')
const relativeTime = computed(() =>
  Math.max(0, Math.min(props.end - props.start, time.value - props.start)),
)
const progressPercent = computed(() => (relativeTime.value / (props.end - props.start)) * 100)
function seekWithinClip(event: Event) {
  const value = Number((event.target as HTMLInputElement).value)
  if (!Number.isFinite(value)) return
  const target = props.start + Math.max(0, Math.min(props.end - props.start, value))
  time.value = target
  props.seek?.(target)
}
const length = computed(() => compactTime(props.end - props.start))
const elapsed = computed(() =>
  compactTime(Math.max(0, Math.min(props.end - props.start, time.value - props.start))),
)
const description = computed(
  () =>
    `片段循环：${props.title || '精彩片段'} · ${momentTime(props.start)} — ${momentTime(props.end)}`,
)
let timer: ReturnType<typeof setInterval>
function startUpdates() {
  clearInterval(timer)
  update()
  timer = setInterval(update, 150)
}
onMounted(startUpdates)
onActivated(startUpdates)
onDeactivated(() => clearInterval(timer))
onBeforeUnmount(() => clearInterval(timer))
</script>
<style scoped>
/* Replace the third-party progress control while looping, preserving its idle/fullscreen container. */
:global(.xgplayer-progress-outer:has(> .loop-progress) > :not(.loop-progress)),
:global(.vjs-progress-holder:has(> .loop-progress) > :not(.loop-progress)) {
  visibility: hidden;
}
:global(.xgplayer:has(.loop-progress) .xgplayer-time),
:global(.video-js:has(.loop-progress) .vjs-current-time),
:global(.video-js:has(.loop-progress) .vjs-duration),
:global(.video-js:has(.loop-progress) .vjs-remaining-time),
:global(.video-js:has(.loop-progress) .vjs-time-divider) {
  display: none !important;
}
.loop-progress {
  position: absolute;
  inset: 0;
  z-index: 5;
  background: #ffffff50;
  border-radius: 4px;
}
.loop-progress-fill {
  position: absolute;
  inset: 0 auto 0 0;
  background: #66b1ff;
  border-radius: inherit;
  pointer-events: none;
}
.loop-seek {
  appearance: none;
  position: absolute;
  inset: -8px 0;
  margin: 0;
  width: 100%;
  height: calc(100% + 16px);
  background: transparent;
  cursor: pointer;
}
.loop-seek::-webkit-slider-runnable-track {
  height: 4px;
  background: transparent;
}
.loop-seek::-webkit-slider-thumb {
  appearance: none;
  width: 12px;
  height: 12px;
  margin-top: -4px;
  background: white;
  border-radius: 50%;
}
.loop-seek::-moz-range-track {
  height: 4px;
  background: transparent;
}
.loop-seek::-moz-range-thumb {
  width: 12px;
  height: 12px;
  border: 0;
  background: white;
  border-radius: 50%;
}
.loop-seek:focus-visible {
  outline: 2px solid #66b1ff;
  outline-offset: 2px;
}
.loop-range {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  height: 3px;
  background: #66b1ff;
  border-radius: 3px;
  z-index: 1;
  pointer-events: none;
}
.loop-range i {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: inherit;
}
.loop-range i:first-child {
  left: -1px;
}
.loop-range i:last-child {
  right: -1px;
}
.loop-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-left: 6px;
  max-width: min(240px, 28vw);
  min-width: 0;
  overflow: hidden;
  font: 11px/1.4 sans-serif;
  color: #a8d4ff;
  white-space: nowrap;
  pointer-events: auto;
}
.loop-name {
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}
.loop-time {
  flex-shrink: 0;
}
.loop-status--fallback {
  position: absolute;
  left: 8px;
  bottom: 62px;
  max-width: calc(100% - 32px);
  padding: 5px 8px;
  border-radius: 5px;
  background: #111c;
  overflow: hidden;
  text-overflow: ellipsis;
}
.loop-status.loop-status--overlay {
  position: absolute;
  left: 12px;
  bottom: calc(100% + 10px);
  flex-direction: row;
  gap: 6px;
  max-width: calc(100% - 32px);
  margin: 0;
  padding: 3px 6px;
  border-radius: 4px;
  background: #111b;
  height: auto;
  line-height: 1.4;
}
@media (max-width: 600px) {
  .loop-status {
    flex-direction: column;
    align-items: flex-start;
    gap: 0;
    max-width: 22vw;
    font-size: 10px;
  }
  .loop-name {
    max-width: 100%;
  }
  .loop-status--fallback {
    display: block;
    max-width: calc(100% - 32px);
  }
}
</style>
