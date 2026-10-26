import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import LibraryMomentsButton from '../LibraryMomentsButton.vue'
import MomentCollection from '../MomentCollection.vue'
import { momentApi } from '@/server/videoMoments.server'
const mocks = vi.hoisted(() => ({ request: vi.fn(), open: vi.fn(), close: vi.fn(), push: vi.fn() }))
vi.mock('@/assets/request', () => ({ default: (...args: unknown[]) => mocks.request(...args) }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: mocks.push }) }))
vi.mock('@/assets/mobile', () => ({ isMobile: () => false }))
vi.mock('@/common/play', () => ({ playUpdate: vi.fn() }))
vi.mock('../videoPlayDialog.vue', () => ({
  default: defineComponent({
    setup(_, { expose }) {
      expose({ open: mocks.open, close: mocks.close })
      return () => h('div')
    },
  }),
}))
const button = defineComponent({
  setup:
    (_, { slots }) =>
    () =>
      h('button', {}, slots.default?.()),
})
const drawer = defineComponent({
  props: ['modelValue', 'direction'],
  setup:
    (props, { slots }) =>
    () =>
      h('aside', { 'data-direction': props.direction }, props.modelValue ? slots.default?.() : []),
})
const input = defineComponent({
  props: ['modelValue'],
  emits: ['update:modelValue'],
  setup:
    (props, { emit }) =>
    () =>
      h('input', {
        value: props.modelValue,
        onInput: (event: Event) =>
          emit('update:modelValue', (event.target as HTMLInputElement).value),
      }),
})
const pagination = defineComponent({
  props: ['currentPage', 'total'],
  emits: ['current-change'],
  setup:
    (props, { emit }) =>
    () =>
      h(
        'button',
        { class: 'next-page', onClick: () => emit('current-change', 2) },
        `${props.currentPage}/${props.total}`,
      ),
})
const global = {
  stubs: {
    'el-button': button,
    'el-drawer': drawer,
    'el-input': input,
    'el-pagination': pagination,
    MomentCover: true,
  },
}
const item = {
  id: 'm',
  resourceId: 'r',
  dramaSeriesId: 'd',
  title: '收藏片段',
  start: 10,
  end: 20,
  available: true,
}
describe('本库片段收藏抽屉', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.request.mockResolvedValue({ status: true, data: { items: [item], total: 65 } })
  })
  it('右侧抽屉只查询当前库，播放复用现有弹窗并传入片段范围', async () => {
    const wrapper = mount(LibraryMomentsButton, { props: { filesBasesId: 'library-a' }, global })
    expect(mocks.request).not.toHaveBeenCalled()
    await wrapper.find('button').trigger('click')
    await flushPromises()
    expect(wrapper.get('aside').attributes('data-direction')).toBe('rtl')
    expect(mocks.request).toHaveBeenCalledWith({
      url: '/videoMoments/library',
      params: { filesBasesId: 'library-a', page: 1, keyword: '' },
    })
    await wrapper.get('.collection-play').trigger('click')
    expect(mocks.open).toHaveBeenCalledWith('r', 'd', { start: 10, end: 20, title: '收藏片段' })
    expect(mocks.push).not.toHaveBeenCalled()
    await wrapper.setProps({ filesBasesId: 'library-b' })
    expect(mocks.close).toHaveBeenCalled()
    wrapper.unmount()
  })
  it('收藏变更后保持当前页，最后一页删空时退回有效页', async () => {
    let total = 65
    mocks.request.mockImplementation(async ({ method }: { method?: string }) => ({
      status: true, data: method ? item : { items: [item], total },
    }))
    const wrapper = mount(MomentCollection, {
      props: { filesBasesId: 'library-a', playInDialog: true }, global,
    })
    await flushPromises()
    await wrapper.get('.next-page').trigger('click')
    await flushPromises()
    await momentApi.save(item)
    await flushPromises()
    expect(mocks.request).toHaveBeenLastCalledWith({
      url: '/videoMoments/library', params: { filesBasesId: 'library-a', page: 2, keyword: '' },
    })
    expect(wrapper.get('.next-page').text()).toBe('2/65')
    total = 50
    await momentApi.remove(item.id)
    await flushPromises()
    expect(mocks.request).toHaveBeenLastCalledWith({
      url: '/videoMoments/library', params: { filesBasesId: 'library-a', page: 1, keyword: '' },
    })
    expect(wrapper.get('.next-page').text()).toBe('1/50')
    wrapper.unmount()
  })
  it('每页50条，翻页替换卡片；搜索重置到第一页', async () => {
    const wrapper = mount(MomentCollection, {
      props: { filesBasesId: 'library-a', playInDialog: true },
      global,
    })
    await flushPromises()
    await wrapper.get('.next-page').trigger('click')
    await flushPromises()
    expect(mocks.request).toHaveBeenLastCalledWith({
      url: '/videoMoments/library',
      params: { filesBasesId: 'library-a', page: 2, keyword: '' },
    })
    expect(wrapper.findAll('.collection-card')).toHaveLength(1)
    await wrapper.get('input').setValue('备注')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(mocks.request).toHaveBeenLastCalledWith({
      url: '/videoMoments/library',
      params: { filesBasesId: 'library-a', page: 1, keyword: '备注' },
    })
    wrapper.unmount()
  })
})
