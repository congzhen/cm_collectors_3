<template>
  <main class="moments-page">
    <header>
      <el-button @click="router.push(mobile ? '/mobile' : '/')">返回片库</el-button>
      <h1>{{ theater ? '片段放映室' : '片段收藏' }}</h1>
    </header>
    <div v-if="mobile && libraryId" class="moments-mobile-toolbar">
      <span class="moments-library-name">{{ app.currentFilesBases.name || '当前文件库' }}</span>
      <el-switch :model-value="theater" :disabled="savingMode" active-text="放映室" inactive-text="收藏" size="small" aria-label="放映室模式" @change="toggleMode" />
    </div>
    <p v-else class="intro">把喜欢的时刻留下，随时回来重看。</p>
    <details v-if="!mobile" v-admin @toggle="loadStatus">
      <summary>预览任务与缓存管理</summary>
      <p v-if="status">
        {{ status.paused ? '队列已暂停' : '队列可运行' }} · {{ status.queued }} 个待完成任务 · 缓存
        {{ (status.bytes / 1024 / 1024).toFixed(1) }} MB
      </p>
      <el-button size="small" @click="control('pause')">暂停</el-button
      ><el-button size="small" @click="control('resume')">恢复</el-button
      ><el-button size="small" @click="loadStatus">刷新状态</el-button
      ><el-button size="small" type="danger" plain @click="clearCache">手动删除全部预览</el-button>
      <p class="intro">
        替换视频不会自动删除或更新预览。删除预览不影响视频和片段收藏。暂停后等待当前抽图结束，再执行删除。
      </p>
    </details>
    <p v-if="!mobile && status?.error" role="alert">{{ status.error }}</p>
    <div v-if="libraryId" class="moments-collection-container">
      <MobileMomentTheater v-if="theater" :key="`theater-${libraryId}`" :files-bases-id="libraryId" />
      <MomentCollection v-else :key="libraryId" :files-bases-id="libraryId" :layout="mobile ? 'list' : 'grid'" />
    </div>
    <p v-else class="intro">请先返回片库选择文件库。</p>
  </main>
</template>
<script setup lang="ts">
import { computed, onActivated, onBeforeUnmount, onDeactivated, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/assets/request'
import { notifyPreviewChanged } from '@/server/videoMoments.server'
import { isMobile } from '@/assets/mobile'
import MomentCollection from '@/components/play/MomentCollection.vue'
import { appStoreData } from '@/storeData/app.storeData'
import MobileMomentTheater from '@/components/play/MobileMomentTheater.vue'
import { filesBasesServer } from '@/server/filesBases.server'
import { createDefaultConfigApp, E_config_type } from '@/dataType/config.dataType'
const router = useRouter()
const app = appStoreData()
const mobile = isMobile()
const libraryId = computed(() => app.currentFilesBases.id || '')
const theater = computed(() => mobile && app.currentFilesBasesAppConfig?.videoMomentLibraryMode !== 'collection')
const savingMode = ref(false)
let active = true, context = 0
onDeactivated(() => { active = false; context++ })
onActivated(() => { active = true })
onBeforeUnmount(() => { active = false; context++ })
async function toggleMode(value: string | number | boolean) {
  if (savingMode.value || !libraryId.value) return
  const id = libraryId.value, operation = context
  const next = value ? 'theater' : 'collection'
  savingMode.value = true
  try {
    const latest = await filesBasesServer.getConfigById(id, E_config_type.app)
    if (id !== libraryId.value || operation !== context || !active) return
    if (!latest.status) throw new Error(latest.msg || '读取文件库设置失败')
    const config = { ...createDefaultConfigApp(), ...JSON.parse(latest.data || '{}'), videoMomentLibraryMode: next }
    const result = await filesBasesServer.setFilesBasesConfigById(id, config)
    if (!result.status) {
      if (active && operation === context && id === libraryId.value && result.msg !== '已取消修改') ElMessage.error(result.msg || '保存默认视图失败')
      return
    }
    if (id === app.currentFilesBases.id) app.currentConfigApp.videoMomentLibraryMode = next
  } catch (cause) {
    if (active && operation === context) ElMessage.error(cause instanceof Error ? cause.message : '保存默认视图失败')
  } finally { savingMode.value = false }
}
const status = ref<{
  paused: boolean
  running: boolean
  queued: number
  bytes: number
  error?: string
}>()
async function loadStatus() {
  const r = await request<NonNullable<typeof status.value>>({ url: '/videoPreviewQueue' })
  if (r.status) status.value = r.data
  else ElMessage.error(r.msg)
}
async function control(action: string) {
  const r = await request({ url: `/videoPreviewQueue/${action}`, method: 'post' })
  if (!r.status) ElMessage.error(r.msg)
  if (r.status) notifyPreviewChanged()
  await loadStatus()
}
async function clearCache() {
  try {
    await ElMessageBox.confirm('删除全部已生成的预览图？源视频和片段收藏会保留。', '手动删除预览', {
      type: 'warning',
    })
  } catch {
    return
  }
  await control('clear')
}
</script>
<style scoped>
.moments-page {
  height: 100%;
  box-sizing: border-box;
  overflow: auto;
  display: flex;
  flex-direction: column;
  padding: 24px max(16px, calc((100vw - 960px) / 2));
  color: var(--el-text-color-primary);
  background: var(--el-bg-color);
}
.moments-collection-container {
  flex: 1;
  min-height: 200px;
}
.moments-page > header,
.moments-page > .intro,
.moments-page > .moments-mobile-toolbar,
.moments-page > details {
  flex-shrink: 0;
}
.moments-mobile-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin: 10px 0 18px; }
.moments-library-name { font-size: 12px; color: var(--el-text-color-secondary); min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
header {
  display: flex;
  gap: 18px;
  align-items: center;
}
h1 {
  font-size: 24px;
}
.intro {
  font-size: 13px;
  opacity: 0.65;
  line-height: 1.8;
}
details {
  padding: 14px;
  margin: 20px 0;
  border: 1px solid var(--el-border-color);
  border-radius: 10px;
}
summary {
  cursor: pointer;
  margin-bottom: 12px;
}
</style>
