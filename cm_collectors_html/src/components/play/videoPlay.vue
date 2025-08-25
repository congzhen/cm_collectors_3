<template>
  <div class="video-player-container">
    <video ref="videoPlayerRef" class="video-js vjs-theme-city" controls preload="auto" width="100%" playsinline
      webkit-playsinline x5-playsinline x5-video-player-type="h5" x5-video-player-fullscreen="true"
      x5-video-orientation="portraint">
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
import { isMobile } from '@/assets/mobile';

const props = defineProps({
  aspectRatio: {
    type: String,
    default: '16:9',
  },
})

const videoPlayerRef = ref<HTMLVideoElement | null>(null)
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const player = ref<any>(null) // 指定更合适的类型
const videoSrc = ref('')
const isHls = ref(false)
// 添加旋转角度状态
const rotation = ref(0)


// 初始化播放器
const initializePlayer = () => {
  const isMobileDevice = isMobile()
  if (videoPlayerRef.value) {

    // 针对移动端优化的配置
    const mobileOptions = {
      preload: 'metadata',
      playsinline: true,
      controls: true,
      autoplay: false,
      muted: false,
      techOrder: ['html5'],
      html5: {
        hls: {
          overrideNative: !isMobileDevice
        },
        nativeVideoTracks: isMobileDevice,
        nativeAudioTracks: isMobileDevice,
        nativeTextTracks: isMobileDevice
      }
    }

    // 桌面端配置
    const desktopOptions = {
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
      }
    }
    const options = isMobileDevice ? mobileOptions : desktopOptions

    player.value = videojs(videoPlayerRef.value, {
      ...options,
      sources: [],
      track: [],
      fill: false,
      aspectRatio: props.aspectRatio,

    }, function () {
      //console.log('Player is ready');
    })

    // (仅在桌面端)
    if (!isMobileDevice) {
      // 添加自定义旋转按钮
      addRotateButton();

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
}

// 添加旋转按钮到控制栏
const addRotateButton = () => {
  if (player.value) {
    // 创建旋转按钮组件
    const Button = videojs.getComponent('Button');

    // 使用新的方式创建组件，替代 videojs.extend
    class RotateButton extends Button {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      constructor(player: any, options: any = {}) {
        super(player, options);
      }

      buildCSSClass() {
        return 'vjs-rotate-button ' + super.buildCSSClass();
      }

      handleClick() {
        rotateVideo(90);
      }
    }

    // 注册组件
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    videojs.registerComponent('RotateButton', RotateButton as any);

    // 添加到控制栏
    player.value.ready(() => {
      player.value.controlBar.addChild('RotateButton', {});
    });
  }
};

// 旋转
const rotateVideo = (degrees: number) => {
  rotation.value = (rotation.value + degrees) % 360;
  if (player.value) {
    const videoElement = player.value.el().querySelector('.vjs-tech');
    if (videoElement) {
      // 根据旋转角度调整视频的变换
      applyRotationTransformation(videoElement);
    }
  }
};

// 设置旋转角度的
const setRotation = (degrees: number) => {
  rotation.value = degrees % 360;
  if (player.value) {
    const videoElement = player.value.el().querySelector('.vjs-tech');
    if (videoElement) {
      // 根据旋转角度调整视频的变换
      applyRotationTransformation(videoElement);
    }
  }
};

// 应用旋转变换
const applyRotationTransformation = (videoElement: HTMLElement) => {
  // 获取视频的原始尺寸
  const videoWidth = player.value.videoWidth();
  const videoHeight = player.value.videoHeight();

  if (videoWidth && videoHeight) {
    // 计算合适的缩放因子
    const scale = calculateScaleFactor(rotation.value, videoWidth, videoHeight, videoElement);

    // 应用旋转和缩放
    videoElement.style.transform = `rotate(${rotation.value}deg) scale(${scale})`;
    videoElement.style.transformOrigin = 'center center';
  } else {
    // 如果没有视频尺寸信息，则仅应用旋转
    videoElement.style.transform = `rotate(${rotation.value}deg)`;
    videoElement.style.transformOrigin = 'center center';
  }
};

// 计算旋转时的缩放因子
const calculateScaleFactor = (rotation: number, videoWidth: number, videoHeight: number, videoElement: HTMLElement) => {
  // 对于0度和180度旋转，不需要缩放
  if (rotation === 0 || rotation === 180) {
    return 1;
  }
  // 获取容器尺寸
  const containerWidth = videoElement.clientWidth;
  const containerHeight = videoElement.clientHeight;

  let w = 0;
  let h = 0;
  if (videoWidth >= videoHeight) {
    w = containerHeight;
    h = containerWidth / w * containerHeight;
  } else {
    h = containerWidth
    w = containerHeight / h * containerWidth;
  }
  return Math.min(containerWidth / w, containerHeight / h);
};



// 获取当前旋转角度
const getRotation = (): number => {
  return rotation.value;
};

//播放
const play = () => {
  player.value?.play();
}
// 暂停
const pause = () => {
  player.value?.pause();
}

// 设置 aspectRatio
const setAspectRatio = (aspectRatio: string) => {
  if (player.value) {
    player.value.aspectRatio(aspectRatio);
  }
};

// 获取播放状态
const isPlaying = (): boolean => {
  if (player.value) {
    return !player.value.paused();
  }
  return false;
};

// 获取当前播放时间
const getCurrentTime = (): number => {
  if (player.value) {
    return player.value.currentTime();
  }
  return 0;
};

// 获取视频总时长
const getDuration = (): number => {
  if (player.value) {
    return player.value.duration();
  }
  return 0;
};

// 获取播放进度 (0-1)
const getProgress = (): number => {
  const duration = getDuration();
  if (duration > 0) {
    return getCurrentTime() / duration;
  }
  return 0;
};

// 设置播放位置
const setCurrentTime = (time: number) => {
  if (player.value) {
    player.value.currentTime(time);
  }
};

// 设置视频源
const setVideoSource = (src: string, type = 'mp4', fn = () => { }) => {
  videoSrc.value = src
  isHls.value = type === 'm3u8' || type === 'hls'

  if (player.value) {
    // 先重置播放器
    //resetPlayer();
    //player.value.reset();

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
      // 重新应用旋转效果
      if (rotation.value !== 0) {
        setTimeout(() => {
          const videoElement = player.value.el().querySelector('.vjs-tech');
          if (videoElement) {
            applyRotationTransformation(videoElement);
          }
        }, 0);
      }
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

      // 重置旋转角度
      rotation.value = 0;
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

// 获取视频尺寸
const getVideoDimensions = (): { width: number; height: number } | null => {
  if (player.value) {
    // 检查视频是否已加载元数据
    if (player.value.readyState() >= 1) { // HAVE_METADATA
      return {
        width: player.value.videoWidth(),
        height: player.value.videoHeight()
      };
    }
  }
  return null;
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
  pause,
  setAspectRatio,
  isPlaying,
  getCurrentTime,
  getDuration,
  getProgress,
  setCurrentTime,
  resetPlayer,
  setVideoSource,
  setVolume,
  getVolume,
  addTextTrack,
  getVideoDimensions,
  // 导出旋转相关方法
  rotateVideo,
  setRotation,
  getRotation
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
  margin: 0 auto;
  width: 100%;
  height: 100%;

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

/* 旋转按钮样式 */
.video-js .vjs-rotate-button .vjs-icon-placeholder:before {
  content: '\f11a';
  /* 使用一个合适的图标 */
  font-family: VideoJS;
  font-weight: normal;
  font-style: normal;
}

.video-js .vjs-rotate-button {
  cursor: pointer;
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
