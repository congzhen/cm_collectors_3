import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import ElementPlus from 'element-plus'
import ResourceMomentsManager from '../ResourceMomentsManager.vue'
import { momentApi } from '@/server/videoMoments.server'

const mocks = vi.hoisted(() => ({ request: vi.fn(), confirm: vi.fn(), success: vi.fn(), error: vi.fn() }))
vi.mock('@/assets/request', () => ({ default: mocks.request }))
vi.mock('element-plus', async original => ({ ...await original<object>(),
  ElMessage: { success: mocks.success, error: mocks.error }, ElMessageBox: { confirm: mocks.confirm },
}))
const item = (id: string) => ({ id, resourceId: 'r', dramaSeriesId: 'd', title: `片段${id}`, start: 5, end: 10, available: true, note: '' })
let items = [item('1'), item('2'), item('3')]
const create = () => mount(ResourceMomentsManager, { props: { resourceId: 'r' }, global: {
  plugins: [ElementPlus], stubs: { MomentCover: true, ElDialog: {
    props: ['modelValue'], template: '<div v-if="modelValue"><slot /><slot name="footer" /></div>',
  } },
} })
beforeEach(() => {
  vi.resetAllMocks()
  items = [item('1'), item('2'), item('3')]
  mocks.confirm.mockResolvedValue('confirm')
  mocks.request.mockImplementation(async ({ url, method }: { url: string; method?: string }) => {
    if (method === 'delete') { items = items.filter(item => !url.endsWith('/' + item.id)); return { status: true, data: true } }
    return { status: true, data: [...items] }
  })
})
it('按钮显示当前资源片段数量，点击后弹窗展示完整列表与即时删除说明', async () => {
  const wrapper = create()
  await flushPromises()
  expect(mocks.request).toHaveBeenCalledWith({ url: '/videoMoments', params: { resourceId: 'r', page: 1, dramaSeriesId: '' } })
  expect(wrapper.find('.moments-preview').exists()).toBe(false)
  expect(wrapper.get('.moments-trigger').text()).toBe('片段收藏（3）')
  await wrapper.get('.moments-trigger').trigger('click')
  expect(wrapper.findAll('.manager-row')).toHaveLength(3)
  expect(wrapper.get('.manager-notice').text()).toContain('关闭资源编辑不会撤销')
  wrapper.unmount()
})
it('删除成功立即更新列表，取消和删除失败不丢失记录', async () => {
  const wrapper = create()
  await flushPromises()
  await wrapper.get('.moments-trigger').trigger('click')
  mocks.confirm.mockRejectedValueOnce('cancel')
  await wrapper.get('[aria-label="删除片段：片段1"]').trigger('click')
  await flushPromises()
  expect(items).toHaveLength(3)
  await wrapper.get('[aria-label="删除片段：片段1"]').trigger('click')
  await flushPromises()
  expect(items).toHaveLength(2)
  expect(wrapper.get('.moments-trigger').text()).toBe('片段收藏（2）')
  expect(wrapper.text()).not.toContain('片段1')
  mocks.request.mockResolvedValueOnce({ status: false, msg: '删除失败' })
  await wrapper.get('[aria-label="删除片段：片段2"]').trigger('click')
  await flushPromises()
  expect(mocks.error).toHaveBeenCalledWith('删除失败')
  expect(wrapper.get('.moments-trigger').text()).toBe('片段收藏（2）')
  wrapper.unmount()
})
it('确认期间切换资源，不再删除旧资源的片段', async () => {
  let confirm!: (value: string) => void
  mocks.confirm.mockImplementation(() => new Promise(resolve => { confirm = resolve }))
  const wrapper = create()
  await flushPromises()
  await wrapper.get('.moments-trigger').trigger('click')
  await wrapper.get('[aria-label="删除片段：片段1"]').trigger('click')
  await wrapper.setProps({ resourceId: 'other' })
  confirm('confirm')
  await flushPromises()
  expect(mocks.request.mock.calls.some(([arg]) => arg.method === 'delete')).toBe(false)
  expect(mocks.request).toHaveBeenLastCalledWith({ url: '/videoMoments', params: { resourceId: 'other', page: 1, dramaSeriesId: '' } })
  wrapper.unmount()
})
it('关闭后忽略未完成的列表响应', async () => {
  let finish!: (value: unknown) => void
  mocks.request.mockImplementationOnce(() => new Promise(resolve => { finish = resolve }))
  const wrapper = create()
  wrapper.unmount()
  finish({ status: true, data: items })
  await flushPromises()
  expect(mocks.error).not.toHaveBeenCalled()
})
it('超过50条可加载更多，删除后重新分页不遗漏片段', async () => {
  items = Array.from({ length: 65 }, (_, index) => item(String(index)))
  mocks.request.mockImplementation(async ({ url, method, params }: { url: string; method?: string; params?: { page: number } }) => {
    if (method === 'delete') { items = items.filter(item => !url.endsWith('/' + item.id)); return { status: true, data: true } }
    const page = params?.page || 1
    return { status: true, data: items.slice((page - 1) * 50, page * 50) }
  })
  const wrapper = create()
  await flushPromises()
  await wrapper.get('.moments-trigger').trigger('click')
  await wrapper.get('.load-more').trigger('click')
  await flushPromises()
  expect(wrapper.findAll('.manager-row')).toHaveLength(65)
  await wrapper.get('.manager-list [aria-label="删除片段：片段60"]').trigger('click')
  await flushPromises()
  expect(wrapper.findAll('.manager-row')).toHaveLength(50)
  await wrapper.get('.load-more').trigger('click')
  await flushPromises()
  expect(wrapper.findAll('.manager-row')).toHaveLength(64)
  expect(wrapper.findAll('.manager-row .moment-title').map(row => row.text())).toEqual(items.map(item => item.title))
  wrapper.unmount()
})
it('长备注默认折叠，可展开和收起，切换资源后恢复折叠', async () => {
  items[0].note = '用于验证长备注的折叠展示。'.repeat(10)
  const wrapper = create()
  await flushPromises()
  await wrapper.get('.moments-trigger').trigger('click')
  expect(wrapper.get('.note-toggle').attributes('aria-expanded')).toBe('false')
  await wrapper.get('.note-toggle').trigger('click')
  expect(wrapper.get('.moment-note').classes()).toContain('is-expanded')
  expect(wrapper.get('.note-toggle').text()).toBe('收起备注')
  await wrapper.get('.note-toggle').trigger('click')
  expect(wrapper.get('.moment-note').classes()).not.toContain('is-expanded')
  await wrapper.setProps({ resourceId: 'other' })
  await flushPromises()
  await wrapper.get('.moments-trigger').trigger('click')
  expect(wrapper.get('.note-toggle').attributes('aria-expanded')).toBe('false')
  wrapper.unmount()
})
it('短备注完整显示，不会被两行截断而无法展开', async () => {
  items[0].note = '短备注在窄屏换行时也应完整展示。'
  const wrapper = create()
  await flushPromises()
  await wrapper.get('.moments-trigger').trigger('click')
  expect(wrapper.get('.moment-note').classes()).toContain('is-expanded')
  expect(wrapper.find('.note-toggle').exists()).toBe(false)
  wrapper.unmount()
})
it('删除已成功但刷新失败时，不再保留已删除卡片', async () => {
  const wrapper = create()
  await flushPromises()
  await wrapper.get('.moments-trigger').trigger('click')
  mocks.request.mockImplementation(async ({ method }: { method?: string }) => method === 'delete'
    ? { status: true, data: true } : { status: false, msg: '刷新失败' })
  await wrapper.get('[aria-label="删除片段：片段1"]').trigger('click')
  await flushPromises()
  expect(wrapper.findAll('.manager-row')).toHaveLength(2)
  expect(wrapper.find('[aria-label="删除片段：片段1"]').exists()).toBe(false)
  expect(wrapper.get('[role="alert"]').text()).toContain('刷新失败')
  wrapper.unmount()
})
it('删除确认期间收到其他片段变化，取消删除后仍补刷列表', async () => {
  let cancel!: () => void
  mocks.confirm.mockImplementation(() => new Promise((_, reject) => { cancel = () => reject('cancel') }))
  const wrapper = create()
  await flushPromises()
  await wrapper.get('.moments-trigger').trigger('click')
  await wrapper.get('[aria-label="删除片段：片段1"]').trigger('click')
  items.push(item('4'))
  await momentApi.save({ resourceId: 'r', dramaSeriesId: 'd', title: '新增片段' })
  cancel()
  await flushPromises()
  expect(wrapper.findAll('.manager-row')).toHaveLength(4)
  expect(wrapper.get('.moments-trigger').text()).toBe('片段收藏（4）')
  wrapper.unmount()
})
