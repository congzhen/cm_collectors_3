import { beforeEach, expect, it, vi } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import HomeStudioView from '../HomeStudioView.vue';
import IndexView from '../IndexView.vue';
const mocks = vi.hoisted(() => ({ change: vi.fn(), search: vi.fn(), content: vi.fn(), error: vi.fn() }));
vi.mock('@/main', () => ({ eventBus: { on: vi.fn(), off: vi.fn() } }));
vi.mock('element-plus', () => ({ ElMessage: { error: mocks.error } }));
vi.mock('@/common/resource', () => ({ resourcePinToTop: vi.fn() }));
vi.mock('@/common/homeMode', () => ({ useHomeMode: () => ({ currentHomeMode: 'classic' }) }));
vi.mock('@/storeData/app.storeData', () => ({ appStoreData: () => ({
  currentFilesBases: { id: 'old', name: 'old' }, currentPerformerBasesIds: [], currentTagClass: [],
  currentConfigApp: { resourceDetailsShowMode: 'right' }, appConfig: {}, detailsViewStatus: true,
  initCurrentFilesBases: mocks.change,
}) }));
vi.mock('@/storeData/filesBases.storeData', () => ({ filesBasesStoreData: () => ({ filesBasesStatus: [] }) }));
vi.mock('@/storeData/search.storeData', () => ({ searchStoreData: () => ({ init: mocks.search }) }));
vi.mock('@/components/home/HomeLibraryRail.vue', () => ({ default: { name: 'HomeLibraryRail', template: '<div />' } }));
vi.mock('@/components/home/HomeStageHero.vue', () => ({ default: { name: 'HomeStageHero', template: '<div />' } }));
vi.mock('@/components/home/HomeToolDrawer.vue', () => ({ default: { name: 'HomeToolDrawer', template: '<div />' } }));
vi.mock('@/components/play/playCloudCheckPromptDialog.vue', () => ({ default: { name: 'playCloudCheckPromptDialog', template: '<div />' } }));
vi.mock('@/components/play/videoPlayDialog.vue', () => ({ default: { name: 'videoPlayDialog', template: '<div />' } }));
vi.mock('@/components/resource/resourceFormDrawer.vue', () => ({ default: { name: 'resourceFormDrawer', template: '<div />' } }));
vi.mock('@/components/resource/resourceSetPerformerDialog.vue', () => ({ default: { name: 'resourceSetPerformerDialog', template: '<div />' } }));
vi.mock('@/components/resource/resourceSetTagDialog.vue', () => ({ default: { name: 'resourceSetTagDialog', template: '<div />' } }));
vi.mock('@/views/ContentView.vue', () => ({ default: { name: 'ContentView', template: '<div />', methods: { init: mocks.content } } }));
vi.mock('@/views/DetailsView.vue', () => ({ default: { name: 'DetailsView', template: '<div />', props: ['resource'] } }));
vi.mock('@/views/HeaderView.vue', () => ({ default: { name: 'HeaderView', template: '<div />' } }));
vi.mock('@/views/TagView.vue', () => ({ default: { name: 'TagView', template: '<div />' } }));
vi.mock('@/views/dataBaseMenuView.vue', () => ({ default: { name: 'dataBaseMenuView', template: '<div />' } }));

function deferred() {
  let resolve!: (value: { status: boolean; message?: string; stale?: boolean }) => void;
  let reject!: (reason: unknown) => void;
  const promise = new Promise<{ status: boolean; message?: string; stale?: boolean }>((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
}
beforeEach(() => { vi.resetAllMocks(); });
const modes = [
  { name: '现代', view: HomeStudioView, menu: 'HomeLibraryRail', event: 'select' },
  { name: '经典', view: IndexView, menu: 'dataBaseMenuView', event: 'select-files-base' },
];
for (const mode of modes) {
  const setup = () => mount(mode.view, { global: { directives: {
    loading: { mounted(el, binding) { el.dataset.loading = String(binding.value); },
      updated(el, binding) { el.dataset.loading = String(binding.value); } },
  } } });
  it.each(['failure', 'exception'])(`${mode.name}切库 %s 保留详情和筛选，关闭遮罩`, async (kind) => {
    const w = setup();
    try {
      w.findComponent({ name: 'ContentView' }).vm.$emit('selectResources', { id: 'old-resource' }, true);
      await flushPromises();
      if (kind === 'failure') mocks.change.mockResolvedValue({ status: false, message: '失败' });
      else mocks.change.mockRejectedValue(new Error('失败'));
      w.findComponent({ name: mode.menu }).vm.$emit(mode.event, { id: 'new' });
      await flushPromises();
      expect(w.findComponent({ name: 'DetailsView' }).props('resource').id).toBe('old-resource');
      expect(mocks.search).not.toHaveBeenCalled();
      expect(mocks.content).not.toHaveBeenCalled();
      expect(mocks.error).toHaveBeenCalledExactlyOnceWith('失败');
      expect(w.get('.index-view').attributes('data-loading')).toBe('false');
    } finally { w.unmount(); }
  });
  it(`${mode.name}旧请求结束不关闭新请求遮罩，也不重置筛选`, async () => {
    const w = setup();
    try {
      const first = deferred(); const second = deferred();
      mocks.change.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
      const menu = w.findComponent({ name: mode.menu });
      menu.vm.$emit(mode.event, { id: 'A' });
      menu.vm.$emit(mode.event, { id: 'B' });
      first.resolve({ status: true, stale: true });
      await flushPromises();
      expect(w.get('.index-view').attributes('data-loading')).toBe('true');
      expect(mocks.search).not.toHaveBeenCalled();
      second.resolve({ status: true });
      await flushPromises();
      expect(mocks.search).toHaveBeenCalledOnce();
      expect(mocks.content).toHaveBeenCalledOnce();
      expect(w.get('.index-view').attributes('data-loading')).toBe('false');
    } finally { w.unmount(); }
  });
}
