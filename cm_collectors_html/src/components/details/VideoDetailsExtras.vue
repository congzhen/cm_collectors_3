<template>
  <section v-if="resource.mode === 'movies'" class="video-extras">
    <MomentCollection
      :resource-id="resource.id"
      :play-mode="resolvedPlayMode === 'page' ? 'page' : 'emit'"
      :layout="layout"
      :show-heading="showHeading"
      :show-resource-title="showResourceTitle"
      :disabled="configLoading"
      @play="play"
    />
    <VideoPlayDialog v-if="resolvedPlayMode === 'dialog'" ref="player" />
  </section>
</template>
<script setup lang="ts">
import type { I_resource } from '@/dataType/resource.dataType'
import MomentCollection from '@/components/play/MomentCollection.vue'
import { computed, onActivated, onDeactivated, onBeforeUnmount, ref, watch } from 'vue'
import VideoPlayDialog from '@/components/play/videoPlayDialog.vue'
import { appStoreData } from '@/storeData/app.storeData'
import { filesBasesServer } from '@/server/filesBases.server'
import { playUpdate } from '@/common/play'
import type { VideoMoment } from '@/server/videoMoments.server'

const props = withDefaults(defineProps<{
  resource: I_resource
  playMode?: 'settings' | 'dialog' | 'page' | 'emit'
  layout?: 'grid' | 'list'
  showHeading?: boolean
  showResourceTitle?: boolean
}>(), { playMode: 'settings', layout: 'grid', showHeading: true, showResourceTitle: false })
const emit = defineEmits<{ play: [moment: VideoMoment] }>()
const app = appStoreData()
const player = ref<InstanceType<typeof VideoPlayDialog>>()
const libraryMode = ref<'dialog' | 'page'>('dialog')
const configLoading = ref(false)
const normalizeMode = (value: unknown) => value === 'page' ? 'page' : 'dialog'
const usesCurrentLibrary = computed(() =>
  !props.resource.filesBases_id || props.resource.filesBases_id === app.currentFilesBases?.id,
)
const resolvedPlayMode = computed(() => props.playMode !== 'settings' ? props.playMode
  : usesCurrentLibrary.value ? normalizeMode(app.currentFilesBasesAppConfig?.videoMomentOpenMode)
    : libraryMode.value)
let configVersion = 0
let inactive = false
async function loadMode() {
  const version = ++configVersion
  player.value?.close()
  libraryMode.value = 'dialog'
  configLoading.value = false
  if (inactive || props.playMode !== 'settings' || usesCurrentLibrary.value) return
  configLoading.value = true
  // Cross-library details follow the resource's own saved preference.
  try {
    const result = await filesBasesServer.infoById(props.resource.filesBases_id)
    if (version !== configVersion || !result.status) return
    const config = JSON.parse(result.data.filesBasesSetting.config_json_data || '{}')
    libraryMode.value = normalizeMode(config.videoMomentOpenMode)
  } catch {
    // Old or invalid settings retain the default dialog player.
  } finally {
    if (version === configVersion) configLoading.value = false
  }
}
watch(
  () => [props.resource.id, props.resource.filesBases_id, props.playMode, usesCurrentLibrary.value],
  loadMode,
  { immediate: true },
)
function play(item: VideoMoment) {
  if (inactive || !item.available || configLoading.value) return
  if (resolvedPlayMode.value === 'emit') {
    emit('play', item)
    return
  }
  void playUpdate(item.resourceId, item.dramaSeriesId)
  if (resolvedPlayMode.value === 'page') return
  player.value?.open(item.resourceId, item.dramaSeriesId, {
    start: item.start, end: item.end, title: item.title,
  })
}
onDeactivated(() => {
  inactive = true
  configVersion++
  configLoading.value = false
  player.value?.close()
})
onActivated(() => {
  if (inactive) {
    inactive = false
    void loadMode()
  }
})
onBeforeUnmount(() => { configVersion++ })
</script>
<style scoped>
.video-extras {
  min-width: 0;
}
</style>
