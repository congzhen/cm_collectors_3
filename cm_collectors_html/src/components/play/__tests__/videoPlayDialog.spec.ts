import { flushPromises, mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, h } from 'vue';
import VideoPlayDialog from '../videoPlayDialog.vue';

const mocks = vi.hoisted(() => ({ info: vi.fn(), destroy: vi.fn() }));
vi.mock('@/server/resource.server', () => ({ resourceServer: { info: mocks.info } }));
vi.mock('../resourceVideoPlayer.vue', () => ({ default: defineComponent({
  props: ['resource', 'dramaSeriesId'],
  unmounted: () => mocks.destroy(),
  setup: props => () => h('div', { class: 'test-player' }, props.dramaSeriesId),
}) }));
const dialogStub = defineComponent({
  props: ['modelValue'], emits: ['update:modelValue', 'close'],
  setup: (props, { slots }) => () => h('div', { 'data-open': props.modelValue }, slots.default?.()),
});
const create = () => mount(VideoPlayDialog, { global: {
  stubs: { 'el-dialog': dialogStub, 'el-icon': true, VideoPlay: true, ArrowUp: true, ArrowDown: true, FullScreen: true, BottomLeft: true, Close: true },
  directives: { loading: {} },
} });

describe('播放弹窗生命周期', () => {
  beforeEach(() => { vi.clearAllMocks(); mocks.info.mockResolvedValue({ status: true, data: { id: 'r', title: '测试', dramaSeries: [{ id: 'e' }] } }); });
  it('打开默认分集，关闭销毁播放器，重新打开按新会话挂载', async () => {
    const wrapper = create();
    wrapper.vm.open('r', ''); await flushPromises();
    expect(wrapper.find('.test-player').text()).toBe('e');
    wrapper.vm.close(); await flushPromises();
    expect(wrapper.find('.test-player').exists()).toBe(false);
    expect(mocks.destroy).toHaveBeenCalledOnce();
    wrapper.findComponent(dialogStub).vm.$emit('close'); await flushPromises();
    wrapper.vm.open('r', 'e2'); await flushPromises();
    expect(wrapper.find('.test-player').text()).toBe('e2');
    wrapper.unmount();
  });
  it('关闭后完成的资源请求不再挂载播放器', async () => {
    let finish!: (value: unknown) => void;
    mocks.info.mockImplementationOnce(() => new Promise(resolve => { finish = resolve; }));
    const wrapper = create(); wrapper.vm.open('r', ''); await flushPromises();
    wrapper.vm.close(); wrapper.findComponent(dialogStub).vm.$emit('close'); await flushPromises();
    finish({ status: true, data: { id: 'r', title: '测试', dramaSeries: [{ id: 'e' }] } }); await flushPromises();
    expect(wrapper.find('.test-player').exists()).toBe(false);
    wrapper.unmount();
  });
});
