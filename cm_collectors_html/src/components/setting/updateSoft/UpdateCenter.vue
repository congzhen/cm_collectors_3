<template>
  <el-dialog v-model="updateVisible" title="软件更新" width="520px" :close-on-click-modal="false">
    <div class="update-center">
      <p v-if="checking">正在检查更新…</p>
      <template v-else-if="info">
        <p>当前版本 {{ info.currentVersion }} · {{ info.platform }}</p>
        <p v-if="info.hasUpdate">
          发现新版本 <strong>{{ info.version }}</strong
          >，是否下载更新？
        </p>
        <p v-else>当前已是最新版本。</p>
      </template>
      <el-alert
        v-if="error || state?.error"
        class="update-error"
        :title="error || state?.error || ''"
        type="error"
        :closable="false"
      />
      <template v-if="state && state.phase !== 'idle'">
        <p>
          {{ phaseText[state.phase] }}<span v-if="state.version"> · {{ state.version }}</span>
        </p>
        <el-progress v-if="state.phase === 'downloading'" :percentage="progress" />
        <p v-if="state.phase === 'ready'">
          更新包已校验。安装会等待当前任务完成、备份数据并短暂重启服务。
        </p>
        <p v-if="state.phase === 'waiting'">
          正在等待当前请求、扫描和转码等任务完成。新任务暂缓启动。
        </p>
        <p v-if="state.phase === 'backingUp'">正在备份数据库和配置，请保持程序运行。</p>
        <p v-if="state.phase === 'installing'">
          正在安装并重新连接，请稍候。若长时间未恢复，可查看程序目录中的更新日志。
        </p>
        <el-button v-if="state.phase === 'installed'" type="primary" @click="reload"
          >更新完成，刷新页面</el-button
        >
      </template>
      <el-checkbox
        v-if="info?.hasUpdate && !active"
        v-model="snooze"
        @change="setUpdatesMuted(snooze)"
        >7 天内不再提示</el-checkbox
      >
      <el-collapse v-if="info?.changelog?.length">
        <el-collapse-item title="更新说明" name="log">
          <ul class="update-changelog">
            <li v-for="(line, index) in info.changelog" :key="index">{{ line }}</li>
          </ul>
        </el-collapse-item>
      </el-collapse>
    </div>
    <template #footer>
      <el-button v-if="!active" :disabled="checking || acting" @click="check(true)"
        >重新检测</el-button
      >
      <el-button v-if="state?.phase === 'waiting'" :disabled="acting" @click="action('cancel')"
        >取消安装</el-button
      >
      <el-button @click="updateVisible = false">{{ active ? '后台继续' : '稍后再说' }}</el-button>
      <el-button v-if="state?.phase === 'ready'" type="primary" :loading="acting" @click="install"
        >安装并重启</el-button
      >
      <el-button
        v-if="
          info?.hasUpdate &&
          !active &&
          (state?.phase !== 'ready' || state.version !== info.version) &&
          (state?.phase !== 'installed' || state.version !== info.version)
        "
        type="primary"
        :loading="acting"
        @click="download"
        >下载更新</el-button
      >
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { appStoreData } from '@/storeData/app.storeData'
import {
  canAutoCheck,
  setUpdatesMuted,
  readUpdatePolicy,
  saveUpdatePolicy,
  updateManualRequest,
  updateRequest,
  UpdateRejectedError,
  updateVisible,
  withUpdateLock,
  type UpdateInfo,
  type UpdateState,
} from '@/common/softwareUpdate'

const route = useRoute()
const store = appStoreData()
const info = ref<UpdateInfo>()
const state = ref<UpdateState>()
const checking = ref(false)
const acting = ref(false)
const snooze = ref(false)
const error = ref('')
const active = computed(() =>
  ['downloading', 'waiting', 'backingUp', 'installing'].includes(state.value?.phase || ''),
)
const progress = computed(() =>
  state.value?.total
    ? Math.min(100, Math.floor((state.value.downloaded * 100) / state.value.total))
    : 0,
)
const phaseText: Record<UpdateState['phase'], string> = {
  idle: '',
  downloading: '正在下载',
  ready: '下载完成',
  waiting: '等待任务完成',
  backingUp: '正在备份',
  installing: '正在安装',
  installed: '更新成功',
  failed: '更新未完成',
}
let autoStarted = false
let manualCheckPending = false
let disposed = false
let polling = false
let reconnectStarted = 0
let stateRequestId = 0
let autoTimer: ReturnType<typeof setTimeout> | undefined
let pollTimer: ReturnType<typeof setTimeout> | undefined
async function refreshState() {
  const requestId = ++stateRequestId
  const next = await updateRequest<UpdateState>('status')
  if (disposed || requestId !== stateRequestId) return
  const previous = state.value?.phase
  state.value = next
  if (
    (state.value.phase === 'ready' &&
      (previous === 'downloading' || (state.value.error && previous !== 'ready'))) ||
    (state.value.phase === 'installed' &&
      ['waiting', 'backingUp', 'installing'].includes(previous || '')) ||
    (state.value.phase === 'failed' && previous && previous !== 'failed')
  )
    updateVisible.value = true
  if (state.value.phase !== 'installing') reconnectStarted = 0
}
async function poll() {
  if (polling || disposed) return
  polling = true
  try {
    await refreshState()
  } catch {
    // A fast installation can restart before a poll observes the installing phase.
    if (state.value && ['waiting', 'backingUp'].includes(state.value.phase))
      state.value.phase = 'installing'
  } finally {
    polling = false
  }
  if (state.value?.phase === 'installing') {
    reconnectStarted ||= Date.now()
    if (Date.now() - reconnectStarted > 180_000) {
      error.value = '服务尚未通过更新验证，请检查服务端更新日志，然后重新打开更新窗口。'
      updateVisible.value = true
      return
    }
  }
  if (active.value && !disposed) pollTimer = setTimeout(poll, 1500)
}
function startPoll() {
  clearTimeout(pollTimer)
  void poll()
}
async function check(manual: boolean) {
  if (disposed) return
  if (checking.value) {
    manualCheckPending ||= manual
    return
  }
  checking.value = true
  if (manual) {
    error.value = ''
    updateVisible.value = true
  }
  try {
    const result = await updateRequest<UpdateInfo>(`check${manual ? '?manual=1' : ''}`)
    if (disposed) return
    info.value = result
    if (!manual) {
      const policy = readUpdatePolicy()
      policy.checked = true
      if (info.value.hasUpdate && !policy.prompted && policy.mutedUntil <= Date.now()) {
        policy.prompted = true
        updateVisible.value = true
      }
      saveUpdatePolicy(policy)
    }
  } catch (e) {
    if (manual) error.value = e instanceof Error ? e.message : '检测失败，请稍后重试'
  } finally {
    checking.value = false
    if (manualCheckPending) {
      manualCheckPending = false
      void check(true)
    }
  }
}
async function automatic() {
  if (disposed) return
  if (!['/', '/mobile'].includes(route.path)) {
    autoStarted = false
    return
  }
  // The server verifies the persisted admin cookie, including after a page reload.
  try {
    await refreshState()
    if (active.value) startPoll()
  } catch {
    autoStarted = false
    return
  }
  await withUpdateLock(async () => {
    if (disposed) return
    const policy = readUpdatePolicy()
    if (!canAutoCheck(policy)) return
    policy.attempts++
    policy.lastAttempt = Date.now()
    saveUpdatePolicy(policy)
    await check(false)
  })
}
watch(
  () => [store.displayAdminFn, route.path],
  () => {
    if (!['/', '/mobile'].includes(route.path)) {
      clearTimeout(autoTimer)
      autoStarted = false
      return
    }
    if (autoStarted) return
    autoStarted = true
    autoTimer = setTimeout(automatic, 3000)
  },
  { immediate: true },
)
watch(updateManualRequest, async () => {
  if (!store.displayAdminFn) return
  reconnectStarted = 0
  try {
    await refreshState()
    if (active.value) startPoll()
  } catch {
    /* check reports errors */
  }
  await check(true)
})
watch(updateVisible, (visible) => {
  if (visible) snooze.value = readUpdatePolicy().mutedUntil > Date.now()
})
async function action(path: string, body: unknown = {}) {
  acting.value = true
  error.value = ''
  try {
    const result = await updateRequest<UpdateState>(path, body)
    if (disposed) return
    stateRequestId++
    state.value = result
  } catch (e) {
    error.value = e instanceof Error ? e.message : '操作失败，请重试'
    if (!(e instanceof UpdateRejectedError)) {
      // The server may have accepted the operation even if its response was lost.
      try {
        await refreshState()
        if (active.value) error.value = ''
      } catch {
        /* Keep the original error when the operation cannot be reconciled. */
      }
    }
  } finally {
    acting.value = false
    if (active.value) startPoll()
  }
}
async function download() {
  if (info.value) await action('download', { version: info.value.version })
}
async function install() {
  if (!state.value || state.value.phase !== 'ready') return
  const selection = { version: state.value.version, job: state.value.job }
  try {
    await ElMessageBox.confirm(
      '安装将等待当前任务完成、备份数据并重启服务，是否继续？',
      '安装更新',
      {
        confirmButtonText: '安装并重启',
        cancelButtonText: '稍后安装',
        type: 'warning',
      },
    )
    await action('install', selection)
  } catch {
    /* user cancelled */
  }
}
function reload() {
  window.location.reload()
}
onBeforeUnmount(() => {
  disposed = true
  clearTimeout(autoTimer)
  clearTimeout(pollTimer)
})
</script>
<style scoped>
.update-center {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.update-center p {
  margin: 0;
  line-height: 1.6;
}
.update-error :deep(.el-alert__content) {
  min-width: 0;
  flex: 1;
}
.update-error :deep(.el-alert__title) {
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  word-break: break-word;
}
.update-changelog {
  max-height: 200px;
  overflow: auto;
  padding-left: 20px;
}
</style>
