<template>
  <el-button size="small" :icon="CollectionTag" :disabled="!filesBasesId" @click="opened = true"
    >片段收藏</el-button
  >
  <el-drawer
    v-model="opened"
    title="片段收藏"
    size="min(1040px, 92vw)"
    direction="rtl"
    body-class="library-moments-body"
    append-to-body
    destroy-on-close
  >
    <div class="library-moments-list">
      <MomentCollection
        v-if="opened && filesBasesId"
        :files-bases-id="filesBasesId"
        play-in-dialog
        @play="play"
      />
    </div>
  </el-drawer>
  <VideoPlayDialog ref="player" />
</template>
<script setup lang="ts">
import { ref, watch } from 'vue'
import { CollectionTag } from '@element-plus/icons-vue'
import { playUpdate } from '@/common/play'
import MomentCollection from './MomentCollection.vue'
import VideoPlayDialog from './videoPlayDialog.vue'
import type { VideoMoment } from '@/server/videoMoments.server'
const props = defineProps<{ filesBasesId: string }>()
const opened = ref(false)
const player = ref<InstanceType<typeof VideoPlayDialog>>()
function play(item: VideoMoment) {
  if (!item.available) return
  void playUpdate(item.resourceId, item.dramaSeriesId)
  player.value?.open(item.resourceId, item.dramaSeriesId, {
    start: item.start,
    end: item.end,
    title: item.title,
  })
}
watch(
  () => props.filesBasesId,
  () => {
    opened.value = false
    player.value?.close()
  },
)
</script>
<style scoped>
:global(.library-moments-body) {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.library-moments-list {
  flex: 1;
  min-height: 0;
  min-width: 0;
}
</style>
