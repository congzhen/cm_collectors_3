import { flushPromises, mount } from '@vue/test-utils';
import { defineComponent, h, KeepAlive } from 'vue';
import { createMemoryHistory, createRouter, RouterView } from 'vue-router';
import { expect, it, vi } from 'vitest';
import MobileView from '@/views/MobileView.vue';
import { readMobileSession } from '@/common/mobileSession';
const state = vi.hoisted(() => ({
  app: { appConfig: {}, currentConfigApp: { pageLimit: 20 }, currentFilesBases: { id: 'library' }, initCurrentFilesBases: vi.fn() },
  search: { init: vi.fn(), searchData: { searchTextSlc: [], sort: 'name', country: { options: [] }, definition: { options: [] }, videoCodec: { options: [] }, year: { options: [] }, star: { options: [] }, performer: { options: [] }, cup: { options: [] }, tag: {} } },
}));
vi.mock('@/storeData/app.storeData', () => ({ appStoreData: () => state.app }));
vi.mock('@/storeData/search.storeData', () => ({ searchStoreData: () => state.search }));
vi.mock('@/storeData/filesBases.storeData', () => ({ filesBasesStoreData: () => ({ filesBasesStatus: [] }) }));
vi.mock('@/server/resource.server', () => ({ resourceServer: { dataList: vi.fn(async () => ({ status: true, data: { dataList: [{ id: 'r' }], total: 1 } })) } }));
vi.mock('@/assets/debounce', () => ({ debounce: (fn: unknown) => fn }));
vi.mock('@/common/play', () => ({ playResource: vi.fn() }));
vi.mock('@/views/TagView.vue', () => ({ default: { render: () => null } }));
vi.mock('@/components/mobile/MobileResourceLayout.vue', () => ({ default: { name: 'MobileResourceLayout', props: ['dataList'], render: () => h('div', { style: 'height:3000px' }) } }));
it('离开前记住位置，缓存 DOM 的位置归零后返回仍恢复到原位置', async () => {
  localStorage.clear();
  const router = createRouter({ history: createMemoryHistory(), routes: [
    { path: '/mobile', component: MobileView }, { path: '/detail', component: { render: () => h('div', '详情') } },
  ] });
  await router.push('/mobile'); await router.isReady();
  const app = defineComponent({ setup: () => () => h(RouterView, {}, { default: ({ Component }: any) => h(KeepAlive, {}, () => h(Component)) }) });
  const wrapper = mount(app, { attachTo: document.body, global: { plugins: [router], directives: { loading: () => {} }, stubs: { ElDrawer: { render: () => null }, ElDropdown: true, ElDropdownMenu: true, ElDropdownItem: true, ElSelect: true, ElOption: true, ElButton: true, ElIcon: true, ElInput: true } } });
  await flushPromises();
  const scroller = wrapper.find('.mobile-content').element as HTMLElement;
  scroller.scrollTop = 860;
  await router.push('/detail'); await flushPromises();
  scroller.scrollTop = 0; scroller.dispatchEvent(new Event('scroll'));
  expect(readMobileSession().list?.scrollTop).toBe(860);
  router.back(); await flushPromises();
  expect((wrapper.find('.mobile-content').element as HTMLElement).scrollTop).toBe(860);
  wrapper.unmount();
});

it('切库失败恢复选择器，保留资源列表、筛选及滚动位置', async () => {
  localStorage.clear();
  state.app.initCurrentFilesBases.mockResolvedValue({ status: false, message: '切库失败' });
  state.search.init.mockClear();
  const router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/mobile', component: MobileView }] });
  await router.push('/mobile'); await router.isReady();
  const wrapper = mount(defineComponent({ render: () => h(RouterView) }), { global: { plugins: [router], directives: {
    loading: { mounted(el, b) { el.dataset.loading = String(b.value); }, updated(el, b) { el.dataset.loading = String(b.value); } },
  }, stubs: { ElDrawer: { render: () => null }, ElDropdown: true, ElDropdownMenu: true, ElDropdownItem: true,
    ElSelect: { name: 'LibrarySelect', props: ['modelValue'], template: '<div />' }, ElOption: true, ElButton: true, ElIcon: true, ElInput: true } } });
  try {
    await flushPromises();
    const scroller = wrapper.get('.mobile-content').element as HTMLElement;
    scroller.scrollTop = 300;
    const select = wrapper.findComponent({ name: 'LibrarySelect' });
    select.vm.$emit('update:modelValue', 'new');
    select.vm.$emit('change', 'new');
    await flushPromises();
    expect(select.props('modelValue')).toBe('library');
    expect(state.search.init).not.toHaveBeenCalled();
    expect(scroller.scrollTop).toBe(300);
    expect(wrapper.get('.mobile-view').attributes('data-loading')).toBe('false');
    expect(wrapper.findComponent({ name: 'MobileResourceLayout' }).props('dataList')).toEqual([{ id: 'r' }]);
  } finally { wrapper.unmount(); }
});
