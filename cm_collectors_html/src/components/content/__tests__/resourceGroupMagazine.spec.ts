import { mount } from '@vue/test-utils';
import { expect, it, vi } from 'vitest';
import ElementPlus from 'element-plus';
import Magazine from '../ResourceGroupMagazine.vue';
import type { I_resource } from '@/dataType/resource.dataType';
vi.mock('@/common/photo', () => ({ getResourceCoverPoster: (r: I_resource) => `/cover/${r.id}` }));
const resources = (count: number) => Array.from({ length: count }, (_, i) => ({ id: String(i), title: `资源 ${i}` } as I_resource));
const create = (count = 8) => mount(Magazine, { props: { members: resources(count), loading: false, error: '' }, global: { plugins: [ElementPlus], stubs: { ElImage: true } } });
it('当前资源成为主封面，不修改原始顺序；鼠标切换不会重排，点击打开对应资源', async () => {
  const w = create();
  try {
    const original = w.props('members').map(r => r.id);
    await w.setProps({ sourceResource: w.props('members')[3] });
    expect(w.find('.is-hero').attributes('aria-label')).toBe('打开资源：资源 3');
    const before = w.findAll('.magazine-tile').map(t => t.attributes('aria-label'));
    await w.findAll('.magazine-tile')[6].trigger('pointerenter');
    expect(w.findAll('.magazine-tile').map(t => t.attributes('aria-label'))).toEqual(before);
    await w.findAll('.magazine-tile')[6].trigger('click');
    expect(w.emitted('select')?.[0]).toEqual([w.props('members')[6]]);
    expect(w.props('members').map(r => r.id)).toEqual(original);
  } finally { w.unmount(); }
});
it('所有成员可达，键盘首尾切换只有一个可聚焦入口，并隔离播放器快捷键', async () => {
  const w = create(60);
  try {
    expect(w.findAll('.magazine-tile')).toHaveLength(60);
    const event = new KeyboardEvent('keydown', { key: 'End', bubbles: true, cancelable: true });
    const parent = vi.fn();
    w.element.addEventListener('keydown', parent);
    w.find('.magazine-scroll').element.dispatchEvent(event);
    await w.vm.$nextTick();
    expect(event.defaultPrevented).toBe(true);
    expect(parent).not.toHaveBeenCalled();
    expect(w.findAll('[tabindex="0"]')).toHaveLength(1);
    expect(w.find('[tabindex="0"]').attributes('aria-label')).toBe('打开资源：资源 59');
    await w.find('.magazine-scroll').trigger('keydown', { key: 'Home' });
    expect(w.find('[tabindex="0"]').attributes('aria-label')).toBe('打开资源：资源 0');
  } finally { w.unmount(); }
});
it('加载与失败不保留旧封面，失败可以重试，空组明确提示', async () => {
  const w = create();
  try {
    await w.setProps({ loading: true });
    expect(w.findAll('.magazine-tile')).toHaveLength(0);
    await w.setProps({ loading: false, error: '加载失败' });
    await w.find('button').trigger('click');
    expect(w.emitted('retry')).toHaveLength(1);
    await w.setProps({ error: '', members: [] });
    expect(w.text()).toContain('暂无关联资源');
  } finally { w.unmount(); }
});
