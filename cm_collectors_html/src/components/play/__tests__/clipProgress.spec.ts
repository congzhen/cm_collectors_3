import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import Controls from '../videoPlayControls.vue'
import LoopPlaybackStatus from '../LoopPlaybackStatus.vue'
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
    expect(root.querySelector('.loop-range')).toBeNull()
    wrapper.unmount()
    expect(root.querySelector('.loop-progress')).toBeNull()
    expect(root.querySelector('.original-slider')).not.toBeNull()
    expect(vi.getTimerCount()).toBe(0)
  })
})
