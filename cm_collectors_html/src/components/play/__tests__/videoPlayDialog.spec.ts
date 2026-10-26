import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import VideoPlayDialog from '../videoPlayDialog.vue'
import ResourceVideoPlayer from '../resourceVideoPlayer.vue'

const mocks = vi.hoisted(() => ({ info: vi.fn(), destroy: vi.fn(), error: vi.fn() }))
vi.mock('element-plus', () => ({ ElMessage: { error: mocks.error } }))
vi.mock('@/server/resource.server', () => ({ resourceServer: { info: mocks.info } }))
vi.mock('../resourceVideoPlayer.vue', () => ({
  default: defineComponent({
    props: ['resource', 'dramaSeriesId', 'initialMoment'],
    unmounted: () => mocks.destroy(),
    setup: (props) => () => h('div', { class: 'test-player' }, props.dramaSeriesId),
  }),
}))
const dialogStub = defineComponent({
  props: ['modelValue'],
  emits: ['update:modelValue', 'close'],
  setup:
    (props, { slots }) =>
    () =>
      h('div', { 'data-open': props.modelValue }, slots.default?.()),
})
const create = () =>
  mount(VideoPlayDialog, {
    global: {
      stubs: {
        'el-dialog': dialogStub,
        'el-icon': true,
        VideoPlay: true,
        ArrowUp: true,
        ArrowDown: true,
        FullScreen: true,
        BottomLeft: true,
        Close: true,
      },
      directives: {
        loading: {
          mounted: (el, binding) => el.setAttribute('data-loading', String(binding.value)),
          updated: (el, binding) => el.setAttribute('data-loading', String(binding.value)),
        },
      },
    },
  })

describe('播放弹窗生命周期', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.info.mockResolvedValue({
      status: true,
      data: { id: 'r', title: '测试', dramaSeries: [{ id: 'e' }] },
    })
  })
  it('打开默认分集，关闭销毁播放器，重新打开按新会话挂载', async () => {
    const wrapper = create()
    wrapper.vm.open('r', '')
    await flushPromises()
    expect(wrapper.find('.test-player').text()).toBe('e')
    wrapper.vm.close()
    await flushPromises()
    expect(wrapper.find('.test-player').exists()).toBe(false)
    expect(mocks.destroy).toHaveBeenCalledOnce()
    wrapper.findComponent(dialogStub).vm.$emit('close')
    await flushPromises()
    wrapper.vm.open('r', 'e2')
    await flushPromises()
    expect(wrapper.find('.test-player').text()).toBe('e2')
    wrapper.unmount()
  })
  it('关闭后完成的资源请求不再挂载播放器', async () => {
    let finish!: (value: unknown) => void
    mocks.info.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          finish = resolve
        }),
    )
    const wrapper = create()
    wrapper.vm.open('r', '')
    await flushPromises()
    wrapper.vm.close()
    wrapper.findComponent(dialogStub).vm.$emit('close')
    await flushPromises()
    finish({ status: true, data: { id: 'r', title: '测试', dramaSeries: [{ id: 'e' }] } })
    await flushPromises()
    expect(wrapper.find('.test-player').exists()).toBe(false)
    wrapper.unmount()
  })
  it('收藏播放传递起止点，同一分集再次选片段也重新挂载', async () => {
    const wrapper = create()
    const initialMoment = { start: 10, end: 20, title: '片段1' }
    wrapper.vm.open('r', 'e', initialMoment)
    await flushPromises()
    expect(wrapper.getComponent(ResourceVideoPlayer).props('initialMoment')).toEqual(initialMoment)
    wrapper.vm.open('r', 'e', { start: 30, end: 40, title: '片段2' })
    await flushPromises()
    expect(wrapper.getComponent(ResourceVideoPlayer).props('initialMoment')?.start).toBe(30)
    expect(mocks.destroy).toHaveBeenCalled()
    wrapper.unmount()
  })
  it('快速切换资源时，旧请求失败不打断新请求的加载状态或弹出错误', async () => {
    let oldDone!: (value: unknown) => void, newDone!: (value: unknown) => void
    mocks.info
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            oldDone = resolve
          }),
      )
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            newDone = resolve
          }),
      )
    const wrapper = create()
    wrapper.vm.open('old', 'e-old')
    await flushPromises()
    wrapper.vm.open('new', 'e-new')
    await flushPromises()
    oldDone({ status: false, msg: '旧请求失败' })
    await flushPromises()
    expect(mocks.error).not.toHaveBeenCalled()
    expect(wrapper.get('.video-play').attributes('data-loading')).toBe('true')
    newDone({ status: true, data: { id: 'new', title: '新资源', dramaSeries: [{ id: 'e-new' }] } })
    await flushPromises()
    expect(wrapper.get('.test-player').text()).toBe('e-new')
    expect(wrapper.get('.video-play').attributes('data-loading')).toBe('false')
    wrapper.unmount()
  })
})
