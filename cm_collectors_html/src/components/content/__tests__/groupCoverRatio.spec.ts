import { mount, flushPromises } from '@vue/test-utils';
import { beforeEach, afterEach, expect, it, vi } from 'vitest';
import ElementPlus from 'element-plus';
import Fan from '../ResourceGroupFan.vue';
import Fold from '../ResourceGroupFold.vue';
import Magazine from '../ResourceGroupMagazine.vue';
import type { I_resource } from '@/dataType/resource.dataType';
vi.mock('@/common/photo', () => ({ getResourceCoverPoster: (r: I_resource) => `/cover/${r.id}` }));
beforeEach(() => vi.stubGlobal('ResizeObserver', class { observe() {} disconnect() {} }));
afterEach(() => vi.unstubAllGlobals());
const members = [
  { id: 'wide', title: '横版', coverPosterWidth: 1600, coverPosterHeight: 900 },
  { id: 'tall', title: '竖版', coverPosterWidth: 900, coverPosterHeight: 1600 },
] as I_resource[];
it.each([Fan, Fold, Magazine])('单资源默认沿用原布局，只有开启适配才调整比例', async component => {
  const w = mount(component, { props: { members: [members[1]], loading: false, error: '' }, global: { plugins: [ElementPlus] } });
  try {
    expect(w.find('.group-single').exists()).toBe(false);
    expect(w.findComponent({ name: 'ElImage' }).props('fit')).toBe('cover');
    await w.setProps({ adaptCoverRatio: true });
    expect(w.findComponent({ name: 'ElImage' }).props('fit')).toBe('contain');
    expect(w.classes()).toContain('is-adaptive');
    await w.find('button').trigger('click');
    expect(w.emitted('select')?.[0]).toEqual([members[1]]);
    await w.setProps({ adaptCoverRatio: false });
    expect(w.findComponent({ name: 'ElImage' }).props('fit')).toBe('cover');
  } finally { w.unmount(); }
});
it('混合封面根据实际加载尺寸修正宽度，关闭后恢复旧布局', async () => {
  const w = mount(Fan, { props: { members, loading: false, error: '', adaptCoverRatio: true }, global: { plugins: [ElementPlus] } });
  try {
    const cards = w.findAll('.fan-card');
    expect(cards[0].attributes('style')).not.toEqual(cards[1].attributes('style'));
    const loaded = new Event('load');
    Object.defineProperty(loaded, 'target', { value: { naturalWidth: 1200, naturalHeight: 1200 } });
    w.findComponent({ name: 'ElImage' }).vm.$emit('load', loaded);
    await flushPromises();
    expect(cards[0].attributes('style')).toContain('* 1)');
    await w.setProps({ adaptCoverRatio: false });
    expect((cards[0].element as HTMLElement).style.width).toBe('');
  } finally { w.unmount(); }
});
