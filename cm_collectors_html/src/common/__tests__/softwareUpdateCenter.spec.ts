import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import UpdateCenter from '@/components/setting/updateSoft/UpdateCenter.vue'
import {
  readUpdatePolicy,
  updateVisible,
  updateManualRequest,
  type UpdateState,
} from '../softwareUpdate'

const route = vi.hoisted(() => ({ current: undefined as { path: string } | undefined }))
vi.mock('vue-router', async () => {
  const { reactive } = await import('vue')
  route.current = reactive({ path: '/' })
  return { useRoute: () => route.current }
})
vi.mock('@/storeData/app.storeData', () => ({ appStoreData: () => ({ displayAdminFn: true }) }))
vi.mock('element-plus', () => ({ ElMessageBox: { confirm: vi.fn().mockResolvedValue('confirm') } }))

let wrapper: VueWrapper | undefined
let state: UpdateState
let disconnected = false
let failedCheck = false
let checks = 0
const info = {
  version: '3.4.0',
  currentVersion: '3.3.0',
  platform: 'linux/amd64',
  hasUpdate: true,
  changelog: [],
}
function create() {
  wrapper = mount(UpdateCenter, {
    global: {
      stubs: {
        ElDialog: {
          props: ['modelValue'],
          template: '<section v-if="modelValue"><slot /><slot name="footer" /></section>',
        },
        ElButton: { template: '<button><slot /></button>' },
        ElAlert: { props: ['title'], template: '<p>{{ title }}</p>' },
        ElCheckbox: {
          props: ['modelValue'],
          emits: ['update:modelValue', 'change'],
          template: `<input type="checkbox" :checked="modelValue" @change="$emit('update:modelValue', $event.target.checked); $emit('change', $event.target.checked)" />`,
        },
        ElProgress: true,
        ElCollapse: true,
        ElCollapseItem: true,
      },
    },
  })
  return wrapper
}
async function click(text: string) {
  const button = wrapper!.findAll('button').find((button) => button.text() === text)
  expect(button, `missing button ${text}`).toBeDefined()
  await button!.trigger('click')
  await flushPromises()
}
beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(new Date(2026, 9, 27, 12))
  localStorage.clear()
  route.current!.path = '/'
  updateVisible.value = false
  updateManualRequest.value = 0
  disconnected = false
  failedCheck = false
  checks = 0
  state = { phase: 'idle', version: '', downloaded: 0, total: 0 }
  vi.stubGlobal(
    'fetch',
    vi.fn(async (url: string) => {
      if (url.endsWith('/status') && disconnected) throw new Error('disconnected')
      let data: unknown = { ...state }
      if (url.includes('/check')) {
        checks++
        if (failedCheck) throw new Error('GitHub unavailable')
        data = info
      }
      if (url.endsWith('/download')) {
        state = { phase: 'downloading', version: '3.4.0', job: 'job-1', downloaded: 50, total: 100 }
        data = { ...state }
      }
      if (url.endsWith('/install')) {
        state.phase = 'waiting'
        data = { ...state }
      }
      return { ok: true, json: async () => ({ status: true, data }) }
    }),
  )
})
afterEach(() => {
  wrapper?.unmount()
  wrapper = undefined
  vi.unstubAllGlobals()
  vi.useRealTimers()
})
describe('update center', () => {
  it('keeps an explicit cancellation rejection visible when installation has advanced', async () => {
    create()
    await vi.advanceTimersByTimeAsync(3200)
    await click('下载更新')
    state.phase = 'ready'
    await vi.advanceTimersByTimeAsync(1600)
    await click('安装并重启')
    vi.mocked(fetch).mockImplementationOnce(async () => {
      state.phase = 'backingUp'
      return {
        ok: true,
        json: async () => ({ status: false, msg: '当前阶段无法取消' }),
      } as Response
    })
    await click('取消安装')
    expect(wrapper!.text()).toContain('当前阶段无法取消')
    await vi.advanceTimersByTimeAsync(1600)
    expect(wrapper!.text()).toContain('正在备份')
    expect(wrapper!.text()).toContain('当前阶段无法取消')
  })
  it('rechecks after leaving and returning home, while keeping the daily attempt limit', async () => {
    failedCheck = true
    create()
    await vi.advanceTimersByTimeAsync(3200)
    for (let visit = 0; visit < 5; visit++) {
      route.current!.path = '/settings'
      await flushPromises()
      await vi.advanceTimersByTimeAsync(60_000)
      route.current!.path = '/'
      await flushPromises()
      await vi.advanceTimersByTimeAsync(3200)
    }
    expect(checks).toBe(5)
    expect(readUpdatePolicy().attempts).toBe(5)
    expect(updateVisible.value).toBe(false)
  })
  it('recovers a download accepted by the server when the POST response is lost', async () => {
    create()
    await vi.advanceTimersByTimeAsync(3200)
    vi.mocked(fetch).mockImplementationOnce(async () => {
      state = { phase: 'downloading', version: '3.4.0', job: 'job-1', downloaded: 0, total: 100 }
      throw new Error('response lost')
    })
    await click('下载更新')
    expect(wrapper!.text()).toContain('正在下载')
    expect(wrapper!.text()).not.toContain('response lost')
    state.phase = 'ready'
    await vi.advanceTimersByTimeAsync(1600)
    expect(wrapper!.text()).toContain('下载完成')
  })
  it('does not let an older status response overwrite an accepted installation', async () => {
    create()
    await vi.advanceTimersByTimeAsync(3200)
    await click('下载更新')
    state.phase = 'ready'
    await vi.advanceTimersByTimeAsync(1600)
    const oldState = { ...state }
    let release!: () => void
    vi.mocked(fetch).mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          release = () =>
            resolve({ ok: true, json: async () => ({ status: true, data: oldState }) } as Response)
        }),
    )
    updateManualRequest.value++
    await flushPromises()
    await click('安装并重启')
    expect(wrapper!.text()).toContain('等待任务完成')
    release()
    await flushPromises()
    expect(wrapper!.text()).toContain('等待任务完成')
  })
  it('clears mute when unchecked and restores the saved checkbox without extending it', async () => {
    create()
    await vi.advanceTimersByTimeAsync(3200)
    await wrapper!.find('input[type="checkbox"]').setValue(true)
    expect(readUpdatePolicy().mutedUntil).toBeGreaterThan(Date.now())
    await wrapper!.find('input[type="checkbox"]').setValue(false)
    expect(readUpdatePolicy().mutedUntil).toBe(0)
    await wrapper!.find('input[type="checkbox"]').setValue(true)
    const until = readUpdatePolicy().mutedUntil
    updateVisible.value = false
    await flushPromises()
    await vi.advanceTimersByTimeAsync(60_000)
    updateVisible.value = true
    await flushPromises()
    expect((wrapper!.find('input[type="checkbox"]').element as HTMLInputElement).checked).toBe(true)
    expect(readUpdatePolicy().mutedUntil).toBe(until)
  })
  it('continues tracking an accepted download when the first status query fails', async () => {
    create()
    await vi.advanceTimersByTimeAsync(3200)
    disconnected = true
    await click('下载更新')
    expect(wrapper!.text()).toContain('正在下载')
    disconnected = false
    state.phase = 'ready'
    await vi.advanceTimersByTimeAsync(1600)
    expect(wrapper!.text()).toContain('下载完成')
  })
  it('silently skips a failed automatic check without retrying on a timer', async () => {
    failedCheck = true
    create()
    await vi.advanceTimersByTimeAsync(3200)
    expect(checks).toBe(1)
    expect(updateVisible.value).toBe(false)
    expect(readUpdatePolicy().attempts).toBe(1)
    await vi.advanceTimersByTimeAsync(600_000)
    expect(checks).toBe(1)
  })
  it('prompts once and skips another check when the homepage is reopened the same day', async () => {
    create()
    await vi.advanceTimersByTimeAsync(3200)
    expect(wrapper!.text()).toContain('发现新版本')
    expect(readUpdatePolicy().prompted).toBe(true)
    wrapper!.unmount()
    updateVisible.value = false
    create()
    await vi.advanceTimersByTimeAsync(3200)
    expect(checks).toBe(1)
    expect(updateVisible.value).toBe(false)
  })
  it('downloads, requests installation separately and reconnects after a fast restart', async () => {
    create()
    await vi.advanceTimersByTimeAsync(3200)
    await click('下载更新')
    expect(state.phase).toBe('downloading')
    state.phase = 'ready'
    await vi.advanceTimersByTimeAsync(1600)
    expect(wrapper!.text()).toContain('下载完成')
    await click('安装并重启')
    expect(state.phase).toBe('waiting')
    disconnected = true
    await vi.advanceTimersByTimeAsync(1600)
    expect(wrapper!.text()).toContain('正在安装')
    disconnected = false
    state.phase = 'installed'
    await vi.advanceTimersByTimeAsync(1600)
    expect(wrapper!.text()).toContain('更新完成，刷新页面')
  })
})
