import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import ElementPlus, { ElInputNumber, ElSelect } from 'element-plus'
import Quick from '../fileDatabaseSetting/CategoryBrowseQuickSettings.vue'
import { appStoreData } from '@/storeData/app.storeData'
import { createDefaultConfigApp } from '@/dataType/config.dataType'
const mocks = vi.hoisted(() => ({ save: vi.fn(), tags: vi.fn(), emit: vi.fn() }))
vi.mock('@/main', () => ({ eventBus: { emit: mocks.emit } }))
vi.mock('@/assets/request', () => ({ default: vi.fn() }))
vi.mock('@/server/tag.server', () => ({ tagServer: { tagClassListByFilesBasesId: mocks.tags } }))
vi.mock('@/server/categoryBrowse.server', async () => {
  const original = await vi.importActual<typeof import('@/server/categoryBrowse.server')>(
    '@/server/categoryBrowse.server',
  )
  return { ...original, categoryBrowseServer: { settings: mocks.save } }
})
vi.mock('@/storeData/app.storeData', async () => {
  const { reactive } = await import('vue')
  const store = reactive({ currentFilesBases: { id: 'A' }, currentConfigApp: {} })
  return { appStoreData: () => store }
})
beforeEach(() => {
  vi.restoreAllMocks(); vi.clearAllMocks();
  mocks.tags.mockResolvedValue({ status: true, data: [] });
  appStoreData().currentFilesBases.id = 'A';
  appStoreData().currentConfigApp = createDefaultConfigApp();
});
const create = () => mount(Quick, { props: { horizontal: true }, global: { plugins: [ElementPlus] } });
it('选择立即预览，不显示应用按钮且不保存，重新加载恢复持久化设置', async () => {
  const persisted = createDefaultConfigApp();
  const w = create();
  try {
    await flushPromises();
    expect(w.find('.category-apply').exists()).toBe(false);
    await w.get('[role="switch"]').trigger('click');
    await flushPromises();
    expect(appStoreData().currentConfigApp.categoryBrowseEnabled).toBe(true);
    w.findAll('.category-field')[2]!.findComponent(ElSelect).vm.$emit('update:modelValue', 'name');
    await flushPromises();
    expect(appStoreData().currentConfigApp.categoryBrowseSort).toBe('name');
    expect(mocks.emit).toHaveBeenCalledWith('category-browse-applied');
    expect(mocks.save).not.toHaveBeenCalled();
    appStoreData().currentConfigApp = { ...persisted }; await flushPromises();
    expect(w.findAll('.category-field')[2]!.findComponent(ElSelect).props('modelValue')).toBe('count');
    expect(appStoreData().currentConfigApp.categoryBrowseEnabled).toBe(false);
  } finally { w.unmount(); }
  expect(mocks.save).not.toHaveBeenCalled();
});
it.each([0, 1, 100])('分组上限 %s 只修改当前显示，不调用持久化接口', async cap => {
  appStoreData().currentConfigApp.categoryBrowseEnabled = true;
  const w = create();
  try {
    await flushPromises();
    w.findComponent(ElInputNumber).vm.$emit('update:modelValue', cap); await flushPromises();
    expect(appStoreData().currentConfigApp.categoryBrowseGroupLimit).toBe(cap);
    expect(mocks.save).not.toHaveBeenCalled();
  } finally { w.unmount(); }
});
it('无效输入不会进入当前配置，修正后立即预览', async () => {
  appStoreData().currentConfigApp.categoryBrowseEnabled = true;
  const w = create();
  try {
    await flushPromises();
    w.findComponent(ElInputNumber).vm.$emit('update:modelValue', undefined); await flushPromises();
    expect(appStoreData().currentConfigApp.categoryBrowseGroupLimit).toBe(12);
    expect(w.text()).toContain('每组资源上限请输入');
    expect(w.vm.validate()).toBe(false);
    w.findComponent(ElInputNumber).vm.$emit('update:modelValue', 8); await flushPromises();
    expect(appStoreData().currentConfigApp.categoryBrowseGroupLimit).toBe(8);
    expect(w.vm.validate()).toBe(true);
    expect(mocks.save).not.toHaveBeenCalled();
  } finally { w.unmount(); }
});
it('切库丢弃无效草稿，不影响新库设置', async () => {
  appStoreData().currentConfigApp.categoryBrowseEnabled = true;
  const w = create();
  try {
    await flushPromises();
    w.findComponent(ElInputNumber).vm.$emit('update:modelValue', undefined); await flushPromises();
    appStoreData().currentFilesBases.id = 'B';
    appStoreData().currentConfigApp = { ...createDefaultConfigApp(), categoryBrowseGroupLimit: 5 };
    await flushPromises();
    expect(w.findComponent(ElInputNumber).props('modelValue')).toBe(5);
    expect(mocks.save).not.toHaveBeenCalled();
  } finally { w.unmount(); }
});
it('标签加载完成后继续预览待选设置，不自动保存', async () => {
  let finish!: (value: unknown) => void;
  mocks.tags.mockReturnValueOnce(new Promise(resolve => { finish = resolve }));
  Object.assign(appStoreData().currentConfigApp, { categoryBrowseEnabled: true, categoryBrowseSource: 'tag:type' });
  const w = create();
  try {
    w.findAll('.category-field')[2]!.findComponent(ElSelect).vm.$emit('update:modelValue', 'name'); await flushPromises();
    expect(appStoreData().currentConfigApp.categoryBrowseSort).toBe('count');
    finish({ status: true, data: [{ id: 'type', name: '类型', status: true }] }); await flushPromises();
    expect(appStoreData().currentConfigApp.categoryBrowseSort).toBe('name');
    expect(mocks.save).not.toHaveBeenCalled();
  } finally { w.unmount(); }
});

it('仅封面下拉框提供单资源开关，只即时预览不保存', async () => {
  appStoreData().currentConfigApp.categoryBrowseEnabled = true;
  const w = create();
  try {
    await flushPromises();
    const cover = w.findComponent({ name: 'selectResourceGroupCoverStyle' });
    const expand = w.findComponent({ name: 'selectResourceGroupExpandStyle' });
    expect(cover.props('hideSingle')).toBe(false);
    cover.vm.$emit('update:hideSingle', true); await flushPromises();
    expect(appStoreData().currentConfigApp.categoryBrowseHideSingle).toBe(true);
    expect(expand.props()).not.toHaveProperty('hideSingle');
    cover.vm.$emit('update:hideSingle', false); await flushPromises();
    expect(cover.props('hideSingle')).toBe(false);
    expect(mocks.save).not.toHaveBeenCalled();
  } finally { w.unmount(); }
});
