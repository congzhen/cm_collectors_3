import { flushPromises, mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, h, KeepAlive, ref, nextTick } from 'vue';
import MobileVideoPlayer from '../mobileVideoPlayer.vue';
import { readMobilePlayback, saveMobilePlayback } from '@/common/mobileSession';

const mocks = vi.hoisted(() => ({ source: vi.fn(), instances: [] as MockPlayer[] }));
interface MockPlayer {
  root: HTMLElement; video: HTMLVideoElement; config: Record<string, unknown>;
  currentTime: number; duration: number; playbackRate: number; paused: boolean;
  muted: boolean; volume: number;
  downloadHook?: () => boolean;
  on: (event: string, callback: () => void) => void;
  fire: (event: string) => void; play: ReturnType<typeof vi.fn>; destroy: ReturnType<typeof vi.fn>;
}
vi.mock('@/common/mobilePlaybackSource', () => ({ resolveMobileSource: mocks.source, MobileSourceError: class extends Error {} }));
vi.mock('xgplayer-hls', () => ({ default: { isSupported: () => true } }));
vi.mock('xgplayer', () => ({
  Events: Object.fromEntries(['LOADED_METADATA', 'LOADED_DATA', 'PLAY', 'PLAYING', 'PAUSE', 'ENDED', 'TIME_UPDATE', 'SEEKING', 'SEEKED', 'RATE_CHANGE', 'VOLUME_CHANGE', 'ERROR', 'WAITING', 'STALLED', 'LOADING'].map(event => [event, event])),
  default: class {
    root: HTMLElement; video = document.createElement('video'); currentTime = 0; duration = 200; playbackRate = 1; paused = true;
    muted = false; volume = 0.6;
    cssfullscreen = false; rotateDeg = 0;
    getCssFullscreen() { this.cssfullscreen = true; }
    exitCssFullscreen() { this.cssfullscreen = false; }
    listeners = new Map<string, (() => void)[]>();
    config: Record<string, unknown>;
    play = vi.fn(async () => { this.paused = false; this.fire('PLAY'); });
    pause = vi.fn(() => { this.paused = true; this.fire('PAUSE'); });
    destroy = vi.fn(); focus = vi.fn(); getPlugin = () => ({ config: {} });
    downloadHook?: () => boolean;
    usePluginHooks(plugin: string, _event: string, callback: () => boolean) {
      if (plugin === 'download') this.downloadHook = callback;
    }
    constructor(config: Record<string, unknown>) {
      this.config = config; this.root = config.el as HTMLElement; mocks.instances.push(this);
      this.volume = config.volume as number;
    }
    on(event: string, callback: () => void) { this.listeners.set(event, [...(this.listeners.get(event) || []), callback]); }
    fire(event: string) { this.listeners.get(event)?.forEach(callback => callback()); }
  },
}));

describe('手机播放恢复与请求隔离', () => {
  beforeEach(() => {
    mocks.instances.length = 0; localStorage.clear();
    mocks.source.mockReset().mockResolvedValue({ playUrl: '/movie.mp4', playType: 'mp4' });
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false }));
    Object.defineProperty(document, 'hidden', { configurable: true, value: false });
  });
  afterEach(() => {
    delete (window as Window & { __cmPhoneShellDownload?: boolean }).__cmPhoneShellDownload;
    vi.unstubAllGlobals(); vi.useRealTimers();
  });
  it('APK 点击下载由桥接接管，浏览器继续内置下载', async () => {
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '视频' } });
    await flushPromises();
    const hook = mocks.instances[0].downloadHook!;
    expect(hook()).toBe(true);
    (window as Window & { __cmPhoneShellDownload?: boolean }).__cmPhoneShellDownload = true;
    const listener = vi.fn();
    window.addEventListener('cm-phone-download', listener);
    try {
      expect(hook()).toBe(false);
      expect(hook()).toBe(false);
      expect(listener).toHaveBeenCalledOnce();
    } finally { window.removeEventListener('cm-phone-download', listener); wrapper.unmount(); }
  });
  it('APK 分片播放不提交伪装成视频的播放列表下载', async () => {
    mocks.source.mockResolvedValue({ playUrl: '/api/video/m3u8/e/v.m3u8', playType: 'm3u8' });
    (window as Window & { __cmPhoneShellDownload?: boolean }).__cmPhoneShellDownload = true;
    const alert = vi.spyOn(window, 'alert').mockImplementation(() => {});
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '视频' } });
    await flushPromises();
    expect(mocks.instances[0].downloadHook!()).toBe(false);
    expect(alert).toHaveBeenCalledWith(expect.stringContaining('分片播放'));
    wrapper.unmount(); alert.mockRestore();
  });
  it('未播放时定位到开头画面，不调用播放也不改变静音设置', async () => {
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '视频' } });
    await flushPromises(); const player = mocks.instances[0]; const muted = player.muted;
    player.fire('LOADED_METADATA');
    expect(player.currentTime).toBe(0.01); expect(player.play).not.toHaveBeenCalled(); expect(player.muted).toBe(muted);
    wrapper.unmount();
  });
  it('页面重建后定位到保存进度，默认不自动出声', async () => {
    saveMobilePlayback('r/e', { time: 60, duration: 200, rate: 1.5, playing: true });
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises();
    const player = mocks.instances[0];
    expect(player.config.startTime).toBe(60);
    player.fire('LOADED_DATA');
    expect(player.currentTime).toBe(60);
    expect(player.play).not.toHaveBeenCalled();
    wrapper.unmount(); await flushPromises();
  });
  it('没有方向 API 的旧内核仍可横屏，并能恢复退出', async () => {
    vi.stubGlobal('screen', {}); vi.stubGlobal('innerWidth', 390); vi.stubGlobal('innerHeight', 844);
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises();
    const player = mocks.instances[0];
    Object.defineProperties(player.video, { videoWidth: { value: 1920 }, videoHeight: { value: 1080 } });
    const toggle = (player.config.fullscreen as { switchCallback: () => Promise<void> }).switchCallback;
    await toggle();
    expect(player.root.classList.contains('phone-fallback-rotate')).toBe(true);
    await toggle();
    expect(player.root.classList.contains('phone-fallback-rotate')).toBe(false);
    expect(document.body.style.overflow).not.toBe('hidden');
    wrapper.unmount(); await flushPromises();
  });
  it('离开缓存的手机列表后，联网事件不重建隐藏播放器', async () => {
    const visible = ref(true);
    const wrapper = mount(defineComponent({ setup: () => () => h(KeepAlive, null, {
      default: () => visible.value ? h(MobileVideoPlayer, { resourceId: 'r', dramaSeriesId: 'e', title: '测试' }) : h('div'),
    }) }));
    await flushPromises(); mocks.instances[0].fire('LOADED_DATA');
    mocks.instances[0].fire('ERROR');
    visible.value = false; await nextTick();
    window.dispatchEvent(new Event('online')); await flushPromises();
    expect(mocks.source).toHaveBeenCalledOnce();
    wrapper.unmount(); await flushPromises();
  });
  it('静音和音量跨视频、跨页面保留，手动取消静音也会记住', async () => {
    const props = { resourceId: 'r', dramaSeriesId: 'e1', title: '测试' };
    const wrapper = mount(MobileVideoPlayer, { props });
    await flushPromises();
    const first = mocks.instances[0];
    first.muted = true; first.volume = 0.25; first.fire('VOLUME_CHANGE');
    await wrapper.setProps({ dramaSeriesId: 'e2', autoplay: true }); await flushPromises();
    const second = mocks.instances[1];
    expect(second.muted).toBe(true); expect(second.volume).toBe(0.25);
    second.fire('LOADED_DATA');
    expect(second.play).toHaveBeenCalledOnce(); expect(second.muted).toBe(true);
    second.muted = false; second.volume = 0; second.fire('VOLUME_CHANGE');
    wrapper.unmount(); await flushPromises();
    const reopened = mount(MobileVideoPlayer, { props }); await flushPromises();
    expect(mocks.instances[2].muted).toBe(false);
    expect(mocks.instances[2].volume).toBe(0);
    reopened.unmount(); await flushPromises();
  });
  it('前后台切换保存进度，正常播放器不重建，之前暂停不会自动播放', async () => {
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises();
    const player = mocks.instances[0]; player.fire('LOADED_DATA'); player.currentTime = 35;
    Object.defineProperty(document, 'hidden', { configurable: true, value: true });
    document.dispatchEvent(new Event('visibilitychange'));
    expect(readMobilePlayback('r/e')?.time).toBe(35);
    Object.defineProperty(document, 'hidden', { configurable: true, value: false });
    document.dispatchEvent(new Event('visibilitychange'));
    expect(player.play).not.toHaveBeenCalled(); expect(mocks.instances).toHaveLength(1);
    wrapper.unmount(); await flushPromises();
  });
  it('切换分集时保存旧分集进度，旧请求不覆盖新分集', async () => {
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e1', title: '测试' } });
    await flushPromises();
    const first = mocks.instances[0]; first.fire('LOADED_DATA'); first.currentTime = 40;
    await wrapper.setProps({ dramaSeriesId: 'e2' }); await flushPromises();
    expect(readMobilePlayback('r/e1')?.time).toBe(40);
    expect(readMobilePlayback('r/e2')).toBeUndefined();
    expect(mocks.instances[1].config.startTime).toBe(0);
    wrapper.unmount(); await flushPromises();
  });
  it('播放源返回前退出页面不会创建播放器', async () => {
    let resolve!: (value: unknown) => void;
    mocks.source.mockImplementation(() => new Promise(done => { resolve = done; }));
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises(); wrapper.unmount();
    resolve({ playUrl: '/late.mp4', playType: 'mp4' }); await flushPromises();
    expect(mocks.instances).toHaveLength(0);
  });
  it('内嵌模式切换资源可继续播放，首次打开仍保持暂停', async () => {
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e1', title: '测试' } });
    await flushPromises();
    mocks.instances[0].fire('LOADED_DATA');
    expect(mocks.instances[0].play).not.toHaveBeenCalled();
    await wrapper.setProps({ dramaSeriesId: 'e2', autoplay: true }); await flushPromises();
    mocks.instances[1].fire('LOADED_DATA');
    expect(mocks.instances[1].play).toHaveBeenCalledOnce();
    wrapper.unmount(); await flushPromises();
  });
  it('后台被系统暂停后返回继续播放，保留同一个播放器', async () => {
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises();
    const player = mocks.instances[0]; player.fire('LOADED_DATA'); player.fire('PLAY');
    Object.defineProperty(document, 'hidden', { configurable: true, value: true });
    document.dispatchEvent(new Event('visibilitychange')); player.fire('PAUSE');
    Object.defineProperty(document, 'hidden', { configurable: true, value: false });
    document.dispatchEvent(new Event('visibilitychange'));
    expect(player.play).toHaveBeenCalledOnce(); expect(mocks.instances).toHaveLength(1);
    wrapper.unmount(); await flushPromises();
  });
  it('拖动后缓冲超过旧超时仍保留播放器，长时间等待可手动重试', async () => {
    vi.useFakeTimers();
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises();
    const player = mocks.instances[0];
    player.fire('LOADED_DATA'); player.fire('PLAY'); player.fire('PLAYING');
    player.muted = true; player.fire('VOLUME_CHANGE');
    player.currentTime = 160; player.fire('SEEKING'); player.fire('WAITING');
    await vi.advanceTimersByTimeAsync(25000);
    expect(mocks.instances).toHaveLength(1); expect(player.destroy).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(40000); await nextTick();
    expect(player.root.textContent).toContain('可继续等待');
    expect(mocks.source).toHaveBeenCalledOnce();
    const retry = player.root.querySelector('button')!;
    retry.click(); await flushPromises();
    expect(mocks.instances).toHaveLength(2);
    expect(mocks.instances[1].config.startTime).toBe(160);
    expect(mocks.instances[1].muted).toBe(true);
    wrapper.unmount(); await flushPromises();
  });
  it('缓冲范围持续增长时不提示失败，恢复播放后清除等待状态', async () => {
    vi.useFakeTimers();
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises(); const player = mocks.instances[0];
    player.fire('LOADED_DATA'); player.fire('PLAY'); player.fire('PLAYING');
    let end = 101;
    Object.defineProperty(player.video, 'buffered', { value: { length: 1, start: () => 100, end: () => end } });
    player.currentTime = 100; player.fire('SEEKING');
    for (let i = 0; i < 4; i++) { await vi.advanceTimersByTimeAsync(30000); end++; }
    await nextTick();
    expect(player.root.textContent).not.toContain('重试');
    player.fire('PLAYING'); await vi.advanceTimersByTimeAsync(70000); await nextTick();
    expect(player.root.textContent).not.toContain('正在加载中');
    expect(mocks.instances).toHaveLength(1);
    wrapper.unmount(); await flushPromises();
  });
  it('连续拖动采用最后位置，暂停定位完成后不会重连或自动播放', async () => {
    vi.useFakeTimers();
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises(); const player = mocks.instances[0]; player.fire('LOADED_DATA');
    for (const time of [40, 120, 180]) { player.currentTime = time; player.fire('SEEKING'); await vi.advanceTimersByTimeAsync(10000); }
    Object.defineProperty(player.video, 'readyState', { value: 2 });
    player.fire('SEEKED'); await vi.advanceTimersByTimeAsync(70000); await nextTick();
    expect(readMobilePlayback('r/e')?.time).toBe(180);
    expect(player.play).not.toHaveBeenCalled(); expect(mocks.instances).toHaveLength(1);
    expect(player.root.textContent).not.toContain('重试');
    wrapper.unmount(); await flushPromises();
  });
  it('首帧慢加载切回前台继续等待原请求，不重新创建播放器', async () => {
    vi.useFakeTimers();
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises();
    await vi.advanceTimersByTimeAsync(25000);
    Object.defineProperty(document, 'hidden', { configurable: true, value: true });
    document.dispatchEvent(new Event('visibilitychange'));
    Object.defineProperty(document, 'hidden', { configurable: true, value: false });
    document.dispatchEvent(new Event('visibilitychange')); await flushPromises();
    expect(mocks.instances).toHaveLength(1); expect(mocks.source).toHaveBeenCalledOnce();
    wrapper.unmount(); await flushPromises();
  });
  it('等待提示出现后用户暂停，切回前台不会擅自续播', async () => {
    vi.useFakeTimers();
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises(); const player = mocks.instances[0];
    player.fire('LOADED_DATA'); player.fire('PLAY'); player.fire('WAITING');
    await vi.advanceTimersByTimeAsync(65000); player.fire('PAUSE');
    Object.defineProperty(document, 'hidden', { configurable: true, value: true });
    document.dispatchEvent(new Event('visibilitychange'));
    Object.defineProperty(document, 'hidden', { configurable: true, value: false });
    document.dispatchEvent(new Event('visibilitychange')); await flushPromises();
    expect(player.play).not.toHaveBeenCalled(); expect(mocks.instances).toHaveLength(1);
    wrapper.unmount(); await flushPromises();
  });
  it('真实解码错误仍提示失败，不自动重建播放器', async () => {
    vi.useFakeTimers();
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises(); const player = mocks.instances[0];
    Object.defineProperty(player.video, 'error', { value: { code: 3 } });
    player.fire('ERROR'); await vi.advanceTimersByTimeAsync(70000); await nextTick();
    expect(player.root.textContent).toContain('视频无法播放');
    expect(mocks.instances).toHaveLength(1);
    wrapper.unmount(); await flushPromises();
  });
  it('首帧、拖动及播放缓冲显示加载提示，恢复后隐藏', async () => {
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises(); const player = mocks.instances[0];
    expect(player.root.textContent).toContain('正在加载中');
    expect(player.root.querySelector('.phone-loading-spinner')).not.toBeNull();
    player.fire('LOADED_DATA'); await nextTick();
    expect(player.root.textContent).not.toContain('正在加载中');
    player.fire('PLAY'); await nextTick();
    expect(player.root.textContent).toContain('正在加载中');
    player.fire('PLAYING'); await nextTick();
    expect(player.root.querySelector('.phone-loading-spinner')).toBeNull();
    for (const event of ['WAITING', 'SEEKING', 'STALLED']) {
      player.fire(event); await nextTick();
      expect(player.root.textContent).toContain('正在加载中');
      player.fire('PLAYING'); await nextTick();
      expect(player.root.textContent).not.toContain('正在加载中');
    }
    wrapper.unmount(); await flushPromises();
  });
  it('暂停后的预加载停顿不会显示加载提示，真实错误也不被覆盖', async () => {
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises(); const player = mocks.instances[0]; player.fire('LOADED_DATA');
    player.fire('PLAY'); player.fire('WAITING'); player.fire('PAUSE'); player.fire('STALLED');
    await nextTick(); expect(player.root.textContent).not.toContain('正在加载中');
    player.fire('PLAY');
    Object.defineProperty(player.video, 'error', { value: { code: 3 } });
    player.fire('ERROR'); player.fire('STALLED'); player.fire('WAITING'); await nextTick();
    expect(player.root.textContent).toContain('视频无法播放');
    expect(player.root.querySelector('.phone-loading-spinner')).toBeNull();
    wrapper.unmount(); await flushPromises();
  });
  it('WebView 未发 waiting 时，播放器加载样式及进度停滞仍显示提示', async () => {
    vi.useFakeTimers();
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises(); const player = mocks.instances[0];
    player.fire('LOADED_DATA'); await player.play(); player.fire('PLAYING');
    Object.defineProperty(player.video, 'readyState', { value: 4 });
    player.root.classList.add('xgplayer-isloading');
    await vi.advanceTimersByTimeAsync(500); await nextTick();
    expect(player.root.textContent).toContain('正在加载中');
    player.root.classList.remove('xgplayer-isloading'); player.currentTime = 1;
    await vi.advanceTimersByTimeAsync(500); await nextTick();
    expect(player.root.textContent).not.toContain('正在加载中');
    await vi.advanceTimersByTimeAsync(3000); await nextTick();
    expect(player.root.textContent).toContain('正在加载中');
    player.currentTime = 2; await vi.advanceTimersByTimeAsync(500); await nextTick();
    expect(player.root.textContent).not.toContain('正在加载中');
    player.paused = true; player.fire('PAUSE'); await vi.advanceTimersByTimeAsync(4000); await nextTick();
    expect(player.root.textContent).not.toContain('正在加载中');
    expect(mocks.instances).toHaveLength(1);
    wrapper.unmount(); await flushPromises();
  });
  it('长缓冲后仅进度恢复而未再发 playing，也清除慢加载提示', async () => {
    vi.useFakeTimers();
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises(); const player = mocks.instances[0];
    player.fire('LOADED_DATA'); await player.play(); player.fire('PLAYING'); player.fire('WAITING');
    await vi.advanceTimersByTimeAsync(65000); await nextTick();
    expect(player.root.textContent).toContain('可继续等待');
    Object.defineProperty(player.video, 'readyState', { value: 4 });
    player.currentTime = 10;
    await vi.advanceTimersByTimeAsync(500); await nextTick();
    expect(player.root.textContent).not.toContain('可继续等待');
    expect(player.root.textContent).not.toContain('重试');
    expect(mocks.instances).toHaveLength(1);
    wrapper.unmount(); await flushPromises();
  });
  it('播放结束清除加载提示和缓冲定时器', async () => {
    vi.useFakeTimers();
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises(); const player = mocks.instances[0];
    player.fire('LOADED_DATA'); await player.play(); player.fire('WAITING');
    player.currentTime = player.duration; player.fire('ENDED');
    await vi.advanceTimersByTimeAsync(70000); await nextTick();
    expect(player.root.querySelector('.phone-player-message')).toBeNull();
    expect(mocks.instances).toHaveLength(1);
    wrapper.unmount(); await flushPromises();
  });
  it('缓存页停用期间首帧就绪，返回时补齐状态并继续保留进度', async () => {
    const visible = ref(true);
    const wrapper = mount(defineComponent({ setup: () => () => h(KeepAlive, null, {
      default: () => visible.value ? h(MobileVideoPlayer, { resourceId: 'r', dramaSeriesId: 'e', title: '测试' }) : h('div'),
    }) }));
    await flushPromises(); const player = mocks.instances[0];
    visible.value = false; await nextTick();
    Object.defineProperty(player.video, 'readyState', { value: 2 }); player.fire('LOADED_DATA');
    visible.value = true; await nextTick(); await flushPromises();
    expect(player.root.textContent).not.toContain('正在加载中');
    player.currentTime = 42; player.fire('SEEKED');
    expect(readMobilePlayback('r/e')?.time).toBe(42);
    expect(mocks.instances).toHaveLength(1); expect(player.play).not.toHaveBeenCalled();
    wrapper.unmount(); await flushPromises();
  });
  it('连接失败最多自动重试三次，避免无限重载', async () => {
    vi.useFakeTimers(); mocks.source.mockRejectedValue(new Error('network'));
    const wrapper = mount(MobileVideoPlayer, { props: { resourceId: 'r', dramaSeriesId: 'e', title: '测试' } });
    await flushPromises();
    for (const delay of [1000, 2000, 4000, 10000]) {
      await vi.advanceTimersByTimeAsync(delay); await flushPromises();
    }
    expect(mocks.source).toHaveBeenCalledTimes(4);
    expect(wrapper.text()).toContain('重试');
    wrapper.unmount(); await flushPromises();
  });
});
