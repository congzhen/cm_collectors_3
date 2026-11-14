import { mount, flushPromises } from '@vue/test-utils'
import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest'
import ImportDialog from '../modeScanDiskImportDataDialog.vue'
import { defualtConfigScanDisk } from '@/dataType/config.dataType'
import { importDataServer } from '@/server/importData.server'

vi.mock('@/components/com/dialog/dialog-common.vue', () => ({ default: { template: '<div />' } }))
vi.mock('@/server/importData.server', () => ({
  importDataServer: { scanDiskImportData: vi.fn(), batch: vi.fn() },
}))
vi.mock('element-plus', () => ({ ElTable: { name: 'ElTable', template: '<div />' } }))
const options = {
  global: {
    stubs: {
      dialogCommon: {
        name: 'TaskDialogStub',
        emits: ['submit', 'closed'],
        methods: { open() {}, disabledSubmit() {} },
        template:
          '<div><button class="start" @click="$emit(\'submit\')">开始</button><slot /></div>',
      },
      ElTable: {
        methods: { getSelectionRows: () => [], toggleAllSelection() {} },
        template: '<div />',
      },
      ElTableColumn: true,
      ElIcon: true,
      Loading: true,
      ElButton: { template: '<button class="cancel"><slot /></button>' },
    },
  },
}
const ok = { status: true, statusCode: 0, msg: '', data: { warnings: [] as string[] } }

describe('导入结果与批次生命周期', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(importDataServer.batch).mockResolvedValue({
      status: true,
      statusCode: 0,
      msg: '',
      data: 'batch',
    })
    vi.mocked(importDataServer.scanDiskImportData).mockResolvedValue(ok)
  })
  afterEach(() => vi.useRealTimers())
  it('320 条已有封面导入逐条完成后推进进度，在途请求不提前计数', async () => {
    let complete: ((value: typeof ok) => void) | undefined
    vi.mocked(importDataServer.scanDiskImportData).mockImplementation(
      () => new Promise((resolve) => { complete = resolve }),
    )
    const w = mount(ImportDialog, options)
    w.vm.open(
      Array.from({ length: 320 }, (_, i) => `/local-posters/${i}/movie.mp4`),
      defualtConfigScanDisk,
      'library',
    )
    await w.get('.start').trigger('click')
    await flushPromises()
    for (let i = 0; i < 320; i++) {
      expect(w.get('[role="progressbar"]').attributes('aria-valuenow')).toBe(String(i))
      expect(w.get('[role="progressbar"]').attributes('aria-valuemax')).toBe('320')
      expect(importDataServer.scanDiskImportData).toHaveBeenCalledTimes(i + 1)
      expect(complete).toBeDefined()
      const resolve = complete!
      complete = undefined
      resolve(ok)
      await flushPromises()
    }
    expect(w.get('[role="progressbar"]').attributes('aria-valuenow')).toBe('320')
    expect(w.text()).toContain('已处理 320 / 320，成功 320，警告 0，失败 0')
    expect(importDataServer.batch).toHaveBeenCalledWith('end', 'library', 'batch')
    w.unmount()
  })
  it('320 条按实际结果汇总，开始与结束各登记一次', async () => {
    vi.mocked(importDataServer.scanDiskImportData)
      .mockResolvedValueOnce({ ...ok, data: { warnings: ['封面超限，使用默认封面'] } })
      .mockResolvedValueOnce({ status: false, statusCode: 1, msg: '写入失败', data: false })
    const w = mount(ImportDialog, options)
    w.vm.open(
      Array.from({ length: 320 }, (_, i) => `file-${i}`),
      defualtConfigScanDisk,
      'library',
    )
    await w.get('.start').trigger('click')
    await flushPromises()
    expect(importDataServer.scanDiskImportData).toHaveBeenCalledTimes(320)
    expect(w.text()).toContain('已处理 320 / 320，成功 318，警告 1，失败 1')
    expect(importDataServer.batch).toHaveBeenCalledWith('end', 'library', 'batch')
    expect(importDataServer.batch).toHaveBeenCalledTimes(2)
    w.unmount()
  })
  it('连接中断后停止后续请求，当前结果待确认', async () => {
    vi.mocked(importDataServer.scanDiskImportData).mockResolvedValueOnce({
      status: false,
      statusCode: 1,
      msg: 'Network Error',
      data: false,
      transportError: true,
    })
    const w = mount(ImportDialog, options)
    w.vm.open(['a', 'b', 'c'], defualtConfigScanDisk, 'library')
    await w.get('.start').trigger('click')
    await flushPromises()
    expect(importDataServer.scanDiskImportData).toHaveBeenCalledTimes(1)
    expect(w.text()).toContain('已处理 0 / 3')
    expect(w.text()).toContain('待确认 1')
    expect(importDataServer.batch).toHaveBeenCalledWith('end', 'library', 'batch')
    w.unmount()
  })
  it('取消会传递 AbortSignal，且仍释放批次', async () => {
    vi.mocked(importDataServer.scanDiskImportData).mockImplementation(
      (_id, _path, _config, _batch, signal) =>
        new Promise((resolve) => {
          signal?.addEventListener('abort', () =>
            resolve({
              status: false,
              statusCode: 1,
              msg: 'canceled',
              data: false,
              transportError: true,
            }),
          )
        }),
    )
    const w = mount(ImportDialog, options)
    w.vm.open(['a', 'b'], defualtConfigScanDisk, 'library')
    await w.get('.start').trigger('click')
    await flushPromises()
    await w.get('.cancel').trigger('click')
    await flushPromises()
    expect(importDataServer.scanDiskImportData).toHaveBeenCalledTimes(1)
    expect(vi.mocked(importDataServer.scanDiskImportData).mock.calls[0][4]?.aborted).toBe(true)
    expect(importDataServer.batch).toHaveBeenCalledWith('end', 'library', 'batch')
    expect(w.text()).toContain('待确认 1')
    w.unmount()
  })
  it('心跳失败会中断在途请求，避免批次过期后继续导入', async () => {
    vi.useFakeTimers()
    vi.mocked(importDataServer.batch)
      .mockResolvedValueOnce({ status: true, statusCode: 0, msg: '', data: 'batch' })
      .mockResolvedValueOnce({ status: false, statusCode: 1, msg: 'expired', data: false })
    vi.mocked(importDataServer.scanDiskImportData).mockImplementation(
      (_id, _path, _config, _batch, signal) =>
        new Promise((resolve) => {
          signal?.addEventListener('abort', () =>
            resolve({
              status: false,
              statusCode: 1,
              msg: 'canceled',
              data: false,
              transportError: true,
            }),
          )
        }),
    )
    const w = mount(ImportDialog, options)
    w.vm.open(['a', 'b'], defualtConfigScanDisk, 'library')
    await w.get('.start').trigger('click')
    await flushPromises()
    await vi.advanceTimersByTimeAsync(20000)
    await flushPromises()
    expect(importDataServer.scanDiskImportData).toHaveBeenCalledTimes(1)
    expect(importDataServer.batch).toHaveBeenCalledWith(
      'heartbeat',
      'library',
      'batch',
      expect.any(AbortSignal),
    )
    expect(importDataServer.batch).toHaveBeenCalledWith('end', 'library', 'batch')
    w.unmount()
  })
  it('只选择部分资源时以本次选择数量作为进度分母', async () => {
    const w = mount(ImportDialog, options)
    w.vm.open(['a', 'b', 'c'], defualtConfigScanDisk, 'library')
    await flushPromises()
    const table = w.getComponent({ ref: 'tableRef' })
    const rows = table.vm.$attrs.data as unknown[]
    vi.spyOn(table.vm, 'getSelectionRows').mockReturnValue(rows.slice(0, 1))
    await w.get('.start').trigger('click')
    await flushPromises()
    expect(importDataServer.scanDiskImportData).toHaveBeenCalledTimes(1)
    expect(w.text()).toContain('已处理 1 / 1，成功 1')
    w.unmount()
  })
  it('旧请求在重新打开后完成，不污染新批次的状态', async () => {
    let finish!: (value: typeof ok) => void
    vi.mocked(importDataServer.scanDiskImportData).mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          finish = resolve
        }),
    )
    const w = mount(ImportDialog, options)
    w.vm.open(['old', 'old-pending'], defualtConfigScanDisk, 'old-library')
    await w.get('.start').trigger('click')
    await flushPromises()
    w.getComponent({ name: 'TaskDialogStub' }).vm.$emit('closed')
    w.vm.open(['new'], defualtConfigScanDisk, 'new-library')
    await w.get('.start').trigger('click')
    await flushPromises()
    finish({ ...ok, data: { warnings: ['old warning'] } })
    await flushPromises()
    expect(importDataServer.scanDiskImportData).toHaveBeenCalledTimes(2)
    expect(w.text()).toContain('已处理 1 / 1，成功 1，警告 0')
    expect(importDataServer.batch).toHaveBeenCalledWith('end', 'old-library', 'batch')
    expect(importDataServer.batch).toHaveBeenCalledWith('end', 'new-library', 'batch')
    w.unmount()
  })
  it('单条请求发现批次过期时立即停止，不继续发送剩余文件', async () => {
    vi.mocked(importDataServer.scanDiskImportData).mockResolvedValue({
      status: false,
      statusCode: 409,
      msg: '批次已过期',
      data: false,
    })
    const w = mount(ImportDialog, options)
    w.vm.open(['a', 'b', 'c'], defualtConfigScanDisk, 'library')
    await w.get('.start').trigger('click')
    await flushPromises()
    expect(importDataServer.scanDiskImportData).toHaveBeenCalledTimes(1)
    expect(w.text()).toContain('已处理 1 / 3，成功 0，警告 0，失败 1')
    expect(w.text()).toContain('已停止后续导入')
    w.unmount()
  })
  it('重试开始后，上一轮迟到的心跳失败不覆盖新状态', async () => {
    vi.useFakeTimers()
    type BatchResponse = Awaited<ReturnType<typeof importDataServer.batch>>
    type ImportResponse = Awaited<ReturnType<typeof importDataServer.scanDiskImportData>>
    let finishHeartbeat!: (value: BatchResponse) => void
    let finishFirst!: (value: ImportResponse) => void
    let finishSecond!: (value: ImportResponse) => void
    vi.mocked(importDataServer.batch).mockImplementation((action) =>
      action === 'heartbeat'
        ? new Promise((resolve) => {
            finishHeartbeat = resolve
          })
        : Promise.resolve({ status: true, statusCode: 0, msg: '', data: 'batch' }),
    )
    vi.mocked(importDataServer.scanDiskImportData)
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            finishFirst = resolve
          }),
      )
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            finishSecond = resolve
          }),
      )
    const w = mount(ImportDialog, options)
    w.vm.open(['a'], defualtConfigScanDisk, 'library')
    await w.get('.start').trigger('click')
    await flushPromises()
    await vi.advanceTimersByTimeAsync(20000)
    finishFirst({ status: false, statusCode: 1, msg: 'failed', data: false })
    await flushPromises()
    await w.get('.start').trigger('click')
    await flushPromises()
    finishHeartbeat({ status: false, statusCode: 1, msg: 'late failure', data: false })
    await flushPromises()
    expect(w.text()).not.toContain('批次连接异常')
    expect(vi.mocked(importDataServer.scanDiskImportData).mock.calls[1][4]?.aborted).toBe(false)
    const heartbeatCall = vi
      .mocked(importDataServer.batch)
      .mock.calls.find((call) => call[0] === 'heartbeat')
    expect(heartbeatCall?.[3]?.aborted).toBe(true)
    finishSecond(ok)
    await flushPromises()
    w.unmount()
  })
})
