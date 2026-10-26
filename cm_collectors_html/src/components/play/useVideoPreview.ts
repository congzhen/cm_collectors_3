import { onActivated, onDeactivated, onBeforeUnmount, ref, watch, type Ref } from 'vue'
import { momentApi, subscribePreviewChanges, type VideoPreview } from '@/server/videoMoments.server'
export function useVideoPreview(id: Ref<string>) {
  const preview = ref<VideoPreview>()
  const error = ref('')
  let active = true
  let generation = 0
  let failures = 0
  let timer: ReturnType<typeof setTimeout> | undefined
  async function refresh(retry = false) {
    if (!retry) failures = 0
    clearTimeout(timer)
    if (!active) return
    const version = ++generation
    if (!id.value) {
      preview.value = undefined
      error.value = ''
      return
    }
    const result = await momentApi.preview(id.value)
    if (version !== generation) return
    if (result.status) {
      preview.value = result.data
      error.value = ''
      failures = 0
    } else {
      error.value = result.msg
      preview.value = undefined
      failures++
    }
    if (result.status && ['queued', 'processing'].includes(result.data.status))
      timer = setTimeout(() => refresh(true), 2500)
    else if (!result.status && failures <= 3)
      timer = setTimeout(() => refresh(true), 5000 * 2 ** (failures - 1))
  }
  watch(
    id,
    () => {
      preview.value = undefined
      error.value = ''
      void refresh()
    },
    { immediate: true },
  )
  onDeactivated(() => {
    active = false
    generation++
    clearTimeout(timer)
  })
  onActivated(() => {
    if (!active) {
      active = true
      void refresh()
    }
  })
  const unsubscribe = subscribePreviewChanges((changedId) => {
    if (!changedId || changedId === id.value) void refresh()
  })
  onBeforeUnmount(() => {
    unsubscribe()
    active = false
    generation++
    clearTimeout(timer)
  })
  return { preview, error, refresh: () => refresh() }
}
