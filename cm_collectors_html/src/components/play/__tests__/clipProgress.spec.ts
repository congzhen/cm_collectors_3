import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, KeepAlive, nextTick, ref } from 'vue'
import Controls from '../videoPlayControls.vue'
import LoopPlaybackStatus from '../LoopPlaybackStatus.vue'
import MobileSeekPreview from '../MobileSeekPreview.vue'
vi.mock('@/assets/request', () => ({
  default: async () => ({
    status: true,
    data: {
      dramaSeriesId: 'd',
      status: 'ready',
      interval: 10,
      count: 100,
      completed: 100,
      duration: 1000,
      sourceKey: 'test',
    },
  }),
}))
afterEach(() => {
  vi.useRealTimers()
  document.body.innerHTML = ''
})
describe('片段专用进度条', () => {
  it('切到后台时清除手机拖动缩略图，避免返回后残留预览', async () => {
    const root = document.createElement('div')
    root.innerHTML = '<div class="xgplayer-progress"></div>'
    document.body.append(root)
    const progress = root.firstElementChild!
    vi.spyOn(progress, 'getBoundingClientRect').mockReturnValue({ left: 0, top: 0, width: 200, height: 20 } as DOMRect)
    const wrapper = mount(MobileSeekPreview, { props: {
      root, dramaSeriesId: 'd', getDuration: () => 1000, getRotation: () => 0,
    } })
    progress.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, clientX: 100 }))
    await flushPromises()
    expect(wrapper.find('.mobile-seek-preview').exists()).toBe(true)
    const hidden = vi.spyOn(document, 'hidden', 'get').mockReturnValue(true)
    try {
      document.dispatchEvent(new Event('visibilitychange'))
      await nextTick()
      expect(wrapper.find('.mobile-seek-preview').exists()).toBe(false)
    } finally { wrapper.unmount(); hidden.mockRestore() }
  })
  it('桌面拖动被后台中断且没有 pointerup 时取消预览，不提交位置', async () => {
    const wrapper = mount(Controls, { props: { duration: 100, currentTime: 10 } })
    const slider = wrapper.get<HTMLInputElement>('.progress-slider')
    await slider.trigger('pointerdown')
    slider.element.value = '50'
    await slider.trigger('input')
    const hidden = vi.spyOn(document, 'hidden', 'get').mockReturnValue(true)
    try {
      document.dispatchEvent(new Event('visibilitychange'))
      await nextTick()
      expect(wrapper.emitted('seek')).toBeUndefined()
      expect(slider.element.value).toBe('10')
      expect(wrapper.emitted('scrub-change')?.at(-1)).toEqual([false])
    } finally { wrapper.unmount(); hidden.mockRestore() }
  })
  it('手机片段进度条后台中断时解除拖动标记，原循环位置保持不变', async () => {
    vi.useFakeTimers()
    const root = document.createElement('div')
    root.innerHTML = '<div class="xgplayer-controls"></div><div class="xgplayer-progress-outer"></div>'
    document.body.append(root)
    const seek = vi.fn()
    const wrapper = mount(LoopPlaybackStatus, { props: {
      root, start: 100, end: 160, visible: true, getTime: () => 110, getDuration: () => 1000, seek,
    } })
    await vi.advanceTimersByTimeAsync(160)
    const slider = root.querySelector<HTMLInputElement>('.loop-seek')!
    slider.dispatchEvent(new Event('pointerdown', { bubbles: true }))
    slider.value = '50'
    slider.dispatchEvent(new Event('input', { bubbles: true }))
    const hidden = vi.spyOn(document, 'hidden', 'get').mockReturnValue(true)
    try {
      document.dispatchEvent(new Event('visibilitychange'))
      await nextTick()
      expect(seek).not.toHaveBeenCalled()
      expect(wrapper.emitted('scrub-change')?.at(-1)).toEqual([false])
      expect(slider.value).toBe('10')
    } finally { wrapper.unmount(); hidden.mockRestore() }
  })
  it('拖动中离开缓存页，回来后没有残留预览且键盘定位仍有效', async () => {
    const active = ref(true)
    const wrapper = mount(defineComponent({ setup: () => () => h(KeepAlive, null, {
      default: () => active.value ? h(Controls, { duration: 100, currentTime: 10 }) : h('div'),
    }) }))
    const controls = wrapper.getComponent(Controls)
    const slider = controls.get<HTMLInputElement>('.progress-slider')
    await slider.trigger('pointerdown')
    slider.element.value = '50'
    await slider.trigger('input')
    active.value = false
    await nextTick()
    active.value = true
    await nextTick()
    expect(slider.element.value).toBe('10')
    await slider.setValue('20')
    expect(controls.emitted('seek')).toEqual([[20]])
    wrapper.unmount()
  })
  it('桌面拖动期间只更新预览位置，松手一次提交最后位置', async () => {
    const wrapper = mount(Controls, { props: {
      duration: 1000, currentTime: 110, loopRange: { start: 100, end: 160, title: '片段' },
    } })
    const slider = wrapper.get<HTMLInputElement>('.progress-slider')
    await slider.trigger('pointerdown')
    slider.element.value = '20'
    await slider.trigger('input')
    slider.element.value = '40'
    await slider.trigger('input')
    expect(wrapper.emitted('seek')).toBeUndefined()
    expect(wrapper.get('.time-display').text()).toBe('0:40 / 1:00')
    await slider.trigger('change')
    expect(wrapper.emitted('seek')).toEqual([[140]])
    wrapper.unmount()
  })
  it('桌面进度条按片段计算点击、拖动和缩略图，退出后恢复全片范围', async () => {
    const wrapper = mount(Controls, {
      props: {
        dramaSeriesId: 'd',
        duration: 1000,
        currentTime: 110,
        loopRange: { start: 100, end: 160, title: '片段' },
      },
    })
    await flushPromises()
    const slider = wrapper.get<HTMLInputElement>('.progress-slider')
    expect(slider.attributes('max')).toBe('60')
    expect(slider.element.value).toBe('10')
    expect(wrapper.get('.time-display').text()).toBe('0:10 / 1:00')
    await slider.setValue('30')
    expect(wrapper.emitted('seek')?.at(-1)).toEqual([130])
    const progress = wrapper.get('.progress-container')
    vi.spyOn(progress.element, 'getBoundingClientRect').mockReturnValue({
      left: 20,
      width: 200,
    } as DOMRect)
    await progress.trigger('click', { clientX: 120 })
    expect(wrapper.emitted('seek')?.at(-1)).toEqual([130])
    await progress.trigger('pointermove', { clientX: 120 })
    expect(wrapper.get('.seek-image img').attributes('src')).toContain('/image/13?')
    expect(wrapper.get('.seek-image span').text()).toBe('0:30')
    await wrapper.setProps({ loopRange: undefined })
    expect(slider.attributes('max')).toBe('1000')
    expect(slider.element.value).toBe('110')
    expect(wrapper.get('.time-display').text()).toBe('1:50 / 16:40')
    await slider.setValue('30')
    expect(wrapper.emitted('seek')?.at(-1)).toEqual([30])
    wrapper.unmount()
  })
  it('手机进度条覆盖整段，拖动转换为源视频时间，退出时移除覆盖层', async () => {
    vi.useFakeTimers()
    const root = document.createElement('div')
    root.innerHTML =
      '<div class="xgplayer-controls"></div><div class="xgplayer-progress-outer"><span class="original-slider"></span></div>'
    document.body.append(root)
    const seek = vi.fn()
    const wrapper = mount(LoopPlaybackStatus, {
      props: {
        root,
        start: 100,
        end: 160,
        title: '片段',
        visible: true,
        getTime: () => 110,
        getDuration: () => 1000,
        seek,
      },
    })
    await vi.advanceTimersByTimeAsync(160)
    const slider = root.querySelector<HTMLInputElement>('.loop-seek')!
    expect(slider.max).toBe('60')
    expect(slider.value).toBe('10')
    slider.value = '35'
    slider.dispatchEvent(new Event('input', { bubbles: true }))
    expect(seek).toHaveBeenLastCalledWith(135)
    seek.mockClear()
    slider.dispatchEvent(new Event('pointerdown', { bubbles: true }))
    slider.value = '40'
    slider.dispatchEvent(new Event('input', { bubbles: true }))
    slider.value = '50'
    slider.dispatchEvent(new Event('input', { bubbles: true }))
    expect(seek).not.toHaveBeenCalled()
    slider.dispatchEvent(new Event('change', { bubbles: true }))
    expect(seek).toHaveBeenCalledExactlyOnceWith(150)
    expect(root.querySelector('.loop-range')).toBeNull()
    wrapper.unmount()
    expect(root.querySelector('.loop-progress')).toBeNull()
    expect(root.querySelector('.original-slider')).not.toBeNull()
    expect(vi.getTimerCount()).toBe(0)
  })
})
