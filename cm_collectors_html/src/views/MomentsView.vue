<template>
  <main class="moments-page">
    <header>
      <el-button @click="router.push(isMobile() ? '/mobile' : '/')">返回片库</el-button>
      <h1>片段收藏</h1>
    </header>
    <p class="intro">把喜欢的时刻留下，随时回来重看。</p>
    <details v-admin @toggle="loadStatus">
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
    <p v-if="status?.error" role="alert">{{ status.error }}</p>
    <MomentCollection />
  </main>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/assets/request'
import { notifyPreviewChanged } from '@/server/videoMoments.server'
import { isMobile } from '@/assets/mobile'
import MomentCollection from '@/components/play/MomentCollection.vue'
const router = useRouter()
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
  padding: 24px max(16px, calc((100vw - 960px) / 2));
  color: var(--el-text-color-primary);
  background: var(--el-bg-color);
}
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
