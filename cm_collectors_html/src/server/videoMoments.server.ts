import request from '@/assets/request'
export interface MomentPlayback {
  autoplay?: boolean
  start: number
  end: number
  title: string
}
export interface VideoMoment {
  id: string
  dramaSeriesId: string
  resourceId: string
  title: string
  note: string
  start: number
  end: number
  available: boolean
  sourceChanged: boolean
  resourceTitle: string
}
export interface VideoPreview {
  dramaSeriesId: string
  sourceKey: string
  status: string
  error: string
  interval: number
  duration: number
  count: number
  width?: number
  completed: number
}
const previewListeners = new Set<(id?: string) => void>()
const momentListeners = new Set<() => void>()
export function subscribeMomentChanges(listener: () => void) {
  momentListeners.add(listener)
  return () => {
    momentListeners.delete(listener)
  }
}
function notifyMomentChanged() {
  momentListeners.forEach((listener) => listener())
}
export function notifyPreviewChanged(id?: string) {
  previewListeners.forEach((listener) => listener(id))
}
export function subscribePreviewChanges(listener: (id?: string) => void) {
  previewListeners.add(listener)
  return () => {
    previewListeners.delete(listener)
  }
}
export const momentApi = {
  list: (resourceId = '', page = 1, dramaSeriesId = '') =>
    request<VideoMoment[]>({ url: '/videoMoments', params: { resourceId, page, dramaSeriesId } }),
  libraryList: (filesBasesId: string, page = 1, keyword = '') =>
    request<{ items: VideoMoment[]; total: number }>({
      url: '/videoMoments/library',
      params: { filesBasesId, page, keyword },
    }),
  save: async (data: Partial<VideoMoment>) => {
    const result = await request<VideoMoment>({
      url: `/videoMoments${data.id ? '/' + encodeURIComponent(data.id) : ''}`,
      method: data.id ? 'put' : 'post',
      data,
    })
    if (result.status) {
      notifyPreviewChanged(data.dramaSeriesId)
      notifyMomentChanged()
    }
    return result
  },
  remove: async (id: string) => {
    const result = await request<boolean>({
      url: `/videoMoments/${encodeURIComponent(id)}`,
      method: 'delete',
    })
    if (result.status) notifyMomentChanged()
    return result
  },
  preview: (id: string) =>
    request<VideoPreview>({ url: `/videoPreviews/${encodeURIComponent(id)}` }),
  queue: async (id: string, interval: number) => {
    const result = await request<VideoPreview>({
      url: `/videoPreviews/${encodeURIComponent(id)}`,
      method: 'post',
      data: { interval },
    })
    if (result.status) notifyPreviewChanged(id)
    return result
  },
}
export const previewImage = (v: VideoPreview, time: number) => {
  if (!Number.isFinite(time) || v.interval <= 0 || v.count <= 0) return ''
  const index = Math.min(v.count - 1, Math.max(0, Math.floor(time / v.interval)))
  // Never show an unrelated frame for a position that has not been generated yet.
  if (index >= v.completed || v.status === 'none') return ''
  return `/api/videoPreviews/${encodeURIComponent(v.dramaSeriesId)}/image/${index}?v=${v.sourceKey}&interval=${v.interval}&width=${v.width || 240}`
}
export const momentDuration = (start: number, end: number) =>
  Number.isFinite(end - start) && end > start ? `共 ${Number((end - start).toFixed(1))} 秒` : ''

export const momentTime = (time: number) => {
  const n = Math.max(0, Math.floor(time || 0))
  return `${Math.floor(n / 3600)
    .toString()
    .padStart(2, '0')}:${Math.floor((n / 60) % 60)
    .toString()
    .padStart(2, '0')}:${(n % 60).toString().padStart(2, '0')}`
}

// One attempt per source in a player session; transient failures may retry on the next play.
export function createPlaybackPreviewTrigger() {
  let requestedId = ''
  return async (id: string) => {
    if (!id || requestedId === id) return
    requestedId = id
    try {
      const result = await request<boolean>({
        url: `/videoPreviews/${encodeURIComponent(id)}/playback`,
        method: 'post',
      })
      if (!result.status) {
        if (requestedId === id) requestedId = ''
        return
      }
      if (result.data) notifyPreviewChanged(id)
    } catch {
      if (requestedId === id) requestedId = ''
    }
  }
}

export const momentCoverImage = (id: string, time: number) =>
  `/api/videoMoments/${encodeURIComponent(id)}/cover?t=${time}`
