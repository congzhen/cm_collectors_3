<template>
  <section class="preview-settings settings-subpage" v-loading="loading">
    <div class="subpage-content">
      <section class="subpage-group" aria-label="时间轴预览设置">
        <div class="group-heading"><h3>时间轴预览设置</h3><p>设置视频播放时间轴的预览缩略图，不影响下方封面候选图。</p></div>
        <el-form class="subpage-panel preview-form" label-position="left" label-width="200px">
          <el-form-item label="缩略图宽度">
            <el-select v-model="setting.width" class="width-select">
              <el-option v-for="w in [160, 240, 320, 480]" :key="w" :value="w" :label="`${w} 像素`" />
            </el-select>
            <div class="help">仅影响之后生成的时间轴预览图。已有图片需删除并重新生成，才会使用新宽度。</div>
          </el-form-item>
          <el-form-item label="缓存容量上限（MB）">
            <el-input-number v-model="setting.maxMB" :min="0" :max="1048576" :step="64" :precision="0" />
            <div class="help">仅限制时间轴预览缓存。0 表示不限，启用上限时至少 64 MB；超出后自动删除最早生成的预览图，不删除视频和片段收藏。</div>
          </el-form-item>
          <div class="save-actions"><el-button type="primary" :loading="saving" @click="save">保存时间轴预览设置</el-button></div>
        </el-form>
      </section>
      <section class="subpage-group" aria-label="时间轴缓存与任务">
        <div class="group-heading"><h3>时间轴缓存与任务</h3><p>查看时间轴预览的缓存占用和生成队列；暂停、恢复仅作用于此队列。</p></div>
        <div class="subpage-panel">
          <dl v-if="status" class="cache-status" aria-live="polite">
            <div><dt>时间轴缓存占用</dt><dd>{{ (status.bytes / 1024 / 1024).toFixed(1) }} <small>MB</small></dd></div>
            <div><dt>待完成任务</dt><dd>{{ status.queued }} <small>项</small></dd></div>
            <div><dt>生成状态</dt><dd class="queue-state">{{ status.paused ? '已暂停' : status.running ? '生成中' : '空闲' }}</dd></div>
          </dl>
          <div class="actions">
            <el-button :loading="openingFolder" title="在运行服务的电脑上打开时间轴预览缓存目录" @click="openCacheFolder">打开时间轴缓存文件夹</el-button>
            <el-button @click="load(false)">刷新状态</el-button>
            <el-button @click="control('pause')">暂停生成</el-button>
            <el-button @click="control('resume')">恢复生成</el-button>
          </div>
          <div class="cache-cleanup">
            <el-button type="danger" plain @click="clear">清空时间轴预览缓存</el-button>
            <p class="help">仅清空时间轴预览图，不影响封面预览缓存、已保存封面、视频和片段收藏。清空前请暂停生成并等待当前任务结束。替换源视频后，可在时间轴中单独删除旧预览再重新生成。</p>
          </div>
          <p v-if="status?.error" class="error-message" role="alert">{{ status.error }}</p>
        </div>
      </section>
      <CoverPreviewCacheSettings />
      <p v-if="error" class="error-message" role="alert">{{ error }}</p>
    </div>
  </section>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import CoverPreviewCacheSettings from './CoverPreviewCacheSettings.vue'
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
    ElMessage.success('时间轴预览设置已保存')
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
    await ElMessageBox.confirm('删除全部时间轴预览图片？不会删除封面预览候选图、已保存的封面、视频、片段收藏和首图。', '清空时间轴预览缓存', {
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
