<template>
  <dialogCommon
    ref="dialogCommonRef"
    title="导入数据列表"
    btnSubmitTitle="导入等待／失败项"
    @submit="submitHandle"
    @closed="closeHandle"
  >
    <section class="import-progress" aria-label="导入进度">
      <p class="sr-only" role="status" aria-live="polite">{{ summary }}</p>
      <div class="progress-heading">
        <div class="progress-heading-main">
          <span class="progress-title">{{
            running ? '正在导入' : batchRows ? '本批次结果' : '准备导入'
          }}</span>
          <span class="progress-count"
            ><strong>{{ stats.processed }}</strong
            ><span> / {{ stats.total }}</span
            ><span class="progress-unit">已处理</span></span
          >
        </div>
        <el-button v-if="running" size="small" plain @click="cancelTask">取消导入</el-button>
      </div>
      <div
        class="progress-track"
        role="progressbar"
        :aria-valuenow="stats.processed"
        :aria-valuemax="stats.total"
        :aria-valuemin="0"
        aria-label="已处理资源"
      >
        <div class="progress-fill" :style="{ width: `${progressPercent}%` }"></div>
      </div>
      <div class="progress-details">
        <div class="result-stats">
          <span class="stat-item" :class="{ 'stat-success': stats.success > 0 }"
            ><i></i>成功 <b>{{ stats.success }}</b></span
          >
          <span class="stat-item" :class="{ 'stat-warning': stats.warning > 0 }"
            ><i></i>警告 <b>{{ stats.warning }}</b></span
          >
          <span class="stat-item" :class="{ 'stat-danger': stats.failed > 0 }"
            ><i></i>失败 <b>{{ stats.failed }}</b></span
          >
          <span class="stat-item" :class="{ 'stat-warning': stats.unknown > 0 }"
            ><i></i>待确认 <b>{{ stats.unknown }}</b></span
          >
        </div>
        <span class="progress-hint">成功与警告均已入库</span>
      </div>
      <div v-if="running" class="current-file">
        <div class="current-file-main">
          <span class="current-file-label">当前文件</span>
          <span class="current-file-name" :title="currentPath">{{
            currentFileName || '正在准备…'
          }}</span>
          <span class="current-file-time">{{ elapsed }} 秒</span>
        </div>
        <div v-if="currentPath" class="current-file-path" :title="currentPath">
          {{ currentPath }}
        </div>
      </div>
    </section>
    <p
      v-if="notice"
      class="import-notice"
      :class="{ 'notice-warning': notice !== '本批次处理结束，请查看成功、警告和失败数量。' }"
    >
      {{ notice }}
    </p>
    <el-table
      ref="tableRef"
      :data="pathList"
      height="400px"
      border
      size="small"
      style="width: 100%"
    >
      <el-table-column type="selection" width="55" :selectable="selectable" />
      <el-table-column type="index" width="50" />
      <el-table-column label="状态" width="150">
        <template #default="scope">
          <el-icon v-if="scope.row.state === 'processing'" class="element-rotating"
            ><Loading
          /></el-icon>
          {{ stateLabels[scope.row.state as RowState] }}
        </template>
      </el-table-column>
      <el-table-column prop="path" label="文件路径">
        <template #default="scope"
          ><p>{{ scope.row.path }}</p>
          <p v-if="scope.row.msg" class="error">{{ scope.row.msg }}</p></template
        >
      </el-table-column>
    </el-table>
  </dialogCommon>
</template>
<script lang="ts" setup>
import { ref, computed, nextTick, onBeforeUnmount } from 'vue'
import dialogCommon from '../com/dialog/dialog-common.vue'
import { importDataServer } from '@/server/importData.server'
import type { I_config_scanDisk } from '@/dataType/config.dataType'
import { ElTable } from 'element-plus'

type RowState = 'waiting' | 'processing' | 'success' | 'warning' | 'failed' | 'unknown'
interface ImportRow {
  path: string
  state: RowState
  msg: string
}
const stateLabels: Record<RowState, string> = {
  waiting: '等待中',
  processing: '正在处理',
  success: '导入成功',
  warning: '已导入，有警告',
  failed: '导入失败',
  unknown: '结果待确认',
}
const emits = defineEmits(['success'])
const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>()
const tableRef = ref<InstanceType<typeof ElTable>>()
const pathList = ref<ImportRow[]>([])
const batchRows = ref<ImportRow[] | null>(null)
const running = ref(false)
const notice = ref('')
const currentPath = ref('')
const elapsed = ref(0)
const stats = computed(() => {
  const rows = batchRows.value ?? pathList.value
  const count = (state: RowState) => rows.filter((row) => row.state === state).length
  const success = count('success'),
    warning = count('warning'),
    failed = count('failed')
  return {
    total: rows.length,
    processed: success + warning + failed,
    success,
    warning,
    failed,
    unknown: count('unknown'),
  }
})
const progressPercent = computed(() =>
  stats.value.total ? (stats.value.processed / stats.value.total) * 100 : 0,
)
const currentFileName = computed(() => currentPath.value.split(/[\\/]/).pop() || '')
const summary = computed(() => {
  const { processed, total, success, warning, failed, unknown } = stats.value
  return `本批次已处理 ${processed} / ${total}，成功 ${success}，警告 ${warning}，失败 ${failed}，待确认 ${unknown}（成功与警告均已入库）`
})

let config: I_config_scanDisk
let library = ''
let generation = 0
let active: AbortController | undefined
let stopRequested = false
const selectable = (row: ImportRow) =>
  !running.value && (row.state === 'waiting' || row.state === 'failed')
const cancelTask = () => {
  stopRequested = true
  active?.abort()
  notice.value = '已取消后续导入。当前请求如未收到结果，请重新扫描核对是否入库。'
}
const closeHandle = () => {
  cancelTask()
  generation++
  running.value = false
}
onBeforeUnmount(closeHandle)

const submitHandle = async () => {
  if (running.value) return
  const selected = (tableRef.value?.getSelectionRows() || []) as ImportRow[]
  const rows = (selected.length ? selected : pathList.value).filter(selectable)
  if (!rows.length) {
    notice.value = '没有可导入项；结果待确认的资源请重新扫描核对，避免重复导入。'
    return
  }
  const taskGeneration = ++generation
  const taskLibrary = library
  const taskConfig = config
  const controller = new AbortController()
  active = controller
  batchRows.value = rows
  rows.forEach((row) => {
    row.state = 'waiting'
    row.msg = ''
  })
  stopRequested = false
  running.value = true
  currentPath.value = ''
  elapsed.value = 0
  notice.value = ''
  dialogCommonRef.value?.disabledSubmit(true)
  let batchId = ''
  let heartbeat: ReturnType<typeof setInterval> | undefined
  let timer: ReturnType<typeof setInterval> | undefined
  let heartbeatPending = false
  let imported = false
  try {
    const begin = await importDataServer.batch('begin', taskLibrary, '', controller.signal)
    if (!begin.status || typeof begin.data !== 'string') {
      if (taskGeneration === generation) notice.value = `无法开始导入：${begin.msg}`
      return
    }
    batchId = begin.data
    if (controller.signal.aborted || taskGeneration !== generation) return
    heartbeat = setInterval(async () => {
      if (heartbeatPending || controller.signal.aborted) return
      heartbeatPending = true
      try {
        const result = await importDataServer.batch(
          'heartbeat',
          taskLibrary,
          batchId,
          controller.signal,
        )
        if (!result.status && !controller.signal.aborted && taskGeneration === generation) {
          notice.value = '批次连接异常，已停止后续导入。请重新扫描核对当前资源。'
          controller.abort()
        }
      } catch {
        if (taskGeneration === generation && !controller.signal.aborted) {
          notice.value = '批次连接异常，已停止后续导入。请重新扫描核对当前资源。'
          controller.abort()
        }
      } finally {
        heartbeatPending = false
      }
    }, 20000)
    for (const row of rows) {
      if (controller.signal.aborted || taskGeneration !== generation) break
      row.state = 'processing'
      row.msg = ''
      currentPath.value = row.path
      elapsed.value = 0
      timer = setInterval(() => {
        if (taskGeneration === generation) elapsed.value++
      }, 1000)
      const result = await importDataServer.scanDiskImportData(
        taskLibrary,
        row.path,
        taskConfig,
        batchId,
        controller.signal,
      )
      clearInterval(timer)
      timer = undefined
      if (result.status) {
        const warnings = typeof result.data === 'object' ? result.data.warnings || [] : []
        row.state = warnings.length ? 'warning' : 'success'
        row.msg = warnings.join('；')
        imported = true
      } else if (result.transportError || controller.signal.aborted) {
        row.state = 'unknown'
        row.msg = '请求中断，是否入库需重新扫描核对。'
        if (taskGeneration === generation && !stopRequested)
          notice.value = '连接异常，已停止后续导入；未发送的资源保留为等待中。'
        break
      } else {
        row.state = 'failed'
        row.msg = result.msg
        if (result.statusCode === 409 || result.statusCode === 401) {
          if (taskGeneration === generation)
            notice.value = '批次已过期或登录失效，已停止后续导入。请重新开始。'
          break
        }
      }
    }
    if (taskGeneration === generation && !notice.value)
      notice.value = '本批次处理结束，请查看成功、警告和失败数量。'
  } catch (error) {
    rows
      .filter((row) => row.state === 'processing')
      .forEach((row) => {
        row.state = 'unknown'
        row.msg = '结果待确认，请重新扫描核对。'
      })
    if (taskGeneration === generation) notice.value = `导入中断：${String(error)}`
  } finally {
    if (heartbeat) clearInterval(heartbeat)
    if (timer) clearInterval(timer)
    controller.abort() // Stop any heartbeat already in flight before the next run.
    // A separate request releases the lease even after the import request was aborted.
    if (batchId) void importDataServer.batch('end', taskLibrary, batchId)
    if (taskGeneration === generation) {
      running.value = false
      active = undefined
      dialogCommonRef.value?.disabledSubmit(false)
    }
    if (imported) emits('success')
  }
}
const open = (paths: string[], draft: I_config_scanDisk, filesBasesId: string) => {
  active?.abort()
  generation++
  running.value = false
  notice.value = ''
  elapsed.value = 0
  batchRows.value = null
  pathList.value = paths.map((path) => ({ path, state: 'waiting', msg: '' }))
  config = JSON.parse(JSON.stringify(draft))
  library = filesBasesId
  dialogCommonRef.value?.disabledSubmit(false)
  dialogCommonRef.value?.open()
  nextTick(() => tableRef.value?.toggleAllSelection())
}
defineExpose({ open })
</script>
<style scoped>
.import-progress {
  padding: 18px 20px;
  margin-bottom: 16px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-fill-color-extra-light);
}
.progress-heading,
.progress-heading-main,
.progress-details,
.result-stats,
.current-file-main {
  display: flex;
  align-items: center;
}
.progress-heading,
.progress-details {
  justify-content: space-between;
  gap: 12px;
}
.progress-heading-main {
  flex-wrap: wrap;
  gap: 16px;
}
.progress-title {
  color: var(--el-text-color-primary);
  font-weight: 600;
}
.progress-count {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  font-variant-numeric: tabular-nums;
}
.progress-count strong {
  color: var(--el-text-color-primary);
  font-size: 22px;
  font-weight: 600;
}
.progress-unit {
  margin-left: 8px;
  font-size: 12px;
}
.progress-track {
  height: 6px;
  margin: 16px 0 14px;
  overflow: hidden;
  border-radius: 4px;
  background: var(--el-fill-color-darker);
}
.progress-fill {
  height: 100%;
  border-radius: inherit;
  background: var(--el-color-primary);
  transition: width 0.25s ease;
}
.result-stats {
  flex-wrap: wrap;
  gap: 10px 20px;
}
.stat-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  white-space: nowrap;
}
.stat-item i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--el-border-color-darker);
}
.stat-item b {
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
.stat-success {
  color: var(--el-color-success);
}
.stat-warning {
  color: var(--el-color-warning);
}
.stat-danger {
  color: var(--el-color-danger);
}
.stat-success i,
.stat-warning i,
.stat-danger i {
  background: currentColor;
}
.progress-hint {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.current-file {
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid var(--el-border-color-lighter);
}
.current-file-main {
  gap: 10px;
  min-width: 0;
}
.current-file-label,
.current-file-time {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.current-file-name {
  flex: 1;
  min-width: 0;
  color: var(--el-text-color-primary);
  font-size: 13px;
  font-weight: 500;
}
.current-file-name,
.current-file-path {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.current-file-path {
  margin-top: 5px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.import-notice {
  padding: 10px 12px;
  margin: 0 0 14px;
  border-radius: 6px;
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-light);
  font-size: 12px;
  line-height: 1.6;
  overflow-wrap: anywhere;
}
.notice-warning {
  color: var(--el-color-warning);
  background: var(--el-color-warning-light-9);
}
.error {
  color: var(--el-color-warning);
  overflow-wrap: anywhere;
}
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}
@media (max-width: 600px) {
  .import-progress {
    padding: 14px;
  }
  .progress-details {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  .progress-heading-main {
    gap: 8px;
  }
  .result-stats {
    gap: 10px 14px;
  }
}
@media (prefers-reduced-motion: reduce) {
  .progress-fill {
    transition: none;
  }
}
</style>
