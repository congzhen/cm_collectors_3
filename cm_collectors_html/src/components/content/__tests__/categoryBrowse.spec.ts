import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import ElementPlus, { ElPagination, ElRadioGroup } from 'element-plus'
import CategoryBrowse from '../CategoryBrowse.vue'
import Card from '../CategoryBrowseCard.vue'
import ContentList from '../contentList.vue'
import { appStoreData } from '@/storeData/app.storeData'
import { searchStoreData } from '@/storeData/search.storeData'
const mocks = vi.hoisted(() => ({ summaries: vi.fn(), members: vi.fn() }))
vi.mock('@/server/categoryBrowse.server', async () => {
  const actual = await vi.importActual<typeof import('@/server/categoryBrowse.server')>(
    '@/server/categoryBrowse.server',
  )
  return { ...actual, categoryBrowseServer: mocks }
})
vi.mock('@/main', () => ({ eventBus: { emit: vi.fn() } }))
vi.mock('@/assets/request', () => ({ default: vi.fn() }))
vi.mock('@/storeData/app.storeData', async () => {
  const { reactive } = await import('vue')
  const store = reactive({
    currentFilesBases: { id: 'lib' },
    currentTagClass: [{ id: 'type', name: '类型' }],
    currentConfigApp: {
      pageLimit: 2,
      categoryBrowseEnabled: true,
      categoryBrowseSource: 'performer',
      categoryBrowseGroupLimit: 4,
      categoryBrowseCoverStyle: 'adaptive',
      categoryBrowseExpandStyle: 'fold',
    },
  })
  return { appStoreData: () => store }
})
vi.mock('@/storeData/search.storeData', async () => {
  const { reactive } = await import('vue')
  const store = reactive({
    searchData: { sort: 'starDesc', star: { logic: 'single', options: ['5'] } },
  })
  return { searchStoreData: () => store }
})
vi.mock('../CategoryBrowseCard.vue', () => ({
  default: {
    props: ['group', 'query', 'coverStyle', 'expandStyle'],
    emits: ['enter'],
    template:
      '<button class="group-card" @click="$emit(\'enter\',group,false)">{{group.name}} {{group.part}}</button>',
  },
}))
vi.mock('../contentList.vue', () => ({
  default: {
    props: { dataList: Array, plain: Boolean },
    emits: ['selectResources'],
    template:
      '<div class="member-list"><button v-for="item in dataList" :key="item.id" @click="$emit(\'selectResources\',item)">{{item.title}}</button></div>',
  },
}))
const group = {
  categoryId: 'a',
  name: '演员甲',
  photo: '',
  part: 2,
  parts: 2,
  count: 1,
  total: 5,
  preview: [],
}
const button = (w: ReturnType<typeof mount>, text: string) =>
  w.findAll('button').find((item) => item.text() === text)!
beforeEach(() => {
  mocks.summaries.mockReset().mockResolvedValue({
    status: true,
    data: { dataList: [group], total: 2, resourceTotal: 5, categoryTotal: 1 },
  })
  mocks.members.mockReset().mockResolvedValue({
    status: true,
    data: { dataList: [{ id: 'r', title: '筛选命中的资源' }], total: 5 },
  })
  appStoreData().currentFilesBases.id = 'lib'
  appStoreData().currentConfigApp.categoryBrowseSource = 'performer'
  appStoreData().currentConfigApp.categoryBrowseSort = 'count'
  appStoreData().currentConfigApp.categoryBrowseGroupLimit = 4
  searchStoreData().searchData.star.options = ['5']
  appStoreData().currentConfigApp.showVideoDuration = false
})
afterEach(() => vi.useRealTimers())
it.each(['performer', 'tag:type'])('分类排序变化重新从第一页请求，且不改变组内排序：%s', async (source) => {
  appStoreData().currentConfigApp.categoryBrowseSource = source
  mocks.summaries.mockResolvedValue({
    status: true, data: { dataList: [group], total: 8, resourceTotal: 20, categoryTotal: 4 },
  })
  const w = mount(CategoryBrowse, { global: { plugins: [ElementPlus] } })
  try {
    await flushPromises()
    expect(mocks.summaries.mock.lastCall?.[0]).toMatchObject({ categorySort: 'count', source })
    w.findComponent(ElPagination).vm.$emit('update:current-page', 2)
    await flushPromises()
    expect(mocks.summaries.mock.lastCall?.[0].page).toBe(2)
    for (const mode of ['name', 'default']) {
      const before = mocks.summaries.mock.calls.length
      appStoreData().currentConfigApp.categoryBrowseSort = mode as 'name' | 'default'
      await flushPromises()
      expect(mocks.summaries).toHaveBeenCalledTimes(before + 1)
      expect(mocks.summaries.mock.lastCall?.[0]).toMatchObject({
        categorySort: mode, source, page: 1, searchData: { sort: 'starDesc' },
      })
    }
  } finally {
    w.unmount()
  }
})
it('0 保持不限量语义，进入分类请求分页列表，不显示拆组和范围切换', async () => {
  appStoreData().currentConfigApp.categoryBrowseGroupLimit = 0
  mocks.summaries.mockResolvedValue({
    status: true,
    data: {
      dataList: [{ ...group, part: 1, parts: 1, count: 230, total: 230 }],
      total: 1,
      resourceTotal: 230,
      categoryTotal: 1,
    },
  })
  mocks.members.mockResolvedValue({
    status: true,
    data: { dataList: [{ id: 'r', title: '资源' }], total: 230 },
  })
  const w = mount(CategoryBrowse, { global: { plugins: [ElementPlus] } })
  try {
    await flushPromises()
    expect(mocks.summaries.mock.lastCall?.[0].groupLimit).toBe(0)
    await w.get('.group-card').trigger('click')
    await flushPromises()
    expect(mocks.members.mock.lastCall?.[0]).toMatchObject({
      groupLimit: 0,
      part: 0,
      page: 1,
      limit: 2,
    })
    expect(w.vm.pagination.total).toBe(230)
    expect(w.vm.pagination.visible).toBe(true)
    expect(w.findComponent(ElRadioGroup).exists()).toBe(false)
    expect(w.get('.category-breadcrumb').text()).not.toContain('第')
    w.vm.setPage(2)
    await flushPromises()
    expect(mocks.members.mock.lastCall?.[0]).toMatchObject({ groupLimit: 0, part: 0, page: 2 })
  } finally {
    w.unmount()
  }
})
it('末组被删空后刷新自动回到有效组，并同步分组数和范围切换入口', async () => {
  const w = mount(CategoryBrowse, { global: { plugins: [ElementPlus] } })
  try {
    await flushPromises()
    await w.get('.group-card').trigger('click')
    await flushPromises()
    expect(mocks.members.mock.lastCall?.[0].part).toBe(2)
    expect(w.findComponent(ElRadioGroup).exists()).toBe(true)
    const remaining = Array.from({ length: 4 }, (_, i) => ({
      id: `remaining-${i}`,
      title: `剩余${i}`,
    }))
    mocks.members
      .mockResolvedValueOnce({ status: true, data: { dataList: [], total: 4 } })
      .mockResolvedValueOnce({ status: true, data: { dataList: remaining, total: 4 } })
    await w.vm.refresh()
    await flushPromises()
    expect(mocks.members.mock.lastCall?.[0].part).toBe(1)
    expect(w.get('.category-breadcrumb').text()).toContain('第 1 组')
    expect(w.findComponent(ElRadioGroup).exists()).toBe(false)
    expect(w.vm.showDataList()).toHaveLength(4)
    expect(w.vm.pagination.summary).toContain('共 4 个资源')
  } finally {
    w.unmount()
  }
})
it('整类资源减少时保留整类范围，原分组更新后可以返回有效本组', async () => {
  const w = mount(CategoryBrowse, { global: { plugins: [ElementPlus] } })
  try {
    await flushPromises()
    await w.get('.group-card').trigger('click')
    await flushPromises()
    w.getComponent(ElRadioGroup).vm.$emit('update:modelValue', true)
    await flushPromises()
    mocks.members.mockResolvedValue({
      status: true,
      data: { dataList: [{ id: 'remaining' }], total: 4 },
    })
    await w.vm.refresh()
    await flushPromises()
    expect(mocks.members.mock.lastCall?.[0].part).toBe(0)
    expect(w.vm.pagination.visible).toBe(true)
    expect(w.find('.member-list').exists()).toBe(true)
    expect(w.findComponent(ElRadioGroup).exists()).toBe(true)
    w.getComponent(ElRadioGroup).vm.$emit('update:modelValue', false)
    await flushPromises()
    expect(mocks.members.mock.lastCall?.[0].part).toBe(1)
    expect(w.vm.pagination.visible).toBe(false)
    expect(w.findComponent(ElRadioGroup).exists()).toBe(false)
  } finally {
    w.unmount()
  }
})
it('按全部筛选查询分类，进入指定组或整类后继续使用普通资源布局，返回分类与退出路径明确', async () => {
  const w = mount(CategoryBrowse, { global: { plugins: [ElementPlus] } })
  try {
    await flushPromises()
    expect(mocks.summaries.mock.calls[0][0]).toMatchObject({
      filesBasesId: 'lib',
      groupLimit: 4,
      searchData: { star: { options: ['5'] } },
    })
    expect(w.findComponent(Card).props('coverStyle')).toBe('adaptive')
    expect(w.findComponent(Card).props('expandStyle')).toBe('fold')
    await w.get('.group-card').trigger('click')
    await flushPromises()
    expect(mocks.members.mock.lastCall?.[0]).toMatchObject({
      categoryId: 'a',
      part: 2,
      searchData: { star: { options: ['5'] } },
    })
    expect(w.findComponent(ContentList).props('plain')).toBe(true)
    await w.get('.member-list button').trigger('click')
    expect(w.emitted('select')?.[0][0]).toMatchObject({ id: 'r' })
    const cleared = w.emitted('clearSelection')!.length
    w.getComponent(ElRadioGroup).vm.$emit('update:modelValue', true)
    await flushPromises()
    expect(w.emitted('clearSelection')).toHaveLength(cleared + 1)
    expect(mocks.members.mock.lastCall?.[0].part).toBe(0)
    await button(w, '分类列表').trigger('click')
    await flushPromises()
    expect(w.emitted('clearSelection')).toHaveLength(cleared + 2)
    expect(w.find('.group-card').exists()).toBe(true)
    await button(w, '返回资源浏览').trigger('click')
    expect(w.emitted('exit')).toHaveLength(1)
  } finally {
    w.unmount()
  }
})
it('切库或改变筛选后清除组内状态，晚到旧请求不能覆盖新页面', async () => {
  const w = mount(CategoryBrowse, { global: { plugins: [ElementPlus] } })
  let resolve!: (result: unknown) => void
  try {
    await flushPromises()
    mocks.members.mockReturnValueOnce(
      new Promise((done) => {
        resolve = done
      }),
    )
    await w.get('.group-card').trigger('click')
    appStoreData().currentFilesBases.id = 'new-lib'
    searchStoreData().searchData.star.options = ['4']
    await flushPromises()
    expect(mocks.summaries.mock.lastCall?.[0]).toMatchObject({
      filesBasesId: 'new-lib',
      searchData: { star: { options: ['4'] } },
    })
    resolve({ status: true, data: { dataList: [{ id: 'old', title: '旧数据' }], total: 1 } })
    await flushPromises()
    expect(w.find('.member-list').exists()).toBe(false)
    expect(w.text()).not.toContain('旧数据')
  } finally {
    w.unmount()
  }
})
it('失败清空旧卡片并提供重试，不能继续操作旧分类', async () => {
  const w = mount(CategoryBrowse, { global: { plugins: [ElementPlus] } })
  try {
    await flushPromises()
    mocks.summaries.mockResolvedValueOnce({ status: false, msg: '查询失败' })
    searchStoreData().searchData.star.options = ['3']
    await flushPromises()
    expect(w.find('.group-card').exists()).toBe(false)
    expect(w.text()).toContain('查询失败')
    await button(w, '重试').trigger('click')
    await flushPromises()
    expect(w.find('.group-card').exists()).toBe(true)
  } finally {
    w.unmount()
  }
})

it('分类页进入后返回原页，重置页码时不重复请求', async () => {
  mocks.summaries.mockResolvedValue({
    status: true,
    data: { dataList: [group], total: 6, resourceTotal: 12, categoryTotal: 2 },
  })
  const w = mount(CategoryBrowse, { global: { plugins: [ElementPlus] } })
  try {
    await flushPromises()
    mocks.summaries.mockClear()
    w.findComponent(ElPagination).vm.$emit('update:current-page', 2)
    await flushPromises()
    expect(mocks.summaries).toHaveBeenCalledTimes(1)
    expect(mocks.summaries.mock.lastCall?.[0].page).toBe(2)
    await w.get('.group-card').trigger('click')
    await flushPromises()
    expect(mocks.members).toHaveBeenCalledTimes(1)
    await button(w, '分类列表').trigger('click')
    await flushPromises()
    expect(mocks.summaries.mock.lastCall?.[0].page).toBe(2)
  } finally {
    w.unmount()
  }
})
it('末页资源减少后恢复到有效分类页和整类页，不显示假空白', async () => {
  mocks.summaries
    .mockResolvedValueOnce({
      status: true,
      data: { dataList: [group], total: 6, resourceTotal: 12, categoryTotal: 2 },
    })
    .mockImplementation(async () => ({
      status: true,
      data: { dataList: [group], total: 2, resourceTotal: 5, categoryTotal: 1 },
    }))
  const w = mount(CategoryBrowse, { global: { plugins: [ElementPlus] } })
  try {
    await flushPromises()
    w.findComponent(ElPagination).vm.$emit('update:current-page', 3)
    await flushPromises()
    expect(mocks.summaries.mock.lastCall?.[0].page).toBe(1)
    expect(w.find('.group-card').exists()).toBe(true)
    await w.get('.group-card').trigger('click')
    await flushPromises()
    w.getComponent(ElRadioGroup).vm.$emit('update:modelValue', true)
    await flushPromises()
    mocks.members.mockResolvedValue({
      status: true,
      data: { dataList: [{ id: 'r', title: '剩余资源' }], total: 2 },
    })
    w.findComponent(ElPagination).vm.$emit('update:current-page', 3)
    await flushPromises()
    expect(mocks.members.mock.lastCall?.[0].page).toBe(1)
    expect(w.text()).toContain('剩余资源')
  } finally {
    w.unmount()
  }
})
it('组内时长补采只更新时长，保留资源对象和本地星级；离开后不再轮询', async () => {
  vi.useFakeTimers()
  appStoreData().currentConfigApp.showVideoDuration = true
  mocks.members
    .mockResolvedValueOnce({
      status: true,
      data: {
        dataList: [
          { id: 'r', stars: 5, title: '资源', dramaSeries: [{ id: 'ds', durationSeconds: 0 }] },
        ],
        total: 5,
      },
    })
    .mockResolvedValue({
      status: true,
      data: {
        dataList: [
          { id: 'r', stars: 1, title: '旧标题', dramaSeries: [{ id: 'ds', durationSeconds: 80 }] },
        ],
        total: 5,
      },
    })
  const w = mount(CategoryBrowse, { global: { plugins: [ElementPlus] } })
  try {
    await flushPromises()
    await w.get('.group-card').trigger('click')
    await flushPromises()
    const original = w.vm.showDataList()[0]
    await vi.advanceTimersByTimeAsync(2500)
    await flushPromises()
    expect(w.vm.showDataList()[0]).toBe(original)
    expect(original.dramaSeries[0].durationSeconds).toBe(80)
    expect(original.stars).toBe(5)
    expect(original.title).toBe('资源')
    await button(w, '分类列表').trigger('click')
    await flushPromises()
    const calls = mocks.members.mock.calls.length
    await vi.advanceTimersByTimeAsync(5000)
    expect(mocks.members).toHaveBeenCalledTimes(calls)
  } finally {
    w.unmount()
  }
})

it('尺寸间距和对齐实时更新分类网格，不重新查询分类数据', async () => {
  const w = mount(CategoryBrowse, { global: { plugins: [ElementPlus] } });
  try {
    await flushPromises();
    const calls = mocks.summaries.mock.calls.length;
    Object.assign(appStoreData().currentConfigApp, { coverPosterGap: 0, contentPadding: 12, resourceJustifyContent: 'center', coverPosterHeightBase: 300 });
    await flushPromises();
    expect(w.get('.category-grid').attributes('style')).toContain('gap: 0px');
    expect(w.get('.category-grid').attributes('style')).toContain('padding-inline: 12%');
    expect(w.get('.category-grid').attributes('style')).toContain('justify-content: center');
    expect(mocks.summaries).toHaveBeenCalledTimes(calls);
  } finally { w.unmount(); }
});
