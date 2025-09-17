<template>
  <div ref="videoPlayContainerElementRef" class="video-player-container"
    :class="{ 'fullscreen-mode': isFullscreenMode }">
    <video ref="videoPlayerRef" class="video-js vjs-theme-city" preload="auto" width="100%" playsinline
      webkit-playsinline x5-playsinline x5-video-player-type="h5" x5-video-player-fullscreen="true"
      x5-video-orientation="portraint">
      <source :src="videoSrc" :type="isHls ? 'application/x-mpegURL' : 'video/mp4'">
    </video>
    <videoPlayControls v-if="useVideoPlayControls && !isMobile()" ref="videoControlsRef" @play="handlePlay"
      @pause="handlePause" @seek="handleSeek" @volume-change="handleVolumeChange" @mute-toggle="handleMuteToggle"
      @playback-rate-change="handlePlaybackRateChange" @rotate="handleRotate" @fullscreen="handleFullscreen"
      @picture-in-picture="handlePictureInPicture" @maximize="toggleFullscreenMode" @open-in-player="handleOpenInPlayer"
      @open-cloud-player="handleOpenCloudPlayer" />
  </div>
</template>

<script lang="ts" setup>
import videojs from 'video.js'
import 'video.js/dist/video-js.css'
import videoPlayControls from './videoPlayControls.vue';
// City
import '@videojs/themes/dist/city/index.css';
// Fantasy
//import '@videojs/themes/dist/fantasy/index.css';
// Forest
//import '@videojs/themes/dist/forest/index.css';
// Sea
//import '@videojs/themes/dist/sea/index.css';

import '@videojs/http-streaming'

import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { ElMessage } from 'element-plus';
import { isMobile } from '@/assets/mobile';
import { openInPlayerDramaSeries } from '@/common/play';

const props = defineProps({
  useVideoPlayControls: {
    type: Boolean,
    default: true,
  },
  aspectRatio: {
    type: String,
    default: '16:9',
  },
})

const videoPlayControlsHeight = 63;

const videoPlayContainerElementRef = ref<HTMLDivElement | null>(null)
const videoPlayerRef = ref<HTMLVideoElement | null>(null)
const videoControlsRef = ref<InstanceType<typeof videoPlayControls> | null>(null)
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const player = ref<any>(null) // 指定更合适的类型
const videoId = ref('');
const videoSrc = ref('');
const isHls = ref(false)
//
const initVideoAspectRatio = ref(props.aspectRatio)
// 添加旋转角度状态
const rotation = ref(0)
const isFullscreen = ref(false)
const isFullscreenMode = ref(false)

// 初始化播放器
const initializePlayer = () => {
  const isMobileDevice = isMobile()
  if (videoPlayerRef.value) {

    // 针对移动端优化的配置
    const mobileOptions = {
      preload: 'metadata',
      playsinline: true,
      controls: true, // 控制条
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
      controls: !props.useVideoPlayControls, // 默认禁用控制条
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

    // 添加事件监听器同步播放器状态到控制组件
    setupPlayerEventListeners();

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
        // 更新控制组件状态
        if (videoControlsRef.value) {
          videoControlsRef.value.volume = currentVolume;
          videoControlsRef.value.isMuted = isMuted;
        }
        // 保存音量到本地存储
        if (!isMuted) {
          saveVolumeToStorage(currentVolume);
        }
      });
    }
  }
}

// 设置播放器事件监听器
const setupPlayerEventListeners = () => {
  if (!player.value || !videoControlsRef.value) return;

  // 监听播放事件
  player.value.on('play', function () {
    if (videoControlsRef.value) {
      videoControlsRef.value.isPlaying = true;
    }
  });

  // 监听暂停事件
  player.value.on('pause', function () {
    if (videoControlsRef.value) {
      videoControlsRef.value.isPlaying = false;
    }
  });

  // 监听时间更新事件
  player.value.on('timeupdate', function () {
    const currentTime = player.value.currentTime();
    const duration = player.value.duration();

    if (videoControlsRef.value) {
      videoControlsRef.value.currentTime = currentTime;
      videoControlsRef.value.duration = duration;
    }
  });

  // 监听加载元数据事件
  player.value.on('loadedmetadata', function () {
    if (videoControlsRef.value) {
      videoControlsRef.value.duration = player.value.duration();
    }
  });

  // 监听全屏变化事件
  player.value.on('fullscreenchange', function () {
    isFullscreen.value = player.value.isFullscreen();
    // 根据全屏状态切换控制条
    if (isFullscreen.value) {
      // 全屏时启用内置控制条
      player.value.controls(true);
    } else {
      // 非全屏时禁用内置控制条
      player.value.controls(false);
    }
  });
};

// 处理播放事件
const handlePlay = () => {
  if (player.value) {
    if (player.value.paused()) {
      player.value.play();
    } else {
      player.value.pause();
    }
  }
};

// 处理暂停事件
const handlePause = () => {
  if (player.value && !player.value.paused()) {
    player.value.pause();
  }
};

// 处理进度条拖动事件
const handleSeek = (time: number) => {
  if (player.value) {
    player.value.currentTime(time);
  }
};

// 处理音量变化事件
const handleVolumeChange = (volume: number) => {
  if (player.value) {
    player.value.volume(volume);
    player.value.muted(volume === 0);
  }
};

// 处理静音切换事件
const handleMuteToggle = (isMuted: boolean) => {
  if (player.value) {
    player.value.muted(isMuted);
  }
};

// 处理播放速度变化事件
const handlePlaybackRateChange = (rate: number) => {
  if (player.value) {
    player.value.playbackRate(rate);
  }
};

// 处理旋转事件
const handleRotate = (degrees: number) => {
  rotateVideo(degrees);
};

// 最大化函数
const toggleFullscreenMode = () => {
  isFullscreenMode.value = !isFullscreenMode.value
  console.log(props.aspectRatio);
  console.log(player.value?.aspectRatio());
  if (isFullscreenMode.value) {
    // 进入最大化模式
    document.body.style.overflow = 'hidden'
    // 可以在这里添加其他需要的样式调整
  } else {
    // 退出最大化模式
    document.body.style.overflow = ''
  }

  nextTick(() => {
    if (isFullscreenMode.value) {
      const ep = videoPlayContainerElementRef.value || undefined;
      if (ep) {
        // 获取html的宽高
        const { width, height } = ep.getBoundingClientRect();
        setAspectRatio(width + ':' + (height - getControllerHeight()))
      }
    } else {
      setAspectRatio(initVideoAspectRatio.value)
    }
  })

}

// 处理全屏事件
const handleFullscreen = () => {
  if (player.value) {
    if (player.value.isFullscreen()) {
      player.value.exitFullscreen();
    } else {
      player.value.requestFullscreen();
    }
  }
};

// 处理画中画事件
const handlePictureInPicture = () => {
  if (player.value) {
    const videoElement = player.value.el().querySelector('video');
    if (videoElement) {
      if (document.pictureInPictureElement) {
        document.exitPictureInPicture();
      } else {
        videoElement.requestPictureInPicture();
      }
    }
  }
};

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

// 获取 aspectRatio
const getAspectRatio = (): string | null => {
  if (player.value) {
    return player.value.aspectRatio();
  }
  return null;
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

// 从视频路径中提取ID
const extractVideoIdFromPath = (path: string): string => {
  // 匹配 /api/video/mp4/{id}/v.mp4 格式
  const match = path.match(/\/api\/video\/(?:mp4|m3u8)\/([^\/]+)\/v\.(?:mp4|m3u8)/);
  return match ? match[1] : '';
};

// 设置视频源
const setVideoSource = (src: string, type = 'mp4', fn = () => { }) => {
  videoId.value = extractVideoIdFromPath(src)
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
      // 同步时长到控制组件
      if (videoControlsRef.value) {
        videoControlsRef.value.duration = player.value.duration();
      }
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

      const _aspectRatio = getAspectRatio()
      if (_aspectRatio) {
        initVideoAspectRatio.value = _aspectRatio
      }

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

// 获取控制器高度
const getControllerHeight = (): number => {
  return videoPlayControlsHeight
}


// 本地视频播放器打开视频
const handleOpenInPlayer = async () => {
  const b = await openInPlayerDramaSeries(videoId.value)
  if (b) {
    pause()
  }
}

// 云播视频
const handleOpenCloudPlayer = async () => {
  // 获取当前服务器地址
  const serverAddress = window.location.origin;
  // 创建云播放协议链接
  const url = `cmcollectorsvideoplay://${serverAddress}${videoSrc.value}`;
  console.log('尝试打开云播放协议链接:', url);

  try {
    // 检查是否有访问父窗口的权限（即是否在 iframe 中）
    let hasParentAccess = false;
    try {
      hasParentAccess = !!(window.top && window.top !== window.self);
    } catch {
      // 如果访问被拒绝（安全错误），则认为在 iframe 中但无访问权限
      hasParentAccess = false;
    }

    if (hasParentAccess) {
      // 在 iframe 中且可以访问父窗口，通过父窗口打开
      if (window.top) {
        window.top.location.href = url;
      } else {
        window.location.href = url;
      }
    } else {
      // 不在 iframe 中或无法访问父窗口，直接打开
      window.location.href = url;
    }
  } catch (error) {
    console.error('打开云播放器失败:', error);
    alert('无法打开云播放器，请确保已正确安装并配置了相关组件。');
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
  getRotation,
  toggleFullscreenMode,
  getControllerHeight,
})
</script>

<style lang="scss">
.video-player-container {
  width: 100%;
  margin: 0 auto;
  overflow: hidden;
}

/* 添加最大化模式样式 */
.video-player-container.fullscreen-mode {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  z-index: 9999;
  background-color: #000;
  display: flex;
  flex-direction: column;
}

.video-player-container.fullscreen-mode .video-js {
  flex: 1;
  width: 100%;
  height: calc(100% - 44px);
  /* 减去控制条的大致高度 */
}

.video-player-container.fullscreen-mode .video-controller {
  /* 保持控制条原有尺寸 */
  flex-shrink: 0;
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
