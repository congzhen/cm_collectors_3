import { mount } from '@vue/test-utils';
import { afterEach, expect, it, vi } from 'vitest';
import { nextTick } from 'vue';
import Fan from '../ResourceGroupFan.vue';
import Magazine from '../ResourceGroupMagazine.vue';
import type { I_resource } from '@/dataType/resource.dataType';
vi.mock('@/common/photo', () => ({ getResourceCoverPoster: (r: I_resource) => `/cover/${r.id}` }));
afterEach(() => vi.unstubAllGlobals());
it.each(['fan', 'strip', 'magazine'] as const)('%s 跟随来源封面尺寸并在关闭自适应后恢复默认', async mode => {
  const source = document.createElement('div');
  let height = 180;
  let resize: (() => void) | undefined;
  vi.spyOn(source, 'getBoundingClientRect').mockImplementation(() => ({ height } as DOMRect));
  vi.stubGlobal('ResizeObserver', class { constructor(callback: () => void) { resize = callback; } observe() {} disconnect() {} });
  const w = mount(mode === 'magazine' ? Magazine : Fan, {
    props: { sourceElement: source, adaptCoverRatio: true, members: [{id:'1',title:'封面',coverPosterWidth:900,coverPosterHeight:1600} as I_resource], loading:false, error:'', expandStyle: mode === 'strip' ? 'strip' : 'fan' },
    global: { stubs: { ElImage:true, ElIcon:true, ElButton:true } },
  });
  try {
    const heightProperty = mode === 'magazine' ? '--magazine-source-height' : '--fan-card-height';
    expect(parseFloat((w.element as HTMLElement).style.getPropertyValue(heightProperty))).toBeCloseTo(162);
    height = 300; resize?.(); await nextTick();
    expect(parseFloat((w.element as HTMLElement).style.getPropertyValue(heightProperty))).toBeCloseTo(270);
    expect(Number(w.emitted('layoutWidth')?.at(-1)?.[0])).toBeGreaterThan(0);
    if (mode !== 'magazine') {
      await w.setProps({members: [0,1].map(id => ({id:String(id),title:'窄封面',coverPosterWidth:100,coverPosterHeight:1000} as I_resource))});
      // Layout/hit zones must use the real narrow width, not a minimum 2:3 card.
      expect(parseFloat((w.element as HTMLElement).style.getPropertyValue('--fan-layout-width'))).toBeCloseTo(27);
    }
    await w.setProps({adaptCoverRatio:false});
    expect(w.emitted('layoutWidth')?.at(-1)).toEqual([undefined]);
    expect((w.element as HTMLElement).style.getPropertyValue(heightProperty)).toBe('');
  } finally { w.unmount(); }
});
