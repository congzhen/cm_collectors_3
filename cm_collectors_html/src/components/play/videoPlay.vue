<template>
  <Teleport to="body" :disabled="!isFullscreenMode">
    <div ref="videoPlayContainerElementRef" class="video-player-container"
      :class="{ 'fullscreen-mode': isFullscreenMode, 'controls-hidden': isControlsHidden }"
      @mousemove="handlePlayerActivity" @mousedown="handlePlayerActivity">
    <div class="video-player-windows" :key="indexkey">
      <div v-if="isLoading && !isMobile()" class="loading-overlay">
        <div class="loading-spinner"></div>
        <p class="loading-text">{{ loadingText }}</p>
      </div>
      <!-- 桌面端使用 video.js -->
      <video v-if="!isMobile()" ref="videoPlayerRef" class="video-js vjs-theme-city" preload="auto" width="100%"
        playsinline webkit-playsinline x5-playsinline x5-video-player-type="h5" x5-video-player-fullscreen="true"
        x5-video-orientation="portraint">
        <source :src="videoSrc" :type="isHls ? 'application/x-mpegURL' : 'video/mp4'">
      </video>
      <!-- 移动端使用原生播放器 -->
      <video v-else ref="nativeVideoRef" class="native-video-player" controls preload="metadata" width="100%"
        playsinline webkit-playsinline x5-playsinline x5-video-player-type="h5" x5-video-player-fullscreen="true"
        x5-video-orientation="portraint" @timeupdate="handleNativeTimeUpdate">
        <source :src="videoSrc" :type="isHls ? 'application/x-mpegURL' : 'video/mp4'">
      </video>
      <div v-if="activeSubtitleText" class="custom-subtitle-display">{{ activeSubtitleText }}</div>
    </div>
    <videoPlayControls v-if="useVideoPlayControls && !isMobile()" ref="videoControlsRef" @play="handlePlay"
      @pause="handlePause" @seek="handleSeek" @volume-change="handleVolumeChange" @mute-toggle="handleMuteToggle"
      @playback-rate-change="handlePlaybackRateChange" @rotate="handleRotate" @fullscreen="handleFullscreen"
      @picture-in-picture="handlePictureInPicture" @maximize="handleMaximize" @open-in-player="handleOpenInPlayer"
      @open-cloud-player="handleOpenCloudPlayer" @mouseenter="handleControlsInteractionStart"
      @mouseleave="handleControlsInteractionEnd" @focusin="handleControlsInteractionStart"
      @focusout="handleControlsInteractionEnd" />
    <playCloudCheckPromptDialog ref="playCloudCheckPromptDialogRef" />
    </div>
  </Teleport>
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

import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { ElMessage } from 'element-plus';
import { isMobile } from '@/assets/mobile';
import { openInPlayerDramaSeries } from '@/common/play';
import { onHostWindowFullscreenChanged, setHostWindowFullscreen } from '@/common/runtimeBridge';
import { appStoreData } from '@/storeData/app.storeData';
import playCloudCheckPromptDialog from './playCloudCheckPromptDialog.vue';

interface I_playerSource {
  src: string;
  type: 'application/x-mpegURL' | 'video/mp4';
}

interface I_subtitleCue {
  startTime: number;
  endTime: number;
  text: string;
}

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

const appStore = appStoreData()
const indexkey = ref(0);
const isLoading = ref(false);
const loadingText = ref('正在加载视频...');

const videoPlayControlsHeight = 63;

const videoPlayContainerElementRef = ref<HTMLDivElement | null>(null)
const videoPlayerRef = ref<HTMLVideoElement | null>(null)
const nativeVideoRef = ref<HTMLVideoElement | null>(null)
const videoControlsRef = ref<InstanceType<typeof videoPlayControls> | null>(null)
const playCloudCheckPromptDialogRef = ref<InstanceType<typeof playCloudCheckPromptDialog> | null>(null)
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const player = ref<any>(null) // 指定更合适的类型
let playerSetupVersion = 0;
let playerSetupTimer: number | undefined;
const videoId = ref('');
const videoSrc = ref('');
const isHls = ref(false)
const subtitleCues = ref<I_subtitleCue[]>([])
const activeSubtitleText = ref('')
let subtitleLoadVersion = 0
//
const initVideoAspectRatio = ref(props.aspectRatio)
// 添加旋转角度状态
const rotation = ref(0)
const isFullscreen = ref(false)
const isFullscreenMode = ref(false)
const isControlsHidden = ref(false)
let browserFullscreenActive = false
let bodyOverflowBeforeMaximize: string | null = null
let controlsHideTimer: number | undefined
let controlsInteractionDepth = 0
let removeHostFullscreenListener: (() => void) | undefined
const controlsHideDelay = 3000
const videoOptions = (isMobileDevice: boolean) => {
  // 公共配置
  const baseOptions = {
    autoplay: false,
    controls: true,
    preload: 'metadata',
    techOrder: ['html5'],
    html5: {
      // 同时支持 hls 和 vhs 配置以确保兼容性
      hls: {
        overrideNative: !isMobileDevice,
        withCredentials: false,
        handleManifestRedirects: true
      },
      vhs: {
        overrideNative: !isMobileDevice,
        withCredentials: false,
        handleManifestRedirects: true,
        smoothQualityChange: true,
        enableLowInitialPlaylist: true,
        fastQualityChange: true,
        limitRenditionByPlayerDimensions: false,
        useDevicePixelRatio: true,
        useNetworkInformationApi: true,
        useDtsForTimestampOffset: true,
      }
    }
  }

  // 移动端特有配置
  if (isMobileDevice) {
    return {
      ...baseOptions,
      playsinline: true,
      controls: true, // 移动端始终启用控制条
      muted: false,
      html5: {
        ...baseOptions.html5,
        // 移动端使用原生轨道
        nativeVideoTracks: true,
        nativeAudioTracks: true,
        nativeTextTracks: true,
        hls: {
          ...baseOptions.html5.hls,
          overrideNative: false // 移动端强制使用原生播放器
        },
        vhs: {
          ...baseOptions.html5.vhs,
          overrideNative: false // 移动端强制使用原生播放器
        }
      }
    }
  }
  // 桌面端特有配置
  else {
    return {
      ...baseOptions,
      controls: !props.useVideoPlayControls, // 桌面端根据props决定是否启用控制条
      responsive: true,
      fluid: true,
      playbackRates: [0.5, 1, 1.5, 2],
      html5: {
        ...baseOptions.html5,
        hls: {
          ...baseOptions.html5.hls,
          overrideNative: true
        },
        vhs: {
          ...baseOptions.html5.vhs,
          overrideNative: true,
          cacheEncryptionKeys: true
        },
        // 桌面端不使用原生轨道
        nativeVideoTracks: false,
        nativeAudioTracks: false,
        nativeTextTracks: false
      }
    }
  }
}

const setupPlayer = (_sources: I_playerSource, callBack?: () => void) => {
  const setupVersion = ++playerSetupVersion;

  // video.js 会在媒体元素之外维护播放实例。切换 key 前必须先释放旧实例，
  // 否则旧元素虽然从 Vue DOM 中移除，音轨仍可能继续播放。
  window.clearTimeout(playerSetupTimer);
  if (player.value) {
    player.value.pause();
    player.value.dispose();
    player.value = null;
  }

  indexkey.value++;
  nextTick(() => {
    playerSetupTimer = window.setTimeout(() => {
      if (setupVersion !== playerSetupVersion) return;
      initializePlayer(_sources, callBack);
    }, 100)
  })
}

// 初始化播放器
const initializePlayer = (_sources: I_playerSource, callBack?: () => void) => {
  const isMobileDevice = isMobile()
  if (videoPlayerRef.value) {
    const options = videoOptions(isMobileDevice)
    player.value = videojs(videoPlayerRef.value, {
      ...options,
      sources: [_sources],
      track: [],
      fill: false,
      aspectRatio: props.aspectRatio,

    }, function () {
      if (callBack) {
        callBack();
      }
    })
    // 添加事件监听器同步播放器状态到控制组件
    setupPlayerEventListeners();
    // (仅在桌面端)
    if (!isMobileDevice) {
      // 添加自定义旋转按钮
      addRotateButton();
    }
  }
}

// 设置播放器事件监听器
const setupPlayerEventListeners = () => {
  if (!player.value) return;

  // 监听播放事件
  player.value.on('play', function () {
    if (videoControlsRef.value) {
      videoControlsRef.value.isPlaying = true;
    }
    // 隐藏loading
    isLoading.value = false;
    showControls()
  });

  // 监听暂停事件
  player.value.on('pause', function () {
    if (videoControlsRef.value) {
      videoControlsRef.value.isPlaying = false;
    }
    showControls(false)
  });

  // 监听时间更新事件
  player.value.on('timeupdate', function () {
    const currentTime = player.value.currentTime();
    const duration = player.value.duration();
    updateActiveSubtitle(currentTime)

    if (videoControlsRef.value) {
      videoControlsRef.value.currentTime = currentTime;
      videoControlsRef.value.duration = duration;
    }
  });
  // 监听加载开始事件
  player.value.on('loadstart', function () {
    // 显示loading
    isLoading.value = true;
    loadingText.value = '正在加载视频...';
  });
  // 监听等待数据事件
  player.value.on('waiting', function () {
    // 显示loading
    isLoading.value = true;
    loadingText.value = '正在缓冲...';
  });
  // 监听加载元数据事件
  player.value.on('loadedmetadata', function () {
    if (videoControlsRef.value) {
      videoControlsRef.value.duration = player.value.duration();
    }
    // 隐藏loading
    isLoading.value = false;
  });
  // 监听可以播放事件
  player.value.on('canplay', function () {
    // 隐藏loading
    isLoading.value = false;
  });

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

const updateActiveSubtitle = (currentTime: number) => {
  activeSubtitleText.value = subtitleCues.value
    .filter((cue) => cue.startTime <= currentTime && currentTime <= cue.endTime)
    .map((cue) => cue.text)
    .join('\n')
}

const handleNativeTimeUpdate = () => {
  updateActiveSubtitle(nativeVideoRef.value?.currentTime || 0)
}

const clearControlsHideTimer = () => {
  window.clearTimeout(controlsHideTimer)
  controlsHideTimer = undefined
}

const showControls = (autoHide = true) => {
  isControlsHidden.value = false
  clearControlsHideTimer()
  if (
    autoHide &&
    controlsInteractionDepth === 0 &&
    isFullscreenMode.value &&
    player.value &&
    !player.value.paused()
  ) {
    controlsHideTimer = window.setTimeout(() => {
      if (controlsInteractionDepth > 0) return
      isControlsHidden.value = true
    }, controlsHideDelay)
  }
}

const handleControlsInteractionStart = () => {
  controlsInteractionDepth++
  showControls(false)
}

const handleControlsInteractionEnd = () => {
  controlsInteractionDepth = Math.max(0, controlsInteractionDepth - 1)
  if (controlsInteractionDepth === 0) {
    showControls()
  }
}

const handlePlayerActivity = () => {
  if (isFullscreenMode.value) {
    showControls()
  }
}

// 最大化函数
const toggleFullscreenMode = (maximized = !isFullscreenMode.value) => {
  if (isFullscreenMode.value === maximized) return

  isFullscreenMode.value = maximized
  if (isFullscreenMode.value) {
    bodyOverflowBeforeMaximize = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    showControls()
  } else {
    document.body.style.overflow = bodyOverflowBeforeMaximize ?? ''
    bodyOverflowBeforeMaximize = null
    showControls(false)
  }

  nextTick(() => {
    videoControlsRef.value?.setMaximized(isFullscreenMode.value)
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

const handleMaximizeKeydown = (event: KeyboardEvent) => {
  handlePlayerActivity()
  if (event.key !== 'Escape') return

  if (isFullscreen.value) {
    if (appStore.runtimeBridgeStatus) {
      void handleFullscreen()
    }
    return
  }

  if (isFullscreenMode.value) {
    toggleFullscreenMode(false)
  }
}

const handleMaximize = (maximized: boolean) => {
  if (!maximized && isFullscreen.value) {
    void handleFullscreen()
    return
  }
  toggleFullscreenMode(maximized)
}

const handleDocumentFullscreenChange = () => {
  if (!browserFullscreenActive) return

  const fullscreen = document.fullscreenElement !== null
  isFullscreen.value = fullscreen
  videoControlsRef.value?.setFullscreen(fullscreen)
  if (!fullscreen) {
    browserFullscreenActive = false
    toggleFullscreenMode(false)
  }
}

const handleHostFullscreenChange = (fullscreen: boolean) => {
  isFullscreen.value = fullscreen
  videoControlsRef.value?.setFullscreen(fullscreen)
  toggleFullscreenMode(fullscreen)
}

// 处理全屏事件
const handleFullscreen = async () => {
  if (appStore.runtimeBridgeStatus) {
    const fullscreen = !isFullscreen.value
    try {
      await setHostWindowFullscreen(fullscreen)
      isFullscreen.value = fullscreen
      videoControlsRef.value?.setFullscreen(fullscreen)
      toggleFullscreenMode(fullscreen)
    } catch (error) {
      console.error('Wails fullscreen toggle failed:', error)
      ElMessage.error('无法切换全屏模式')
    }
    return
  }

  try {
    if (browserFullscreenActive && document.fullscreenElement) {
      await document.exitFullscreen()
    } else {
      browserFullscreenActive = true
      toggleFullscreenMode(true)
      await document.documentElement.requestFullscreen()
    }
  } catch (error) {
    browserFullscreenActive = false
    isFullscreen.value = false
    videoControlsRef.value?.setFullscreen(false)
    toggleFullscreenMode(false)
    console.error('Browser fullscreen toggle failed:', error)
    ElMessage.error('当前环境不支持全屏播放')
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
  if (isMobile()) {
    nativeVideoRef.value?.play();
    return;
  }
  player.value?.play();
}
// 暂停
const pause = () => {
  if (isMobile()) {
    nativeVideoRef.value?.pause();
    return;
  }
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
  if (isMobile()) {
    return Boolean(nativeVideoRef.value && !nativeVideoRef.value.paused && !nativeVideoRef.value.ended);
  }
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
const setVideoSource = (src: string, type = 'mp4', fn = () => { }, retryCount = 0) => {
  clearSubtitleCues()
  videoId.value = extractVideoIdFromPath(src)
  videoSrc.value = src
  isHls.value = type === 'm3u8' || type === 'hls'
  const source: I_playerSource = {
    src: src,
    type: isHls.value ? 'application/x-mpegURL' : 'video/mp4'
  }
  // 移动端直接使用原生播放器
  if (isMobile()) {
    if (nativeVideoRef.value) {
      // 记录当前播放状态
      const wasPlaying = !nativeVideoRef.value.paused && !nativeVideoRef.value.ended;
      // 为了解决某些浏览器下切换视频后UI卡住的问题，我们先暂停并重置播放器
      nativeVideoRef.value.pause();
      nativeVideoRef.value.removeAttribute('src');
      void nativeVideoRef.value.load();

      // 使用 nextTick 确保 DOM 更新后再设置新源
      nextTick(() => {
        nativeVideoRef.value!.src = src;
        nativeVideoRef.value!.addEventListener('loadedmetadata', fn, { once: true })
        // 强制重新加载视频
        void nativeVideoRef.value!.load();
        // 如果之前正在播放，则在可以播放时自动播放
        if (wasPlaying) {
          const playHandler = () => {
            nativeVideoRef.value?.play();
            nativeVideoRef?.value?.removeEventListener('canplay', playHandler);
          };
          nativeVideoRef.value?.addEventListener('canplay', playHandler);
        }
      });
    }
    return;
  }
  isLoading.value = true;
  loadingText.value = retryCount > 0 ? `重试中 (${retryCount}/3)...` : '正在加载视频...';
  setupPlayer(source, () => {
    if (player.value) {
      // 从本地存储读取并设置音量
      const savedVolume = getVolumeFromStorage();
      setVolume(savedVolume)
      const handleLoadedMetadata = () => {
        // 同步时长到控制组件
        if (videoControlsRef.value) {
          videoControlsRef.value.duration = player.value.duration();
        }
        fn();
        // 隐藏loading
        isLoading.value = false;
      }
      // 本地视频或缓存命中时，元数据可能在 ready 回调前已经加载完成。
      // 此时应立即执行后续逻辑，避免错过字幕轨道的添加。
      if (player.value.readyState() >= 1) {
        handleLoadedMetadata()
      } else {
        player.value.one('loadedmetadata', handleLoadedMetadata)
      }
      // 添加错误处理
      player.value.on('error', function () {
        const error = player.value.error()
        if (retryCount < 3) {
          retryCount++
          console.log('重试加载视频：', retryCount)
          // 添加延迟重试，避免频繁请求
          setTimeout(() => {
            setVideoSource(src, type, fn, retryCount)
          }, 1000)
          return
        }
        // 隐藏loading
        isLoading.value = false;
        ElMessage({
          showClose: true,
          message: error.message,
          type: 'error',
          duration: 5000,
        })
      })
      // 添加 loadeddata 事件监听
      player.value.on('loadeddata', function () {
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
        // 隐藏loading
        isLoading.value = false;
      })
    }
  })

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

const parseSubtitleTimestamp = (value: string): number => {
  const parts = value.trim().replace(',', '.').split(':').map(Number)
  if (parts.some(Number.isNaN)) return -1
  if (parts.length === 3) return parts[0] * 3600 + parts[1] * 60 + parts[2]
  if (parts.length === 2) return parts[0] * 60 + parts[1]
  return -1
}

const decodeSubtitleText = (value: string): string => {
  const textarea = document.createElement('textarea')
  textarea.innerHTML = value.replace(/<[^>]+>/g, '')
  return textarea.value
}

const parseWebVtt = (content: string): I_subtitleCue[] => {
  const blocks = content.replace(/^\uFEFF/, '').replace(/\r\n?/g, '\n').split(/\n{2,}/)
  const cues: I_subtitleCue[] = []

  for (const block of blocks) {
    const lines = block.split('\n').map((line) => line.trimEnd())
    const timingIndex = lines.findIndex((line) => line.includes('-->'))
    if (timingIndex < 0) continue

    const timing = lines[timingIndex].match(/^(\S+)\s+-->\s+(\S+)/)
    if (!timing) continue
    const startTime = parseSubtitleTimestamp(timing[1])
    const endTime = parseSubtitleTimestamp(timing[2])
    if (startTime < 0 || endTime < startTime) continue

    const text = decodeSubtitleText(lines.slice(timingIndex + 1).join('\n')).trim()
    if (text) cues.push({ startTime, endTime, text })
  }

  return cues.sort((a, b) => a.startTime - b.startTime)
}

// 字幕与视频流独立加载，统一使用自定义覆盖层显示，避免 WebView 文本轨道兼容性差异。
const addTextTrack = async (src: string, _label: string, _language: string, _isDefault = false) => {
  const loadVersion = ++subtitleLoadVersion
  try {
    const response = await fetch(src)
    if (!response.ok) {
      if (response.status !== 404) {
        console.error(`Subtitle request failed: ${response.status} ${response.statusText}`)
      }
      return
    }
    const cues = parseWebVtt(await response.text())
    if (loadVersion !== subtitleLoadVersion) return
    subtitleCues.value = cues
    const currentTime = player.value?.currentTime() || nativeVideoRef.value?.currentTime || 0
    updateActiveSubtitle(currentTime)
  } catch (error) {
    if (loadVersion === subtitleLoadVersion) {
      console.error('Subtitle load failed:', error)
    }
  }
}

const clearSubtitleCues = () => {
  subtitleLoadVersion++
  subtitleCues.value = []
  activeSubtitleText.value = ''
}

// 清除所有字幕轨道
const removeAllTextTracks = () => {
  clearSubtitleCues()
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
  playCloudCheckPromptDialogRef.value?.open(videoId.value)

}



// 组件挂载时初始化播放器
onMounted(() => {
  document.addEventListener('keydown', handleMaximizeKeydown)
  document.addEventListener('fullscreenchange', handleDocumentFullscreenChange)
  if (appStore.runtimeBridgeStatus) {
    removeHostFullscreenListener = onHostWindowFullscreenChanged(handleHostFullscreenChange)
  }
  //initializePlayer()
})

// 组件销毁前释放播放器资源
onBeforeUnmount(() => {
  const shouldExitBrowserFullscreen = browserFullscreenActive && document.fullscreenElement !== null
  const shouldExitHostFullscreen = appStore.runtimeBridgeStatus && isFullscreen.value
  document.removeEventListener('keydown', handleMaximizeKeydown)
  document.removeEventListener('fullscreenchange', handleDocumentFullscreenChange)
  removeHostFullscreenListener?.()
  removeHostFullscreenListener = undefined
  clearControlsHideTimer()
  if (isFullscreenMode.value) {
    document.body.style.overflow = bodyOverflowBeforeMaximize ?? ''
    bodyOverflowBeforeMaximize = null
  }
  if (shouldExitBrowserFullscreen) {
    void document.exitFullscreen().catch((error) => {
      console.error('Failed to exit browser fullscreen while closing player:', error)
    })
  } else if (shouldExitHostFullscreen) {
    void setHostWindowFullscreen(false).catch((error) => {
      console.error('Failed to exit Wails fullscreen while closing player:', error)
    })
  }
  playerSetupVersion++;
  window.clearTimeout(playerSetupTimer);
  if (player.value) {
    player.value.pause()
    player.value.dispose()
    player.value = null
  }
  if (nativeVideoRef.value) {
    nativeVideoRef.value.pause()
    nativeVideoRef.value.removeAttribute('src')
    nativeVideoRef.value.load()
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
  height: 100%;
  margin: 0 auto;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.video-player-windows {
  flex: 1;
  position: relative;
}

.custom-subtitle-display {
  position: absolute;
  right: 8%;
  bottom: 3%;
  left: 8%;
  z-index: 15;
  color: #fff;
  font-size: clamp(18px, 2.2vw, 34px);
  font-family: Arial, Helvetica, sans-serif;
  line-height: 1.35;
  text-align: center;
  white-space: pre-line;
  pointer-events: none;
  text-shadow:
    -2px -2px 2px #000,
    2px -2px 2px #000,
    -2px 2px 2px #000,
    2px 2px 2px #000;
  transition: bottom 0.2s ease;
}

/* Loading样式 */
.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  z-index: 10;
}

.loading-spinner {
  border: 4px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top: 4px solid #fff;
  width: 40px;
  height: 40px;
  animation: spin 1s linear infinite;
  margin-bottom: 15px;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }

  100% {
    transform: rotate(360deg);
  }
}

.loading-text {
  color: white;
  font-size: 16px;
  margin: 0;
}

/* 添加最大化模式样式 */
.video-player-container.fullscreen-mode {
  position: fixed;
  inset: 0;
  width: 100vw;
  height: 100vh;
  height: 100dvh;
  z-index: 9999;
  background-color: #000;
  display: flex;
  flex-direction: column;
}

.video-player-container.fullscreen-mode .video-js {
  flex: 1;
  width: 100%;
  height: 100%;
}

.video-player-container.fullscreen-mode .video-controller {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 20;
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.video-player-container.fullscreen-mode.controls-hidden {
  cursor: none;
}

.video-player-container.fullscreen-mode.controls-hidden .video-controller {
  opacity: 0;
  pointer-events: none;
  transform: translateY(100%);
}

.video-player-container.fullscreen-mode:not(.controls-hidden) .custom-subtitle-display {
  bottom: 80px;
}

.video-player-container.fullscreen-mode.controls-hidden .custom-subtitle-display {
  bottom: 24px;
}

/* 可选：自定义视频播放器样式 */
.video-js {
  background-color: #000;
  margin: 0 auto;
  width: 100% !important;
  height: 100% !important;

  video {
    /* 保证视频完整显示 */
    object-fit: contain;
  }
}

.native-video-player {
  width: 100% !important;
  height: 100% !important;

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

/* 隐藏video.js自带的loading spinner */
.video-js .vjs-loading-spinner {
  display: none !important;
}

/* 隐藏video.js自带的大播放按钮 */
.video-js .vjs-big-play-button {
  display: none !important;
}
</style>
