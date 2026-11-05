import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, expect, it, vi } from 'vitest'
import Card from '../CategoryBrowseCard.vue'
import type { CategoryGroup, CategoryQuery } from '@/server/categoryBrowse.server'
const mocks = vi.hoisted(() => ({ members: vi.fn(), play: vi.fn() }))
vi.mock('@/server/categoryBrowse.server', () => ({
  categoryBrowseServer: { members: mocks.members },
}))
vi.mock('@/common/play', () => ({ playResource: mocks.play }))
vi.mock('@/common/photo', () => ({ getResourceCoverPoster: () => '/cover.jpg' }))
vi.mock('@/storeData/app.storeData', () => ({ appStoreData: () => ({ currentConfigApp: {} }) }))
const group = {
  categoryId: 'a',
  name: '演员甲',
  part: 1,
  parts: 1,
  count: 4,
  total: 4,
  preview: [{ id: 'r', filesBases_id: 'lib', title: '资源' }],
} as CategoryGroup
const query = {
  filesBasesId: 'lib',
  source: 'performer',
  groupLimit: 4,
  page: 1,
  limit: 2,
  searchData: {},
} as CategoryQuery
const create = () =>
  mount(Card, {
    props: { group, query, coverStyle: 'stack', expandStyle: 'fan' },
    global: {
      stubs: {
        ElAvatar: true,
        ElPopover: { props: ['visible'], template: '<div><slot v-if="visible" /></div>' },
        ElButton: { template: '<button><slot /></button>' },
        ResourceGroupCover: true,
        ResourceGroupFan: {
          props: ['loading', 'members', 'error'],
          template: '<div class="preview-content">{{loading ? "loading" : members.length}}</div>',
        },
      },
    },
  })
afterEach(() => {
  vi.useRealTimers()
  mocks.members.mockReset()
})
it.each(['stack', 'triptych', 'book', 'adaptive'] as const)('%s 单资源小组不保留叠牌缩进，分类总量和导航保持原样', async (style) => {
  const w = create()
  try {
    const single = { ...group, part: 3, parts: 3, count: 1, total: 9 }
    await w.setProps({ coverStyle: style, group: single })
    expect(w.find('.category-front.is-stack').exists()).toBe(false)
    expect(w.find('.category-back').exists()).toBe(false)
    expect(w.get('.category-count').text()).toBe('1 个')
    expect(w.get('.category-meta').text()).toContain('共 9 个资源')
    await w.get('.category-cover').trigger('click')
    expect(w.emitted('enter')?.[0]).toEqual([single, false])
    await w.setProps({ group: { ...group, count: 2, preview: [...group.preview, { id: 'second' } as CategoryGroup['preview'][number]] } })
    expect(w.find('.category-front.is-stack').exists()).toBe(style === 'stack')
    expect(w.findAll('.category-back')).toHaveLength(style === 'stack' ? 1 : 0)
    await w.setProps({ group: single })
    expect(w.find('.category-front.is-stack').exists()).toBe(false)
    expect(w.find('.category-back').exists()).toBe(false)
    expect(w.get('.category-part').text()).toBe('3 / 3')
    expect(w.get('.category-meta').text()).toContain('共 9 个资源')
  } finally {
    w.unmount()
  }
})
it('慢请求未完成时移出再进入不重复请求，切换分类后可重新加载并忽略旧结果', async () => {
  vi.useFakeTimers()
  let resolve!: (data: unknown) => void
  mocks.members
    .mockReturnValueOnce(
      new Promise((done) => {
        resolve = done
      }),
    )
    .mockResolvedValue({ status: true, data: { dataList: [{ id: 'new' }] } })
  const w = create()
  try {
    await w.trigger('pointerenter', { clientX: 10, clientY: 10 })
    await vi.advanceTimersByTimeAsync(180)
    expect(mocks.members).toHaveBeenCalledTimes(1)
    await w.trigger('pointerleave', { clientX: 1000, clientY: 1000 })
    await vi.advanceTimersByTimeAsync(400)
    await w.trigger('pointerenter', { clientX: 10, clientY: 10 })
    await vi.advanceTimersByTimeAsync(180)
    expect(mocks.members).toHaveBeenCalledTimes(1)
    await w.setProps({ group: { ...group, categoryId: 'b', name: '演员乙' } })
    await w.trigger('pointerenter', { clientX: 10, clientY: 10 })
    await vi.advanceTimersByTimeAsync(180)
    await flushPromises()
    expect(mocks.members).toHaveBeenCalledTimes(2)
    expect(w.get('.preview-content').text()).toBe('1')
    resolve({ status: true, data: { dataList: [{ id: 'old' }, { id: 'old2' }] } })
    await flushPromises()
    expect(w.get('.preview-content').text()).toBe('1')
  } finally {
    w.unmount()
  }
})
it('不限量分类展开一次读取全部成员，不额外分页或截成 100 个', async () => {
  vi.useFakeTimers()
  mocks.members.mockResolvedValue({
    status: true,
    data: { dataList: Array.from({ length: 230 }, (_, i) => ({ id: `r${i}` })), total: 230 },
  })
  const w = create()
  try {
    await w.setProps({
      query: { ...query, groupLimit: 0 },
      group: { ...group, parts: 1, count: 230, total: 230 },
    })
    await w.trigger('pointerenter', { clientX: 10, clientY: 10 })
    await vi.advanceTimersByTimeAsync(180)
    await flushPromises()
    expect(mocks.members).toHaveBeenCalledOnce()
    expect(mocks.members.mock.lastCall?.[0]).toMatchObject({ groupLimit: 0, part: 1, page: 1 })
    expect(w.get('.preview-content').text()).toBe('230')
  } finally {
    w.unmount()
  }
})
