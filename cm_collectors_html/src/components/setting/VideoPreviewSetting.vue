<template>
  <section class="preview-settings" v-loading="loading">
    <h3>视频缩略图与预览缓存</h3>
    <el-form label-position="top">
      <el-form-item label="缩略图宽度"
        ><el-select v-model="setting.width" style="width: 220px"
          ><el-option v-for="w in [160, 240, 320, 480]" :key="w" :value="w" :label="`${w} 像素`"
        /></el-select>
        <div class="help">
          仅影响之后生成的预览。已有图片保留，手动删除后重新生成才会使用新尺寸。
        </div></el-form-item
      >
      <el-form-item label="缓存容量上限（MB）"
        ><el-input-number
          v-model="setting.maxMB"
          :min="0"
          :max="1048576"
          :step="64"
          :precision="0"
        />
        <div class="help">
          0 表示不限；启用上限时至少 64
          MB。超过上限，自动删除最早生成的预览缓存，保留视频和片段收藏。
        </div></el-form-item
      >
      <el-button type="primary" :loading="saving" @click="save">保存设置</el-button>
    </el-form>
    <el-divider />
    <p v-if="status">
      当前缓存：{{ (status.bytes / 1024 / 1024).toFixed(1) }} MB · 待完成：{{ status.queued }} ·
      {{ status.paused ? '已暂停' : status.running ? '生成中' : '空闲' }}
    </p>
    <div class="actions">
      <el-button
        :loading="openingFolder"
        title="在运行服务的电脑上打开缓存目录"
        @click="openCacheFolder"
        >打开缓存文件夹</el-button
      >
      <el-button @click="load(false)">刷新状态</el-button
      ><el-button @click="control('pause')">暂停生成</el-button
      ><el-button @click="control('resume')">恢复生成</el-button
      ><el-button type="danger" plain @click="clear">清空预览缓存</el-button>
    </div>
    <p class="help">
      清空前请暂停并等待当前抽图结束。替换源视频不会自动删除或更新预览；你可以在视频时间轴中单独删除预览。
    </p>
    <p v-if="status?.error" role="alert">{{ status.error }}</p>
    <p v-if="error" role="alert">{{ error }}</p>
  </section>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/assets/request'
import { notifyPreviewChanged } from '@/server/videoMoments.server'
const openingFolder = ref(false)
async function openCacheFolder() {
  if (openingFolder.value) return
  openingFolder.value = true
  const result = await request({ url: '/videoPreviewCache/openFolder', method: 'post' })
  openingFolder.value = false
  if (!result.status) ElMessage.error(result.msg)
}
const setting = ref({ width: 240, maxMB: 1024 })
const status = ref<{
  bytes: number
  queued: number
  paused: boolean
  running: boolean
  error?: string
}>()
const loading = ref(false),
  saving = ref(false),
  error = ref('')
async function load(includeSettings = true) {
  loading.value = true
  const [s, q] = await Promise.all([
    includeSettings
      ? request<typeof setting.value>({ url: '/videoPreviewSetting' })
      : Promise.resolve(undefined),
    request<NonNullable<typeof status.value>>({ url: '/videoPreviewQueue' }),
  ])
  loading.value = false
  error.value = s && !s.status ? s.msg : !q.status ? q.msg : ''
  if (s?.status && includeSettings) setting.value = s.data
  if (q.status) status.value = q.data
}
async function save() {
  if (saving.value) return
  saving.value = true
  const r = await request({
    url: '/videoPreviewSetting',
    method: 'put',
    data: { ...setting.value },
  })
  saving.value = false
  if (!r.status) ElMessage.error(r.msg)
  else {
    ElMessage.success('预览设置已保存')
    notifyPreviewChanged()
    await load(false)
  }
}
async function control(action: string) {
  const r = await request({ url: `/videoPreviewQueue/${action}`, method: 'post' })
  if (!r.status) ElMessage.error(r.msg)
  else notifyPreviewChanged()
  await load(false)
}
async function clear() {
  try {
    await ElMessageBox.confirm('删除全部预览图片？不会删除视频、片段收藏和首图。', '清空预览缓存', {
      type: 'warning',
    })
  } catch {
    return
  }
  await control('clear')
}
onMounted(() => load())
</script>
<style scoped>
.preview-settings {
  max-width: 850px;
  padding: 12px;
}
.help {
  width: 100%;
  font-size: 13px;
  line-height: 1.8;
  opacity: 0.65;
  margin: 8px 0;
}
.actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.actions .el-button {
  margin: 0;
}
</style>
