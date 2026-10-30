import { flushPromises, mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import ElementPlus, { ElPagination, ElPopover } from 'element-plus';
import ContentView from '../ContentView.vue';
import ContentListAdmin from '@/components/content/contentListAdmin.vue';

const mocks = vi.hoisted(() => ({ list: vi.fn(), stats: vi.fn() }));
vi.mock('@/main', () => ({ eventBus: { on: vi.fn(), off: vi.fn() } }));
vi.mock('@/assets/debounce', () => ({ debounce: (fn: (...args: unknown[]) => unknown) => fn }));
vi.mock('@/server/resource.server', () => ({ resourceServer: { dataList: mocks.list, fileSizeStats: mocks.stats } }));
vi.mock('@/storeData/app.storeData', async () => {
  const { reactive } = await import('vue');
  const store = reactive({ currentFilesBases: { id: 'library' }, currentConfigApp: { pageLimit: 2 }, appConfig: {}, adminResourceStatus: true });
  return { appStoreData: () => store };
});
vi.mock('@/storeData/search.storeData', () => ({ searchStoreData: () => ({ searchData: {} }) }));
vi.mock('@/components/content/contentList.vue', () => ({ default: { template: '<div />' } }));
vi.mock('@/components/content/contentListAdmin.vue', () => ({ default: { template: '<div />', emits: ['updateData'], methods: { change: vi.fn() } } }));
vi.mock('@/components/play/LibraryMomentsButton.vue', () => ({ default: { template: '<div />' } }));
vi.mock('@/components/setting/fileDatabaseSetting/coverAdjuster.vue', () => ({ default: { template: '<div />' } }));
vi.mock('@/components/playList/playListBtn.vue', () => ({ default: { template: '<div />' } }));

const response = (ids: string[], total: number) => ({ status: true, data: { dataList: ids.map(id => ({ id })), total } });
const mountView = () => mount(ContentView, { global: { plugins: [ElementPlus], directives: { admin: {} }, stubs: { InfoFilled: true } } });
beforeEach(() => {
  vi.clearAllMocks();
  mocks.list.mockResolvedValue(response(['a', 'b'], 3));
  mocks.stats.mockResolvedValue({ status: true, data: { totalSize: 1, totalFiles: 1, countedFiles: 1, uncountedFiles: 0 } });
});
afterEach(() => vi.restoreAllMocks());

describe('管理操作后的列表刷新', () => {
  it('批量操作后重新统计文件大小，不复用操作前缓存', async () => {
    const wrapper = mountView();
    try {
      await flushPromises();
      wrapper.findComponent(ElPopover).vm.$emit('before-enter');
      await flushPromises();
      expect(mocks.stats).toHaveBeenCalledTimes(1);
      wrapper.findComponent(ContentListAdmin).vm.$emit('updateData');
      await flushPromises();
      wrapper.findComponent(ElPopover).vm.$emit('before-enter');
      await flushPromises();
      expect(mocks.stats).toHaveBeenCalledTimes(2);
    } finally { wrapper.unmount(); }
  });

  it('翻页请求未返回时立即移除上一页资源，失败后也不保留旧页', async () => {
    const wrapper = mountView();
    try {
      await flushPromises();
      let finish!: (value: ReturnType<typeof response>) => void;
      mocks.list.mockImplementationOnce(() => new Promise(resolve => { finish = resolve; }));
      await wrapper.findComponent(ElPagination).findAll('.el-pager li').find(item => item.text() === '2')!.trigger('click');
      expect(wrapper.vm.showDataList()).toEqual([]);
      finish({ ...response([], 0), status: false });
      await flushPromises();
      expect(wrapper.vm.showDataList()).toEqual([]);
    } finally { wrapper.unmount(); }
  });

  it('多次刷新乱序返回时只显示最新请求的资源', async () => {
    const wrapper = mountView();
    try {
      await flushPromises();
      let oldResult!: (value: ReturnType<typeof response>) => void;
      let newResult!: (value: ReturnType<typeof response>) => void;
      mocks.list.mockImplementationOnce(() => new Promise(resolve => { oldResult = resolve; }));
      mocks.list.mockImplementationOnce(() => new Promise(resolve => { newResult = resolve; }));
      const oldCallback = vi.fn();
      void wrapper.vm.init_DataList(oldCallback);
      void wrapper.vm.init_DataList();
      newResult(response(['latest'], 3));
      await flushPromises();
      oldResult(response(['stale'], 3));
      await flushPromises();
      expect(wrapper.vm.showDataList()).toEqual([{ id: 'latest' }]);
      expect(oldCallback).not.toHaveBeenCalled();
    } finally { wrapper.unmount(); }
  });

  it('删除末页最后一个资源后自动回到有效页并加载资源', async () => {
    const wrapper = mountView();
    try {
      await flushPromises();
      mocks.list.mockResolvedValueOnce(response(['c'], 0));
      const pagination = wrapper.findComponent(ElPagination);
      await pagination.findAll('.el-pager li').find(item => item.text() === '2')!.trigger('click');
      await flushPromises();
      mocks.list.mockImplementation(async (_library, _count, page) => page === 2 ? response([], 2) : response(['a', 'b'], 2));
      wrapper.findComponent(ContentListAdmin).vm.$emit('updateData');
      await flushPromises();
      expect(pagination.props('currentPage')).toBe(1);
      expect(wrapper.vm.showDataList().map((item: { id: string }) => item.id)).toEqual(['a', 'b']);
    } finally { wrapper.unmount(); }
  });
});
