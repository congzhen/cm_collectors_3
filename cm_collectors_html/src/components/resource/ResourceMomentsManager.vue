<template>
  <span class="resource-moments-manager">
    <el-button class="moments-trigger" @click="opened = true">片段收藏（{{ error ? '读取失败' : loading && !items.length ? '…' : items.length + (hasMore ? '+' : '') }}）</el-button>
    <el-dialog v-model="opened" title="管理片段收藏" width="min(720px, 94vw)" top="8vh" append-to-body destroy-on-close :close-on-click-modal="false">
      <template #header>
        <div class="manager-heading"><span class="manager-heading-icon"><el-icon><CollectionTag /></el-icon></span><strong>管理片段收藏</strong><span class="manager-count">{{ items.length }}{{ hasMore ? '+' : '' }}</span></div>
      </template>
      <p class="manager-notice"><el-icon><InfoFilled /></el-icon><span>删除立即生效，关闭资源编辑不会撤销；不会删除源视频。</span></p>
      <div class="manager-list" v-loading="loading">
        <p v-if="error" role="alert">{{ error }} <el-button size="small" @click="load()">重试</el-button></p>
        <el-empty v-else-if="!loading && !items.length" description="暂无片段收藏" :image-size="64" />
        <article v-for="item in items" :key="item.id" class="manager-row">
          <div class="manager-cover">
            <MomentCover :moment-id="item.id" :time="item.start" />
            <span class="moment-duration">{{ item.end > item.start ? momentDuration(item.start, item.end).replace(/^共 /, '').replace(' 秒', 's') : '时间点' }}</span>
          </div>
          <div class="manager-row-info">
            <strong class="moment-title" :title="item.title">{{ item.title }}</strong>
            <span class="moment-time">{{ timeRange(item) }}</span>
            <div v-if="item.note" class="moment-note-wrap">
              <p :id="`moment-note-${item.id}`" class="moment-note" :class="{ 'is-expanded': expandedNotes[item.id] || !canToggleNote(item.note) }">{{ item.note }}</p>
              <button v-if="canToggleNote(item.note)" class="note-toggle" type="button" :aria-expanded="!!expandedNotes[item.id]" :aria-controls="`moment-note-${item.id}`" @click="expandedNotes[item.id] = !expandedNotes[item.id]">{{ expandedNotes[item.id] ? '收起备注' : '展开备注' }}</button>
            </div>
            <span v-if="!item.available" class="moment-unavailable">源视频不可用</span>
          </div>
          <el-button class="manager-delete" size="small" text circle :icon="Delete" :loading="deleting === item.id" :disabled="!!deleting && deleting !== item.id" :aria-label="`删除片段：${item.title}`" title="删除片段" @click="remove(item)" />
        </article>
        <el-button v-if="hasMore" class="load-more" :loading="loading" @click="load(true)">加载更多</el-button>
      </div>
      <template #footer><el-button type="primary" @click="opened = false">完成</el-button></template>
    </el-dialog>
  </span>
</template>
<script setup lang="ts">
import { onActivated, onBeforeUnmount, onDeactivated, ref, watch } from 'vue'
import { CollectionTag, Delete, InfoFilled } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { momentApi, momentTime, momentDuration, subscribeMomentChanges, type VideoMoment } from '@/server/videoMoments.server'
import MomentCover from '@/components/play/MomentCover.vue'

const props = defineProps<{ resourceId: string }>()
const opened = ref(false)
const items = ref<VideoMoment[]>([])
const loading = ref(false)
const deleting = ref('')
const error = ref('')
const hasMore = ref(false)
const expandedNotes = ref<Record<string, boolean>>({})
let page = 1
let version = 0
let context = 0
let active = true
let refreshPending = false
const timeRange = (item: VideoMoment) => `${momentTime(item.start)}${item.end ? ' — ' + momentTime(item.end) : ''}`
const canToggleNote = (note: string) => note.length > 50 || note.includes('\n')

async function load(append = false) {
  if (!active || !props.resourceId || (append && loading.value)) return
  const request = ++version
  const targetPage = append ? page + 1 : 1
  loading.value = true
  error.value = ''
  try {
    const result = await momentApi.list(props.resourceId, targetPage)
    if (request !== version) return
    if (!result.status) throw new Error(result.msg || '片段加载失败')
    items.value = append ? [...items.value, ...result.data] : result.data
    page = targetPage
    hasMore.value = result.data.length === 50
  } catch (cause) {
    if (request === version) error.value = cause instanceof Error ? cause.message : '片段加载失败，请重试'
  } finally {
    if (request === version) loading.value = false
  }
}
async function remove(item: VideoMoment) {
  if (deleting.value) return
  const operation = context
  deleting.value = item.id
  try {
    try {
      await ElMessageBox.confirm(`删除片段“${item.title}”？删除立即生效，不会删除源视频。`, '删除片段收藏', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' })
    } catch { return }
    if (operation !== context) return
    const result = await momentApi.remove(item.id)
    if (operation !== context) return
    if (!result.status) throw new Error(result.msg || '删除片段失败')
    // 服务端删除已成功；即使随后的列表刷新失败，也不再展示这条记录。
    items.value = items.value.filter(existing => existing.id !== item.id)
    refreshPending = false
    // 删除后重新读取第一页，避免偏移分页跳过下一页的第一条。
    await load()
    if (operation === context) ElMessage.success('片段已删除')
  } catch (cause) {
    if (operation === context) ElMessage.error(cause instanceof Error ? cause.message : '删除片段失败，请重试')
  } finally {
    deleting.value = ''
    if (refreshPending && active) {
      refreshPending = false
      void load()
    }
  }
}
watch(() => props.resourceId, () => {
  context++
  version++
  opened.value = false
  expandedNotes.value = {}
  refreshPending = false
  items.value = []
  hasMore.value = false
  error.value = ''
  page = 1
  void load()
}, { immediate: true })
const unsubscribe = subscribeMomentChanges(() => {
  if (deleting.value) refreshPending = true
  else void load()
})
onBeforeUnmount(() => { active = false; context++; version++; unsubscribe() })
onDeactivated(() => { active = false; context++; version++; opened.value = false })
onActivated(() => { if (!active) { active = true; void load() } })
</script>
<style scoped>
.resource-moments-manager { display: inline-flex; margin-left: 12px; vertical-align: middle; }
.manager-heading { display: flex; align-items: center; gap: 10px; padding-right: 28px; }
.manager-heading strong { font-size: 16px; font-weight: 600; }
.manager-heading-icon { display: grid; place-items: center; width: 32px; height: 32px; border-radius: 9px; color: var(--el-color-primary); background: var(--el-color-primary-light-9); }
.manager-count { padding: 2px 8px; border-radius: 10px; font-size: 11px; color: var(--el-text-color-secondary); background: var(--el-fill-color); }
.manager-notice { display: flex; align-items: flex-start; gap: 7px; margin: 0 0 16px; font-size: 12px; line-height: 1.7; color: var(--el-text-color-secondary); }
.manager-notice .el-icon { flex-shrink: 0; margin-top: 4px; }
.manager-list { max-height: 58vh; min-height: 80px; overflow-y: auto; overflow-x: hidden; scrollbar-gutter: stable; padding-right: 3px; }
.manager-row { display: flex; align-items: flex-start; gap: 14px; width: 100%; min-width: 0; box-sizing: border-box; padding: 14px; margin-bottom: 10px; border: 1px solid var(--el-border-color-lighter); border-radius: 12px; background: var(--el-fill-color-extra-light); transition: border-color .15s; }
.manager-row:last-of-type { margin-bottom: 0; }
.manager-row:hover { border-color: var(--el-border-color); }
.manager-cover { position: relative; flex: 0 0 112px; width: 112px; height: 72px; overflow: hidden; border-radius: 8px; background: var(--el-fill-color); }
.manager-cover :deep(img), .manager-cover :deep(.cover-placeholder) { display: block; width: 100%; height: 100%; border-radius: inherit; object-fit: cover; }
.moment-duration { position: absolute; right: 4px; bottom: 4px; padding: 1px 5px; border-radius: 4px; background: #000b; color: #fff; font-size: 10px; font-variant-numeric: tabular-nums; }
.manager-row-info { flex: 1; min-width: 0; display: flex; flex-direction: column; align-items: flex-start; gap: 7px; }
.moment-title { display: -webkit-box; max-width: 100%; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; overflow-wrap: anywhere; font-size: 13px; line-height: 1.5; font-weight: 500; }
.moment-time { padding: 2px 7px; border-radius: 5px; background: var(--el-fill-color); font-size: 11px; color: var(--el-text-color-secondary); font-variant-numeric: tabular-nums; }
.moment-note-wrap { width: 100%; min-width: 0; }
.moment-note { display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; margin: 0; font-size: 12px; line-height: 1.7; color: var(--el-text-color-secondary); white-space: pre-wrap; overflow-wrap: anywhere; }
.moment-note.is-expanded { display: block; }
.note-toggle { margin-top: 4px; padding: 0; border: 0; background: transparent; color: var(--el-color-primary); font: inherit; font-size: 11px; cursor: pointer; }
.manager-row > .manager-delete { flex-shrink: 0; margin: 0; color: var(--el-text-color-secondary); }
.manager-row > .manager-delete:hover { color: var(--el-color-danger); background: var(--el-color-danger-light-9); }
.moment-unavailable { font-size: 11px; color: var(--el-color-warning); }
.load-more { display: block; margin: 16px auto 0; }
@media (max-width: 480px) { .manager-row { gap: 10px; padding: 10px; } .manager-cover { flex-basis: 80px; width: 80px; height: 56px; } .moment-time { padding: 2px 4px; font-size: 10px; } }
</style>
