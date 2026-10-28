<template>
  <el-button size="small" :icon="CollectionTag" :disabled="!filesBasesId" @click="open">
    {{ defaultMode === 'theater' ? '片段放映室' : '片段收藏' }}
  </el-button>
  <component :is="mode === 'theater' ? ElDialog : ElDrawer" v-model="opened" :title="mode === 'theater' ? '片段放映室' : '片段收藏'"
    width="min(1500px, 96vw)" size="min(1040px, 92vw)" direction="rtl" top="4vh" class="library-moments-dialog"
    :body-class="mode === 'collection' ? 'library-moments-drawer-body' : ''"
    append-to-body destroy-on-close :close-on-click-modal="false" @close="onClose">
    <template #header>
      <div class="library-moments-header">
        <strong>{{ mode === 'theater' ? '片段放映室' : '片段收藏' }}</strong>
        <div class="moment-view-toggle">
          <span :class="{ 'is-active': mode === 'collection' }">收藏</span>
          <button type="button" class="moment-view-switch" role="switch"
            :disabled="savingMode" :aria-busy="savingMode"
            :aria-checked="mode === 'theater'" aria-label="放映室模式"
            :title="mode === 'theater' ? '切换到片段收藏' : '切换到片段放映室'"
            @click="switchMode"><span class="moment-view-switch-thumb" /></button>
          <span :class="{ 'is-active': mode === 'theater' }">放映室</span>
        </div>
      </div>
    </template>
    <div v-if="opened" class="library-moments-content" :class="{ 'is-theater': mode === 'theater', 'is-collection': mode === 'collection' }">
      <section v-if="mode === 'theater'" class="theater-stage" aria-label="片段播放器">
        <div v-if="resource && selected" class="theater-screen">
        <ResourceVideoPlayer :key="session" :resource="resource" fit-container
          :drama-series-id="selected.dramaSeriesId"
          :initial-moment="initialPlayback" @playing="recordPlayback" />
        </div>
        <div v-else class="theater-placeholder" :role="error ? 'alert' : 'status'">
          {{ loading ? '正在加载片段…' : error || '从片段列表选择一个片段开始放映' }}
          <el-button v-if="error && selected" size="small" @click="play(selected, autoplay)">重试</el-button>
        </div>
      </section>
      <div :ref="setListTarget" class="library-moments-list" />
    </div>
  </component>
  <Teleport v-if="opened && listTarget && filesBasesId" :to="listTarget">
        <MomentCollection :files-bases-id="filesBasesId"
          :layout="mode === 'theater' ? 'list' : 'grid'"
          :active-moment-id="mode === 'theater' ? selected?.id : undefined"
          play-in-dialog @play="play" @items-change="onItems" />
  </Teleport>
  <VideoPlayDialog ref="player" />
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, onDeactivated, ref, shallowRef, watch } from 'vue'
import { ElDialog, ElDrawer, ElMessage } from 'element-plus'
import { CollectionTag } from '@element-plus/icons-vue'
import { playUpdate } from '@/common/play'
import { appStoreData } from '@/storeData/app.storeData'
import { resourceServer } from '@/server/resource.server'
import { filesBasesServer } from '@/server/filesBases.server'
import { createDefaultConfigApp, E_config_type } from '@/dataType/config.dataType'
import type { I_resource } from '@/dataType/resource.dataType'
import type { VideoMoment } from '@/server/videoMoments.server'
import MomentCollection from './MomentCollection.vue'
import ResourceVideoPlayer from './resourceVideoPlayer.vue'
import VideoPlayDialog from './videoPlayDialog.vue'
const props = defineProps<{ filesBasesId: string }>()
const app = appStoreData()
const defaultMode = computed(() => app.currentFilesBasesAppConfig.videoMomentLibraryMode === 'collection' ? 'collection' : 'theater')
const mode = ref<'theater' | 'collection'>('theater')
const opened = ref(false)
const player = ref<InstanceType<typeof VideoPlayDialog>>()
const resource = ref<I_resource>()
const selected = ref<VideoMoment>()
const loading = ref(false)
const error = ref('')
const session = ref(0)
const autoplay = ref(false)
const initialPlayback = computed(() => selected.value ? {
  start: selected.value.start, end: selected.value.end, title: selected.value.title, autoplay: autoplay.value,
} : undefined)
let historyRecorded = false
let contextVersion = 0
function recordPlayback() {
  if (!selected.value || historyRecorded) return
  historyRecorded = true
  void playUpdate(selected.value.resourceId, selected.value.dramaSeriesId)
}
const savingMode = ref(false)
const listTarget = shallowRef<HTMLElement>()
let loadedItems: VideoMoment[] = []
function setListTarget(element: unknown) {
  if (element instanceof HTMLElement) listTarget.value = element
}
function onItems(items: VideoMoment[]) {
  loadedItems = items
  if (mode.value === 'theater' && !selected.value) {
    const first = items.find(item => item.available)
    if (first) void play(first, false)
  }
}
function onClose() {
  if (!opened.value) stopPlayback()
}
let version = 0
function stopPlayback() {
  version++
  contextVersion++
  resource.value = undefined
  selected.value = undefined
  loading.value = false
  error.value = ''
  player.value?.close()
}
function open() {
  stopPlayback()
  listTarget.value = undefined
  loadedItems = []
  mode.value = defaultMode.value
  opened.value = true
}
async function switchMode() {
  if (savingMode.value || !props.filesBasesId) return
  const libraryId = props.filesBasesId
  const context = contextVersion
  const nextMode = mode.value === 'theater' ? 'collection' : 'theater'
  savingMode.value = true
  try {
    // 从服务端读取最新配置，避免快捷切换覆盖其他入口保存的设置。
    const latest = await filesBasesServer.getConfigById(libraryId, E_config_type.app)
    if (!latest.status) throw new Error(latest.msg || '读取文件库设置失败')
    if (context !== contextVersion || libraryId !== props.filesBasesId || !opened.value) return
    const config = { ...createDefaultConfigApp(), ...JSON.parse(latest.data || '{}'), videoMomentLibraryMode: nextMode }
    const result = await filesBasesServer.setFilesBasesConfigById(libraryId, config)
    if (!result.status) {
      if (result.msg !== '已取消修改') ElMessage.error(result.msg || '保存片段默认视图失败')
      return
    }
    if (app.currentFilesBases.id === libraryId) app.currentConfigApp.videoMomentLibraryMode = nextMode
    if (context !== contextVersion || libraryId !== props.filesBasesId || !opened.value) return
    stopPlayback()
    mode.value = nextMode
    if (mode.value === 'theater') onItems(loadedItems)
  } catch (cause) {
    ElMessage.error(cause instanceof Error ? cause.message : '保存片段默认视图失败，请重试')
  } finally {
    savingMode.value = false
  }
}
async function play(item: VideoMoment, shouldAutoplay = true) {
  if (!item.available) return
  if (mode.value === 'collection') {
    void playUpdate(item.resourceId, item.dramaSeriesId)
    player.value?.open(item.resourceId, item.dramaSeriesId, { start: item.start, end: item.end, title: item.title })
    return
  }
  const request = ++version
  historyRecorded = false
  autoplay.value = shouldAutoplay
  selected.value = item
  resource.value = undefined
  error.value = ''
  loading.value = true
  try {
    const result = await resourceServer.info(item.resourceId)
    if (request !== version || !opened.value) return
    if (!result.status) throw new Error(result.msg || '资源加载失败')
    if (!result.data.dramaSeries.some(episode => episode.id === item.dramaSeriesId)) {
      throw new Error('该片段的源视频已移除，请刷新列表')
    }
    session.value++
    resource.value = result.data
  } catch (cause) {
    if (request === version) error.value = cause instanceof Error ? cause.message : '片段加载失败，请重试'
  } finally {
    if (request === version) loading.value = false
  }
}
watch(() => props.filesBasesId, () => {
  opened.value = false
  stopPlayback()
})
onBeforeUnmount(stopPlayback)
onDeactivated(() => {
  opened.value = false
  stopPlayback()
})
</script>
<style scoped>
.library-moments-header { display: flex; align-items: center; gap: 16px; padding-right: 30px; }
.moment-view-toggle { display: flex; align-items: center; gap: 7px; font-size: 12px; color: var(--el-text-color-secondary); }
.moment-view-toggle .is-active { color: var(--el-text-color-primary); }
.moment-view-switch { display: inline-flex; align-items: center; flex-shrink: 0; width: 30px; height: 18px; padding: 2px; border: 0; border-radius: 12px; background: var(--el-border-color-darker); cursor: pointer; transition: background .15s; }
.moment-view-switch[aria-checked="true"] { background: var(--el-color-primary); }
.moment-view-switch-thumb { width: 14px; height: 14px; border-radius: 50%; background: #fff; transition: transform .15s; }
.moment-view-switch[aria-checked="true"] .moment-view-switch-thumb { transform: translateX(12px); }
.moment-view-switch:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: 3px; }
.moment-view-switch:disabled { cursor: wait; opacity: .6; }
.library-moments-content { height: 80vh; min-height: 0; }
:global(.library-moments-drawer-body) { display: flex; flex-direction: column; min-height: 0; overflow: hidden; }
.library-moments-content.is-collection { flex: 1; height: 100%; display: flex; flex-direction: column; }
.is-collection .library-moments-list { flex: 1; height: auto; }
.is-theater { display: grid; grid-template-columns: minmax(0, 1fr) minmax(340px, 420px); gap: 20px; }
.theater-stage { min-width: 0; min-height: 0; overflow: hidden; border-radius: 12px; }
.theater-screen { height: 100%; background: #000; }
.theater-screen :deep(.resource-video-player) { height: 100%; }
.theater-screen :deep(.phone-player-stage) { height: 100%; aspect-ratio: auto; }
.theater-placeholder { min-height: 340px; display: flex; align-items: center; justify-content: center; flex-direction: column; gap: 14px; padding: 24px; text-align: center; color: var(--el-text-color-secondary); }
.library-moments-list { height: 100%; min-height: 0; min-width: 0; }
@media (max-width: 900px) {
  .is-theater { display: flex; flex-direction: column; gap: 12px; }
  .theater-stage { flex: 0 0 auto; max-height: 45%; }
  .theater-placeholder { min-height: 150px; }
  .theater-screen { height: clamp(180px, 30vh, 300px); }
  .is-theater .library-moments-list { flex: 1; }
}
</style>
