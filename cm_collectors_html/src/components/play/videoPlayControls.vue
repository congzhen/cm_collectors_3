<template>
  <div class="video-controller">
    <!-- 播放进度条 -->
    <div class="progress-container">
      <div class="progress-slider-container" @click="onProgressClick">
        <input type="range" min="0" :max="duration || 100" :value="currentTime" class="progress-slider"
          @input="onSeek" />
        <div class="progress-filled" :style="{ width: progressPercent + '%' }"></div>
      </div>
    </div>

    <div class="controls-row">
      <!-- 左侧控件 -->
      <div class="controls-left">
        <!-- 播放速度 -->
        <div class="speed-control">
          <button class="control-button" @click="toggleSpeedMenu" :title="`播放速度: ${playbackRate}x`">
            {{ playbackRate }}x
          </button>
          <div v-if="showSpeedMenu" class="speed-menu">
            <button v-for="rate in speedOptions" :key="rate" :class="{ active: playbackRate === rate }"
              @click="setPlaybackRate(rate)">
              {{ rate }}x
            </button>
          </div>
        </div>

        <!-- 播放时间 -->
        <div class="time-display">
          {{ formatTime(currentTime) }} / {{ formatTime(duration) }}
        </div>
      </div>

      <!-- 中间控件 -->
      <div class="controls-center">
        <!-- 播放/暂停按钮 -->
        <button class="control-button" @click="togglePlay" :title="isPlaying ? '暂停' : '播放'">
          <svg v-if="isPlaying" class="icon" viewBox="0 0 24 24" width="24" height="24">
            <path fill="currentColor" d="M6 4h4v16H6V4zm8 0h4v16h-4V4z" />
          </svg>
          <svg v-else class="icon" viewBox="0 0 24 24" width="24" height="24">
            <path fill="currentColor" d="M8 5v14l11-7z" />
          </svg>
        </button>
      </div>

      <!-- 右侧控件 -->
      <div class="controls-right">
        <!-- 音量控制 -->
        <div class="volume-control">
          <button class="control-button" @click="toggleMute" :title="isMuted ? '取消静音' : '静音'">
            <svg v-if="isMuted || volume === 0" class="icon" viewBox="0 0 24 24" width="24" height="24">
              <path fill="currentColor"
                d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51C20.63 14.91 21 13.5 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3L3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.38-.31 2.63-.95 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4L9.91 6.09 12 8.18V4z" />
            </svg>
            <svg v-else-if="volume < 0.5" class="icon" viewBox="0 0 24 24" width="24" height="24">
              <path fill="currentColor"
                d="M18.5 12c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM5 9v6h4l5 5V4l-5 5H5z" />
            </svg>
            <svg v-else class="icon" viewBox="0 0 24 24" width="24" height="24">
              <path fill="currentColor"
                d="M3 9v6h4l5 5V4l-5 5H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z" />
            </svg>
          </button>
          <div class="volume-slider-container">
            <input type="range" min="0" max="1" step="0.01" v-model="volume" class="volume-slider"
              @input="onVolumeChange" />
          </div>
        </div>

        <!-- 旋转按钮 -->
        <button class="control-button" @click="rotateVideo" :title="`旋转: ${rotation}°`">
          <svg class="icon" viewBox="0 0 24 24" width="24" height="24">
            <path fill="currentColor"
              d="M12 21c-4.97 0-9-4.03-9-9 0-1.94.62-3.72 1.65-5.16L2.51 4.7A11.94 11.94 0 0 0 0 12c0 6.63 5.37 12 12 12s12-5.37 12-12-5.37-12-12-12c-1.53 0-3 .29-4.37.82l1.23 1.23C10.1 1.62 11.03 1.38 12 1.38c5.77 0 10.62 4.85 10.62 10.62S17.77 22.62 12 22.62 1.38 17.77 1.38 12H3c0 4.97 4.03 9 9 9z" />
            <path fill="currentColor" d="M12 6v4.38l3.54 3.54 1.41-1.41L13.41 9H12V6z" />
          </svg>
        </button>

        <!-- 本地云播放按钮 -->
        <button class="control-button" @click="openCloudPlayer" title="本地云播放">
          <svg class="icon" viewBox="0 0 24 24" width="24" height="24">
            <path fill="currentColor"
              d="M19.35 10.04A7.49 7.49 0 0 0 12 4C9.11 4 6.6 5.64 5.35 8.04A5.994 5.994 0 0 0 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96zM10 15.5v-7l6 3.5-6 3.5z" />
          </svg>
        </button>


        <!-- 用本地播放器打开视频按钮 -->
        <button class="control-button" @click="openInPlayer" title="本地播放器打开">
          <svg class="icon" viewBox="0 0 24 24" width="24" height="24">
            <path fill="currentColor"
              d="M19 4H5c-1.11 0-2 .9-2 2v12c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V6c0-1.1-.89-2-2-2zm0 14H5V6h14v12z" />
            <path fill="currentColor" d="M8 9.5v5l4-2.5-4-2.5z" />
            <path fill="currentColor"
              d="M16.5 12c0-.28-.22-.5-.5-.5s-.5.22-.5.5.22.5.5.5.5-.22.5-.5zm1.5 0c0 .55-.45 1-1 1s-1-.45-1-1 .45-1 1-1 1 .45 1 1zm-1 3.5c-1.66 0-3-1.34-3-3s1.34-3 3-3 3 1.34 3 3-1.34 3-3 3z" />
          </svg>
        </button>

        <!-- 画中画按钮 -->
        <button class="control-button" @click="togglePictureInPicture" :title="isPictureInPicture ? '退出画中画' : '画中画'">
          <svg class="icon" viewBox="0 0 24 24" width="24" height="24">
            <path fill="currentColor"
              d="M19 7h-8v6h8V7zm2-4H3c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h18c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H3V5h18v14z" />
          </svg>
        </button>

        <!-- 最大化/恢复按钮 -->
        <button class="control-button" @click="toggleMaximize" :title="isMaximized ? '恢复' : '最大化'">
          <svg v-if="!isMaximized" class="icon" viewBox="0 0 24 24" width="24" height="24">
            <path fill="currentColor" d="M4 4h16v16H4V4m2 2v12h12V6H6zm4 4h4v2h2v4h-4v-2h-2v-4z" />
          </svg>
          <svg v-else class="icon" viewBox="0 0 24 24" width="24" height="24">
            <path fill="currentColor" d="M4 4h16v16H4V4m2 2v12h12V6H6zm4 4v2h2v2h2v-2h2v-2h-2v-2h-2v2h-2z" />
          </svg>
        </button>

        <!-- 全屏按钮 -->
        <button class="control-button" @click="toggleFullscreen" :title="isFullscreen ? '退出全屏' : '全屏'">
          <svg v-if="!isFullscreen" class="icon" viewBox="0 0 24 24" width="24" height="24">
            <path fill="currentColor"
              d="M7 14H5v5h5v-2H7v-3zm-2-4h2V7h3V5H5v5zm12 7h-3v2h5v-5h-2v3zM14 5v2h3v3h2V5h-5z" />
          </svg>
          <svg v-else class="icon" viewBox="0 0 24 24" width="24" height="24">
            <path fill="currentColor"
              d="M5 16h3v3h2v-5H5v2zm3-8H5v2h5V5H8v3zm6 11h2v-3h3v-2h-5v5zm2-11V5h-2v5h5V8h-3z" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'

// 状态定义
const isPlaying = defineModel<boolean>('isPlaying', { default: false })
const isMuted = defineModel<boolean>('isMuted', { default: false })
const volume = defineModel<number>('volume', { default: 1 })
const currentTime = defineModel<number>('currentTime', { default: 0 })
const duration = defineModel<number>('duration', { default: 0 })
const playbackRate = defineModel<number>('playbackRate', { default: 1 })
const rotation = defineModel<number>('rotation', { default: 0 })
const isFullscreen = ref(false)
const isPictureInPicture = ref(false)
const isMaximized = ref(false)
const showSpeedMenu = ref(false)

// 播放速度选项
const speedOptions = [0.5, 0.75, 1, 1.25, 1.5, 2]

// 计算进度条填充百分比
const progressPercent = computed(() => {
  if (duration.value <= 0) return 0
  return (currentTime.value / duration.value) * 100
})

// 格式化时间显示
const formatTime = (seconds: number) => {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs < 10 ? '0' : ''}${secs}`
}

// 播放/暂停控制
const togglePlay = () => {
  emit('play')
}

// 音量控制
const toggleMute = () => {
  emit('mute-toggle', !isMuted.value)
}

const onVolumeChange = () => {
  emit('volume-change', volume.value)
  isMuted.value = volume.value === 0
}

// 进度条控制
const onSeek = (event: Event) => {
  const target = event.target as HTMLInputElement
  const time = parseFloat(target.value)
  emit('seek', time)
}

// 添加进度条点击功能
const onProgressClick = (event: MouseEvent) => {
  const progressBar = event.currentTarget as HTMLElement
  const rect = progressBar.getBoundingClientRect()
  const pos = (event.clientX - rect.left) / rect.width
  const time = pos * (duration.value || 100)
  emit('seek', time)
}

// 播放速度控制
const toggleSpeedMenu = () => {
  showSpeedMenu.value = !showSpeedMenu.value
}

const setPlaybackRate = (rate: number) => {
  playbackRate.value = rate
  emit('playback-rate-change', rate)
  showSpeedMenu.value = false
}

// 视频旋转
const rotateVideo = () => {
  rotation.value = (rotation.value + 90) % 360
  emit('rotate', 90)
}

// 用本地播放器打开视频
const openInPlayer = () => {
  emit('open-in-player')
}

// 本地云播放
const openCloudPlayer = () => {
  emit('open-cloud-player')
}

// 画中画控制
const togglePictureInPicture = () => {
  isPictureInPicture.value = !isPictureInPicture.value
  emit('picture-in-picture')
}

// 最大化控制
const toggleMaximize = () => {
  isMaximized.value = !isMaximized.value
  emit('maximize', isMaximized.value)
}

// 全屏控制
const toggleFullscreen = () => {
  isFullscreen.value = !isFullscreen.value
  emit('fullscreen')
}

// 点击其他地方关闭速度菜单
const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (!target.closest('.speed-control')) {
    showSpeedMenu.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})

// 定义事件发射器
const emit = defineEmits<{
  (e: 'play'): void
  (e: 'pause'): void
  (e: 'seek', time: number): void
  (e: 'volume-change', volume: number): void
  (e: 'mute-toggle', isMuted: boolean): void
  (e: 'playback-rate-change', rate: number): void
  (e: 'rotate', degrees: number): void
  (e: 'fullscreen'): void
  (e: 'picture-in-picture'): void
  (e: 'maximize', isMaximized: boolean): void
  (e: 'open-in-player'): void
  (e: 'open-cloud-player'): void
}>()

// 暴露方法给父组件
defineExpose({
  togglePlay,
  toggleMute,
  setPlaybackRate,
  rotateVideo,
  togglePictureInPicture,
  toggleFullscreen,
  formatTime,
  isPlaying,
  isMuted,
  volume,
  currentTime,
  duration,
  playbackRate,
  rotation,
  isFullscreen,
  isPictureInPicture,
  isMaximized
})
</script>


<style scoped>
.video-controller {
  display: flex;
  flex-direction: column;
  padding: 5px 10px;
  background: rgba(0, 0, 0, 0.7);
  border-radius: 4px;
  color: white;
  min-width: 200px;
  width: calc(100% - 20px);
  user-select: none;
}

.controls-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.controls-left,
.controls-center,
.controls-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.controls-center {
  flex: 1;
  justify-content: center;
}

.controls-right {
  justify-content: flex-end;
}

.control-button {
  background: transparent;
  border: none;
  color: white;
  cursor: pointer;
  padding: 5px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
  flex-shrink: 1;
  width: auto;
  min-width: 30px;
  height: 36px;
}

.control-button:hover {
  background: rgba(255, 255, 255, 0.2);
}

.icon {
  width: 100%;
  height: 100%;
  max-width: 24px;
  max-height: 24px;
}

.volume-control {
  display: flex;
  align-items: center;
  gap: 5px;
  flex-shrink: 1;
}

.volume-slider-container {
  position: relative;
  width: 80px;
  margin-top: -4px;
  flex-shrink: 1;
}

.volume-slider {
  width: 100%;
  height: 5px;
  -webkit-appearance: none;
  background: rgba(255, 255, 255, 0.3);
  border-radius: 5px;
  outline: none;
  position: relative;
}

.volume-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 12px;
  height: 12px;
  margin-top: -3px;
  border-radius: 50%;
  background: white;
  cursor: pointer;
}

/* 音量滑块填充效果 */
.volume-slider::-webkit-slider-container {
  height: 5px;
}

.volume-slider::-webkit-slider-runnable-track {
  height: 5px;
  -webkit-appearance: none;
  background: linear-gradient(to right,
      white 0%,
      white v-bind(volume * 100 + '%'),
      rgba(255, 255, 255, 0.3) v-bind(volume * 100 + '%'),
      rgba(255, 255, 255, 0.3) 100%);
  border-radius: 5px;
}

.time-display {
  font-size: 14px;
  min-width: 80px;
  text-align: center;
  flex-shrink: 1;
}

.progress-container {
  width: 100%;
}

.progress-slider-container {
  position: relative;
  width: 100%;
  height: 5px;
  margin: 6px 0;
  cursor: pointer;
}

/* 进度条背景轨道 */
.progress-slider-container::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(255, 255, 255, 0.3);
  border-radius: 5px;
  z-index: 0;
}

.progress-slider {
  width: 100%;
  height: 5px;
  -webkit-appearance: none;
  background: transparent;
  border-radius: 5px;
  outline: none;
  position: relative;
  z-index: 2;
}

.progress-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 12px;
  height: 12px;
  margin-top: -15px;
  border-radius: 50%;
  background: white;
  cursor: pointer;
  position: relative;
  z-index: 3;
}

/* 进度条填充效果 */
.progress-filled {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  background-color: white;
  border-radius: 5px;
  pointer-events: none;
  z-index: 1;
  transition: width 0.1s linear;
}

.speed-control {
  position: relative;
  flex-shrink: 1;
}

.speed-control .control-button {
  min-width: 40px;
  padding: 5px 6px;
}

.speed-menu {
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(30, 30, 30, 0.9);
  border-radius: 4px;
  padding: 5px 0;
  margin-bottom: 5px;
  display: flex;
  flex-direction: column;
  min-width: 60px;
  z-index: 10;
}

.speed-menu button {
  background: transparent;
  border: none;
  color: white;
  padding: 8px 12px;
  text-align: center;
  cursor: pointer;
}

.speed-menu button:hover,
.speed-menu button.active {
  background: rgba(255, 255, 255, 0.2);
}

/* 当屏幕宽度较小时，缩小控件 */
@media (max-width: 768px) {

  .controls-left,
  .controls-center,
  .controls-right {
    gap: 5px;
  }

  .control-button {
    padding: 3px;
    height: 32px;
    min-width: 25px;
  }

  .time-display {
    font-size: 12px;
    min-width: 60px;
  }

  .volume-slider-container {
    width: 60px;
  }

  .speed-control .control-button {
    padding: 3px 6px;
  }
}

/* 当屏幕宽度小于500px时，进一步缩小控件 */
@media (max-width: 500px) {

  .controls-left,
  .controls-center,
  .controls-right {
    gap: 3px;
  }

  .control-button {
    padding: 2px;
    height: 28px;
    min-width: 20px;
  }

  .time-display {
    font-size: 11px;
    min-width: 50px;
  }

  .volume-slider-container {
    width: 50px;
  }

  .speed-control .control-button {
    padding: 2px 5px;
    min-width: 35px;
  }
}

/* 当屏幕宽度小于400px时，进一步缩小控件 */
@media (max-width: 400px) {

  .controls-left,
  .controls-center,
  .controls-right {
    gap: 2px;
  }

  .control-button {
    padding: 1px;
    height: 24px;
    min-width: 18px;
  }

  .time-display {
    font-size: 10px;
    min-width: 45px;
  }

  .volume-slider-container {
    width: 40px;
  }

  .speed-control .control-button {
    padding: 1px 4px;
    min-width: 30px;
    font-size: 12px;
  }
}
</style>
