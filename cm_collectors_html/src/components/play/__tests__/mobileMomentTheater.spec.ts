import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import { defineComponent, h, KeepAlive, ref } from 'vue'
import ElementPlus from 'element-plus'
import MobileMomentTheater from '../MobileMomentTheater.vue'
import { momentApi } from '@/server/videoMoments.server'
const mocks = vi.hoisted(() => ({ request: vi.fn(), playUpdate: vi.fn() }))
vi.mock('@/assets/request', () => ({ default: mocks.request }))
vi.mock('@/common/play', () => ({ playUpdate: mocks.playUpdate }))
vi.mock('../mobileVideoPlayer.vue', () => ({ default: defineComponent({
  name: 'MobileVideoPlayer', props: ['resourceId', 'dramaSeriesId', 'initialMoment'], emits: ['playing'], setup: () => () => h('div', { class: 'phone-player-stub' }),
}) }))
const global = { plugins: [ElementPlus], stubs: { MomentCover: true } }
const item = (id: string, available = true) => ({ id, title: id, available, resourceId: 'r', dramaSeriesId: id, start: 2, end: 6 })
beforeEach(() => {
  vi.resetAllMocks()
  mocks.request.mockResolvedValue({ status: true, data: { items: [item('missing', false), item('first'), item('next')], total: 3 } })
})
it('按库加载，首个可用片段待播，点击卡片播放，实际播放才记一次历史', async () => {
  const wrapper = mount(MobileMomentTheater, { props: { filesBasesId: 'library-a' }, global })
  await flushPromises()
  expect(mocks.request).toHaveBeenCalledWith({ url: '/videoMoments/library', params: { filesBasesId: 'library-a', page: 1, keyword: '' } })
  let player = wrapper.findComponent({ name: 'MobileVideoPlayer' })
  expect(player.props('dramaSeriesId')).toBe('first')
  expect(player.props('initialMoment').autoplay).toBe(false)
  expect(mocks.playUpdate).not.toHaveBeenCalled()
  expect(wrapper.findAll('.filmstrip-card')[0].attributes('disabled')).toBeDefined()
  await wrapper.findAll('.filmstrip-card')[2].trigger('click')
  player = wrapper.findComponent({ name: 'MobileVideoPlayer' })
  expect(player.props('dramaSeriesId')).toBe('next')
  expect(player.props('initialMoment').autoplay).toBe(true)
  player.vm.$emit('playing'); player.vm.$emit('playing')
  expect(mocks.playUpdate).toHaveBeenCalledExactlyOnceWith('r', 'next')
  wrapper.unmount()
})
it('换库后清空旧播放器，迟到的旧库响应不能覆盖新库', async () => {
  let finish!: (value: unknown) => void
  mocks.request.mockImplementation(({ params }: { params: { filesBasesId: string } }) => params.filesBasesId === 'library-a'
    ? new Promise(resolve => { finish = resolve }) : Promise.resolve({ status: true, data: { items: [item('new-library')], total: 1 } }))
  const wrapper = mount(MobileMomentTheater, { props: { filesBasesId: 'library-a' }, global })
  await wrapper.setProps({ filesBasesId: 'library-b' })
  await flushPromises()
  finish({ status: true, data: { items: [item('old-library')], total: 1 } })
  await flushPromises()
  expect(wrapper.findComponent({ name: 'MobileVideoPlayer' }).props('dramaSeriesId')).toBe('new-library')
  expect(wrapper.text()).not.toContain('old-library')
  wrapper.unmount()
})
it('加载失败可重试，成功后首个片段保持待播', async () => {
  mocks.request.mockRejectedValueOnce(new Error('网络暂不可用'))
  const wrapper = mount(MobileMomentTheater, { props: { filesBasesId: 'library-a' }, global })
  await flushPromises()
  expect(wrapper.get('[role="alert"]').text()).toContain('网络暂不可用')
  expect(wrapper.findComponent({ name: 'MobileVideoPlayer' }).exists()).toBe(false)
  await wrapper.get('[role="alert"] button').trigger('click')
  await flushPromises()
  expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  expect(wrapper.findComponent({ name: 'MobileVideoPlayer' }).props('initialMoment').autoplay).toBe(false)
  wrapper.unmount()
})
it('追加页的迟到响应不会覆盖更新后的列表与片段范围', async () => {
  let finish!: (value: unknown) => void
  let changed = false
  mocks.request.mockImplementation(({ params, method }: { params?: { page: number }; method?: string }) => {
    if (method) return Promise.resolve({ status: true, data: {} })
    if (params?.page === 2) return new Promise(resolve => { finish = resolve })
    return Promise.resolve({ status: true, data: { total: 65, items: [{ ...item('first'), end: changed ? 9 : 6 }] } })
  })
  const wrapper = mount(MobileMomentTheater, { props: { filesBasesId: 'library-a' }, global })
  await flushPromises()
  await wrapper.get('.filmstrip-more').trigger('click')
  changed = true
  await momentApi.save({ dramaSeriesId: 'first', title: '修改范围', end: 9 })
  await flushPromises()
  finish({ status: true, data: { total: 65, items: [item('stale')] } })
  await flushPromises()
  expect(wrapper.findAll('.filmstrip-card')).toHaveLength(1)
  expect(wrapper.text()).not.toContain('stale')
  expect(wrapper.findComponent({ name: 'MobileVideoPlayer' }).props('initialMoment')).toMatchObject({ end: 9, autoplay: false })
  expect(wrapper.get('.filmstrip-more').attributes('disabled')).toBeUndefined()
  wrapper.unmount()
})
it('横向列表加载更多时保留已选片段和待播状态', async () => {
  mocks.request.mockImplementation(async ({ params }: { params: { page: number } }) => ({ status: true,
    data: { total: 65, items: Array.from({ length: params.page === 1 ? 50 : 15 }, (_, i) => item('m' + ((params.page - 1) * 50 + i))) },
  }))
  const wrapper = mount(MobileMomentTheater, { props: { filesBasesId: 'library-a' }, global })
  await flushPromises()
  const initial = wrapper.findComponent({ name: 'MobileVideoPlayer' }).props('initialMoment')
  await wrapper.get('.filmstrip-more').trigger('click')
  await flushPromises()
  expect(wrapper.findAll('.filmstrip-card')).toHaveLength(65)
  expect(wrapper.findComponent({ name: 'MobileVideoPlayer' }).props('initialMoment')).toBe(initial)
  expect(wrapper.find('.filmstrip-more').exists()).toBe(false)
  wrapper.unmount()
})
it('多页刷新中途失败保留当前播放器和列表，下次刷新成功可恢复', async () => {
  let fail = false
  mocks.request.mockImplementation(async ({ params, method }: { params?: { page: number }; method?: string }) => {
    if (method) return { status: true, data: {} }
    if (fail && params?.page === 2) return { status: false, msg: '第二页请求失败' }
    return { status: true, data: { total: 65, items: Array.from({ length: params?.page === 2 ? 15 : 50 }, (_, i) => item('m' + (((params?.page || 1) - 1) * 50 + i))) } }
  })
  const wrapper = mount(MobileMomentTheater, { props: { filesBasesId: 'library-a' }, global })
  await flushPromises()
  await wrapper.get('.filmstrip-more').trigger('click'); await flushPromises()
  await wrapper.findAll('.filmstrip-card')[60].trigger('click')
  const player = wrapper.findComponent({ name: 'MobileVideoPlayer' }).vm
  const initial = wrapper.findComponent({ name: 'MobileVideoPlayer' }).props('initialMoment')
  fail = true
  await momentApi.save({ dramaSeriesId: 'other', title: '其他片段更新' }); await flushPromises()
  expect(wrapper.get('[role="alert"]').text()).toContain('第二页请求失败')
  expect(wrapper.findAll('.filmstrip-card')).toHaveLength(65)
  expect(wrapper.findComponent({ name: 'MobileVideoPlayer' }).vm).toBe(player)
  expect(wrapper.findComponent({ name: 'MobileVideoPlayer' }).props('initialMoment')).toBe(initial)
  fail = false
  await wrapper.get('[role="alert"] button').trigger('click'); await flushPromises()
  expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  expect(wrapper.findComponent({ name: 'MobileVideoPlayer' }).vm).toBe(player)
  wrapper.unmount()
})
it('离开缓存页会移除播放器，再进入时重新预载并保持暂停', async () => {
  const active = ref(true)
  const wrapper = mount(defineComponent({ setup: () => () => h(KeepAlive, null, {
    default: () => active.value ? h(MobileMomentTheater, { filesBasesId: 'library-a' }) : h('div'),
  }) }), { global })
  await flushPromises()
  await wrapper.findAll('.filmstrip-card')[2].trigger('click')
  active.value = false
  await flushPromises()
  active.value = true
  await flushPromises()
  expect(wrapper.findComponent({ name: 'MobileVideoPlayer' }).props('initialMoment').autoplay).toBe(false)
  expect(wrapper.findComponent({ name: 'MobileVideoPlayer' }).props('dramaSeriesId')).toBe('first')
  wrapper.unmount()
})
it('数据刷新保留第二批片段、当前播放器和横向位置；当前片段失效才重新选择', async () => {
  let removed = false
  mocks.request.mockImplementation(async ({ params, method }: { params?: { page: number }; method?: string }) => {
    if (method) return { status: true, data: {} }
    const data = Array.from({ length: 65 }, (_, i) => item('m' + i)).filter(row => !removed || row.id !== 'm60')
    const page = params?.page || 1
    return { status: true, data: { total: data.length, items: data.slice((page - 1) * 50, page * 50) } }
  })
  const wrapper = mount(MobileMomentTheater, { props: { filesBasesId: 'library-a' }, global })
  await flushPromises()
  await wrapper.get('.filmstrip-more').trigger('click')
  await flushPromises()
  await wrapper.findAll('.filmstrip-card')[60].trigger('click')
  const initial = wrapper.findComponent({ name: 'MobileVideoPlayer' }).props('initialMoment')
  const strip = wrapper.get('.moment-filmstrip').element
  strip.scrollLeft = 420
  await momentApi.save({ dramaSeriesId: 'other', title: '其他片段更新' })
  await flushPromises()
  expect(wrapper.findAll('.filmstrip-card')).toHaveLength(65)
  expect(wrapper.findComponent({ name: 'MobileVideoPlayer' }).props('initialMoment')).toBe(initial)
  expect(strip.scrollLeft).toBe(420)
  removed = true
  await momentApi.remove('m60')
  await flushPromises()
  expect(wrapper.findAll('.filmstrip-card')).toHaveLength(64)
  expect(wrapper.findComponent({ name: 'MobileVideoPlayer' }).props('dramaSeriesId')).toBe('m0')
  expect(wrapper.findComponent({ name: 'MobileVideoPlayer' }).props('initialMoment').autoplay).toBe(false)
  wrapper.unmount()
})
