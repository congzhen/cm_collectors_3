import { mount } from '@vue/test-utils';
import { nextTick, defineComponent, h } from 'vue';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import VideoPlay from '../videoPlay.vue';

const mocks = vi.hoisted(() => ({
  fullscreen: (_value: boolean) => {},
  paused: vi.fn(() => false), pause: vi.fn(), play: vi.fn(),
}));
vi.mock('@/storeData/app.storeData', () => ({ appStoreData: () => ({ runtimeBridgeStatus: true }) }));
vi.mock('@/common/play', () => ({ openInPlayerDramaSeries: vi.fn() }));
vi.mock('@/common/runtimeBridge', () => ({
  setHostWindowFullscreen: vi.fn().mockResolvedValue(undefined),
  onHostWindowFullscreenChanged: (callback: (value: boolean) => void) => { mocks.fullscreen = callback; return () => {}; },
}));
vi.mock('video.js', () => ({ default: Object.assign(vi.fn(() => ({
  on: vi.fn(), ready: vi.fn(), aspectRatio: vi.fn(), paused: mocks.paused,
  pause: mocks.pause, play: mocks.play, dispose: vi.fn(),
})), { getComponent: () => class {}, registerComponent: vi.fn() }) }));
vi.mock('../playCloudCheckPromptDialog.vue', () => ({ default: defineComponent({ setup: () => () => h('div') }) }));
vi.mock('../videoPlayControls.vue', () => ({ default: defineComponent({
  setup(_, { expose }) {
    expose({ setFullscreen: vi.fn(), setMaximized: vi.fn() });
    return () => h('div', { class: 'test-controls' }, h('button', { onTouchstart: (event: Event) => event.stopPropagation() }, '进度条'));
  },
}) }));

function touch(target: Element, type: string, count: number) {
  const event = new Event(type, { bubbles: true, cancelable: true });
  Object.defineProperty(event, 'touches', { value: Array.from({ length: count }, () => ({})) });
  target.dispatchEvent(event);
  return event;
}
async function create() {
  const wrapper = mount(VideoPlay, { props: { forceDesktop: true }, attachTo: document.body, global: { stubs: { 'el-icon': true } } });
  wrapper.vm.setVideoSource('/test.mp4', 'mp4', () => {});
  await nextTick(); await vi.advanceTimersByTimeAsync(100);
  mocks.fullscreen(true); await nextTick();
  const root = document.querySelector('.video-player-container')!;
  await vi.advanceTimersByTimeAsync(3001);
  expect(root.classList.contains('controls-hidden')).toBe(true);
  return { wrapper, root };
}

describe('PC 全屏播放器触控唤出控制条', () => {
  beforeEach(() => { vi.useFakeTimers(); vi.clearAllMocks(); mocks.paused.mockReturnValue(false); });
  afterEach(() => { vi.useRealTimers(); });
  it('轻触、滑动唤醒，持续触摸不隐藏，松手后恢复自动隐藏', async () => {
    const { wrapper, root } = await create();
    const video = root.querySelector('video')!;
    expect(touch(video, 'touchstart', 1).defaultPrevented).toBe(false);
    await nextTick(); expect(root.classList.contains('controls-hidden')).toBe(false);
    touch(video, 'touchmove', 1);
    await vi.advanceTimersByTimeAsync(5000);
    expect(root.classList.contains('controls-hidden')).toBe(false);
    touch(video, 'touchend', 0);
    await vi.advanceTimersByTimeAsync(2999); expect(root.classList.contains('controls-hidden')).toBe(false);
    await vi.advanceTimersByTimeAsync(2); expect(root.classList.contains('controls-hidden')).toBe(true);
    // 单独的滑动活动也可以唤醒，不依赖模拟鼠标事件。
    touch(video, 'touchmove', 1); await nextTick();
    expect(root.classList.contains('controls-hidden')).toBe(false);
    expect(mocks.play).not.toHaveBeenCalled(); expect(mocks.pause).not.toHaveBeenCalled();
    wrapper.unmount();
  });
  it('子控件阻止冒泡仍能保持显示，多指和取消触摸能正确恢复计时', async () => {
    const { wrapper, root } = await create();
    const control = root.querySelector('.test-controls button')!;
    touch(control, 'touchstart', 2); await nextTick();
    expect(root.classList.contains('controls-hidden')).toBe(false);
    touch(control, 'touchend', 1);
    await vi.advanceTimersByTimeAsync(4000); expect(root.classList.contains('controls-hidden')).toBe(false);
    touch(control, 'touchcancel', 0);
    await vi.advanceTimersByTimeAsync(3001); expect(root.classList.contains('controls-hidden')).toBe(true);
    wrapper.unmount();
  });
  it('暂停时触摸后不自动隐藏，鼠标仍可唤醒', async () => {
    const { wrapper, root } = await create();
    root.dispatchEvent(new MouseEvent('mousemove', { bubbles: true })); await nextTick();
    expect(root.classList.contains('controls-hidden')).toBe(false);
    mocks.paused.mockReturnValue(true);
    touch(root, 'touchstart', 1); touch(root, 'touchend', 0);
    await vi.advanceTimersByTimeAsync(5000); expect(root.classList.contains('controls-hidden')).toBe(false);
    wrapper.unmount();
  });
});
