<template>
  <img v-if="!failed" :src="src" loading="lazy" alt="片段画面" @error="failed = true" />
  <span v-else class="cover-placeholder">暂无画面</span>
</template>
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { momentCoverImage } from '@/server/videoMoments.server'
const props = defineProps<{ momentId: string; time: number; refreshToken?: number }>()
const failed = ref(false)
const src = computed(() => momentCoverImage(props.momentId, props.time))
watch([src, () => props.refreshToken], () => {
  failed.value = false
})
</script>
<style scoped>
img,
.cover-placeholder {
  width: 144px;
  height: 82px;
  object-fit: contain;
  background: var(--el-fill-color-darker);
  border-radius: 7px;
}
.cover-placeholder {
  display: grid;
  place-items: center;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
</style>
