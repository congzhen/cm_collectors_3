import { flushPromises, mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import ElementPlus, { ElPagination, ElPopover } from 'element-plus';
import ContentView from '../ContentView.vue';
import ContentList from '@/components/content/contentList.vue';
import ContentListAdmin from '@/components/content/contentListAdmin.vue';
import { appStoreData } from '@/storeData/app.storeData';

const mocks = vi.hoisted(() => ({ list: vi.fn(), stats: vi.fn(), setDisplay: vi.fn(), summaries: vi.fn() }));
vi.mock('@/main', () => ({ eventBus: { on: vi.fn(), off: vi.fn() } }));
vi.mock('@/assets/debounce', () => ({ debounce: (fn: (...args: unknown[]) => unknown) => fn }));
vi.mock('@/server/resourceGroup.server', () => ({ resourceGroupServer: { setDisplay: mocks.setDisplay, summaries: mocks.summaries } }));
vi.mock('@/server/resource.server', () => ({ resourceServer: { dataList: mocks.list, fileSizeStats: mocks.stats } }));
vi.mock('@/storeData/app.storeData', async () => {
  const { reactive } = await import('vue');
  const store = reactive({ currentFilesBases: { id: 'library' }, currentConfigApp: { pageLimit: 2 }, appConfig: {}, adminResourceStatus: true });
  return { appStoreData: () => store };
});
vi.mock('@/storeData/search.storeData', () => ({ searchStoreData: () => ({ searchData: {} }) }));
vi.mock('@/components/content/contentList.vue', () => ({ default: { props: ['dataList','groupPage'], template: '<div />', methods: { change: vi.fn() } } }));
vi.mock('@/components/content/contentListAdmin.vue', () => ({ default: { template: '<div />', emits: ['updateData'], methods: { change: vi.fn() } } }));
vi.mock('@/components/play/LibraryMomentsButton.vue', () => ({ default: { template: '<div />' } }));
vi.mock('@/components/setting/fileDatabaseSetting/coverAdjuster.vue', () => ({ default: { template: '<div />' } }));
vi.mock('@/components/playList/playListBtn.vue', () => ({ default: { template: '<div />' } }));

const response = (ids: string[], total: number) => ({ status: true, data: { dataList: ids.map(id => ({ id })), total } });
const mountView = () => mount(ContentView, { global: { plugins: [ElementPlus], directives: { admin: {} }, stubs: { InfoFilled: true } } });
beforeEach(() => {
  vi.clearAllMocks();
  appStoreData().adminResourceStatus = true;
  appStoreData().currentFilesBases.id = 'library';
  appStoreData().currentConfigApp.mergeResourceGroups = true;
  appStoreData().currentConfigApp.showVideoDuration = false;
  mocks.summaries.mockResolvedValue({status:true,data:{dataList:[],total:0}});
  mocks.setDisplay.mockResolvedValue({ status: true, data: true });
  mocks.list.mockResolvedValue(response(['a', 'b'], 3));
  mocks.stats.mockResolvedValue({ status: true, data: { totalSize: 1, totalFiles: 1, countedFiles: 1, uncountedFiles: 0 } });
});
afterEach(() => { vi.restoreAllMocks(); vi.useRealTimers(); });

describe('管理操作后的列表刷新', () => {
  it('分页开关保存独立设置，不重新请求资源列表，管理模式暂停开关', async () => {
    appStoreData().adminResourceStatus = false;
    const wrapper = mountView();
    try {
      await flushPromises();
      const calls = mocks.list.mock.calls.length;
      await wrapper.find('.group-display button').trigger('click');
      await flushPromises();
      expect(mocks.setDisplay).toHaveBeenCalledWith('library', false);
      expect(appStoreData().currentConfigApp.mergeResourceGroups).toBe(false);
      expect(mocks.list).toHaveBeenCalledTimes(calls);
      appStoreData().adminResourceStatus = true;
      await flushPromises();
      expect(wrapper.find('.group-display button').attributes('disabled')).toBeDefined();
      expect(appStoreData().currentConfigApp.mergeResourceGroups).toBe(false);
    } finally { wrapper.unmount(); }
  });

  it('保存失败保留原开关状态，切库后旧请求不回写新库', async () => {
    appStoreData().adminResourceStatus = false;
    const wrapper = mountView();
    try {
      await flushPromises();
      mocks.setDisplay.mockResolvedValueOnce({status:false,msg:'failed'});
      await wrapper.find('.group-display button').trigger('click');
      await flushPromises();
      expect(appStoreData().currentConfigApp.mergeResourceGroups).toBe(true);
      let finish!: (value: { status: boolean; data: boolean }) => void;
      mocks.setDisplay.mockImplementationOnce(() => new Promise(resolve => { finish = resolve; }));
      await wrapper.find('.group-display button').trigger('click');
      appStoreData().currentFilesBases.id = 'other';
      finish({status:true,data:true});
      await flushPromises();
      expect(appStoreData().currentConfigApp.mergeResourceGroups).toBe(true);
    } finally { wrapper.unmount(); }
  });

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

it('资源与组摘要一起发布，摘要请求未结束前不显示临时拼封',async()=>{
 appStoreData().adminResourceStatus=false;
 const item={id:'a',associationGroupId:'g',dramaSeries:[]};
 mocks.list.mockResolvedValue({status:true,data:{dataList:[item],total:1}});
 let finish!: (value:unknown)=>void;
 mocks.summaries.mockImplementationOnce(()=>new Promise(resolve=>{finish=resolve;}));
 const wrapper=mountView();
 try{
  await flushPromises();expect(wrapper.vm.showDataList()).toEqual([]);
  finish({status:true,data:{dataList:[{id:'g',count:7,preview:[item]}],total:1}});await flushPromises();
  expect(wrapper.vm.showDataList()).toEqual([item]);
  expect(wrapper.findComponent(ContentList).props('groupPage')?.summaries[0].count).toBe(7);
  expect(mocks.summaries).toHaveBeenCalledTimes(1);
 }finally{wrapper.unmount();}
});
it('2.5秒时长补采只更新时长，保留列表、资源、分集和组摘要对象，不重新请求组数据',async()=>{
 vi.useFakeTimers();appStoreData().adminResourceStatus=false;appStoreData().currentConfigApp.showVideoDuration=true;
 const item={id:'a',associationGroupId:'g',dramaSeries:[{id:'video',durationSeconds:0}]};
 mocks.list.mockResolvedValueOnce({status:true,data:{dataList:[item],total:1}});
 mocks.list.mockResolvedValue({status:true,data:{dataList:[{...item,dramaSeries:[{id:'video',durationSeconds:80}]}],total:1}});
 const wrapper=mountView();
 try{
  await flushPromises();const list=wrapper.vm.showDataList(),resource=list[0],series=resource.dramaSeries[0];
  const page=wrapper.findComponent(ContentList).props('groupPage');
  await vi.advanceTimersByTimeAsync(2500);await flushPromises();
  expect(wrapper.vm.showDataList()).toBe(list);expect(list[0]).toBe(resource);expect(resource.dramaSeries[0]).toBe(series);
  expect(series.durationSeconds).toBe(80);expect(wrapper.findComponent(ContentList).props('groupPage')).toBe(page);
  expect(mocks.summaries).toHaveBeenCalledTimes(1);
 }finally{wrapper.unmount();}
});
it('组摘要失败不阻断资源显示，等待摘要时旧页请求也不能覆盖新页',async()=>{
 appStoreData().adminResourceStatus=false;
 const item={id:'a',associationGroupId:'g'};
 mocks.list.mockResolvedValue({status:true,data:{dataList:[item],total:1}});
 let finish!: (value:unknown)=>void;
 mocks.summaries.mockImplementationOnce(()=>new Promise(resolve=>{finish=resolve;}));
 const wrapper=mountView();
 try{
  await flushPromises();mocks.list.mockResolvedValue(response(['latest'],1));void wrapper.vm.init_DataList();await flushPromises();
  expect(wrapper.vm.showDataList()).toEqual([{id:'latest'}]);
  finish({status:true,data:{dataList:[{id:'g',count:7}],total:1}});await flushPromises();
  expect(wrapper.vm.showDataList()).toEqual([{id:'latest'}]);
  mocks.list.mockResolvedValue({status:true,data:{dataList:[item],total:1}});mocks.summaries.mockRejectedValueOnce(new Error('offline'));
  void wrapper.vm.init_DataList();await flushPromises();expect(wrapper.vm.showDataList()).toEqual([item]);
 }finally{wrapper.unmount();}
});

it('后台时长补采网络失败保留当前资源和时长，不产生未处理的异步异常',async()=>{
 vi.useFakeTimers();appStoreData().adminResourceStatus=false;appStoreData().currentConfigApp.showVideoDuration=true;
 const item={id:'a',dramaSeries:[{id:'video',durationSeconds:80}]};
 mocks.list.mockResolvedValueOnce({status:true,data:{dataList:[item],total:1}});
 mocks.list.mockRejectedValue(new Error('offline'));
 const wrapper=mountView();
 try{
  await flushPromises();const list=wrapper.vm.showDataList();
  await vi.advanceTimersByTimeAsync(2500);await flushPromises();
  expect(wrapper.vm.showDataList()).toBe(list);expect(list[0].dramaSeries[0].durationSeconds).toBe(80);
 }finally{wrapper.unmount();}
});

it.each(['failed','superseded'] as const)('首次初始化 %s 后重新加载成功，分页仍可继续请求',async(kind)=>{
 let finish: ((value:ReturnType<typeof response>)=>void)|undefined;
 if(kind==='superseded') mocks.list.mockImplementationOnce(()=>new Promise(resolve=>{finish=resolve;}));
 else mocks.list.mockResolvedValueOnce({...response([],0),status:false,msg:'first request failed'});
 const wrapper=mountView();
 try{
  await flushPromises();
  mocks.list.mockImplementation(async(_library,_count,page)=>response(page===2?['c']:['a','b'],3));
  void wrapper.vm.init_DataList();await flushPromises();
  expect(wrapper.vm.showDataList().map(item=>item.id)).toEqual(['a','b']);
  await wrapper.findComponent(ElPagination).findAll('.el-pager li').find(item=>item.text()==='2')!.trigger('click');
  await flushPromises();expect(mocks.list).toHaveBeenCalledTimes(3);
  expect(wrapper.vm.showDataList().map(item=>item.id)).toEqual(['c']);
  finish?.(response(['stale'],3));await flushPromises();
  expect(wrapper.vm.showDataList().map(item=>item.id)).toEqual(['c']);
 }finally{wrapper.unmount();}
});
it('关闭时长显示后，已排队的补采不再发送额外资源请求',async()=>{
 vi.useFakeTimers();appStoreData().adminResourceStatus=false;appStoreData().currentConfigApp.showVideoDuration=true;
 const wrapper=mountView();
 try{
  await flushPromises();expect(mocks.list).toHaveBeenCalledTimes(1);
  appStoreData().currentConfigApp.showVideoDuration=false;
  await vi.advanceTimersByTimeAsync(2500);await flushPromises();
  expect(mocks.list).toHaveBeenCalledTimes(1);
 }finally{wrapper.unmount();}
});
