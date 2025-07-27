<template>
  <div class="video-player-container">
    <video ref="videoPlayerRef" class="video-js vjs-theme-city" controls preload="auto" width="100%" height="400">
      <source :src="videoSrc" :type="isHls ? 'application/x-mpegURL' : 'video/mp4'">
    </video>
  </div>
</template>

<script lang="ts" setup>
import videojs from 'video.js'
import 'video.js/dist/video-js.css'

// City
import '@videojs/themes/dist/city/index.css';
// Fantasy
//import '@videojs/themes/dist/fantasy/index.css';
// Forest
//import '@videojs/themes/dist/forest/index.css';
// Sea
//import '@videojs/themes/dist/sea/index.css';

import '@videojs/http-streaming'

import { ref, onMounted, onBeforeUnmount } from 'vue'


const videoPlayerRef = ref<HTMLVideoElement | null>(null)
const player = ref<any>(null) // 指定更合适的类型
const videoSrc = ref('')
const isHls = ref(false)

// 初始化播放器
const initializePlayer = () => {
  if (videoPlayerRef.value) {
    player.value = videojs(videoPlayerRef.value, {
      autoplay: false,
      controls: true,
      responsive: true,
      fluid: true,
      playbackRates: [0.5, 1, 1.5, 2],
      techOrder: ['html5'],
      html5: {
        hls: {
          overrideNative: true
        },
        nativeVideoTracks: false,
        nativeAudioTracks: false,
        nativeTextTracks: false
      },
      sources: [],
    }, function () {
      //console.log('Player is ready');
    })
  }
}

// 设置视频源
const setVideoSource = (src: string, type = 'mp4') => {
  videoSrc.value = src
  isHls.value = type === 'm3u8' || type === 'hls'

  if (player.value) {
    // 先重置播放器
    player.value.reset()

    // 设置新的源
    player.value.src({
      src: src,
      type: isHls.value ? 'application/x-mpegURL' : 'video/mp4'
    })

    // 添加 loadedmetadata 事件监听
    player.value.on('loadedmetadata', function () {
      console.log('Video metadata loaded:', {
        videoWidth: player.value.videoWidth(),
        videoHeight: player.value.videoHeight(),
        duration: player.value.duration()
      })
    })

    // 添加错误处理
    player.value.on('error', function () {
      const error = player.value.error()
      console.error('Video.js Error:', error.code, error.message)
    })

    // 添加 loadeddata 事件监听
    player.value.on('loadeddata', function () {
      console.log('Video data loaded successfully')
    })
  }
}
// 添加这个函数来设置音量
const setVolume = (volumeLevel: number) => {
  if (player.value) {
    // 确保音量值在有效范围内
    const validVolume = Math.min(1, Math.max(0, volumeLevel))
    player.value.volume(validVolume)
  }
}

// 组件挂载时初始化播放器
onMounted(() => {
  initializePlayer()
})

// 组件销毁前释放播放器资源
onBeforeUnmount(() => {
  if (player.value) {
    player.value.dispose()
    player.value = null
  }
})

// 导出方法供外部调用
defineExpose({
  setVideoSource,
  setVolume
})
</script>

<style lang="scss" scoped>
.video-player-container {
  width: 100%;
  margin: 0 auto;
}

/* 可选：自定义视频播放器样式 */
.video-js {
  background-color: #000;
}

.video-js .vjs-control-bar {
  background: rgba(0, 0, 0, 0.7);
}
</style>
