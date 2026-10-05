<template>
  <div ref="stageRef" class="phone-player-stage">
    <div ref="hostRef" class="phone-player-host" />
    <Teleport v-if="playerRoot" :to="playerRoot">
      <div v-if="subtitleText" class="phone-subtitle">{{ subtitleText }}</div>
      <div v-if="displayMessage" class="phone-player-message" role="status">
        <span v-if="isLoading" class="phone-loading-spinner" aria-hidden="true" />
        <span>{{ displayMessage }}</span>
        <button v-if="failed" type="button" @click.stop="retry">重试</button>
        <button v-if="needsPlay" type="button" @click.stop="continuePlay">继续播放</button>
      </div>
    </Teleport>
    <div v-if="!playerRoot && displayMessage" class="phone-player-message" role="status">
      <span v-if="isLoading" class="phone-loading-spinner" aria-hidden="true" />
        <span>{{ displayMessage }}</span><button v-if="failed" type="button" @click="retry">重试</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, shallowRef, watch } from 'vue';
import Player, { Events } from 'xgplayer';
import HlsPlugin from 'xgplayer-hls';
import 'xgplayer/dist/index.min.css';
import { MobileSourceError, resolveMobileSource, type MobileSource } from '@/common/mobilePlaybackSource';
import { readMobilePlayback, readMobileSound, resumeTime, saveMobilePlayback, saveMobileSound } from '@/common/mobileSession';
import { parseMobileVtt } from '@/common/mobileSubtitles';
import { downloadMobileShell, notifyMobileShell } from '@/common/mobileShell';

const props = withDefaults(defineProps<{ resourceId: string; dramaSeriesId: string; title: string; autoplay?: boolean;
  sourceLoader?: (id: string, signal?: AbortSignal) => Promise<MobileSource> }>(), { sourceLoader: resolveMobileSource });
const stageRef = ref<HTMLElement>();
const hostRef = ref<HTMLElement>();
const playerRoot = shallowRef<HTMLElement>();
const loadingText = '正在加载中…';
const message = ref('');
const playbackStalled = ref(false);
const displayMessage = computed(() => message.value || (playbackStalled.value ? loadingText : ''));
const isLoading = computed(() => displayMessage.value === loadingText);
const failed = ref(false);
const needsPlay = ref(false);
const full = ref(false);
const subtitleText = ref('');
let player: Player | undefined;
let version = 0;
let activeId = '';
let activeResourceId = '';
let ready = false;
let wantedPlaying = false;
let hidden = document.hidden;
let retries = 0;
let retryTimer: number | undefined;
let stallTimer: number | undefined;
let progressTimer: number | undefined;
let lastSave = 0;
let savedOverflow = '';
let browserFullscreen = false;
let disposed = false;
let inactive = false;
let resumePosition = 0;
let cues: ReturnType<typeof parseMobileVtt> = [];
let subtitleAbort: AbortController | undefined;
let sourceAbort: AbortController | undefined;
let retryableFailure = false;
let watchBuffering: (() => void) | undefined;
let syncLoadedData: (() => void) | undefined;
const playbackId = () => `${activeResourceId}/${activeId}`;

const saveProgress = () => {
  if (!player || !activeId || !ready) return;
  saveMobilePlayback(playbackId(), {
    time: Math.max(0, player.currentTime || 0), duration: player.duration || 0,
    rate: player.playbackRate || 1, playing: wantedPlaying,
  });
};
const continuePlay = async () => {
  const instance = player;
  if (!instance || disposed || inactive) return;
  wantedPlaying = true;
  try {
    await instance.play();
    if (instance !== player) return;
    needsPlay.value = false;
    // 等待 PLAYING 确认真正开始播放后再隐藏提示。
  } catch {
    if (instance !== player) return;
    needsPlay.value = true;
    message.value = '已恢复播放位置，点击继续播放';
  }
};

const clearFullscreenLayout = () => {
  if (player?.isRotateFullscreen) player.exitRotateFullscreen();
  else if (player?.cssfullscreen) player.exitCssFullscreen();
  if (player?.root?.classList.contains('phone-fallback-rotate')) {
    player.root.classList.remove('phone-fallback-rotate');
    player.root.style.removeProperty('--phone-full-width');
    player.root.style.removeProperty('--phone-full-height');
    player.rotateDeg = 0;
  }
};
const exitFullscreen = async () => {
  if (!full.value) return;
  full.value = false;
  notifyMobileShell(false);
  browserFullscreen = false;
  clearFullscreenLayout();
  document.body.style.overflow = savedOverflow;
  try { screen.orientation?.unlock?.(); } catch { /* 旧 WebView 可能禁止方向操作 */ }
  if (document.fullscreenElement === stageRef.value) await document.exitFullscreen().catch(() => undefined);
  updateGestures();
};
const updateGestures = () => {
  const mobile = player?.getPlugin('mobile');
  if (mobile) mobile.config.gestureX = full.value;
};
const applyFullscreenLayout = () => {
  if (!player || !full.value) return;
  clearFullscreenLayout();
  const video = player.video as HTMLVideoElement;
  if (video.videoWidth > video.videoHeight && window.innerHeight > window.innerWidth) {
    if (screen.orientation) player.getRotateFullscreen();
    else {
      // xgplayer 的旋转方法直接读取 screen.orientation.angle；旧内核改用 CSS 全屏旋转。
      player.getCssFullscreen();
      player.root?.classList.add('phone-fallback-rotate');
      player.root?.style.setProperty('--phone-full-width', `${window.innerHeight}px`);
      player.root?.style.setProperty('--phone-full-height', `${window.innerWidth}px`);
      player.rotateDeg = 90;
    }
  } else player.getCssFullscreen();
  updateGestures();
};
const toggleFullscreen = async () => {
  if (full.value) { await exitFullscreen(); return; }
  if (!player || disposed) return;
  const instance = player;
  savedOverflow = document.body.style.overflow;
  full.value = true;
  document.body.style.overflow = 'hidden';
  // 全屏 API 不可用的套壳仍使用西瓜播放器提供的网页横屏。
  const landscape = (instance.video as HTMLVideoElement).videoWidth > (instance.video as HTMLVideoElement).videoHeight;
  const shellHandlesFullscreen = notifyMobileShell(true, landscape);
  if (!shellHandlesFullscreen) {
    try { await stageRef.value?.requestFullscreen?.(); } catch { /* 使用网页全屏 */ }
    browserFullscreen = document.fullscreenElement === stageRef.value;
  }
  if (disposed || player !== instance || !full.value) return;
  if (landscape && !shellHandlesFullscreen) {
    const orientation = screen.orientation as ScreenOrientation & { lock?: (value: string) => Promise<void> };
    try { await orientation?.lock?.('landscape'); } catch { /* 使用旋转布局 */ }
  }
  if (!disposed && player === instance && full.value) applyFullscreenLayout();
};
const onFullscreenChange = () => {
  if (browserFullscreen && !document.fullscreenElement && full.value) void exitFullscreen();
};
const onKeydown = (event: KeyboardEvent) => {
  if (full.value && event.key === 'Escape') { event.preventDefault(); void exitFullscreen(); }
};

const clearTimers = () => {
  window.clearTimeout(retryTimer);
  window.clearTimeout(stallTimer);
};
const recover = () => {
  if (hidden || disposed || inactive || !activeId || retries >= 3) return;
  window.clearTimeout(retryTimer);
  retryTimer = window.setTimeout(() => {
    retries++;
    void load(true);
  }, 1000 * 2 ** retries);
};
const reportFailure = (retryable: boolean) => {
  if (disposed) return;
  window.clearTimeout(stallTimer);
  failed.value = true;
  retryableFailure = retryable;
  message.value = navigator.onLine === false ? '网络已断开，播放位置已保留' :
    retryable ? '连接暂不可用，播放位置已保留' : '视频无法播放，请检查文件、格式或访问权限';
  player?.focus({ autoHide: false });
  if (retryable) recover();
};
const retry = () => { retries = 0; void load(true); };
const destroyPlayer = async () => {
  clearTimers();
  watchBuffering = undefined;
  syncLoadedData = undefined;
  window.clearInterval(progressTimer);
  playbackStalled.value = false;
  subtitleAbort?.abort();
  const previous = player;
  const exiting = exitFullscreen();
  player = undefined;
  ready = false;
  playerRoot.value = undefined;
  await nextTick();
  previous?.destroy();
  await exiting;
};
const load = async (recovering = false) => {
  if (!props.dramaSeriesId || disposed) return;
  saveProgress();
  const request = ++version;
  sourceAbort?.abort();
  const target = props.dramaSeriesId;
  const shouldPlay = recovering ? wantedPlaying : Boolean(props.autoplay);
  const previous = readMobilePlayback(`${props.resourceId}/${target}`);
  const position = recovering && player && ready ? player.currentTime : resumeTime(previous);
  await destroyPlayer();
  if (request !== version || disposed) return;
  activeId = target;
  activeResourceId = props.resourceId;
  resumePosition = position;
  wantedPlaying = shouldPlay && !inactive;
  failed.value = false;
  retryableFailure = false;
  needsPlay.value = false;
  subtitleText.value = '';
  cues = [];
  message.value = recovering ? '正在重新连接…' : loadingText;
  sourceAbort = new AbortController();
  const controller = sourceAbort;
  const timeout = window.setTimeout(() => controller.abort(), 15000);
  try {
    const source = await props.sourceLoader(target, controller.signal);
    window.clearTimeout(timeout);
    if (request !== version || disposed) return;
    if (!source.playUrl) throw new Error('播放地址暂不可用');
    const host = hostRef.value;
    if (!host) return;
    const mount = document.createElement('div');
    host.replaceChildren(mount);
    const nativeHls = document.createElement('video').canPlayType('application/vnd.apple.mpegurl');
    const hls = source.playType === 'm3u8' && !nativeHls;
    if (hls && !HlsPlugin.isSupported()) {
      message.value = '当前浏览器不支持此视频的播放格式'; failed.value = true; return;
    }
    const sound = readMobileSound();
    const instance = new Player({
      el: mount, url: source.playUrl, width: '100%', height: '100%', lang: 'zh-cn',
      autoplay: false, videoInit: true, playsinline: true, preload: 'metadata',
      volume: sound.volume, videoConfig: { muted: sound.muted },
      isMobileSimulateMode: 'mobile', seekedStatus: 'auto', startTime: resumePosition,
      defaultPlaybackRate: previous?.rate || 1, closePauseVideoFocus: false,
      playbackRate: [0.5, 0.75, 1, 1.25, 1.5, 2],
      download: true,
      fullscreen: { switchCallback: toggleFullscreen },
      mobile: { gestureX: false, gestureY: false, disablePress: true, isTouchingSeek: false, moveDuration: 90000 },
      plugins: hls ? [HlsPlugin] : [],
      hls: { retryCount: 2, retryDelay: 1000, loadTimeout: 60000 },
      ignores: ['error', 'pc', 'keyboard'],
    });
    player = instance;
    let lastDownloadClick = -Infinity;
    instance.usePluginHooks('download', 'click', () => {
      if (!(window as Window & { __cmPhoneShellDownload?: boolean }).__cmPhoneShellDownload) return true;
      // 同一次触摸可能同时触发 touchend 和 click，避免建立两个原生下载任务。
      const time = Date.now();
      if (time - lastDownloadClick < 600) return false;
      lastDownloadClick = time;
      if (source.playType === 'm3u8') {
        window.alert('当前视频使用分片播放，暂不支持下载完整视频');
        return false;
      }
      return !downloadMobileShell(source.playUrl, `${props.title || '视频'}.mp4`);
    });
    const media = instance.video as HTMLVideoElement;
    instance.muted = sound.muted;
    playerRoot.value = instance.root || undefined;
    const current = () => !disposed && !inactive && player === instance && request === version;
    // 与原 videoPlay 的移动预览逻辑一致：定位到开头附近，只解码预览、不调用 play。
    let previewInitialized = false;
    const previewFirstFrame = () => {
      if (previewInitialized || !current()) return;
      previewInitialized = true;
      if (resumePosition > 0 || wantedPlaying || !instance.paused) return;
      if (Number.isFinite(instance.duration) && instance.duration > 0 && instance.currentTime === 0) {
        try { instance.currentTime = Math.min(0.01, instance.duration / 2); }
        catch { /* 个别 WebView 暂不支持定位时保留默认预览，不影响手动播放。 */ }
      }
    };
    instance.on(Events.LOADED_METADATA, previewFirstFrame);
    if (media.readyState >= 1) previewFirstFrame();
    // 声音是手机播放器的共用偏好，不随资源或分集切换重置。
    instance.on(Events.VOLUME_CHANGE, () => {
      if (current()) saveMobileSound({ muted: instance.muted, volume: instance.volume });
    });
    const onLoadedData = () => {
      if (!current()) return;
      ready = true;
      failed.value = false;
      retryableFailure = false;
      window.clearTimeout(stallTimer);
      if (resumePosition > 0 && Number.isFinite(instance.duration)) {
        instance.currentTime = Math.min(resumePosition, Math.max(0, instance.duration - 1));
      }
      resumePosition = 0;
      message.value = wantedPlaying ? loadingText : '';
      if (wantedPlaying && !hidden) void continuePlay();
      instance.focus({ autoHide: false });
    };
    syncLoadedData = onLoadedData;
    instance.on(Events.LOADED_DATA, onLoadedData);
    instance.on(Events.PLAY, () => {
      if (!current()) return;
      wantedPlaying = true;
      needsPlay.value = false;
      if (!failed.value && media.readyState < 3) message.value = loadingText;
    });
    instance.on(Events.PLAYING, () => {
      if (!current()) return;
      playbackStalled.value = false;
      retries = 0; failed.value = false; retryableFailure = false; message.value = ''; clearTimers();
    });
    instance.on(Events.PAUSE, () => {
      if (!current()) return;
      playbackStalled.value = false;
      if (!document.hidden && !(failed.value && retryableFailure)) wantedPlaying = false;
      window.clearTimeout(stallTimer);
      if (!document.hidden && ready && !media.seeking && isLoading.value) message.value = '';
      saveProgress();
    });
    instance.on(Events.ENDED, () => {
      if (!current()) return;
      wantedPlaying = false;
      playbackStalled.value = false;
      failed.value = false;
      retryableFailure = false;
      needsPlay.value = false;
      message.value = '';
      clearTimers();
      saveProgress();
    });
    instance.on(Events.TIME_UPDATE, () => {
      if (!current()) return;
      const time = instance.currentTime;
      subtitleText.value = cues.filter(cue => time >= cue.start && time <= cue.end).map(cue => cue.text).join('\n');
      if (Date.now() - lastSave > 3000) { lastSave = Date.now(); saveProgress(); }
    });
    instance.on(Events.SEEKED, saveProgress);
    instance.on(Events.RATE_CHANGE, saveProgress);
    instance.on(Events.ERROR, () => {
      if (!current()) return;
      const code = (instance.video as HTMLVideoElement).error?.code;
      reportFailure(code !== 3 && code !== 4);
    });
    // 慢取流/解码不等于连接失败。观察缓冲变化，超时只提供手动重试，
    // 避免大文件拖动时不断销毁播放器、重复请求同一位置。
    let lastProgressAt = Date.now();
    let lastBuffer = '';
    const bufferSnapshot = () => {
      const ranges = media.buffered;
      const parts = [String(instance.currentTime), String(media.readyState)];
      for (let i = 0; i < ranges.length; i++) parts.push(`${ranges.start(i)}:${ranges.end(i)}`);
      return parts.join('|');
    };
    const checkBuffering = () => {
      window.clearTimeout(stallTimer);
      if (!current() || hidden || retryableFailure || media.error) return;
      const snapshot = bufferSnapshot();
      if (snapshot !== lastBuffer) {
        lastBuffer = snapshot;
        lastProgressAt = Date.now();
      }
      if (Date.now() - lastProgressAt >= 60000) {
        // 不设置连接失败，避免切回前台或网络事件触发无谓重建。
        failed.value = true;
        message.value = '视频加载较慢，可继续等待或手动重试';
      }
      stallTimer = window.setTimeout(checkBuffering, 5000);
    };
    const buffering = () => {
      if (!current() || hidden) return;
      lastProgressAt = Date.now();
      lastBuffer = bufferSnapshot();
      if (!failed.value && !needsPlay.value) message.value = loadingText;
      checkBuffering();
    };
    watchBuffering = buffering;
    instance.on(Events.WAITING, () => { if (wantedPlaying) buffering(); });
    // STALLED 也可能发生在暂停且已有预览画面时，不应把它当作失败。
    instance.on(Events.STALLED, () => {
      if (wantedPlaying || !ready) {
        if (current() && !hidden && media.readyState < 3 && !failed.value && !needsPlay.value) message.value = loadingText;
        checkBuffering();
      }
    });
    instance.on(Events.SEEKING, () => {
      if (!current()) return;
      clearTimers();
      failed.value = false;
      retryableFailure = false;
      saveProgress();
      buffering();
    });
    instance.on(Events.SEEKED, () => {
      if (!current()) return;
      if (media.readyState >= (wantedPlaying ? 3 : 2)) {
        window.clearTimeout(stallTimer);
        failed.value = false;
        message.value = '';
      }
    });
    // 部分 WebView 缓冲时只改变播放器加载样式，或者不再发出 waiting。
    // 同步真实播放器状态，并以播放时间停滞兜底，不重连、不改变播放状态。
    let observedTime = instance.currentTime;
    let advancedAt = Date.now();
    progressTimer = window.setInterval(() => {
      if (!current() || hidden || !wantedPlaying || instance.paused || media.ended || retryableFailure || media.error || needsPlay.value) {
        playbackStalled.value = false;
        observedTime = instance.currentTime;
        advancedAt = Date.now();
        return;
      }
      const time = instance.currentTime;
      if (time !== observedTime) {
        observedTime = time;
        advancedAt = Date.now();
        if (!media.seeking && media.readyState >= 3) {
          // 慢加载提示不是媒体错误；部分 WebView 恢复后只更新进度、不再发 playing。
          if (failed.value || message.value === loadingText) message.value = '';
          failed.value = false;
          clearTimers();
        }
      }
      playbackStalled.value = Boolean(instance.root?.classList.contains('xgplayer-isloading')) || Date.now() - advancedAt >= 2500;
    }, 500);
    instance.on(Events.LOADING, () => {
      if (current() && !hidden && wantedPlaying && !failed.value && !needsPlay.value) message.value = loadingText;
    });
    checkBuffering();
    subtitleAbort = new AbortController();
    void fetch(`/api/video/subtitle/${encodeURIComponent(target)}`, { signal: subtitleAbort.signal })
      .then(async response => {
        if (!response.ok) return;
        const text = await response.text();
        if (current()) cues = parseMobileVtt(text);
      }).catch(() => undefined);
  } catch (error) {
    if (request === version && !disposed) {
      reportFailure(error instanceof MobileSourceError ? error.retryable : true);
      if (error instanceof MobileSourceError && !error.retryable) message.value = error.message;
    }
  } finally { window.clearTimeout(timeout); }
};
const onVisibility = () => {
  if (inactive) return;
  hidden = document.hidden;
  if (hidden) { saveProgress(); clearTimers(); }
  else if (failed.value && retryableFailure) { retries = 0; void load(true); }
  else {
    // 缓存页停用期间可能已经收到首帧事件，激活时补齐就绪状态。
    if (player && !ready && (player.video as HTMLVideoElement).readyState >= 2) syncLoadedData?.();
    if (player && (!ready || (wantedPlaying && (player.video as HTMLVideoElement).readyState < 3))) watchBuffering?.();
    if (wantedPlaying && player?.paused) void continuePlay();
  }
};
const onOnline = () => { if (!inactive && !hidden && failed.value && retryableFailure) retry(); };
const onOffline = () => { if (!inactive) { saveProgress(); reportFailure(true); } };
const onPageShow = () => { if (!document.hidden) onVisibility(); };
const onShellExit = () => { void exitFullscreen(); };
const onShellQuery = () => {
  if (full.value && player) {
    const video = player.video as HTMLVideoElement;
    notifyMobileShell(true, video.videoWidth > video.videoHeight);
  }
};
const pause = () => {
  wantedPlaying = false;
  player?.pause();
  saveProgress();
};
defineExpose({ isPlaying: () => wantedPlaying, pause });
onDeactivated(() => { inactive = true; pause(); clearTimers(); void exitFullscreen(); });
onActivated(() => {
  if (!inactive) return;
  inactive = false;
  onVisibility();
});
watch(() => [props.resourceId, props.dramaSeriesId], () => { retries = 0; void load(); });
onMounted(() => {
  void load();
  document.addEventListener('visibilitychange', onVisibility);
  document.addEventListener('fullscreenchange', onFullscreenChange);
  window.addEventListener('pagehide', saveProgress);
  window.addEventListener('pageshow', onPageShow);
  window.addEventListener('online', onOnline);
  window.addEventListener('offline', onOffline);
  window.addEventListener('resize', applyFullscreenLayout);
  window.addEventListener('keydown', onKeydown);
  window.addEventListener('cm-phone-exit-fullscreen', onShellExit);
  window.addEventListener('cm-phone-query-fullscreen', onShellQuery);
});
onBeforeUnmount(() => {
  saveProgress(); disposed = true; version++;
  sourceAbort?.abort();
  void destroyPlayer();
  document.removeEventListener('visibilitychange', onVisibility);
  document.removeEventListener('fullscreenchange', onFullscreenChange);
  window.removeEventListener('pagehide', saveProgress);
  window.removeEventListener('pageshow', onPageShow);
  window.removeEventListener('online', onOnline);
  window.removeEventListener('offline', onOffline);
  window.removeEventListener('resize', applyFullscreenLayout);
  window.removeEventListener('keydown', onKeydown);
  window.removeEventListener('cm-phone-exit-fullscreen', onShellExit);
  window.removeEventListener('cm-phone-query-fullscreen', onShellQuery);
});
</script>

<style scoped>
.phone-player-stage { width: 100%; aspect-ratio: 16 / 9; background: #000; position: relative; }
.phone-player-stage:fullscreen { width: 100%; height: 100%; aspect-ratio: auto; }
.phone-player-host { width: 100%; height: 100%; }
.phone-subtitle { position: absolute; bottom: 68px; left: 5%; width: 90%; text-align: center; white-space: pre-line; pointer-events: none; color: white; text-shadow: 0 1px 3px #000; font-size: clamp(16px, 3vw, 25px); z-index: 15; }
.phone-player-message { position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); width: max-content; max-width: 90%; border-radius: 8px; display: flex; justify-content: center; align-items: center; gap: 12px; padding: 10px; background: #000b; color: #fff; z-index: 20; font-size: 14px; }
.phone-loading-spinner { width: 16px; height: 16px; flex: 0 0 16px; border: 2px solid #ffffff40; border-top-color: #fff; border-radius: 50%; animation: phone-loading-spin .8s linear infinite; }
@keyframes phone-loading-spin { to { transform: rotate(360deg); } }
@media (prefers-reduced-motion: reduce) { .phone-loading-spinner { animation: none; } }
.phone-player-message { pointer-events: none; }
.phone-player-message button { pointer-events: auto; background: #252c34; color: #fff; border: 1px solid #ffffff60; border-radius: 6px; padding: 10px 14px; cursor: pointer; }
:deep(.xgplayer-controls) { padding-bottom: env(safe-area-inset-bottom); }
:deep(.phone-fallback-rotate) {
  width: var(--phone-full-width) !important; height: var(--phone-full-height) !important;
  top: 50% !important; left: 50% !important;
  transform: translate(-50%, -50%) rotate(90deg) !important; transform-origin: center !important;
}
</style>
