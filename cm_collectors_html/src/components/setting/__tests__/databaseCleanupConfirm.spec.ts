import { shallowMount, flushPromises } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import Cleanup from '../databaseCleanup/databaseCleanup.vue'
import { appDataServer } from '@/server/app.server'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { I_databaseCleanupForm } from '@/dataType/other.dataType'

vi.mock('@/assets/debounce', () => ({ debounceNow: (fn: () => unknown) => fn }))
vi.mock('@/server/app.server', () => ({ appDataServer: { dbBackupList: vi.fn(), databaseCleanup: vi.fn(), deleteDbBackup: vi.fn() } }))
vi.mock('@/storeData/filesBases.storeData', () => ({ filesBasesStoreData: () => ({ getFilesBasesNameById: (id: string) => ({ a: '媒体库 A', b: '媒体库 B' })[id as 'a' | 'b'] || '' }) }))
vi.mock('@/components/com/form/selectFilesBases.vue', () => ({ default: { template: '<div />' } }))
vi.mock('element-plus', () => ({ ElMessage: { warning: vi.fn(), error: vi.fn(), success: vi.fn() }, ElMessageBox: { confirm: vi.fn() } }))
const confirmed = 'confirm' as Awaited<ReturnType<typeof ElMessageBox.confirm>>
const ok = <T>(data: T) => ({ status: true, statusCode: 200, msg: '', data })
let wrapper: ReturnType<typeof shallowMount>
const mountPage = async () => {
  wrapper = shallowMount(Cleanup, { global: { directives: { loading: {} }, stubs: Object.fromEntries(['ElForm','ElFormItem','ElCheckboxGroup','ElCheckbox','ElButton','ElText'].map(n => [n, true])) } })
  await flushPromises()
  return wrapper.vm as unknown as { formData: I_databaseCleanupForm; clearHandle: () => Promise<void>; loading: boolean }
}
const text = (value: any): string => typeof value === 'string' ? value : Array.isArray(value) ? value.map(text).join('') : value && typeof value === 'object' ? text(value.children) : ''
const message = () => text(vi.mocked(ElMessageBox.confirm).mock.calls[0][0])
beforeEach(() => {
  vi.resetAllMocks()
  vi.mocked(appDataServer.dbBackupList).mockResolvedValue(ok([]))
  vi.mocked(appDataServer.databaseCleanup).mockResolvedValue(ok(true))
  vi.mocked(ElMessageBox.confirm).mockResolvedValue(confirmed)
})
afterEach(() => wrapper?.unmount())
describe('数据库清理二次确认', () => {
  it('未选清理内容时提示，不弹窗也不提交', async () => {
    const vm = await mountPage(); await vm.clearHandle()
    expect(ElMessage.warning).toHaveBeenCalledWith('请选择要清理的数据')
    expect(ElMessageBox.confirm).not.toHaveBeenCalled(); expect(appDataServer.databaseCleanup).not.toHaveBeenCalled()
  })
  it.each(['cancel', 'close'])('%s 不提交，保留选择并恢复交互', async action => {
    const vm = await mountPage(); vm.formData.clear_items = ['resource']
    vi.mocked(ElMessageBox.confirm).mockRejectedValue(action)
    await vm.clearHandle()
    expect(message()).toContain('全部文件库（未指定文件库）'); expect(message()).toContain('清理内容：资源')
    expect(appDataServer.databaseCleanup).not.toHaveBeenCalled(); expect(vm.formData.clear_items).toEqual(['resource']); expect(vm.loading).toBe(false)
  })
  it('显示指定库名和全部选中内容，确认后沿用原请求并刷新备份', async () => {
    const vm = await mountPage(); vm.formData = { filesBases_ids: ['a','b'], clear_items: ['resource','tags','tagClass','cronJobs'] }
    await vm.clearHandle()
    expect(message()).toContain('指定 2 个文件库：媒体库 A、媒体库 B')
    expect(message()).toContain('资源、标签、标签分类、计划任务')
    expect(appDataServer.databaseCleanup).toHaveBeenCalledExactlyOnceWith(vm.formData)
    expect(appDataServer.dbBackupList).toHaveBeenCalledTimes(2)
  })
  it('显示全局配置与共用演员集影响，未知库保留 ID', async () => {
    const vm = await mountPage(); vm.formData = { filesBases_ids: ['missing'], clear_items: ['performer','generalConfig'] }
    await vm.clearHandle()
    expect(message()).toContain('未知文件库（missing）'); expect(message()).toContain('不受所选文件库范围限制'); expect(message()).toContain('其他文件库')
  })
  it('确认期间防止重复提交，使用已展示的范围快照', async () => {
    const vm = await mountPage(); vm.formData = { filesBases_ids: ['a'], clear_items: ['resource'] }
    let resolve!: () => void
    vi.mocked(ElMessageBox.confirm).mockReturnValue(new Promise(r => { resolve = () => r(confirmed) }))
    const pending = vm.clearHandle(); await vm.clearHandle()
    vm.formData.filesBases_ids.push('b'); vm.formData.clear_items.push('generalConfig')
    expect(appDataServer.databaseCleanup).not.toHaveBeenCalled(); expect(ElMessageBox.confirm).toHaveBeenCalledTimes(1)
    resolve(); await pending
    expect(appDataServer.databaseCleanup).toHaveBeenCalledExactlyOnceWith({ filesBases_ids: ['a'], clear_items: ['resource'] })
  })
  it('接口拒绝时保留表单且可重试，不误报成功', async () => {
    const vm = await mountPage(); vm.formData.clear_items = ['tags']
    vi.mocked(appDataServer.databaseCleanup).mockResolvedValue({ ...ok(false), status: false, msg: '清理失败' })
    await vm.clearHandle()
    expect(ElMessage.error).toHaveBeenCalledWith('清理失败'); expect(ElMessage.success).not.toHaveBeenCalled(); expect(vm.loading).toBe(false); expect(vm.formData.clear_items).toEqual(['tags'])
  })
})
