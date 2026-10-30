import { flushPromises, mount, shallowMount } from '@vue/test-utils';
import { computed, defineComponent, inject, ref } from 'vue';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import * as icons from '@element-plus/icons-vue';
import ElementPlus, { ElMessage, ElMessageBox, ElRadioGroup, ElTable } from 'element-plus';
import ContentListAdmin from '../contentListAdmin.vue';
import ContentRightClickMenu from '../contentRightClickMenu.vue';
import LayoutTable from '../layoutTable.vue';
import ShortVideo from '../layoutShortVideo.vue';
import ShortVideoTopBottom from '../layoutShortVideoTopBottom.vue';
import MosaicShortVideo from '../layoutCoverPosterMosaicShortVideo.vue';
import { resourceSelectionKey } from '../resourceSelection';
import type { I_resource } from '@/dataType/resource.dataType';

const mocks = vi.hoisted(() => ({ open: vi.fn(), remove: vi.fn(), play: vi.fn(), addTranscode: vi.fn(), playUrl: vi.fn() }));
vi.mock('@/main', () => ({ eventBus: { emit: vi.fn() } }));
vi.mock('@/common/play', () => ({ getPlayVideoURLAndType: mocks.playUrl, playResource: mocks.play, playOpenResourceFolder: vi.fn() }));
vi.mock('@/common/resource', () => ({ resourceDelete: vi.fn(), resourceBatchDelete: mocks.remove }));
vi.mock('@/common/playList', () => ({ playListAdd: vi.fn() }));
vi.mock('@/common/photo', () => ({ getResourceCoverPoster: () => '' }));
vi.mock('@/language/app.lang', () => ({ AppLang: () => ({ country: () => '' }) }));
vi.mock('@/storeData/app.storeData', () => ({ appStoreData: () => ({ currentFilesBases: { id: 'library' }, currentConfigApp: {}, displayAdminFn: false }) }));
vi.mock('@/storeData/search.storeData', () => ({ searchStoreData: () => ({ searchData: { sort: '' } }) }));
vi.mock('@/server/videoTranscode.server', () => ({ videoTranscodeServer: { add: mocks.addTranscode } }));
vi.mock('@/server/tvboxRecommend.server', () => ({ tvboxRecommendServer: { add: vi.fn() } }));
vi.mock('@/server/resource.server', () => ({ resourceServer: {} }));
vi.mock('../../resource/resourceFormDrawer.vue', () => ({ default: { template: '<div />' } }));
vi.mock('../../resource/resourceSetTagBatchDialog.vue', () => ({ default: { template: '<div />', methods: { open: mocks.open } } }));
vi.mock('../../resource/resourceSetPerformerBatchDialog.vue', () => ({ default: { template: '<div />', methods: { open: mocks.open } } }));
vi.mock('../../resource/resourceSetStarsBatchDialog.vue', () => ({ default: { template: '<div />', methods: { open: mocks.open } } }));

const resources = ['a', 'b'].map(id => ({ id, title: id, coverPoster: '', tags: [], performers: [], dramaSeries: [] }) as unknown as I_resource);
const LayoutStub = defineComponent({
  props: ['dataList'],
  setup() { return { selection: inject(resourceSelectionKey)! }; },
  template: '<div><button v-for="resource in dataList" :key="resource.id" :data-id="resource.id" @click="selection.toggle(resource)">{{ resource.title }}</button></div>',
});
const mountAdmin = () => mount(ContentListAdmin, {
  props: { dataList: resources },
  global: { plugins: [ElementPlus], components: icons, stubs: { contentList: LayoutStub } },
});
const button = (wrapper: ReturnType<typeof mountAdmin>, label: string) => wrapper.findAll('button').find(item => item.text() === label)!;

beforeEach(() => {
  vi.clearAllMocks();
  vi.stubGlobal('ResizeObserver', class { observe() {} disconnect() {} });
});
afterEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals(); });
describe('资源管理视图', () => {
  it('两种视图共享选择，批量操作使用同一份资源清单', async () => {
    const wrapper = mountAdmin();
    try {
      expect(button(wrapper, '添加标签').attributes('disabled')).toBeDefined();
      await wrapper.get('[data-id="a"]').trigger('click');
      await wrapper.findComponent(ElRadioGroup).setValue('table');
      await flushPromises();
      const table = wrapper.findComponent(ElTable);
      expect(table.vm.getSelectionRows().map((item: I_resource) => item.id)).toEqual(['a']);
      table.vm.toggleRowSelection(resources[1], true);
      await flushPromises();
      await button(wrapper, '添加标签').trigger('click');
      expect(mocks.open).toHaveBeenCalledWith(resources, 'add');
      await wrapper.findComponent(ElRadioGroup).setValue('layout');
      await flushPromises();
      expect(wrapper.get('.selection-count').text()).toContain('2');
      await button(wrapper, '删除资源').trigger('click');
      expect(mocks.remove).toHaveBeenCalledWith(resources, expect.any(Function));
      await button(wrapper, '清空选择').trigger('click');
      expect(wrapper.get('.selection-count').text()).toContain('0');
    } finally { wrapper.unmount(); }
  });

  it('悬停查看已选资源，移除成员同步到表格，最后一项移除后关闭列表', async () => {
    const wrapper = mountAdmin();
    try {
      expect(button(wrapper, '查看已选').attributes('disabled')).toBeDefined();
      await wrapper.get('[data-id="a"]').trigger('click');
      await wrapper.get('[data-id="b"]').trigger('click');
      await wrapper.findComponent(ElRadioGroup).setValue('table');
      await flushPromises();
      await button(wrapper, '查看已选').trigger('mouseenter');
      await vi.waitFor(() => expect(document.body.querySelectorAll('.selection-preview__item')).toHaveLength(2));
      const remove = document.body.querySelector('[aria-label="取消选择：a"]') as HTMLButtonElement;
      remove.click();
      await flushPromises();
      expect(wrapper.findComponent(ElTable).vm.getSelectionRows().map((resource: I_resource) => resource.id)).toEqual(['b']);
      expect(document.body.querySelectorAll('.selection-preview__item')).toHaveLength(1);
      (document.body.querySelector('[aria-label="取消选择：b"]') as HTMLButtonElement).click();
      await flushPromises();
      expect(wrapper.findComponent(ElTable).vm.getSelectionRows()).toEqual([]);
      expect(button(wrapper, '查看已选').attributes('disabled')).toBeDefined();
      expect(wrapper.findComponent({ name: 'ElPopover' }).props('visible')).toBe(false);
    } finally { wrapper.unmount(); }
  });

  it('鼠标从入口移入已选面板后保持展开，清空和刷新同步更新', async () => {
    const wrapper = mountAdmin();
    try {
      await wrapper.get('[data-id="a"]').trigger('click');
      const trigger = button(wrapper, '查看已选');
      await trigger.trigger('mouseenter');
      await vi.waitFor(() => expect(document.body.querySelectorAll('.selection-preview__item')).toHaveLength(1));
      await trigger.trigger('mouseleave');
      document.body.querySelector('.selection-preview')!.closest('.el-popper')!.dispatchEvent(new MouseEvent('mouseenter'));
      await new Promise(resolve => setTimeout(resolve, 240));
      expect(wrapper.findComponent({ name: 'ElPopover' }).props('visible')).toBe(true);
      await wrapper.setProps({ dataList: [{ ...resources[0], title: 'updated-title' }, resources[1]] });
      await flushPromises();
      expect(document.body.querySelector('.selection-preview__title')!.textContent).toBe('updated-title');
      await button(wrapper, '清空选择').trigger('click');
      await flushPromises();
      expect(wrapper.findComponent({ name: 'ElPopover' }).props('visible')).toBe(false);
      expect(button(wrapper, '查看已选').attributes('disabled')).toBeDefined();
    } finally { wrapper.unmount(); }
  });

  it('悬停延时尚未结束就清空或翻页，不能弹出旧选择列表', async () => {
    const wrapper = mountAdmin();
    try {
      await wrapper.get('[data-id="a"]').trigger('click');
      await button(wrapper, '查看已选').trigger('mouseenter');
      await button(wrapper, '清空选择').trigger('click');
      await new Promise(resolve => setTimeout(resolve, 240));
      expect(wrapper.findComponent({ name: 'ElPopover' }).props('visible')).toBe(false);
      await wrapper.get('[data-id="b"]').trigger('click');
      await button(wrapper, '查看已选').trigger('mouseenter');
      await wrapper.setProps({ dataList: [] });
      await new Promise(resolve => setTimeout(resolve, 240));
      expect(wrapper.findComponent({ name: 'ElPopover' }).props('visible')).toBe(false);
      expect(wrapper.get('.selection-count').text()).toContain('0');
    } finally { wrapper.unmount(); }
  });

  it('管理表格全选后在已选列表移除资源，两个全选框均变为半选', async () => {
    const wrapper = mountAdmin();
    try {
      await wrapper.findComponent(ElRadioGroup).setValue('table');
      await flushPromises();
      await wrapper.get('.management-selection input').setValue(true);
      await flushPromises();
      expect(wrapper.findComponent(ElTable).vm.getSelectionRows()).toHaveLength(2);
      await button(wrapper, '查看已选').trigger('mouseenter');
      await vi.waitFor(() => expect(document.body.querySelectorAll('.selection-preview__item')).toHaveLength(2));
      (document.body.querySelector('[aria-label="取消选择：a"]') as HTMLButtonElement).click();
      await flushPromises();
      expect(wrapper.findComponent(ElTable).vm.getSelectionRows().map((item: I_resource) => item.id)).toEqual(['b']);
      expect(wrapper.get('.management-selection .el-checkbox__input').classes()).toContain('is-indeterminate');
      expect(wrapper.get('.el-table__header .el-checkbox__input').classes()).toContain('is-indeterminate');
    } finally { wrapper.unmount(); }
  });

  it('已选列表展开时切换视图，移除后通过表格表头重新全选和取消', async () => {
    const wrapper = mountAdmin();
    try {
      await wrapper.get('[data-id="a"]').trigger('click');
      await wrapper.get('[data-id="b"]').trigger('click');
      await button(wrapper, '查看已选').trigger('mouseenter');
      await vi.waitFor(() => expect(document.body.querySelectorAll('.selection-preview__item')).toHaveLength(2));
      await wrapper.findComponent(ElRadioGroup).setValue('table');
      await flushPromises();
      (document.body.querySelector('[aria-label="取消选择：a"]') as HTMLButtonElement).click();
      await flushPromises();
      const table = wrapper.findComponent(ElTable);
      const headerCheckbox = () => wrapper.get('.el-table__header input[type=checkbox]');
      await headerCheckbox().setValue(true);
      await flushPromises();
      await vi.waitFor(() => expect(table.vm.getSelectionRows().map((resource: I_resource) => resource.id).sort()).toEqual(['a', 'b']));
      expect(document.body.querySelectorAll('.selection-preview__item')).toHaveLength(2);
      await headerCheckbox().setValue(false);
      await vi.waitFor(() => expect(table.vm.getSelectionRows()).toEqual([]));
      await flushPromises();
      expect(wrapper.findComponent({ name: 'ElPopover' }).props('visible')).toBe(false);
      await wrapper.findComponent(ElRadioGroup).setValue('layout');
      await flushPromises();
      expect(wrapper.get('.selection-count').text()).toContain('0');
      expect(button(wrapper, '删除资源').attributes('disabled')).toBeDefined();
    } finally { wrapper.unmount(); }
  });

  it('转码业务失败不清空选择，修正后可再次提交成功', async () => {
    vi.spyOn(ElMessageBox, 'confirm').mockResolvedValue('confirm' as Awaited<ReturnType<typeof ElMessageBox.confirm>>);
    const error = vi.spyOn(ElMessage, 'error').mockImplementation(() => ({ close: vi.fn() }));
    const success = vi.spyOn(ElMessage, 'success').mockImplementation(() => ({ close: vi.fn() }));
    mocks.addTranscode.mockResolvedValueOnce({ status: false, msg: 'service unavailable' });
    mocks.addTranscode.mockResolvedValueOnce({ status: true, data: { added: 1, skippedDuplicate: 0, skippedMissing: 0 } });
    const wrapper = mountAdmin();
    try {
      await wrapper.get('[data-id="a"]').trigger('click');
      await button(wrapper, '加入转码与剪辑').trigger('click');
      await flushPromises();
      expect(error).toHaveBeenCalledWith('service unavailable');
      expect(wrapper.get('.selection-count').text()).toContain('1');
      expect(button(wrapper, '加入转码与剪辑').classes()).not.toContain('is-loading');
      await button(wrapper, '加入转码与剪辑').trigger('click');
      await flushPromises();
      expect(mocks.addTranscode).toHaveBeenCalledTimes(2);
      expect(mocks.addTranscode).toHaveBeenNthCalledWith(2, { resourceIds: ['a'] });
      expect(success).toHaveBeenCalledOnce();
      expect(wrapper.get('.selection-count').text()).toContain('1');
    } finally { wrapper.unmount(); }
  });

  it('管理表格刷新为相同 ID 的新对象时保留选择并使用最新数据', async () => {
    const wrapper = mountAdmin();
    try {
      await wrapper.get('[data-id="a"]').trigger('click');
      await wrapper.findComponent(ElRadioGroup).setValue('table');
      await flushPromises();
      const updated = resources.map(resource => ({ ...resource, title: `${resource.title}-new` }));
      await wrapper.setProps({ dataList: updated });
      await flushPromises();
      expect(wrapper.findComponent(ElTable).vm.getSelectionRows().map((item: I_resource) => item.id)).toEqual(['a']);
      await button(wrapper, '添加标签').trigger('click');
      expect(mocks.open).toHaveBeenLastCalledWith([updated[0]], 'add');
    } finally { wrapper.unmount(); }
  });

  it('快速切换视图并同时替换列表，选择状态最终一致', async () => {
    const wrapper = mountAdmin();
    try {
      await wrapper.get('[data-id="a"]').trigger('click');
      const switcher = wrapper.findComponent(ElRadioGroup);
      switcher.vm.$emit('update:modelValue', 'table');
      switcher.vm.$emit('update:modelValue', 'layout');
      switcher.vm.$emit('update:modelValue', 'table');
      const updated = [{ ...resources[0], title: 'latest' }, resources[1]];
      await wrapper.setProps({ dataList: updated });
      await flushPromises();
      const table = wrapper.findComponent(ElTable);
      expect(table.vm.getSelectionRows()).toEqual([updated[0]]);
      table.vm.toggleRowSelection(updated[0], false);
      await wrapper.setProps({ dataList: updated.map(resource => ({ ...resource })) });
      await flushPromises();
      expect(table.vm.getSelectionRows()).toEqual([]);
      expect(wrapper.get('.selection-count').text()).toContain('0');
    } finally { wrapper.unmount(); }
  });

  it('确认转码期间修改选择不会改变待提交资源，也不能重复提交', async () => {
    let confirm!: (value: Awaited<ReturnType<typeof ElMessageBox.confirm>>) => void;
    vi.spyOn(ElMessageBox, 'confirm').mockImplementation(() => new Promise(resolve => { confirm = resolve; }));
    mocks.addTranscode.mockResolvedValueOnce({ status: true, data: { added: 1, skippedDuplicate: 0, skippedMissing: 0 } });
    const wrapper = mountAdmin();
    try {
      await wrapper.get('[data-id="a"]').trigger('click');
      await button(wrapper, '加入转码与剪辑').trigger('click');
      await button(wrapper, '加入转码与剪辑').trigger('click');
      await wrapper.get('[data-id="b"]').trigger('click');
      await button(wrapper, '清空选择').trigger('click');
      confirm('confirm' as Awaited<ReturnType<typeof ElMessageBox.confirm>>);
      await flushPromises();
      expect(ElMessageBox.confirm).toHaveBeenCalledTimes(1);
      expect(mocks.addTranscode).toHaveBeenCalledExactlyOnceWith({ resourceIds: ['a'] });
    } finally { wrapper.unmount(); }
  });

  it('转码请求失败时提示错误，恢复按钮并允许重试', async () => {
    vi.spyOn(ElMessageBox, 'confirm').mockResolvedValue('confirm' as Awaited<ReturnType<typeof ElMessageBox.confirm>>);
    const error = vi.spyOn(ElMessage, 'error').mockImplementation(() => ({ close: vi.fn() }));
    mocks.addTranscode.mockRejectedValueOnce(new Error('network'));
    const wrapper = mountAdmin();
    try {
      await wrapper.get('[data-id="a"]').trigger('click');
      await button(wrapper, '加入转码与剪辑').trigger('click');
      await flushPromises();
      expect(mocks.addTranscode).toHaveBeenCalledWith({ resourceIds: ['a'] });
      expect(error).toHaveBeenCalledWith('批量加入转码与剪辑列表失败，请稍后重试');
      expect(button(wrapper, '加入转码与剪辑').classes()).not.toContain('is-loading');
      vi.mocked(ElMessageBox.confirm).mockRejectedValueOnce('cancel');
      await button(wrapper, '加入转码与剪辑').trigger('click');
      await flushPromises();
      expect(mocks.addTranscode).toHaveBeenCalledTimes(1);
      expect(error).toHaveBeenCalledTimes(1);
    } finally { wrapper.unmount(); }
  });

  it('全选只作用于本页，更新和翻页后不会保留不可见资源', async () => {
    const wrapper = mountAdmin();
    try {
      await wrapper.get('.management-selection input').setValue(true);
      expect(wrapper.get('.selection-count').text()).toContain('2');
      const updated = { ...resources[0], title: 'updated' };
      await wrapper.setProps({ dataList: [updated] });
      await button(wrapper, '添加标签').trigger('click');
      expect(mocks.open).toHaveBeenLastCalledWith([updated], 'add');
      await wrapper.setProps({ dataList: [resources[1]] });
      expect(wrapper.get('.selection-count').text()).toContain('0');
    } finally { wrapper.unmount(); }
  });
});

describe('当前布局的选择隔离', () => {
  it('管理时拦截卡片内播放和右键，支持键盘，浏览时恢复点击', async () => {
    const active = ref(true);
    const toggle = vi.fn();
    const wrapper = mount(ContentRightClickMenu, {
      props: { resource: resources[0] },
      slots: { default: defineComponent({ setup: () => ({ play: mocks.play }), template: '<button @click.stop="play">播放</button>' }) },
      global: { plugins: [ElementPlus], components: icons, provide: { [resourceSelectionKey as symbol]: {
        active: computed(() => active.value), selectedIds: computed(() => new Set(['a'])), toggle,
      } }, stubs: { contentMenu: { template: '<div><slot /></div>' } } },
    });
    try {
      expect(wrapper.attributes('aria-checked')).toBe('true');
      await wrapper.get('button').trigger('click');
      expect(toggle).toHaveBeenCalledWith(resources[0]);
      expect(mocks.play).not.toHaveBeenCalled();
      await wrapper.trigger('keydown', { key: ' ' });
      await wrapper.trigger('keydown', { key: ' ', repeat: true });
      expect(toggle).toHaveBeenCalledTimes(2);
      const event = new MouseEvent('contextmenu', { bubbles: true, cancelable: true });
      wrapper.element.dispatchEvent(event);
      expect(event.defaultPrevented).toBe(true);
      active.value = false;
      await flushPromises();
      await wrapper.get('button').trigger('click');
      expect(mocks.play).toHaveBeenCalledOnce();
      expect(wrapper.attributes('role')).toBeUndefined();
    } finally { wrapper.unmount(); }
  });

  it('普通表格布局中点击播放入口也只选择资源', async () => {
    const toggle = vi.fn();
    const wrapper = mount(LayoutTable, {
      props: { dataList: resources },
      global: { plugins: [ElementPlus], components: icons, mocks: { $t: () => '' }, provide: { [resourceSelectionKey as symbol]: {
        active: computed(() => true), selectedIds: computed(() => new Set<string>()), toggle,
      } } },
    });
    try {
      await flushPromises();
      await wrapper.findAll('.playDiv')[0].trigger('click');
      expect(toggle).toHaveBeenCalledWith(resources[0]);
      expect(mocks.play).not.toHaveBeenCalled();
      toggle.mockClear();
      await wrapper.findAll('input[type=checkbox]')[0].setValue(true);
      expect(toggle).toHaveBeenCalledExactlyOnceWith(resources[0]);
    } finally { wrapper.unmount(); }
  });
});

describe('管理状态下的短视频布局', () => {
  it.each([
    ['短视频', ShortVideo], ['上下短视频', ShortVideoTopBottom], ['拼贴短视频', MosaicShortVideo],
  ] as const)('%s 不初始化播放器或请求播放地址', async (_, component) => {
    const playable = resources.map(resource => ({ ...resource, dramaSeries: [{ id: `video-${resource.id}`, src: '/video.mp4' }] })) as I_resource[];
    const wrapper = shallowMount(component, {
      props: { dataList: playable },
      global: { plugins: [ElementPlus], components: icons, stubs: { ElScrollbar: defineComponent({ template: '<div><slot /></div>', methods: { update() {}, setScrollTop() {}, setScrollLeft() {} } }) }, provide: { [resourceSelectionKey as symbol]: {
        active: computed(() => true), selectedIds: computed(() => new Set<string>()), toggle: vi.fn(),
      } } },
    });
    try {
      await flushPromises();
      expect(wrapper.find('video-play-stub').exists()).toBe(false);
      expect(wrapper.find('.management-player-placeholder').exists()).toBe(true);
      await wrapper.setProps({ dataList: playable.map(resource => ({ ...resource })) });
      await flushPromises();
      expect(mocks.playUrl).not.toHaveBeenCalled();
    } finally { wrapper.unmount(); }
  });
});
