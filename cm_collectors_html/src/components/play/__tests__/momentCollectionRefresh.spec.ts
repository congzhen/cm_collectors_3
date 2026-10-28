import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import MomentCollection from '../MomentCollection.vue'
import { momentApi } from '@/server/videoMoments.server'
import { defineComponent, h, KeepAlive, ref } from 'vue'
import ElementPlus, { ElPagination } from 'element-plus'

const mocks = vi.hoisted(() => ({ request: vi.fn() }))
vi.mock('@/assets/request', () => ({ default: (...args: unknown[]) => mocks.request(...args) }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))
vi.mock('@/assets/mobile', () => ({ isMobile: () => false }))
const item = {
  id: 'm',
  dramaSeriesId: 'd',
  resourceId: 'r',
  title: '新增片段',
  start: 10,
  end: 20,
  available: true,
}

describe('收藏列表自动同步', () => {
  let items: (typeof item)[]
  let success: boolean
  beforeEach(() => {
    items = []
    success = true
    mocks.request.mockReset()
    mocks.request.mockImplementation(async ({ method }: { method?: string }) => {
      if (!method) return { status: true, data: [...items] }
      return { status: success, data: method === 'delete' ? true : item }
    })
  })
  const render = () =>
    mount(MomentCollection, {
      props: { resourceId: 'r' },
      global: { stubs: { MomentCover: true, 'el-button': true } },
    })

  it('新增、编辑、删除成功后自动更新卡片，无需点击刷新', async () => {
    const wrapper = render()
    await flushPromises()
    expect(wrapper.findAll('.collection-card')).toHaveLength(0)
    items = [item]
    await momentApi.save({ dramaSeriesId: 'd', title: item.title, start: 10, end: 20 })
    await flushPromises()
    expect(wrapper.text()).toContain('新增片段')
    items = [{ ...item, title: '编辑后的片段' }]
    await momentApi.save({ ...items[0] })
    await flushPromises()
    expect(wrapper.text()).toContain('编辑后的片段')
    expect(wrapper.text()).not.toContain('新增片段')
    items = []
    await momentApi.remove('m')
    await flushPromises()
    expect(wrapper.findAll('.collection-card')).toHaveLength(0)
    wrapper.unmount()
  })

  it('失败和预览生成不会触发列表刷新，卸载后取消订阅', async () => {
    const wrapper = render()
    await flushPromises()
    mocks.request.mockClear()
    success = false
    await momentApi.save(item)
    await momentApi.remove('m')
    success = true
    await momentApi.queue('d', 30)
    await flushPromises()
    expect(mocks.request.mock.calls.filter(([args]) => !args.method)).toHaveLength(0)
    wrapper.unmount()
    mocks.request.mockClear()
    await momentApi.save(item)
    await flushPromises()
    expect(mocks.request).toHaveBeenCalledTimes(1)
  })
  it('从播放页返回缓存收藏列表时保留当前库页码和滚动位置', async () => {
    mocks.request.mockImplementation(async ({ params }: { params: { page: number } }) => ({ status: true,
      data: { total: 65, items: [{ ...item, title: '第' + params.page + '页' }] },
    }))
    const active = ref(true)
    const wrapper = mount(defineComponent({ setup: () => () => h(KeepAlive, null, {
      default: () => active.value ? h(MomentCollection, { filesBasesId: 'library-a' }) : h('div'),
    }) }), { global: { plugins: [ElementPlus], stubs: { MomentCover: true } } })
    await flushPromises()
    wrapper.findComponent(ElPagination).vm.$emit('current-change', 2)
    await flushPromises()
    const list = wrapper.get('.collection-list').element
    Object.defineProperty(list, 'scrollTo', { value: ({ top }: { top: number }) => { list.scrollTop = top } })
    list.scrollTop = 137
    active.value = false
    await flushPromises()
    active.value = true
    await flushPromises()
    expect(wrapper.findComponent(ElPagination).props('currentPage')).toBe(2)
    expect(list.scrollTop).toBe(137)
    expect(wrapper.get('.collection-card').text()).toContain('第2页')
    wrapper.unmount()
  })
})
