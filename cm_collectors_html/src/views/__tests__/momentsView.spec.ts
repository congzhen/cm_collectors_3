import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import { defineComponent, h, KeepAlive, ref } from 'vue'
import ElementPlus from 'element-plus'
import MomentsView from '../MomentsView.vue'
import { appStoreData } from '@/storeData/app.storeData'

const mocks = vi.hoisted(() => ({ request: vi.fn(), push: vi.fn(), mobile: true, getConfig: vi.fn(), saveConfig: vi.fn() }))
vi.mock('@/server/filesBases.server', () => ({ filesBasesServer: { getConfigById: mocks.getConfig, setFilesBasesConfigById: mocks.saveConfig } }))
vi.mock('@/components/play/MobileMomentTheater.vue', () => ({ default: { props: ['filesBasesId'], template: '<div class="mobile-theater-stub">{{ filesBasesId }}</div>' } }))
vi.mock('@/assets/request', () => ({ default: mocks.request }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: mocks.push }) }))
vi.mock('@/assets/mobile', () => ({ isMobile: () => mocks.mobile }))
vi.mock('@/storeData/app.storeData', async () => {
  const { reactive } = await import('vue')
  const state = reactive({ currentFilesBases: { id: 'library-a' }, currentConfigApp: { videoMomentLibraryMode: 'collection' }, get currentFilesBasesAppConfig() { return this.currentConfigApp } })
  return { appStoreData: () => state }
})
const global = { plugins: [ElementPlus], directives: { admin: () => {} }, stubs: { MomentCover: true } }
const result = (title: string) => ({ status: true, data: { total: 1, items: [{
  id: title, title, available: true, start: 1, end: 3, resourceId: 'r', dramaSeriesId: 'd',
}] } })
beforeEach(() => {
  vi.resetAllMocks()
  mocks.mobile = true
  appStoreData().currentFilesBases.id = 'library-a'
  appStoreData().currentConfigApp.videoMomentLibraryMode = 'collection'
  mocks.getConfig.mockResolvedValue({ status: true, data: '{"sampleFolder":"preserved"}' })
  mocks.saveConfig.mockResolvedValue({ status: true, data: true })
  mocks.request.mockResolvedValue(result('当前库片段'))
})
it('手机版隐藏缓存管理，仅查询当前文件库，并返回移动片库', async () => {
  const wrapper = mount(MomentsView, { global })
  await flushPromises()
  expect(wrapper.find('details').exists()).toBe(false)
  expect(wrapper.text()).not.toContain('预览任务与缓存管理')
  expect(mocks.request).toHaveBeenCalledWith({ url: '/videoMoments/library', params: { filesBasesId: 'library-a', page: 1, keyword: '' } })
  expect(mocks.request.mock.calls.every(([request]) => request.url === '/videoMoments/library')).toBe(true)
  await wrapper.get('header button').trigger('click')
  expect(mocks.push).toHaveBeenCalledWith('/mobile')
  wrapper.unmount()
})
it('切换库时丢弃旧库列表与迟到响应', async () => {
  let finish!: (value: unknown) => void
  mocks.request.mockImplementation(({ params }: { params: { filesBasesId: string } }) => params.filesBasesId === 'library-a'
    ? new Promise(resolve => { finish = resolve }) : Promise.resolve(result('新库片段')))
  const wrapper = mount(MomentsView, { global })
  appStoreData().currentFilesBases.id = 'library-b'
  await flushPromises()
  finish(result('旧库片段'))
  await flushPromises()
  expect(wrapper.findAll('.collection-card')).toHaveLength(1)
  expect(wrapper.get('.collection-card').text()).toContain('新库片段')
  expect(wrapper.text()).not.toContain('旧库片段')
  wrapper.unmount()
})
it('片段页缓存后，在片库换库再返回也展示新库', async () => {
  mocks.request.mockImplementation(({ params }: { params: { filesBasesId: string } }) => Promise.resolve(result(params.filesBasesId)))
  const active = ref(true)
  const wrapper = mount(defineComponent({ setup: () => () => h(KeepAlive, null, {
    default: () => active.value ? h(MomentsView) : h('div'),
  }) }), { global })
  await flushPromises()
  active.value = false
  await flushPromises()
  appStoreData().currentFilesBases.id = 'library-b'
  active.value = true
  await flushPromises()
  expect(wrapper.get('.collection-card').text()).toContain('library-b')
  expect(wrapper.get('.collection-card').text()).not.toContain('library-a')
  wrapper.unmount()
})
it('没有选中文件库时不回退到全库查询；桌面仍保留缓存管理', async () => {
  appStoreData().currentFilesBases.id = ''
  mocks.mobile = false
  const wrapper = mount(MomentsView, { global })
  await flushPromises()
  expect(wrapper.find('details').exists()).toBe(true)
  expect(wrapper.text()).toContain('请先返回片库选择文件库')
  expect(mocks.request).not.toHaveBeenCalled()
  wrapper.unmount()
})
it('手机可切换放映室并保存库偏好，失败时保留原模式', async () => {
  const wrapper = mount(MomentsView, { global })
  await flushPromises()
  await wrapper.get('[role="switch"]').trigger('click')
  await flushPromises()
  expect(wrapper.get('.mobile-theater-stub').text()).toBe('library-a')
  expect(mocks.saveConfig).toHaveBeenCalledWith('library-a', expect.objectContaining({ videoMomentLibraryMode: 'theater', sampleFolder: 'preserved' }))
  expect(wrapper.find('.collection-card').exists()).toBe(false)
  mocks.saveConfig.mockResolvedValue({ status: false, msg: '已取消修改' })
  await wrapper.get('[role="switch"]').trigger('click')
  await flushPromises()
  expect(wrapper.find('.mobile-theater-stub').exists()).toBe(true)
  wrapper.unmount()
})
it('读取偏好期间切库，不把旧库的切换保存到新库', async () => {
  let finish!: (value: unknown) => void
  mocks.getConfig.mockImplementationOnce(() => new Promise(resolve => { finish = resolve }))
  const wrapper = mount(MomentsView, { global })
  await flushPromises()
  await wrapper.get('[role="switch"]').trigger('click')
  appStoreData().currentFilesBases.id = 'library-b'
  finish({ status: true, data: '{}' })
  await flushPromises()
  expect(mocks.saveConfig).not.toHaveBeenCalled()
  expect(appStoreData().currentConfigApp.videoMomentLibraryMode).toBe('collection')
  expect(wrapper.get('[role="switch"]').attributes('aria-disabled')).not.toBe('true')
  wrapper.unmount()
})
