<template>
  <section class="subpage-group" aria-label="封面预览缓存" v-loading="busy">
    <div class="group-heading"><h3>封面预览缓存</h3><p>管理资源添加、编辑窗口中“获取视频预览图”生成的封面候选图。</p></div>
    <el-form class="subpage-panel cover-cache-form" label-position="left" label-width="200px">
      <el-form-item label="最多缓存资源数">
        <el-input-number v-if="ready" v-model="maxResources" :min="1" :max="100" :precision="0" :disabled="busy" />
        <div class="help">默认 5 个，可设置 1–100 个。超过上限淘汰最久未使用的资源；调小并保存后立即清理超额缓存。</div>
      </el-form-item>
      <el-form-item label="当前缓存">
        <span v-if="status">{{ status.count }} 个资源 · {{ (status.bytes / 1024 / 1024).toFixed(1) }} MB</span>
        <span v-else>暂未读取</span>
      </el-form-item>
      <div class="cover-cache-actions">
        <el-button type="primary" :disabled="!ready || busy" @click="save">保存封面缓存设置</el-button>
        <el-button :disabled="busy" @click="load(false)">刷新状态</el-button>
        <el-button type="danger" plain :disabled="!ready || busy" @click="clear">清空封面预览缓存</el-button>
      </div>
      <p class="help">仅清理封面候选图，不影响视频、已保存的封面、时间轴预览和片段收藏。再次打开优先使用缓存，点击预览窗口的“刷新”可强制重新获取。</p>
      <p v-if="error" role="alert" class="cache-error">{{ error }}</p>
    </el-form>
  </section>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/assets/request'
const maxResources = ref<number | undefined>(5)
const status = ref<{ count: number; bytes: number }>()
const ready = ref(false), busy = ref(false), error = ref('')
async function load(includeSettings = true) {
  if (busy.value) return
  busy.value = true
  try {
    const [setting, result] = await Promise.all([
      includeSettings || !ready.value ? request<{maxResources: number}>({url: '/coverPreviewCache/settings'}) : Promise.resolve(undefined),
      request<NonNullable<typeof status.value>>({url: '/coverPreviewCache/status'}),
    ])
    if (setting) {
      ready.value = setting.status
      if (setting.status) maxResources.value = setting.data.maxResources
    }
    status.value = result.status ? result.data : undefined
    error.value = setting && !setting.status ? setting.msg : !result.status ? result.msg : ''
  } catch { error.value = '读取封面预览缓存失败，请刷新重试' }
  finally { busy.value = false }
}
async function save() {
  if (busy.value || !ready.value) return
  if (!Number.isInteger(maxResources.value) || maxResources.value! < 1 || maxResources.value! > 100) {
    ElMessage.warning('缓存资源数应为 1 至 100 的整数'); return
  }
  busy.value = true
  try {
    const result = await request({url: '/coverPreviewCache/settings', method: 'put', data: {maxResources: maxResources.value}})
    if (!result.status) { ElMessage.error(result.msg); return }
    ElMessage.success('封面预览缓存设置已保存')
  } catch { ElMessage.error('保存失败，请重试') }
  finally { busy.value = false }
  await load(false)
}
async function clear() {
  if (busy.value || !ready.value) return
  busy.value = true
  try {
    try {
      await ElMessageBox.confirm('清空所有资源的封面预览候选图？不会删除视频、已保存的封面、时间轴预览或片段收藏。再次获取时会重新生成；如正在抽图，将等待完成后清空。', '清空封面预览缓存', {type:'warning', confirmButtonText:'确认清空', cancelButtonText:'取消', closeOnClickModal:false})
    } catch { return }
    const result = await request({url:'/coverPreviewCache/clear', method:'post'})
    if (!result.status) { ElMessage.error(result.msg); return }
    ElMessage.success('封面预览缓存已清空')
  } catch { ElMessage.error('清空失败，请重试') }
  finally { busy.value = false }
  await load(false)
}
onMounted(() => load())
</script>
<style scoped lang="scss">
@use './settingsSubpage';
.cover-cache-form {
  .el-form-item { margin-bottom: 16px; }
  .help { width: 100%; margin: 8px 0; font-size: 12px; line-height: 1.7; color: var(--el-text-color-secondary); }
  .cover-cache-actions { display: flex; flex-wrap: wrap; gap: 8px; .el-button { margin: 0; } }
  .cache-error { color: var(--el-color-danger); }
}
</style>
