import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import LoopPlaybackStatus from '../LoopPlaybackStatus.vue'

afterEach(() => {
  vi.useRealTimers()
  document.body.innerHTML = ''
})

describe('片段循环控制条提示', () => {
  it.each([
    ['progress-slider-container', 'controls-left'],
    ['xgplayer-progress-outer', 'xgplayer-controls'],
  ])('标记范围并更新片段时间，兼容 %s', async (progressClass, timeClass) => {
    vi.useFakeTimers()
    const root = document.createElement('div')
    root.innerHTML = `<div class="video-js vjs-controls-disabled"><div class="vjs-control-bar" style="display:none"><div class="vjs-progress-holder"></div></div></div><div class="${timeClass === 'controls-left' ? 'video-controller' : 'xgplayer'}"><div class="${progressClass}"></div><div class="${timeClass}"></div></div>`
    document.body.append(root)
    let time = 20
    const wrapper = mount(LoopPlaybackStatus, {
      props: {
        root,
        start: 10,
        end: 30,
        title: '收藏片段',
        visible: true,
        getTime: () => time,
        getDuration: () => 100,
      },
    })
    await vi.advanceTimersByTimeAsync(160)
    const range = root.querySelector<HTMLElement>('.loop-range')!
    expect(range.style.left).toBe('10%')
    expect(range.style.width).toBe('20%')
    expect(root.querySelector('.vjs-control-bar .loop-range')).toBeNull()
    expect(root.querySelector('.vjs-control-bar .loop-status')).toBeNull()
    expect(root.querySelector(`.${progressClass} .loop-range`)).not.toBeNull()
    expect(root.querySelector(`.${timeClass} .loop-status`)).not.toBeNull()
    expect(root.querySelector('.loop-status')?.textContent).toContain('收藏片段')
    expect(root.querySelector('.loop-status')?.textContent).toContain('00:10 / 00:20')
    time = 10
    await vi.advanceTimersByTimeAsync(160)
    expect(root.querySelector('.loop-status')?.textContent).toContain('00:00 / 00:20')
    wrapper.unmount()
    expect(root.querySelector('.loop-range')).toBeNull()
    expect(root.querySelector('.loop-status')).toBeNull()
    expect(vi.getTimerCount()).toBe(0)
  })

  it('元数据未就绪时不显示无效范围，原生控制条回退提示跟随显隐', async () => {
    const wrapper = mount(LoopPlaybackStatus, {
      props: {
        start: 10,
        end: 30,
        visible: true,
        getTime: () => 0,
        getDuration: () => 0,
      },
    })
    expect(wrapper.find('.loop-range').exists()).toBe(false)
    expect(wrapper.text()).toContain('片段循环')
    expect(wrapper.text()).toContain('00:00 / 00:20')
    await wrapper.setProps({ visible: false })
    expect(wrapper.find('.loop-status').exists()).toBe(false)
    wrapper.unmount()
  })
})
