import { flushPromises, mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { defineComponent, h } from 'vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { appStoreData } from '@/storeData/app.storeData';
import ResourceVideoPlayer from '../resourceVideoPlayer.vue';

const mocks = vi.hoisted(() => ({ source: vi.fn(), setSource: vi.fn(), subtitle: vi.fn(), destroy: vi.fn() }));
vi.mock('@/storeData/app.storeData', async () => {
  const { defineStore } = await import('pinia');
  return { appStoreData: defineStore('app', { state: () => ({ appConfig: {} }) }) };
});
vi.mock('@/common/play', () => ({ getPlayVideoURLAndType: mocks.source }));
vi.mock('../videoPlay.vue', () => ({ default: defineComponent({
  setup(_, { expose }) {
    expose({ isPlaying: () => false, setVideoSource: mocks.setSource, addTextTrack: mocks.subtitle, getVideoDimensions: () => ({ width: 1920, height: 1080 }) });
    return () => h('div', { class: 'desktop-player' });
  },
}) }));
vi.mock('../mobileVideoPlayer.vue', () => ({ __esModule: true, default: defineComponent({
  props: ['resourceId', 'dramaSeriesId', 'title'],
  unmounted: () => mocks.destroy(),
  setup: props => () => h('div', { class: 'mobile-player' }, props.dramaSeriesId),
}) }));
const resource = { id: 'r', title: '测试', dramaSeries: [{ id: 'e', src: 'video.mp4' }] } as I_resource;

describe('桌面页面复用移动播放器', () => {
  beforeEach(() => {
    setActivePinia(createPinia()); localStorage.clear(); vi.clearAllMocks();
    vi.stubGlobal('navigator', { userAgent: 'Android Tablet' });
    vi.stubGlobal('innerWidth', 800); vi.stubGlobal('innerHeight', 1280);
    mocks.source.mockResolvedValue({ playUrl: '/video.mp4', playType: 'mp4' });
  });
  afterEach(() => vi.unstubAllGlobals());
  it('旧配置加载桌面视频并保留字幕', async () => {
    const wrapper = mount(ResourceVideoPlayer, { props: { resource, dramaSeriesId: 'e' } });
    await flushPromises();
    expect(wrapper.find('.desktop-player').exists()).toBe(true);
    expect(mocks.source).toHaveBeenCalledWith('e');
    mocks.setSource.mock.calls[0][2]();
    expect(mocks.subtitle).toHaveBeenCalledWith('/api/video/subtitle/e', '默认字幕', 'zh', true);
    wrapper.unmount();
  });
  it('平板使用移动播放器，旋转、配置更新、切集不会切换播放器类型', async () => {
    appStoreData().appConfig.largeMobilePlayer = 'mobile';
    const wrapper = mount(ResourceVideoPlayer, { props: { resource, dramaSeriesId: 'e' } });
    await flushPromises();
    expect(wrapper.find('.mobile-player').exists()).toBe(true);
    expect(mocks.source).not.toHaveBeenCalled();
    vi.stubGlobal('innerWidth', 1280); vi.stubGlobal('innerHeight', 800);
    window.dispatchEvent(new Event('resize'));
    appStoreData().appConfig.largeMobilePlayer = 'desktop';
    await wrapper.setProps({ dramaSeriesId: 'e2' });
    expect(wrapper.find('.mobile-player').text()).toBe('e2');
    expect(mocks.destroy).not.toHaveBeenCalled();
    wrapper.unmount();
    expect(mocks.destroy).toHaveBeenCalledOnce();
  });
  it('没有本设备选择栏，后台修改在重新打开时生效', async () => {
    localStorage.setItem('video_player_preference', 'mobile');
    const wrapper = mount(ResourceVideoPlayer, { props: { resource, dramaSeriesId: 'e' } });
    await flushPromises();
    expect(wrapper.find('select').exists()).toBe(false);
    expect(wrapper.text()).not.toContain('本设备播放器');
    expect(wrapper.find('.desktop-player').exists()).toBe(true);
    appStoreData().appConfig.largeMobilePlayer = 'mobile';
    await flushPromises();
    expect(wrapper.find('.desktop-player').exists()).toBe(true);
    wrapper.unmount();
    const reopened = mount(ResourceVideoPlayer, { props: { resource, dramaSeriesId: 'e' } });
    await flushPromises(); expect(reopened.find('.mobile-player').exists()).toBe(true);
    reopened.unmount();
  });
  it('切集后旧请求不能覆盖新地址', async () => {
    let resolveOld!: (value: { playUrl: string; playType: string }) => void;
    mocks.source.mockImplementationOnce(() => new Promise(resolve => { resolveOld = resolve; }));
    const wrapper = mount(ResourceVideoPlayer, { props: { resource, dramaSeriesId: 'e' } });
    await flushPromises(); await wrapper.setProps({ dramaSeriesId: 'e2' }); await flushPromises();
    resolveOld({ playUrl: '/old.mp4', playType: 'mp4' }); await flushPromises();
    expect(mocks.setSource).toHaveBeenCalledOnce();
    expect(mocks.setSource.mock.calls[0][0]).toBe('/video.mp4');
    wrapper.unmount();
  });
});
