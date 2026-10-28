import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, KeepAlive, ref } from 'vue'
import LibraryMomentsButton from '../LibraryMomentsButton.vue'
import MomentCollection from '../MomentCollection.vue'
import { momentApi } from '@/server/videoMoments.server'
const mocks = vi.hoisted(() => ({ request: vi.fn(), open: vi.fn(), close: vi.fn(), push: vi.fn(), getConfig: vi.fn(), saveConfig: vi.fn(), playUpdate: vi.fn() }))
const store = vi.hoisted(() => ({
  currentFilesBases: { id: 'library-a' },
  currentConfigApp: { videoMomentLibraryMode: 'theater' },
  get currentFilesBasesAppConfig() { return this.currentConfigApp },
}))
vi.mock('@/storeData/app.storeData', async () => {
  const { reactive } = await import('vue')
  const state = reactive(store)
  return { appStoreData: () => state }
})
vi.mock('@/server/filesBases.server', () => ({ filesBasesServer: {
  getConfigById: mocks.getConfig, setFilesBasesConfigById: mocks.saveConfig,
} }))
vi.mock('../resourceVideoPlayer.vue', () => ({ default: defineComponent({
  name: 'ResourceVideoPlayer',
  props: ['resource', 'dramaSeriesId', 'initialMoment'],
  setup: () => () => h('div', { class: 'inline-player' }),
}) }))
vi.mock('@/assets/request', () => ({ default: (...args: unknown[]) => mocks.request(...args) }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: mocks.push }) }))
vi.mock('@/assets/mobile', () => ({ isMobile: () => false }))
vi.mock('@/common/play', () => ({ playUpdate: mocks.playUpdate }))
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
      h('aside', { 'data-direction': props.direction }, props.modelValue ? [slots.header?.(), slots.default?.()] : []),
})
const dialog = defineComponent({
  props: ['modelValue'],
  setup: (props, { slots }) => () => h('section', { class: 'dialog-surface' }, props.modelValue ? [slots.header?.(), slots.default?.()] : []),
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
    'el-dialog': dialog,
    'el-drawer': drawer,
    teleport: false,
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
describe('本库片段放映室与收藏', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    store.currentFilesBasesAppConfig.videoMomentLibraryMode = 'theater'
    store.currentFilesBases.id = 'library-a'
    mocks.getConfig.mockResolvedValue({ status: true, data: '{"sampleFolder":"keep-folder","pageLimit":72}' })
    mocks.saveConfig.mockResolvedValue({ status: true, data: true })
    mocks.request.mockImplementation(async ({ url }: { url: string }) => ({ status: true,
      data: url.startsWith('/resource/info/') ? { id: 'r', dramaSeries: [{ id: 'd' }] } : { items: [item], total: 65 },
    }))
  })
  it('收藏默认模式只查询当前库，播放复用现有弹窗并传入片段范围', async () => {
    store.currentFilesBasesAppConfig.videoMomentLibraryMode = 'collection'
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
  it('切换视图保存到文件库并更新默认入口，保留搜索和分页，重新打开仍使用保存值', async () => {
    mocks.request.mockImplementation(async ({ url }: { url: string }) => ({ status: true,
      data: url.startsWith('/resource/info/') ? { id: 'r', dramaSeries: [{ id: 'd' }] } : { items: [item], total: 65 },
    }))
    const wrapper = mount(LibraryMomentsButton, { props: { filesBasesId: 'library-a' }, global })
    expect(wrapper.find('button').text()).toBe('片段放映室')
    await wrapper.find('button').trigger('click')
    await flushPromises()
    expect(wrapper.find('.is-theater').exists()).toBe(true)
    expect(wrapper.get('[role="switch"]').attributes('aria-checked')).toBe('true')
    expect(wrapper.find('.inline-player').exists()).toBe(true)
    expect(wrapper.findComponent({ name: 'ResourceVideoPlayer' }).props('initialMoment')).toEqual({ start: 10, end: 20, title: '收藏片段', autoplay: false })
    await wrapper.get('.collection-play').trigger('click')
    await flushPromises()
    expect(wrapper.find('.inline-player').exists()).toBe(true)
    expect(wrapper.findComponent({ name: 'ResourceVideoPlayer' }).props('initialMoment').autoplay).toBe(true)
    expect(wrapper.get('.collection-card').attributes('aria-current')).toBe('true')
    expect(mocks.open).not.toHaveBeenCalled()
    await wrapper.get('input').setValue('镜头')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    await wrapper.get('.next-page').trigger('click')
    await flushPromises()
    await wrapper.get('.library-moments-header button').trigger('click')
    await flushPromises()
    expect(mocks.saveConfig).toHaveBeenCalledWith('library-a', expect.objectContaining({
      videoMomentLibraryMode: 'collection', sampleFolder: 'keep-folder', pageLimit: 72,
    }))
    expect(wrapper.find('.inline-player').exists()).toBe(false)
    expect(wrapper.find('.is-theater').exists()).toBe(false)
    expect(wrapper.get('[role="switch"]').attributes('aria-checked')).toBe('false')
    expect(wrapper.get('aside').attributes('data-direction')).toBe('rtl')
    expect(wrapper.get('input').element.value).toBe('镜头')
    expect(wrapper.get('.next-page').text()).toBe('2/65')
    await wrapper.get('.collection-play').trigger('click')
    expect(mocks.open).toHaveBeenCalledWith('r', 'd', { start: 10, end: 20, title: '收藏片段' })
    await wrapper.findComponent(drawer).vm.$emit('update:modelValue', false)
    await wrapper.find('button').trigger('click')
    expect(wrapper.find('.is-theater').exists()).toBe(false)
    expect(wrapper.get('aside').attributes('data-direction')).toBe('rtl')
    wrapper.unmount()
  })
  it('保存失败或取消时保留原视图与文件库默认值', async () => {
    mocks.saveConfig.mockResolvedValue({ status: false, msg: '已取消修改' })
    const wrapper = mount(LibraryMomentsButton, { props: { filesBasesId: 'library-a' }, global })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    await wrapper.get('[role="switch"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('.is-theater').exists()).toBe(true)
    expect(store.currentConfigApp.videoMomentLibraryMode).toBe('theater')
    expect(wrapper.get('[role="switch"]').element.hasAttribute('disabled')).toBe(false)
    wrapper.unmount()
  })
  it('首片段待播不记历史，开始播放后只记录一次，循环与恢复播放不重复记录', async () => {
    const wrapper = mount(LibraryMomentsButton, { props: { filesBasesId: 'library-a' }, global })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    expect(mocks.playUpdate).not.toHaveBeenCalled()
    const player = wrapper.findComponent({ name: 'ResourceVideoPlayer' })
    player.vm.$emit('playing')
    player.vm.$emit('playing')
    expect(mocks.playUpdate).toHaveBeenCalledExactlyOnceWith('r', 'd')
    wrapper.unmount()
  })
  it('同一视频切片段或重复点击保留播放器实例和缓冲，不再请求资源', async () => {
    mocks.request.mockImplementation(async ({ url }: { url: string }) => ({ status: true,
      data: url.startsWith('/resource/info/') ? { id: 'r', dramaSeries: [{ id: 'd' }] }
        : { items: [item, { ...item, id: 'm2', start: 30, end: 40 }], total: 2 },
    }))
    const wrapper = mount(LibraryMomentsButton, { props: { filesBasesId: 'library-a' }, global })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    const player = wrapper.getComponent({ name: 'ResourceVideoPlayer' })
    const instance = player.vm
    const initial = player.props('initialMoment')
    await wrapper.findAll('.collection-play')[1].trigger('click')
    await flushPromises()
    expect(wrapper.getComponent({ name: 'ResourceVideoPlayer' }).vm).toBe(instance)
    expect(player.props('initialMoment')).toEqual({ start: 30, end: 40, title: item.title, autoplay: true })
    const repeated = player.props('initialMoment')
    expect(repeated).not.toBe(initial)
    await wrapper.findAll('.collection-play')[1].trigger('click')
    expect(player.props('initialMoment')).not.toBe(repeated)
    expect(mocks.request.mock.calls.filter(([request]) => request.url.startsWith('/resource/info/'))).toHaveLength(1)
    wrapper.unmount()
  })
  it('保存期间不重设片段位置，旧保存结果不切换重新打开的窗口', async () => {
    let finish!: (value: unknown) => void
    mocks.saveConfig.mockImplementation(() => new Promise(resolve => { finish = resolve }))
    const wrapper = mount(LibraryMomentsButton, { props: { filesBasesId: 'library-a' }, global })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    const initial = wrapper.findComponent({ name: 'ResourceVideoPlayer' }).props('initialMoment')
    await wrapper.get('[role="switch"]').trigger('click')
    await flushPromises()
    expect(wrapper.findComponent({ name: 'ResourceVideoPlayer' }).props('initialMoment')).toBe(initial)
    await wrapper.findComponent(dialog).vm.$emit('update:modelValue', false)
    await wrapper.findComponent(dialog).vm.$emit('close')
    await wrapper.find('button').trigger('click')
    await flushPromises()
    finish({ status: true, data: true })
    await flushPromises()
    expect(wrapper.find('.is-theater').exists()).toBe(true)
    expect(store.currentConfigApp.videoMomentLibraryMode).toBe('collection')
    wrapper.unmount()
  })
  it('离开被缓存的页面时关闭放映室并销毁播放器', async () => {
    const active = ref(true)
    const wrapper = mount(defineComponent({ setup: () => () => h(KeepAlive, null, {
      default: () => active.value ? h(LibraryMomentsButton, { filesBasesId: 'library-a' }) : h('div'),
    }) }), { global })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    expect(wrapper.find('.inline-player').exists()).toBe(true)
    mocks.close.mockClear()
    active.value = false
    await flushPromises()
    expect(mocks.close).toHaveBeenCalled()
    active.value = true
    await flushPromises()
    expect(wrapper.find('.inline-player').exists()).toBe(false)
    expect(wrapper.find('.is-theater').exists()).toBe(false)
    wrapper.unmount()
  })
  it('切库或关闭后忽略尚未完成的播放请求', async () => {
    let finish!: (value: unknown) => void
    mocks.request.mockImplementation(({ url }: { url: string }) => url.startsWith('/resource/info/')
      ? new Promise(resolve => { finish = resolve })
      : Promise.resolve({ status: true, data: { items: [item], total: 1 } }))
    const wrapper = mount(LibraryMomentsButton, { props: { filesBasesId: 'library-a' }, global })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    await wrapper.get('.collection-play').trigger('click')
    await wrapper.setProps({ filesBasesId: 'library-b' })
    finish({ status: true, data: { id: 'r', dramaSeries: [{ id: 'd' }] } })
    await flushPromises()
    expect(wrapper.find('.inline-player').exists()).toBe(false)
    expect(mocks.open).not.toHaveBeenCalled()
    wrapper.unmount()
  })
  it('快速切换片段时，晚到的旧视频响应不能覆盖当前选择', async () => {
    let finishOld!: (value: unknown) => void
    mocks.request.mockImplementation(({ url }: { url: string }) => {
      if (url === '/resource/info/r') return new Promise(resolve => { finishOld = resolve })
      if (url === '/resource/info/r2') return Promise.resolve({ status: true, data: { id: 'r2', dramaSeries: [{ id: 'd2' }] } })
      return Promise.resolve({ status: true, data: { items: [item, { ...item, id: 'm2', resourceId: 'r2', dramaSeriesId: 'd2' }], total: 2 } })
    })
    const wrapper = mount(LibraryMomentsButton, { props: { filesBasesId: 'library-a' }, global })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    await wrapper.findAll('.collection-play')[1].trigger('click')
    await flushPromises()
    finishOld({ status: true, data: { id: 'r', dramaSeries: [{ id: 'd' }] } })
    await flushPromises()
    expect(wrapper.findComponent({ name: 'ResourceVideoPlayer' }).props('dramaSeriesId')).toBe('d2')
    expect(wrapper.findAll('.collection-card')[1].attributes('aria-current')).toBe('true')
    wrapper.unmount()
  })
  it.each([{ items: [] }, { items: [{ ...item, available: false }] }])('空列表或全部失效时不请求视频，也不记录播放历史', async ({ items }) => {
    mocks.request.mockResolvedValue({ status: true, data: { items, total: items.length } })
    const wrapper = mount(LibraryMomentsButton, { props: { filesBasesId: 'library-a' }, global })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    expect(wrapper.find('.inline-player').exists()).toBe(false)
    expect(mocks.request.mock.calls.some(([request]) => request.url.startsWith('/resource/info/'))).toBe(false)
    expect(mocks.playUpdate).not.toHaveBeenCalled()
    wrapper.unmount()
  })
  it('读取默认设置期间关闭窗口，不继续提交保存', async () => {
    let finish!: (value: unknown) => void
    mocks.getConfig.mockImplementationOnce(() => new Promise(resolve => { finish = resolve }))
    const wrapper = mount(LibraryMomentsButton, { props: { filesBasesId: 'library-a' }, global })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    await wrapper.get('[role="switch"]').trigger('click')
    await wrapper.findComponent(dialog).vm.$emit('update:modelValue', false)
    await wrapper.findComponent(dialog).vm.$emit('close')
    finish({ status: true, data: '{}' })
    await flushPromises()
    expect(mocks.saveConfig).not.toHaveBeenCalled()
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
