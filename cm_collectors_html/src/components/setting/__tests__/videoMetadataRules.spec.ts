import { shallowMount, flushPromises } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import VideoMetadataSetting from '../videoMetadata/videoMetadataSetting.vue'
import { videoMetadataServer } from '@/server/videoMetadata.server'
import { ElMessage, ElMessageBox } from 'element-plus'

vi.mock('@/server/videoMetadata.server', () => ({ videoMetadataServer: {
  setClassification: vi.fn(), setting: vi.fn(), stats: vi.fn(), taskStatus: vi.fn(), cleanupStatus: vi.fn(),
  failures: vi.fn(), saveSetting: vi.fn(), cleanupPreview: vi.fn(), startCleanup: vi.fn(), bulk: vi.fn(),
} }))
vi.mock('@/components/com/form/selectFilesBases.vue', () => ({ default: { template: '<div />' } }))
vi.mock('element-plus', () => ({
  ElMessage: { error: vi.fn(), success: vi.fn(), info: vi.fn(), warning: vi.fn() },
  ElMessageBox: { confirm: vi.fn() },
}))
const ok = <T>(data: T) => ({ status: true, statusCode: 200, msg: '', data })
const row = { dramaSeriesId: 'one', src: 'file.pdf' }
let wrapper: ReturnType<typeof shallowMount>
const mountPage = async () => {
  wrapper = shallowMount(VideoMetadataSetting, {
    global: { directives: { loading: {} }, stubs: {
      ElCard: { template: '<section><slot /><slot name="header" /></section>' },
      ElButton: { template: '<button><slot /></button>' },
      ElTableColumn: true, ElDialog: true,
      ...Object.fromEntries(['ElSwitch','ElFormItem','ElInput','ElForm','ElAlert','ElRadioButton','ElRadioGroup','ElInputNumber','ElOption','ElSelect','ElProgress','ElTable','ElTabPane','ElPagination','ElEmpty','ElTabs'].map(name => [name, true])),
    } },
  })
  await flushPromises()
  return wrapper.vm as unknown as {
    settingData: { setting: { excludedExtensions: string } },
    selectedFailures: typeof row[],
    failureQuery: { excluded: boolean }, failures: typeof row[],
    saveSetting: () => Promise<void>, previewCleanup: () => Promise<void>,
    markNonVideo: (item: typeof row) => Promise<void>,
    bulkAction: (action: string) => Promise<void>, loadFailures: () => Promise<void>,
  }
}
beforeEach(() => {
  vi.resetAllMocks()
  vi.mocked(videoMetadataServer.setting).mockResolvedValue(ok({ setting: {
    id: 'default', autoExcludeNonVideo: true, excludedExtensions: '', collectOnNewOrChanged: true,
    collectOnDetailOrPlay: true, collectOnList: false, idleBackfillEnabled: false, idleScopeMode: 'selected',
    idleWaitMinutes: 5, probeIntervalMilliseconds: 1500, idleBatchSize: 20, paused: false,
  }, filesBasesIds: [] }))
  vi.mocked(videoMetadataServer.stats).mockResolvedValue(ok([]))
  vi.mocked(videoMetadataServer.taskStatus).mockResolvedValue(ok({ id: '', scopeMode: 'selected', runMode: 'missing_stale', status: '', total: 0, success: 0, failed: 0, skipped: 0, currentSrc: '', lastError: '', createdAt: '', startedAt: '', finishedAt: '' }))
  vi.mocked(videoMetadataServer.cleanupStatus).mockResolvedValue(ok({ status: 'idle', processed: 0, total: 0, error: '' }))
  vi.mocked(videoMetadataServer.failures).mockResolvedValue(ok({ dataList: [], total: 0 }))
  vi.mocked(ElMessageBox.confirm).mockResolvedValue('confirm' as Awaited<ReturnType<typeof ElMessageBox.confirm>>)
})
afterEach(() => wrapper?.unmount())
describe('视频采集排除设置', () => {
  it('用户取消排除视频格式时不保存', async () => {
    const vm = await mountPage()
    vm.settingData.setting.excludedExtensions = '.MP4, pdf'
    vi.mocked(ElMessageBox.confirm).mockRejectedValueOnce('cancel')
    await vm.saveSetting()
    expect(ElMessageBox.confirm).toHaveBeenCalledWith(expect.stringContaining('mp4'), '排除视频格式', expect.anything())
    expect(videoMetadataServer.saveSetting).not.toHaveBeenCalled()
  })
  it('未保存的排除规则不能触发历史整理', async () => {
    const vm = await mountPage()
    vm.settingData.setting.excludedExtensions = 'pdf'
    await vm.previewCleanup()
    expect(videoMetadataServer.cleanupPreview).not.toHaveBeenCalled()
    expect(ElMessage.warning).toHaveBeenCalledWith(expect.stringContaining('尚未保存'))
  })
  it('预览取消不整理，确认后传递预览规则版本', async () => {
    const vm = await mountPage()
    vi.mocked(videoMetadataServer.cleanupPreview).mockResolvedValue(ok({ total: 5001, byExtension: { '.jpg': 5001 }, token: 'rules-v1' }))
    vi.mocked(ElMessageBox.confirm).mockRejectedValueOnce('cancel')
    await vm.previewCleanup()
    expect(videoMetadataServer.startCleanup).not.toHaveBeenCalled()
    vi.mocked(videoMetadataServer.startCleanup).mockResolvedValue(ok(true))
    await vm.previewCleanup()
    expect(videoMetadataServer.startCleanup).toHaveBeenCalledWith('rules-v1')
    expect(ElMessageBox.confirm).toHaveBeenCalledWith(expect.stringContaining('5001'), '预览历史记录整理', expect.anything())
  })
  it('批量恢复允许部分失败并刷新已排除列表', async () => {
    const vm = await mountPage()
    vm.failureQuery.excluded = true
    vm.selectedFailures = [row, { ...row, dramaSeriesId: 'two' }]
    vi.mocked(videoMetadataServer.bulk).mockResolvedValue(ok({ succeeded: 1, failed: { two: '文件不存在' } }))
    await vm.bulkAction('video')
    expect(videoMetadataServer.bulk).toHaveBeenCalledWith(['one', 'two'], 'video')
    expect(ElMessage.warning).toHaveBeenCalledWith(expect.stringContaining('成功 1 项，失败 1 项'))
    expect(vm.selectedFailures).toEqual([])
    expect(videoMetadataServer.failures).toHaveBeenLastCalledWith(expect.objectContaining({ excluded: true }))
  })
  it('切换列表后迟到的旧请求不能覆盖当前结果', async () => {
    const vm = await mountPage()
    let resolveOld!: (value: Awaited<ReturnType<typeof videoMetadataServer.failures>>) => void
    vi.mocked(videoMetadataServer.failures).mockImplementationOnce(() => new Promise(resolve => { resolveOld = resolve }))
    const old = vm.loadFailures()
    vm.failureQuery.excluded = true
    await vm.loadFailures()
    resolveOld(ok({ dataList: [row as never], total: 1 }))
    await old
    expect(vm.failures).toEqual([])
  })
})

it('有勾选项时，单条标记非视频后刷新列表并清空选择', async () => {
  const vm = await mountPage()
  vm.failures = [row]
  vm.selectedFailures = [row]
  vi.mocked(videoMetadataServer.setClassification).mockResolvedValue(ok(true))
  await vm.markNonVideo(row)
  expect(vm.selectedFailures).toEqual([])
  expect(vm.failures).toEqual([])
})

it('自定义规则保存采用服务端规范化结果，随后可以预览整理', async () => {
  const vm = await mountPage()
  const setting = (await videoMetadataServer.setting()).data
  vm.settingData.setting.excludedExtensions = '.PDF；zip, PDF'
  vi.mocked(videoMetadataServer.saveSetting).mockResolvedValue(ok({ ...setting, setting: { ...setting.setting, excludedExtensions: 'pdf,zip' } }))
  await vm.saveSetting()
  expect(videoMetadataServer.saveSetting).toHaveBeenCalledWith(expect.objectContaining({ setting: expect.objectContaining({ excludedExtensions: '.PDF；zip, PDF' }) }))
  expect(vm.settingData.setting.excludedExtensions).toBe('pdf,zip')
  vi.mocked(videoMetadataServer.cleanupPreview).mockResolvedValue(ok({ total: 0, byExtension: {}, token: 'new-rules' }))
  await vm.previewCleanup()
  expect(videoMetadataServer.cleanupPreview).toHaveBeenCalledTimes(1)
  expect(videoMetadataServer.startCleanup).not.toHaveBeenCalled()
})
