import { beforeEach, expect, it, vi } from 'vitest'
import { filesBasesServer } from '../filesBases.server'
import type { I_config_app } from '@/dataType/config.dataType'

const mocks = vi.hoisted(() => ({ request: vi.fn(), status: vi.fn(), follow: vi.fn(), confirm: vi.fn() }))
vi.mock('@/assets/request', () => ({ default: mocks.request }))
vi.mock('../sharedConfig.server', () => ({ sharedConfigServer: { status: mocks.status, follow: mocks.follow } }))
vi.mock('element-plus', () => ({ ElMessageBox: { confirm: mocks.confirm } }))
const config = { videoMomentLibraryMode: 'collection', sampleFolder: 'keep-folder' } as I_config_app
beforeEach(() => {
  vi.resetAllMocks()
  mocks.status.mockResolvedValue({ status: true, data: {
    following: true, fields: ['videoMomentLibraryMode'], config: { videoMomentLibraryMode: 'theater' }, revision: 7,
  } })
  mocks.follow.mockResolvedValue({ status: true, data: true })
  mocks.confirm.mockResolvedValue('confirm')
  mocks.request.mockResolvedValue({ status: true, data: true })
})
it('跟随公共配置时只将当前库改为自定义，然后保存模式与其他配置', async () => {
  await filesBasesServer.setFilesBasesConfigById('library-a', config)
  expect(mocks.confirm).toHaveBeenCalledOnce()
  expect(mocks.follow).toHaveBeenCalledExactlyOnceWith('library-a', 'display', false, 7)
  expect(mocks.follow.mock.invocationCallOrder[0]).toBeLessThan(mocks.request.mock.invocationCallOrder[0])
  expect(mocks.request).toHaveBeenCalledExactlyOnceWith({ url: '/filesBases/setConfig/filesBases', method: 'put', data: {
    id: 'library-a', config: JSON.stringify(config),
  } })
})
it('取消改为自定义时不改变跟随关系，也不保存配置', async () => {
  mocks.confirm.mockRejectedValue('cancel')
  const result = await filesBasesServer.setFilesBasesConfigById('library-a', config)
  expect(result.status).toBe(false)
  expect(mocks.follow).not.toHaveBeenCalled()
  expect(mocks.request).not.toHaveBeenCalled()
})
it('跟随状态版本冲突时不继续保存', async () => {
  mocks.follow.mockResolvedValue({ status: false, msg: '公共配置已更新，请重试' })
  const result = await filesBasesServer.setFilesBasesConfigById('library-a', config)
  expect(result.status).toBe(false)
  expect(mocks.request).not.toHaveBeenCalled()
})
