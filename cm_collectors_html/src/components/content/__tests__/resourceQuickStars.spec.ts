import { defineComponent, nextTick, reactive } from 'vue';
import { mount, flushPromises } from '@vue/test-utils';
import { beforeEach, expect, it, vi } from 'vitest';
import ResourceQuickStars from '../ResourceQuickStars.vue';
import { trackResourceStars, pendingResourceStars } from '@/common/resourceStars';
import type { I_resource } from '@/dataType/resource.dataType';
import { resourceSelectionKey } from '../resourceSelection';
import { useQuickResourceStars } from '@/common/resourceStars';
import { computed, ref } from 'vue';

const mocks = vi.hoisted(() => ({ save: vi.fn(), error: vi.fn(), mobile: false }));
vi.mock('@/server/resource.server', () => ({ resourceServer: { batchSetStars: mocks.save } }));
vi.mock('element-plus', () => ({ ElMessage: { error: mocks.error } }));
vi.mock('@/assets/mobile', () => ({ isMobile: () => mocks.mobile, isMobileDevice: () => mocks.mobile }));
vi.mock('@/storeData/app.storeData', async () => {
  const { reactive } = await import('vue');
  const store = reactive({ currentConfigApp: { resourceQuickStars: true }, currentFilesBases: { id: 'lib' }, displayAdminFn: true });
  return { appStoreData: () => store };
});
import { appStoreData } from '@/storeData/app.storeData';
const testStore = appStoreData() as unknown as { currentConfigApp: { resourceQuickStars: boolean }; currentFilesBases: { id: string }; displayAdminFn: boolean };
const resource = (id = 'a', library = 'lib') => reactive({ id, filesBases_id: library, stars: 2 } as I_resource);
const mountStars = (item = resource()) => {
  const wrapper = mount(ResourceQuickStars, { props: { resource: item } });
  const element = wrapper.find('.quick-stars-track').exists() ? wrapper.get('.quick-stars-track').element : undefined;
  if (element) {
    vi.spyOn(element, 'getBoundingClientRect').mockReturnValue({ left: 0, width: 100 } as DOMRect);
    Object.assign(element, { setPointerCapture: vi.fn(), hasPointerCapture: () => false });
  }
  return wrapper;
};
beforeEach(() => {
  mocks.save.mockReset().mockResolvedValue({ status: true });
  mocks.error.mockReset(); mocks.mobile = false;
  pendingResourceStars.clear();
  testStore.currentConfigApp.resourceQuickStars = true;
  testStore.displayAdminFn = true;
  testStore.currentFilesBases.id = 'lib';
});

it('悬停只预览，鼠标拖动松手仅保存最终星级，不冒泡打开资源', async () => {
  const item = resource();
  const click = vi.fn();
  const w = mountStars(item);
  const parent = document.createElement('div');
  parent.appendChild(w.element);
  parent.addEventListener('click', click);
  try {
    const track = w.get('.quick-stars-track');
    await track.trigger('pointermove', { pointerType: 'mouse', clientX: 90 });
    expect(w.findAll('.filled')).toHaveLength(5);
    expect(mocks.save).not.toHaveBeenCalled();
    await track.trigger('pointerleave');
    expect(w.findAll('.filled')).toHaveLength(2);
    await track.trigger('pointerdown', { pointerType: 'mouse', pointerId: 1, button: 0, clientX: 10 });
    await track.trigger('pointermove', { pointerType: 'mouse', pointerId: 1, clientX: 75 });
    expect(mocks.save).not.toHaveBeenCalled();
    await track.trigger('pointerup', { pointerType: 'mouse', pointerId: 1, clientX: 75 });
    w.get('[aria-label="评为 4 星"]').element.dispatchEvent(new MouseEvent('click', { bubbles: true, detail: 1 }));
    await flushPromises();
    expect(mocks.save).toHaveBeenCalledExactlyOnceWith(['a'], 4);
    expect(item.stars).toBe(4);
    expect(click).not.toHaveBeenCalled();
  } finally { w.unmount(); }
});

it('取消手势、触摸和非左键都不保存', async () => {
  const w = mountStars();
  try {
    const track = w.get('.quick-stars-track');
    await track.trigger('pointerdown', { pointerType: 'mouse', pointerId: 1, button: 0, clientX: 90 });
    await track.trigger('pointercancel', { pointerId: 1 });
    await track.trigger('pointerup', { pointerId: 1, clientX: 90 });
    for (const [pointerType, button] of [['touch', 0], ['mouse', 2]] as const) {
      await track.trigger('pointerdown', { pointerType, pointerId: 2, button, clientX: 90 });
      await track.trigger('pointerup', { pointerType, pointerId: 2, clientX: 90 });
    }
    expect(mocks.save).not.toHaveBeenCalled();
    expect(w.findAll('.filled')).toHaveLength(2);
  } finally { w.unmount(); }
});

it('键盘可选星及清空；失败恢复原星级并提示', async () => {
  const item = resource();
  const w = mountStars(item);
  try {
    mocks.save.mockResolvedValueOnce({ status: false, msg: '保存失败' });
    await w.get('[aria-label="评为 5 星"]').trigger('click');
    await flushPromises();
    expect(item.stars).toBe(2);
    expect(w.findAll('.filled')).toHaveLength(2);
    expect(mocks.error).toHaveBeenCalledWith('保存失败');
    await w.get('[aria-label="清除评星"]').trigger('click');
    await flushPromises();
    expect(item.stars).toBe(0);
    expect(mocks.save).toHaveBeenLastCalledWith(['a'], 0);
  } finally { w.unmount(); }
});

it('同资源多处显示同步且阻止并发保存，其他文件库保持原值', async () => {
  const item = resource();
  const copy = resource();
  const otherLibrary = resource('a', 'other');
  const observer = mount(defineComponent({ setup() { trackResourceStars(() => [copy, otherLibrary]); return () => null; } }));
  const first = mountStars(item);
  const second = mountStars(copy);
  let resolve!: (result: { status: boolean }) => void;
  mocks.save.mockReturnValueOnce(new Promise(done => { resolve = done; }));
  try {
    await first.get('[aria-label="评为 4 星"]').trigger('click');
    await second.get('[aria-label="评为 5 星"]').trigger('click');
    expect(mocks.save).toHaveBeenCalledTimes(1);
    resolve({ status: true }); await flushPromises();
    expect(item.stars).toBe(4); expect(copy.stars).toBe(4); expect(otherLibrary.stars).toBe(2);
  } finally { first.unmount(); second.unmount(); observer.unmount(); }
});

it('未开启及手机隐藏入口，访客只读，权限丢失后不能保存', async () => {
  testStore.currentConfigApp.resourceQuickStars = false;
  const disabled = mountStars(); expect(disabled.find('.resource-quick-stars').exists()).toBe(false); disabled.unmount();
  testStore.currentConfigApp.resourceQuickStars = true;
  mocks.mobile = true;
  const mobile = mountStars(); expect(mobile.find('.resource-quick-stars').exists()).toBe(false); mobile.unmount();
  mocks.mobile = false;
  const w = mountStars();
  try {
    await w.get('.quick-stars-track').trigger('pointerdown', { pointerType: 'mouse', pointerId: 1, button: 0, clientX: 90 });
    testStore.displayAdminFn = false; await nextTick();
    await w.get('.quick-stars-track').trigger('pointerup', { pointerId: 1, clientX: 90 });
    expect(w.findAll('button:disabled')).toHaveLength(5);
    expect(w.find('[aria-label="清除评星"]').exists()).toBe(false);
    expect(mocks.save).not.toHaveBeenCalled();
  } finally { w.unmount(); }
});

it('保存中释放指针仍显示目标星级；切库后旧请求不改动新库的预览或星级', async () => {
  const oldResource = resource();
  const newResource = resource('a', 'new-library');
  const w = mountStars(oldResource);
  const resolvers: Array<(result: { status: boolean }) => void> = [];
  mocks.save.mockImplementation(() => new Promise(resolve => resolvers.push(resolve)));
  try {
    const track = w.get('.quick-stars-track');
    await track.trigger('pointerdown', { pointerType: 'mouse', pointerId: 1, button: 0, clientX: 90 });
    await track.trigger('pointerup', { pointerType: 'mouse', pointerId: 1, clientX: 90 });
    await track.trigger('lostpointercapture', { pointerId: 1 });
    await track.trigger('pointerleave');
    expect(w.findAll('.filled')).toHaveLength(5);
    testStore.currentFilesBases.id = 'new-library';
    await w.setProps({ resource: newResource });
    expect(w.findAll('.filled')).toHaveLength(2);
    await w.get('[aria-label="评为 4 星"]').trigger('click');
    resolvers[0]({ status: true }); await flushPromises();
    expect(oldResource.stars).toBe(5); expect(newResource.stars).toBe(2);
    expect(w.findAll('.filled')).toHaveLength(4);
    resolvers[1]({ status: true }); await flushPromises();
    expect(newResource.stars).toBe(4);
  } finally { w.unmount(); }
});

it('管理模式统一关闭快捷评星分支，使表格和宽图卡片恢复只读星级', async () => {
  const active = ref(false);
  const w = mount(defineComponent({
    setup() { const { enabled } = useQuickResourceStars(); return { enabled }; },
    template: '<span>{{ enabled ? "quick" : "readonly" }}</span>',
  }), { global: { provide: { [resourceSelectionKey as symbol]: { active: computed(() => active.value) } } } });
  try {
    expect(w.text()).toBe('quick');
    active.value = true; await nextTick(); expect(w.text()).toBe('readonly');
    active.value = false; await nextTick(); expect(w.text()).toBe('quick');
  } finally { w.unmount(); }
});

it('拖动中窗口失焦或按 Escape 取消；切库旧列表不能继续写入', async () => {
  const w = mountStars();
  try {
    const track = w.get('.quick-stars-track');
    for (const action of ['blur', 'escape']) {
      await track.trigger('pointerdown', { pointerType: 'mouse', pointerId: 1, button: 0, clientX: 90 });
      if (action === 'blur') window.dispatchEvent(new Event('blur'));
      else await w.get('.resource-quick-stars').trigger('keydown', { key: 'Escape' });
      await track.trigger('pointerup', { pointerType: 'mouse', pointerId: 1, clientX: 90 });
      expect(w.findAll('.filled')).toHaveLength(2);
    }
    testStore.currentFilesBases.id = 'new-library'; await nextTick();
    await w.get('[aria-label="评为 5 星"]').trigger('click');
    expect(mocks.save).not.toHaveBeenCalled();
  } finally { w.unmount(); }
});

it('拖动后控件被隐藏时不提交无效星级，网络异常保留原值并可重试', async () => {
  const item = resource();
  const w = mountStars(item);
  try {
    const track = w.get('.quick-stars-track');
    await track.trigger('pointerdown', { pointerType: 'mouse', pointerId: 1, button: 0, clientX: 90 });
    vi.mocked(track.element.getBoundingClientRect).mockReturnValueOnce({ left: 0, width: 0 } as DOMRect);
    await track.trigger('pointerup', { pointerType: 'mouse', pointerId: 1, clientX: 90 });
    expect(mocks.save).not.toHaveBeenCalled();
    mocks.save.mockRejectedValueOnce(new Error('网络连接失败'));
    await w.get('[aria-label="评为 5 星"]').trigger('click'); await flushPromises();
    expect(item.stars).toBe(2); expect(pendingResourceStars.size).toBe(0);
    expect(mocks.error).toHaveBeenCalledWith('网络连接失败');
    await w.get('[aria-label="评为 5 星"]').trigger('click'); await flushPromises();
    expect(item.stars).toBe(5);
  } finally { w.unmount(); }
});

it('保存中翻页卸载旧控件，新挂载副本能收到结果，已卸载副本不再同步', async () => {
  const original = resource();
  const disposedCopy = resource();
  const observer = mount(defineComponent({ setup() { trackResourceStars(() => [disposedCopy]); return () => null; } }));
  const first = mountStars(original);
  let resolve!: (result: { status: boolean }) => void;
  mocks.save.mockReturnValueOnce(new Promise(done => { resolve = done; }));
  await first.get('[aria-label="评为 4 星"]').trigger('click');
  first.unmount(); observer.unmount();
  const current = resource();
  const second = mountStars(current);
  try {
    expect(second.findAll('button:disabled')).toHaveLength(6);
    resolve({ status: true }); await flushPromises();
    expect(current.stars).toBe(4);
    expect(disposedCopy.stars).toBe(2);
    expect(pendingResourceStars.size).toBe(0);
    expect(second.get('[aria-label="评为 5 星"]').attributes('disabled')).toBeUndefined();
  } finally { second.unmount(); }
});

it.each(['切库', '卸载'])('旧保存请求在%s后失败，不在当前页面弹出过期错误', async (action) => {
  const w = mountStars();
  let reject!: (error: Error) => void;
  mocks.save.mockReturnValueOnce(new Promise((_resolve, fail) => { reject = fail; }));
  await w.get('[aria-label="评为 4 星"]').trigger('click');
  if (action === '卸载') w.unmount();
  else { testStore.currentFilesBases.id = 'new-library'; await nextTick(); }
  try {
    reject(new Error('旧请求失败')); await flushPromises();
    expect(mocks.error).not.toHaveBeenCalled();
    expect(pendingResourceStars.size).toBe(0);
  } finally { if (action !== '卸载') w.unmount(); }
});

it.each([true, false])('同资源其他位置保存中，离开后不残留悬停预览（保存成功=%s）', async (success) => {
  const first = mountStars(resource());
  const second = mountStars(resource());
  let resolve!: (result: { status: boolean }) => void;
  mocks.save.mockReturnValueOnce(new Promise(done => { resolve = done; }));
  try {
    await second.get('.quick-stars-track').trigger('pointermove', { pointerType: 'mouse', clientX: 90 });
    expect(second.findAll('.filled')).toHaveLength(5);
    await first.get('[aria-label="评为 4 星"]').trigger('click');
    await second.get('.quick-stars-track').trigger('pointerleave');
    resolve({ status: success }); await flushPromises();
    expect(first.findAll('.filled')).toHaveLength(success ? 4 : 2);
    expect(second.findAll('.filled')).toHaveLength(success ? 4 : 2);
  } finally { first.unmount(); second.unmount(); }
});
