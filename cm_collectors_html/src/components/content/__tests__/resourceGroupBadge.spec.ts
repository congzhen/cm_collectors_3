import { computed, nextTick, ref } from 'vue';
import { mount } from '@vue/test-utils';
import { beforeEach, expect, it, vi } from 'vitest';
import Badge from '../contentVideoDurationBadge.vue';
import { resourceGroupContextKey } from '../resourceGroups';
import { appStoreData } from '@/storeData/app.storeData';
import type { I_resource } from '@/dataType/resource.dataType';
import { E_resourceDramaSeriesType } from '@/dataType/app.dataType';
vi.mock('@/storeData/app.storeData', async () => {
  const { reactive } = await import('vue');
  const store = reactive({ currentConfigApp: { showVideoDuration: true } });
  return { appStoreData: () => store };
});
const resource = {id:'a',associationGroupId:'g',mode:E_resourceDramaSeriesType.Movies,dramaSeries:[{durationSeconds:80}]} as I_resource;
beforeEach(() => { appStoreData().currentConfigApp.showVideoDuration = true; });
it('显示整组数量而非预览数量；关闭时长、非视频和未知时长仍保留组数量', async () => {
  const summaries = ref(new Map([['g',{count:18,preview:[resource]}]]));
  const w = mount(Badge,{props:{resource},global:{provide:{[resourceGroupContextKey as symbol]:{enabled:computed(()=>true),pageIds:computed(()=>new Set(['a'])),summaries}}}});
  try {
    expect(w.find('.group-count').text()).toBe('18 项');
    expect(w.find('.duration-text').text()).toBe('1:20');
    expect(w.find('.group-count').attributes('aria-label')).toBe('关联组共 18 个资源');
    appStoreData().currentConfigApp.showVideoDuration = false; await nextTick();
    expect(w.find('.group-count').exists()).toBe(true); expect(w.find('.badge-divider').exists()).toBe(false);
    appStoreData().currentConfigApp.showVideoDuration = true;
    await w.setProps({resource:{...resource,dramaSeries:[]}});
    expect(w.find('.group-count').exists()).toBe(true); expect(w.find('.duration-text').exists()).toBe(false);
    summaries.value = new Map([['g',{count:8,preview:[]}]]); await nextTick();
    expect(w.find('.group-count').text()).toBe('8 项');
  } finally { w.unmount(); }
});
it('平铺、管理或非主列表不显示组数量，普通资源继续沿用时长开关', async () => {
  const enabled = ref(true);
  const w = mount(Badge,{props:{resource},global:{provide:{[resourceGroupContextKey as symbol]:{enabled:computed(()=>enabled.value),pageIds:computed(()=>new Set(['a'])),summaries:ref(new Map([['g',{count:7}]]))}}}});
  try {
    enabled.value = false; await nextTick();
    expect(w.find('.group-count').exists()).toBe(false); expect(w.text()).toBe('1:20');
    appStoreData().currentConfigApp.showVideoDuration = false; await nextTick(); expect(w.find('.video-duration-badge').exists()).toBe(false);
  } finally { w.unmount(); }
  const plain = mount(Badge,{props:{resource}});
  expect(plain.find('.group-count').exists()).toBe(false); plain.unmount();
});
