import { computed, defineComponent, nextTick, reactive } from 'vue';
import { mount } from '@vue/test-utils';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import Popover from '../ResourceQuickStarsPopover.vue';
import { appStoreData } from '@/storeData/app.storeData';
import { resourceSelectionKey } from '../resourceSelection';
import type { I_resource } from '@/dataType/resource.dataType';
vi.mock('@/storeData/app.storeData', async () => {
  const { reactive } = await import('vue');
  const store = reactive({currentConfigApp:{resourceQuickStars:true},currentFilesBases:{id:'lib'},displayAdminFn:true});
  return {appStoreData:()=>store};
});
vi.mock('@/assets/mobile',()=>({isMobile:()=>false,isMobileDevice:()=>false}));
vi.mock('@/server/resource.server',()=>({resourceServer:{batchSetStars:vi.fn()}}));
const panelStub = defineComponent({name:'ElPopover',props:['visible','placement'],template:'<div><slot name="reference" /><slot v-if="visible" /></div>'});
const ratingStub = defineComponent({name:'ResourceQuickStars',props:['resource'],emits:['dragging'],template:'<div class="rating-editor" />'});
const resource = reactive({id:'a',filesBases_id:'lib',title:'资源 A',stars:3} as I_resource);
const mountPopover = (extra = {}) => mount(Popover,{props:{resource},global:{stubs:{ElPopover:panelStub,ResourceQuickStars:ratingStub},...extra}});
beforeEach(()=>{vi.useFakeTimers();appStoreData().currentConfigApp.resourceQuickStars=true;});
afterEach(()=>{vi.useRealTimers();});
it('默认仅有星级入口，悬停展开含资源名的面板；鼠标跨越间隙移入后保持展开，离开延迟关闭',async()=>{
  const w=mountPopover();
  try {
    expect(w.find('.rating-editor').exists()).toBe(false);
    expect(w.get('.quick-stars-trigger').text()).toBe('3');
    await w.trigger('pointerenter');await vi.advanceTimersByTimeAsync(150);
    expect(w.get('.quick-stars-resource').text()).toBe('资源 A');
    await w.trigger('pointerleave');await vi.advanceTimersByTimeAsync(100);
    await w.get('.quick-stars-panel').trigger('pointerenter');await vi.advanceTimersByTimeAsync(300);
    expect(w.find('.rating-editor').exists()).toBe(true);
    await w.get('.quick-stars-panel').trigger('pointerleave');await vi.advanceTimersByTimeAsync(250);
    expect(w.find('.rating-editor').exists()).toBe(false);
  } finally {w.unmount();}
});
it('点击和键盘可打开，Escape 关闭；同时只展开一个资源面板，关闭配置会收起',async()=>{
  const first=mountPopover();const second=mountPopover();
  try {
    await first.get('.quick-stars-trigger').trigger('click');
    expect(first.find('.rating-editor').exists()).toBe(true);
    await second.get('.quick-stars-trigger').trigger('click');await nextTick();
    expect(first.find('.rating-editor').exists()).toBe(false);
    await second.get('.quick-stars-panel').trigger('keydown',{key:'Escape'});
    expect(second.find('.rating-editor').exists()).toBe(false);
    const trigger=second.get('.quick-stars-trigger');
    vi.spyOn(trigger.element,'matches').mockReturnValue(true);
    await trigger.trigger('focusin');await vi.advanceTimersByTimeAsync(150);
    expect(second.find('.rating-editor').exists()).toBe(true);
    appStoreData().currentConfigApp.resourceQuickStars=false;await nextTick();
    expect(second.find('.quick-stars-trigger').exists()).toBe(false);
  } finally {first.unmount();second.unmount();}
});
it('拖动移出面板不提前收起，松手后延迟关闭；管理模式无入口',async()=>{
  const w=mountPopover();
  try {
    await w.get('.quick-stars-trigger').trigger('click');
    w.findComponent(ratingStub).vm.$emit('dragging',true);await nextTick();
    await w.get('.quick-stars-panel').trigger('pointerleave');await vi.advanceTimersByTimeAsync(300);
    expect(w.find('.rating-editor').exists()).toBe(true);
    w.findComponent(ratingStub).vm.$emit('dragging',false);await nextTick();await vi.advanceTimersByTimeAsync(250);
    expect(w.find('.rating-editor').exists()).toBe(false);
  } finally {w.unmount();}
  const management=mountPopover({provide:{[resourceSelectionKey as symbol]:{active:computed(()=>true)}}});
  expect(management.find('.quick-stars-trigger').exists()).toBe(false);management.unmount();
});
