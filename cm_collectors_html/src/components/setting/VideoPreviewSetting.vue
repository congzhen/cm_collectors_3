<template>
  <section class="preview-settings settings-subpage" v-loading="loading">
    <div class="subpage-content">
      <section class="subpage-group" aria-label="预览设置">
        <div class="group-heading"><h3>预览设置</h3><p>调整之后生成的缩略图与缓存保留策略。</p></div>
        <el-form class="subpage-panel preview-form" label-position="left" label-width="200px">
          <el-form-item label="缩略图宽度">
            <el-select v-model="setting.width" class="width-select">
              <el-option v-for="w in [160, 240, 320, 480]" :key="w" :value="w" :label="`${w} 像素`" />
            </el-select>
            <div class="help">仅影响之后生成的预览。已有图片保留，手动删除后重新生成才会使用新尺寸。</div>
          </el-form-item>
          <el-form-item label="缓存容量上限（MB）">
            <el-input-number v-model="setting.maxMB" :min="0" :max="1048576" :step="64" :precision="0" />
            <div class="help">0 表示不限；启用上限时至少 64 MB。超过上限，自动删除最早生成的预览缓存，保留视频和片段收藏。</div>
          </el-form-item>
          <div class="save-actions"><el-button type="primary" :loading="saving" @click="save">保存设置</el-button></div>
        </el-form>
      </section>
      <section class="subpage-group" aria-label="缓存与任务">
        <div class="group-heading"><h3>缓存与任务</h3><p>查看占用空间与任务队列，控制预览生成。</p></div>
        <div class="subpage-panel">
          <dl v-if="status" class="cache-status" aria-live="polite">
            <div><dt>当前缓存</dt><dd>{{ (status.bytes / 1024 / 1024).toFixed(1) }} <small>MB</small></dd></div>
            <div><dt>待完成</dt><dd>{{ status.queued }} <small>项</small></dd></div>
            <div><dt>生成状态</dt><dd class="queue-state">{{ status.paused ? '已暂停' : status.running ? '生成中' : '空闲' }}</dd></div>
          </dl>
          <div class="actions">
            <el-button :loading="openingFolder" title="在运行服务的电脑上打开缓存目录" @click="openCacheFolder">打开缓存文件夹</el-button>
            <el-button @click="load(false)">刷新状态</el-button>
            <el-button @click="control('pause')">暂停生成</el-button>
            <el-button @click="control('resume')">恢复生成</el-button>
          </div>
          <div class="cache-cleanup">
            <el-button type="danger" plain @click="clear">清空预览缓存</el-button>
            <p class="help">清空前请暂停并等待当前抽图结束。替换源视频不会自动删除或更新预览；你可以在视频时间轴中单独删除预览。</p>
          </div>
          <p v-if="status?.error" class="error-message" role="alert">{{ status.error }}</p>
        </div>
      </section>
      <p v-if="error" class="error-message" role="alert">{{ error }}</p>
    </div>
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
<style scoped lang="scss">
@use './settingsSubpage';
.preview-form {
  :deep(.el-form-item) { margin: 0 0 18px; padding-bottom: 18px; border-bottom: 1px solid var(--el-border-color-extra-light); }
  :deep(.el-form-item__label) { height: auto; line-height: 24px; padding-top: 4px; color: var(--el-text-color-regular); }
  :deep(.el-form-item__content) { min-width: 0; }
  .width-select { width: 240px; max-width: 100%; }
}
.save-actions { padding-left: 200px; }
.cache-status {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin: 0 0 22px;
  dt { font-size: 12px; color: var(--el-text-color-secondary); }
  dd { margin: 6px 0 0; font-size: 24px; font-weight: 600; line-height: 32px; font-variant-numeric: tabular-nums; }
  small { font-size: 12px; font-weight: 400; color: var(--el-text-color-secondary); }
  .queue-state { font-size: 18px; }
}
.actions { display: flex; gap: 8px; flex-wrap: wrap; .el-button { margin: 0; } }
.cache-cleanup { margin-top: 12px; padding-top: 12px; border-top: 1px solid var(--el-border-color-extra-light); }
.error-message { margin-top: 12px; padding: 10px 14px; border-radius: 6px; color: var(--el-color-danger); background: var(--el-color-danger-light-9); overflow-wrap: anywhere; }
@media (max-width: 700px) {
  .preview-form :deep(.el-form-item) { display: block; }
  .preview-form :deep(.el-form-item__label) { width: auto !important; display: block; padding: 0 0 8px; }
  .save-actions { padding-left: 0; }
  .cache-status { gap: 10px; dd { font-size: 20px; } .queue-state { font-size: 16px; } }
}
</style>
