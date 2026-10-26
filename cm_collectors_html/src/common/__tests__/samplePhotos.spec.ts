import { describe, expect, it, vi } from 'vitest'
import { getSamplePhoto } from '../photo'
import { resourceServer } from '@/server/resource.server'
import type { I_resource_base } from '@/dataType/resource.dataType'
const mocks = vi.hoisted(() => ({ request: vi.fn(async () => ({ status: true, data: [] })) }))
vi.mock('@/assets/request', () => ({ default: mocks.request }))
vi.mock('@/storeData/app.storeData', () => ({ appStoreData: () => ({ currentConfigApp: { sampleFolder: 'old-folder' } }) }))
describe('剧照固定使用视频所在目录', () => {
  it('列表查询不携带子文件夹，图片地址不追加旧配置', async () => {
    await resourceServer.sampleImages('resource')
    expect(mocks.request).toHaveBeenCalledWith({ url: '/resource/sampleImages/resource', method: 'get' })
    const resource = { id: 'resource', coverPoster: 'poster.jpg' } as I_resource_base
    expect(getSamplePhoto(resource, '剧照 1.jpg')).toBe(`/api/resource/sampleData/resource?q=${encodeURIComponent('剧照 1.jpg')}`)
  })
})
