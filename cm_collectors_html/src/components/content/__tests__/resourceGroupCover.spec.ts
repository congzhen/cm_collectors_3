import { computed, nextTick, ref } from 'vue';
import { mount } from '@vue/test-utils';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import ElementPlus from 'element-plus';
import Cover from '../ResourceGroupCover.vue';
import { resourceGroupContextKey } from '../resourceGroups';
import { appStoreData } from '@/storeData/app.storeData';
import type { I_resource } from '@/dataType/resource.dataType';
vi.mock('@/common/photo', () => ({ getResourceCoverPoster: (r: I_resource) => `/cover/${r.id}` }));
vi.mock('@/storeData/app.storeData', async () => {
  const { reactive } = await import('vue');
  const store = reactive({ currentConfigApp: { resourceGroupCoverStyle: 'stack' } });
  return { appStoreData: () => store };
});
const a = { id: 'a', associationGroupId: 'g' } as I_resource;
const b = { id: 'b' } as I_resource;
const c = { id: 'c' } as I_resource;
beforeEach(() => { appStoreData().currentConfigApp.resourceGroupCoverStyle = 'stack'; });
afterEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals(); });
it.each(['triptych', 'book'] as const)('%s 只在本页合并组使用，排除当前封面，保留原加载事件且不复制业务内容', async style => {
  const enabled = ref(true);
  const summaries = ref(new Map([['g', { id: 'g', filesBasesId: 'lib', count: 3, preview: [b, a, c] }]]));
  const w = mount(Cover, { props: { resource: a, src: '/cover/a', fit: 'contain' }, global: { plugins: [ElementPlus], provide: {
    [resourceGroupContextKey as symbol]: { enabled: computed(() => enabled.value), pageIds: computed(() => new Set(['a'])), summaries },
  } } });
  try {
    expect(w.find('.group-cover').exists()).toBe(false);
    appStoreData().currentConfigApp.resourceGroupCoverStyle = style;
    await nextTick();
    expect(w.find(`.group-cover--${style}`).exists()).toBe(true);
    expect(w.findAll('.group-cover-leaf').map(i => i.findComponent({ name: 'ElImage' }).props('src'))).toEqual(['/cover/b', '/cover/c']);
    w.findComponent({ name: 'ElImage' }).vm.$emit('load', new Event('load'));
    expect(w.emitted('load')).toHaveLength(1);
    summaries.value = new Map([['g', { id: 'g', filesBasesId: 'lib', count: 2, preview: [a, b] }]]);
    await nextTick();
    expect(w.findAll('.group-cover-leaf')).toHaveLength(1);
    enabled.value = false; await nextTick();
    expect(w.find('.group-cover').exists()).toBe(false);
    expect(w.findComponent({ name: 'ElImage' }).props('fit')).toBe('contain');
  } finally { w.unmount(); }
});
it('旧配置和推荐区域保持普通封面；摘要未到达时使用当前封面预留切片', async () => {
  appStoreData().currentConfigApp.resourceGroupCoverStyle = 'triptych';
  const plain = mount(Cover, { props: { resource: a, src: '/a' }, global: { plugins: [ElementPlus] } });
  expect(plain.find('.group-cover').exists()).toBe(false); plain.unmount();
  const grouped = mount(Cover, { props: { resource: a, src: '/a' }, global: { plugins: [ElementPlus], provide: {
    [resourceGroupContextKey as symbol]: { enabled: computed(() => true), pageIds: computed(() => new Set(['a'])), summaries: ref(new Map()) },
  } } });
  expect(grouped.findAll('.group-cover-leaf')).toHaveLength(1); grouped.unmount();
});
it.each([2,3,4,5,6,7,12])('自适应拼封 %s 个成员按顺序展示，最多七张，当前代表不必位于摘要内', async count => {
  appStoreData().currentConfigApp.resourceGroupCoverStyle = 'adaptive';
  const members = Array.from({ length: count - 1 }, (_, i) => ({ id: String(i) } as I_resource));
  const w = mount(Cover, { props: { resource: a, src: '/a' }, global: { plugins: [ElementPlus], provide: {
    [resourceGroupContextKey as symbol]: { enabled: computed(() => true), pageIds: computed(() => new Set(['a'])), summaries: ref(new Map([['g', { count, preview: members.slice(0,7) }]])) },
  } } });
  try {
    await nextTick(); await nextTick();
    expect(w.findAll('.group-cover-leaf')).toHaveLength(Math.min(count - 1, 6));
    expect(w.classes()).toContain(`group-cover--count-${Math.min(count,7)}`);
    expect(w.findAll('.group-cover-leaf').map(i => i.findComponent({name:'ElImage'}).props('src'))).toEqual(members.slice(0,6).map(r => `/cover/${r.id}`));
  } finally { w.unmount(); }
});
it.each([6,7,10,30])('%s 个成员在紧凑、矮宽及默认封面中保留对应构图，不随尺寸降为五张', async count => {
  const preview = [a, ...Array.from({length:Math.min(count-1,6)},(_,i)=>({id:String(i)} as I_resource))];
  appStoreData().currentConfigApp.resourceGroupCoverStyle = 'adaptive';
  const w = mount(Cover, { props: { resource: a, src: '/a' }, global: { plugins: [ElementPlus], provide: {
    [resourceGroupContextKey as symbol]: { enabled: computed(() => true), pageIds: computed(() => new Set(['a'])), summaries: ref(new Map([['g', { count, preview }]])) },
  } } });
  try {
    for (const [width,height] of [[316,218],[218,316],[180,140],[120,100]]) {
      w.element.setAttribute('style',`width:${width}px;height:${height}px`); await nextTick();
      expect(w.findAll('.group-cover-leaf')).toHaveLength(Math.min(count-1,6));
      expect(w.classes()).toContain(`group-cover--count-${Math.min(count,7)}`);
    }
    appStoreData().currentConfigApp.resourceGroupCoverStyle = 'triptych'; await nextTick();
    expect(w.findAll('.group-cover-leaf')).toHaveLength(2);
  } finally { w.unmount(); }
});
