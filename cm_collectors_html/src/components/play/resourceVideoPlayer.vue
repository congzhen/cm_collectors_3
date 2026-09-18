<template>
  <div class="resource-video-player">
    <MobileVideoPlayer v-if="mode === 'mobile'" :resource-id="resource.id" :drama-series-id="dramaSeriesId" :title="resource.title" />
    <VideoPlay v-else ref="desktop" force-desktop />
    <div v-if="error" role="alert">{{ error }} <button @click="loadDesktop">重试</button></div>
  </div>
</template>

<script setup lang="ts">
import { defineAsyncComponent, onBeforeUnmount, ref, watch, nextTick } from 'vue';
import VideoPlay from './videoPlay.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { appStoreData } from '@/storeData/app.storeData';
import { getPlayVideoURLAndType } from '@/common/play';
import { resolvePlayerMode } from '@/common/playerPreference';

const MobileVideoPlayer = defineAsyncComponent(() => import('./mobileVideoPlayer.vue'));
const props = defineProps<{ resource: I_resource; dramaSeriesId: string }>();
const emit = defineEmits<{ dimensions: [width: number, height: number] }>();
const desktop = ref<InstanceType<typeof VideoPlay>>();
// 每次打开确定一次策略，切集、旋转、全屏和后台配置更新都不重建播放器。
const mode = resolvePlayerMode(appStoreData().appConfig);
const error = ref('');
let version = 0;
const loadDesktop = async () => {
  const request = ++version;
  error.value = '';
  if (mode !== 'desktop' || !props.dramaSeriesId) return;
  await nextTick();
  const vp = desktop.value;
  if (!vp) return;
  const id = props.dramaSeriesId;
  const wasPlaying = vp.isPlaying();
  try {
    const { playUrl, playType } = await getPlayVideoURLAndType(id);
    if (request !== version) return;
    if (!playUrl) throw new Error('播放地址加载失败');
    const episode = props.resource.dramaSeries.find(item => item.id === id);
    vp.setVideoSource(playUrl, playType, () => {
      if (request !== version) return;
      vp.addTextTrack(`/api/video/subtitle/${encodeURIComponent(id)}`, '默认字幕', 'zh', true);
      const dimensions = vp.getVideoDimensions();
      if (dimensions) emit('dimensions', dimensions.width, dimensions.height);
      if (wasPlaying) vp.play();
    }, episode?.src || props.resource.title);
  } catch {
    if (request === version) error.value = '播放地址加载失败，请重试';
  }
};
watch(() => [props.resource.id, props.dramaSeriesId], loadDesktop, { immediate: true });
onBeforeUnmount(() => { version++; });
defineExpose({
  setAspectRatio: (ratio: string) => desktop.value?.setAspectRatio(ratio),
  getVideoDimensions: () => desktop.value?.getVideoDimensions(),
  getControllerHeight: () => desktop.value?.getControllerHeight() || 0,
});
</script>

<style scoped>
.resource-video-player { width: 100%; min-width: 0; }
</style>
