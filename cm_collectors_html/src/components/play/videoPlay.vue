<template>
  <div class="video-player-container">
    <video ref="videoPlayerRef" class="video-js vjs-theme-city" controls preload="auto" width="100%">
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
import { ElMessage } from 'element-plus';

const props = defineProps({
  aspectRatio: {
    type: String,
    default: '16:9',
  },
})

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
      track: [],
      fill: false,
      aspectRatio: props.aspectRatio,
    }, function () {
      //console.log('Player is ready');
    })

    //监控音量变化
    player.value.on('volumechange', function () {
      // 获取当前音量
      const currentVolume = player.value.volume();
      // 获取当前静音状态
      const isMuted = player.value.muted();
      // 保存音量到本地存储
      if (!isMuted) {
        saveVolumeToStorage(currentVolume);
      }
    });
  }
}

//播放
const play = () => {
  player.value?.play();
}

// 设置视频源
const setVideoSource = (src: string, type = 'mp4', fn = () => { }) => {
  videoSrc.value = src
  isHls.value = type === 'm3u8' || type === 'hls'

  if (player.value) {
    // 先重置播放器
    resetPlayer();

    // 从本地存储读取并设置音量
    const savedVolume = getVolumeFromStorage();
    setVolume(savedVolume)

    // 设置新的源
    player.value.src({
      src: src,
      type: isHls.value ? 'application/x-mpegURL' : 'video/mp4'
    })

    // 添加 loadedmetadata 事件监听
    player.value.on('loadedmetadata', function () {
      /*
      console.log('Video metadata loaded:', {
        videoWidth: player.value.videoWidth(),
        videoHeight: player.value.videoHeight(),
        duration: player.value.duration()
      })
        */
      fn();
    })

    // 添加错误处理
    player.value.on('error', function () {
      const error = player.value.error()
      console.error('Video.js Error:', error.code, error.message)
      ElMessage({
        showClose: true,
        message: error.message,
        type: 'error',
        duration: 5000,
      })
    })

    // 添加 loadeddata 事件监听
    player.value.on('loadeddata', function () {
      //console.log('Video data loaded successfully')
    })
  }
}
// 设置音量（0~1）
const setVolume = (volumeLevel: number) => {
  if (player.value) {
    // 确保音量值在有效范围内
    const validVolume = Math.min(1, Math.max(0, volumeLevel))
    player.value.volume(validVolume)
    // 触发音量变化事件，更新UI
    player.value.trigger('volumechange')
    console.log('设置声音');
  }
}
const getVolume = () => {
  if (player.value) {
    return player.value.volume()
  }
  return 0
}

// 添加字幕轨道
const addTextTrack = (src: string, label: string, language: string, isDefault = false) => {
  if (player.value) {
    // 先移除已有的字幕轨道
    removeAllTextTracks()

    // 添加新的字幕轨道
    const track = player.value.addRemoteTextTrack({
      kind: 'subtitles',
      src: src,
      srclang: language,
      label: label,
      default: isDefault
    }, true) // 注意这里改为 true，立即加载

    // 监听字幕加载完成事件
    if (track && track.track) {
      // 字幕数据加载完成
      track.track.addEventListener('load', function () {
        //console.log('Subtitle track loaded successfully');
      });

      // 字幕数据加载错误
      track.track.addEventListener('error', function (e: ErrorEvent) {
        console.error('Subtitle track load error:', e);
      });

      // 字幕就绪状态改变
      track.track.addEventListener('cuechange', function () {
        //console.log('Subtitle cue changed');
        // 强制触发字幕更新
        player.value.trigger('texttrackchange');
      });
    }

    // 强制显示字幕（如果设置为默认）
    if (isDefault) {
      setTimeout(() => {
        const textTracks = player.value.textTracks();
        for (let i = 0; i < textTracks.length; i++) {
          if (textTracks[i].language === language) {
            textTracks[i].mode = 'showing';
            //console.log('Subtitle showing for language:', language);
            // 触发更新
            player.value.trigger('texttrackchange');
            break;
          }
        }
      }, 500);
    }

    return track
  }
}

// 清除所有字幕轨道
const removeAllTextTracks = () => {
  if (player.value) {
    const tracks = player.value.remoteTextTracks() || []
    for (let i = tracks.length - 1; i >= 0; i--) {
      player.value.removeRemoteTextTrack(tracks[i])
    }
  }
}

// 添加重置播放器的方法  该方法有可能 触发  player.value.on('error'
const resetPlayer = () => {
  if (player.value) {
    try {
      // 保存当前状态
      const currentVolume = player.value.volume();
      const isMuted = player.value.muted();

      // 清理所有事件监听
      player.value.off('loadedmetadata')
      player.value.off('error')
      player.value.off('loadeddata')

      // 暂停并重置
      player.value.pause()
      player.value.load()
      //player.value.reset()

      // 恢复音量状态
      player.value.volume(currentVolume);
      player.value.muted(isMuted);

      // 清理字幕轨道
      removeAllTextTracks()
    } catch (e) {
      console.warn('Error resetting player:', e)
    }
  }
}

// 定义本地存储的键名
const VOLUME_STORAGE_KEY = 'video-player-volume';

// 保存音量到本地存储
const saveVolumeToStorage = (volume: number) => {
  try {
    localStorage.setItem(VOLUME_STORAGE_KEY, volume.toString());
  } catch (e) {
    console.warn('无法保存音量到本地存储:', e);
  }
};

// 从本地存储读取音量
const getVolumeFromStorage = (): number => {
  try {
    const savedVolume = localStorage.getItem(VOLUME_STORAGE_KEY);
    return savedVolume ? parseFloat(savedVolume) : 1; // 默认音量为1
  } catch (e) {
    console.warn('无法从本地存储读取音量:', e);
    return 1;
  }
};


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
  play,
  resetPlayer,
  setVideoSource,
  setVolume,
  getVolume,
  addTextTrack
})
</script>

<style lang="scss">
.video-player-container {
  width: 100%;
  margin: 0 auto;
  overflow: hidden;
}

/* 可选：自定义视频播放器样式 */
.video-js {
  background-color: #000;

  video {
    /* 保证视频完整显示 */
    object-fit: contain;
  }
}

.video-js .vjs-control-bar {
  background: rgba(0, 0, 0, 0.7);
}

.vjs-playback-rate .vjs-playback-rate-value {
  padding-top: 9px;
}


/* 字幕样式设置 - 白色字体带黑色边框 */
.video-js video::-webkit-media-text-track-display {
  font-size: 1.2em !important;
  text-align: center !important;
}

.video-js .vjs-text-track-display div,
.video-js .vjs-text-track-cue>div {
  font-size: 1.4em !important;
  text-align: center !important;
  color: white !important;
  text-shadow:
    -1px -1px 0 #000,
    1px -1px 0 #000,
    -1px 1px 0 #000,
    1px 1px 0 #000,
    -2px -2px 0 #000,
    2px -2px 0 #000,
    -2px 2px 0 #000,
    2px 2px 0 #000 !important;
  background-color: transparent !important;
  background: transparent !important;
  font-family: Arial, Helvetica, sans-serif !important;
}

/* 字幕容器背景 */
.video-js .vjs-text-track-display {
  background-color: rgba(0, 0, 0, 0) !important;
}
</style>
