import { flushPromises, mount, shallowMount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, KeepAlive, reactive, ref } from 'vue'
import VideoDetailsExtras from '@/components/details/VideoDetailsExtras.vue'
import MomentCollection from '../MomentCollection.vue'
import type { I_resource } from '@/dataType/resource.dataType'
import { momentApi } from '@/server/videoMoments.server'

const mocks = vi.hoisted(() => ({ open: vi.fn(), close: vi.fn(), info: vi.fn(), history: vi.fn(), request: vi.fn() }))
const store = reactive({
  currentFilesBases: { id: 'library-a' },
  currentFilesBasesAppConfig: {} as { videoMomentOpenMode?: string },
})
vi.mock('@/storeData/app.storeData', () => ({ appStoreData: () => store }))
vi.mock('@/server/filesBases.server', () => ({ filesBasesServer: { infoById: mocks.info } }))
vi.mock('@/assets/request', () => ({ default: mocks.request }))
vi.mock('@/common/play', () => ({ playUpdate: mocks.history }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))
vi.mock('../videoPlayDialog.vue', () => ({ default: defineComponent({
  setup(_, { expose }) {
    expose({ open: mocks.open, close: mocks.close })
    return () => h('div')
  },
}) }))
const resource = (id = 'r', library = 'library-a') => ({ id, filesBases_id: library, mode: 'movies' }) as I_resource
const item = { id: 'm', resourceId: 'r', dramaSeriesId: 'd', title: '收藏', start: 10, end: 20, available: true }
const render = (extra = {}) => shallowMount(VideoDetailsExtras, {
  props: { resource: resource(), ...extra }, global: { stubs: { VideoPlayDialog: false } },
})
describe('资源片段收藏打开方式', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    store.currentFilesBases.id = 'library-a'
    store.currentFilesBasesAppConfig = {}
    mocks.info.mockResolvedValue({ status: true, data: { filesBasesSetting: { config_json_data: '{}' } } })
    mocks.request.mockResolvedValue({ status: true, data: [] })
  })
  it('详情无收藏时隐藏整个区块，新增后显示，删除最后一条后再次隐藏', async () => {
    let items: typeof item[] = []
    mocks.request.mockImplementation(async ({ method }: { method?: string }) => ({
      status: true, data: method ? item : [...items],
    }))
    const wrapper = mount(VideoDetailsExtras, {
      props: { resource: resource() },
      global: { stubs: { VideoPlayDialog: false, MomentCover: true, 'el-button': true } },
    })
    expect(wrapper.get('.video-extras').attributes('style')).toContain('display: none')
    await flushPromises()
    expect(wrapper.get('.video-extras').attributes('style')).toContain('display: none')
    expect(wrapper.findComponent(MomentCollection).exists()).toBe(true)
    items = [item]
    await momentApi.save(item)
    await flushPromises()
    expect(wrapper.get('.video-extras').attributes('style')).not.toContain('display: none')
    expect(wrapper.text()).toContain('片段收藏')
    items = []
    await momentApi.remove(item.id)
    await flushPromises()
    expect(wrapper.get('.video-extras').attributes('style')).toContain('display: none')
    wrapper.unmount()
  })
  it('首次加载失败保留刷新入口，成功确认空列表后隐藏', async () => {
    mocks.request.mockResolvedValueOnce({ status: false, msg: '暂时无法连接' })
    const wrapper = mount(VideoDetailsExtras, {
      props: { resource: resource() },
      global: { stubs: { VideoPlayDialog: false, MomentCover: true, 'el-button': true } },
    })
    await flushPromises()
    expect(wrapper.get('.video-extras').attributes('style')).not.toContain('display: none')
    expect(wrapper.get('[role="alert"]').text()).toBe('暂时无法连接')
    await wrapper.get('.collection-refresh').trigger('click')
    await flushPromises()
    expect(wrapper.get('.video-extras').attributes('style')).toContain('display: none')
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    wrapper.unmount()
  })
  it('切换资源清除旧错误，旧列表请求不能误显示新资源区块', async () => {
    let finish!: (value: unknown) => void
    mocks.request.mockImplementationOnce(() => new Promise(resolve => { finish = resolve }))
    const wrapper = mount(VideoDetailsExtras, {
      props: { resource: resource() },
      global: { stubs: { VideoPlayDialog: false, MomentCover: true, 'el-button': true } },
    })
    await wrapper.setProps({ resource: resource('other') })
    await flushPromises()
    finish({ status: true, data: [item] })
    await flushPromises()
    expect(wrapper.get('.video-extras').attributes('style')).toContain('display: none')
    expect(wrapper.findAll('.collection-card')).toHaveLength(0)
    mocks.request.mockResolvedValueOnce({ status: false, msg: '旧资源错误' })
    await wrapper.get('.collection-refresh').trigger('click')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('旧资源错误')
    await wrapper.setProps({ resource: resource('next') })
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    await flushPromises()
    expect(wrapper.get('.video-extras').attributes('style')).toContain('display: none')
    wrapper.unmount()
  })
  it('旧库默认使用弹窗播放器并传入片段范围、更新播放历史', async () => {
    const wrapper = render()
    expect(wrapper.getComponent(MomentCollection).props('playMode')).toBe('emit')
    wrapper.getComponent(MomentCollection).vm.$emit('play', item)
    await flushPromises()
    expect(mocks.open).toHaveBeenCalledWith('r', 'd', { start: 10, end: 20, title: '收藏' })
    expect(mocks.history).toHaveBeenCalledWith('r', 'd')
    expect(mocks.info).not.toHaveBeenCalled()
    await wrapper.setProps({ resource: resource('other') })
    expect(mocks.close).toHaveBeenCalled()
    wrapper.unmount()
  })
  it('本库选择进入播放页后使用页面跳转，组件显式参数可覆盖库设置', async () => {
    store.currentFilesBasesAppConfig.videoMomentOpenMode = 'page'
    const wrapper = render()
    expect(wrapper.getComponent(MomentCollection).props('playMode')).toBe('page')
    expect(wrapper.findComponent({ name: 'VideoPlayDialog' }).exists()).toBe(false)
    wrapper.getComponent(MomentCollection).vm.$emit('play', item)
    expect(mocks.history).toHaveBeenCalledWith('r', 'd')
    expect(mocks.open).not.toHaveBeenCalled()
    await wrapper.setProps({ playMode: 'dialog', layout: 'list', showHeading: false, showResourceTitle: true })
    const list = wrapper.getComponent(MomentCollection)
    expect(list.props()).toMatchObject({ playMode: 'emit', layout: 'list', showHeading: false, showResourceTitle: true })
    list.vm.$emit('play', item)
    await flushPromises()
    expect(mocks.open).toHaveBeenCalledOnce()
    wrapper.unmount()
  })
  it('独立播放页显式指定 page，不受库的弹窗偏好影响', () => {
    const wrapper = render({ playMode: 'page' })
    expect(wrapper.getComponent(MomentCollection).props('playMode')).toBe('page')
    expect(mocks.info).not.toHaveBeenCalled()
    wrapper.unmount()
  })
  it('跨库资源读取自身配置，切换资源时忽略旧配置请求', async () => {
    let finish!: (value: unknown) => void
    mocks.info.mockImplementationOnce(() => new Promise(resolve => { finish = resolve }))
    const wrapper = render({ resource: resource('foreign', 'library-b') })
    expect(wrapper.getComponent(MomentCollection).props('disabled')).toBe(true)
    expect(mocks.info).toHaveBeenCalledWith('library-b')
    await wrapper.setProps({ resource: resource() })
    finish({ status: true, data: { filesBasesSetting: { config_json_data: '{"videoMomentOpenMode":"page"}' } } })
    await flushPromises()
    expect(wrapper.getComponent(MomentCollection).props()).toMatchObject({ playMode: 'emit', disabled: false })
    mocks.info.mockResolvedValueOnce({ status: true, data: { filesBasesSetting: { config_json_data: '{"videoMomentOpenMode":"page"}' } } })
    await wrapper.setProps({ resource: resource('third', 'library-c') })
    await flushPromises()
    expect(wrapper.getComponent(MomentCollection).props('playMode')).toBe('page')
    wrapper.unmount()
  })
  it('跨库读取失败时恢复默认弹窗，解除加载期间的禁用', async () => {
    mocks.info.mockResolvedValueOnce({ status: false, msg: '离线' })
    const wrapper = render({ resource: resource('foreign', 'library-b') })
    await flushPromises()
    expect(wrapper.getComponent(MomentCollection).props()).toMatchObject({ playMode: 'emit', disabled: false })
    wrapper.getComponent(MomentCollection).vm.$emit('play', { ...item, available: false })
    expect(mocks.open).not.toHaveBeenCalled()
    wrapper.unmount()
  })
  it('缓存详情离开时关闭播放器，返回时重新读取跨库配置', async () => {
    const visible = ref(true)
    const wrapper = mount(defineComponent({
      setup: () => () => h(KeepAlive, {}, () => visible.value
        ? h(VideoDetailsExtras, { resource: resource('foreign', 'library-b') }) : null),
    }), { global: { stubs: { MomentCollection: true, VideoPlayDialog: false } } })
    await flushPromises()
    wrapper.getComponent(MomentCollection).vm.$emit('play', item)
    await flushPromises()
    expect(mocks.open).toHaveBeenCalledOnce()
    visible.value = false
    await flushPromises()
    expect(mocks.close).toHaveBeenCalledOnce()
    mocks.info.mockResolvedValueOnce({ status: true, data: { filesBasesSetting: { config_json_data: '{"videoMomentOpenMode":"page"}' } } })
    visible.value = true
    await flushPromises()
    expect(mocks.info).toHaveBeenCalledTimes(2)
    expect(wrapper.getComponent(MomentCollection).props('playMode')).toBe('page')
    wrapper.unmount()
  })
  it('父组件接管播放时仅发事件，不打开播放器和写历史', async () => {
    const wrapper = render({ playMode: 'emit' })
    wrapper.getComponent(MomentCollection).vm.$emit('play', item)
    await flushPromises()
    expect(wrapper.emitted('play')).toEqual([[item]])
    expect(mocks.open).not.toHaveBeenCalled()
    expect(mocks.history).not.toHaveBeenCalled()
    wrapper.unmount()
  })
})
