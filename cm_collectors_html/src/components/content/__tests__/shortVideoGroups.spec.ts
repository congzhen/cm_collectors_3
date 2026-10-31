import { defineComponent, inject } from 'vue';
import { flushPromises, mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import * as icons from '@element-plus/icons-vue';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import ShortVideo from '../layoutShortVideo.vue';
import TopBottom from '../layoutShortVideoTopBottom.vue';
import MosaicVideo from '../layoutCoverPosterMosaicShortVideo.vue';
import Mosaic from '../layoutCoverPosterMosaic.vue';
import { resourceGroupKeepOpenKey, resourceGroupSelectKey } from '../resourceGroups';
import type { I_resource } from '@/dataType/resource.dataType';
import { appStoreData } from '@/storeData/app.storeData';
const mocks=vi.hoisted(()=>({url:vi.fn(),source:vi.fn(),scroll:vi.fn(),pause:vi.fn()}));
vi.mock('@/main',()=>({eventBus:{emit:vi.fn()}}));
vi.mock('@/common/photo',()=>({getResourceCoverPoster:()=>''}));
vi.mock('@/common/play',()=>({getPlayVideoURLAndType:mocks.url,playResource:vi.fn()}));
vi.mock('@/storeData/app.storeData',async()=>{
  const {reactive}=await import('vue');
  const store=reactive({currentFilesBases:{id:'lib'},currentConfigApp:{coverPosterGap:4}});
  return {appStoreData:()=>store};
});
vi.mock('@/server/resourceGroup.server',()=>({resourceGroupServer:{}}));
const resource=(id:string)=>({id,title:id,filesBases_id:'lib',dramaSeries:[{id:`video-${id}`,src:`/${id}.mp4`}],tags:[],performers:[]} as unknown as I_resource);
const a=resource('a'),b=resource('b'),c=resource('c');
let groupAction:((resource:I_resource)=>void)|undefined;
let keepOpenAfterSelect=false;
const MenuStub=defineComponent({setup(){groupAction=inject(resourceGroupSelectKey);keepOpenAfterSelect=inject(resourceGroupKeepOpenKey,false);},template:'<div class="visible-resource"><slot /></div>'});
const VideoStub=defineComponent({setup(_, {expose}){expose({pause:mocks.pause,isPlaying:()=>false,setVideoSource:mocks.source,addTextTrack:vi.fn(),getControllerHeight:()=>0,setAspectRatio:vi.fn(),resetPlayer:vi.fn()});},template:'<div />'});
const ScrollbarStub=defineComponent({template:'<div><slot /></div>',methods:{update(){},setScrollLeft(){},setScrollTop(){}}});
const WaterfallStub=defineComponent({props:['list'],template:'<div><slot v-for="(item,index) in list" :item="item" :index="index" /></div>',methods:{renderer(){}}});
const MosaicStub=defineComponent({props:['dataList'],emits:['playResource'],setup(){keepOpenAfterSelect=inject(resourceGroupKeepOpenKey,false);},template:'<div class="mosaic-stub" />',methods:{scrollToResource:mocks.scroll,change(){}}});
beforeEach(()=>{
  vi.clearAllMocks(); groupAction=undefined;keepOpenAfterSelect=false;
  appStoreData().currentFilesBases.id='lib';
  vi.stubGlobal('ResizeObserver',class{observe(){}disconnect(){}});
  mocks.url.mockImplementation(async(id:string)=>({playUrl:`/${id}.mp4`,playType:'video/mp4'}));
});

it.each([['短视频',ShortVideo],['上下短视频',TopBottom],['拼贴短视频',MosaicVideo]] as const)('%s 选择页外成员后切库，旧页尚未清空也不能重新播放旧库资源',async(_,component)=>{
  const wrapper=mount(component,{props:{dataList:[a]},global:{plugins:[ElementPlus],components:icons,stubs:{contentRightClickMenu:MenuStub,videoPlay:VideoStub,ElScrollbar:ScrollbarStub,Waterfall:WaterfallStub,layoutCoverPosterMosaic:MosaicStub,ElImage:true,ElBacktop:true}}});
  try {
    await flushPromises();
    if(component===MosaicVideo)wrapper.findComponent(Mosaic).vm.$emit('playResource',b);
    else groupAction!(b);
    await flushPromises();
    expect(mocks.source.mock.calls.at(-1)?.[0]).toBe('/video-b.mp4');
    mocks.url.mockClear();mocks.source.mockClear();
    appStoreData().currentFilesBases.id='other';
    await flushPromises();
    expect(mocks.url).not.toHaveBeenCalled();
    expect(mocks.source).not.toHaveBeenCalled();
    await wrapper.setProps({dataList:[{...resource('new'),filesBases_id:'other'}]});
    await flushPromises();
    expect(mocks.url).toHaveBeenLastCalledWith('video-new');
    expect(mocks.source.mock.calls.at(-1)?.[0]).toBe('/video-new.mp4');
  }finally{wrapper.unmount();}
});
afterEach(()=>vi.unstubAllGlobals());
it.each([['短视频',ShortVideo],['上下短视频',TopBottom],['拼贴短视频',MosaicVideo]] as const)('%s 可连续选择两个本页外成员，不回退到旧播放器索引',async(_,component)=>{
  const wrapper=mount(component,{props:{dataList:[a]},global:{plugins:[ElementPlus],components:icons,stubs:{contentRightClickMenu:MenuStub,videoPlay:VideoStub,ElScrollbar:ScrollbarStub,Waterfall:WaterfallStub,layoutCoverPosterMosaic:MosaicStub,ElImage:true,ElBacktop:true}}});
  const choose=async(resource:I_resource)=>{
    if(component===MosaicVideo)wrapper.findComponent(Mosaic).vm.$emit('playResource',resource);
    else groupAction!(resource);
    await flushPromises();
  };
  try{
    await flushPromises();
    expect(keepOpenAfterSelect).toBe(true);
    mocks.scroll.mockClear();
    await choose(b);
    expect(mocks.url).toHaveBeenLastCalledWith('video-b');
    await choose(c);
    expect(mocks.url).toHaveBeenLastCalledWith('video-c');
    expect(mocks.source.mock.calls.at(-1)?.[0]).toBe('/video-c.mp4');
    expect(wrapper.props('dataList')).toEqual([a]);
    expect(mocks.scroll).not.toHaveBeenCalled();
  }finally{wrapper.unmount();}
});

it.each([['短视频',ShortVideo],['上下短视频',TopBottom],['拼贴短视频',MosaicVideo]] as const)('%s 选择无可播放分集的组成员后，后台列表更新不能跳回首项播放',async(_,component)=>{
 const wrapper=mount(component,{props:{dataList:[a]},global:{plugins:[ElementPlus],components:icons,stubs:{contentRightClickMenu:MenuStub,videoPlay:VideoStub,ElScrollbar:ScrollbarStub,Waterfall:WaterfallStub,layoutCoverPosterMosaic:MosaicStub,ElImage:true,ElBacktop:true}}});
 try{
  await flushPromises();const unavailable={...b,dramaSeries:[]};
  if(component===MosaicVideo)wrapper.findComponent(Mosaic).vm.$emit('playResource',unavailable);
  else groupAction!(unavailable);
  await flushPromises();mocks.url.mockClear();mocks.source.mockClear();
  await wrapper.setProps({dataList:[{...a,dramaSeries:a.dramaSeries.map(series=>({...series,durationSeconds:80}))}]});await flushPromises();
  expect(mocks.url).not.toHaveBeenCalled();expect(mocks.source).not.toHaveBeenCalled();
  if(component===MosaicVideo)wrapper.findComponent(Mosaic).vm.$emit('playResource',c);else groupAction!(c);
  await flushPromises();expect(mocks.url).toHaveBeenLastCalledWith('video-c');
 }finally{wrapper.unmount();}
});

it.each([['短视频',ShortVideo],['上下短视频',TopBottom],['拼贴短视频',MosaicVideo]] as const)('%s 组内成员播放地址加载失败后，再点同一成员可以重试',async(_,component)=>{
 const wrapper=mount(component,{props:{dataList:[a]},global:{plugins:[ElementPlus],components:icons,stubs:{contentRightClickMenu:MenuStub,videoPlay:VideoStub,ElScrollbar:ScrollbarStub,Waterfall:WaterfallStub,layoutCoverPosterMosaic:MosaicStub,ElImage:true,ElBacktop:true}}});
 const choose=async()=>{if(component===MosaicVideo)wrapper.findComponent(Mosaic).vm.$emit('playResource',b);else groupAction!(b);await flushPromises();};
 try{
  await flushPromises();mocks.url.mockClear();mocks.source.mockClear();
  mocks.url.mockResolvedValueOnce({playUrl:'',playType:'video/mp4'});
  await choose();expect(mocks.source).not.toHaveBeenCalled();expect(mocks.url).toHaveBeenCalledTimes(1);
  await choose();expect(mocks.url).toHaveBeenCalledTimes(2);expect(mocks.source.mock.calls.at(-1)?.[0]).toBe('/video-b.mp4');
 }finally{wrapper.unmount();}
});

it.each([['短视频',ShortVideo],['上下短视频',TopBottom],['拼贴短视频',MosaicVideo]] as const)('%s 播放地址网络异常可重试；旧成员异常不能清除新成员播放状态',async(_,component)=>{
 const wrapper=mount(component,{props:{dataList:[a]},global:{plugins:[ElementPlus],components:icons,stubs:{contentRightClickMenu:MenuStub,videoPlay:VideoStub,ElScrollbar:ScrollbarStub,Waterfall:WaterfallStub,layoutCoverPosterMosaic:MosaicStub,ElImage:true,ElBacktop:true}}});
 const choose=async(resource:I_resource)=>{if(component===MosaicVideo)wrapper.findComponent(Mosaic).vm.$emit('playResource',resource);else groupAction!(resource);await flushPromises();};
 try{
  await flushPromises();mocks.url.mockClear();mocks.source.mockClear();
  mocks.url.mockRejectedValueOnce(new Error('offline'));await choose(b);
  expect(mocks.source).not.toHaveBeenCalled();expect(mocks.url).toHaveBeenCalledTimes(1);
  await choose(b);expect(mocks.url).toHaveBeenCalledTimes(2);expect(mocks.source.mock.calls.at(-1)?.[0]).toBe('/video-b.mp4');
  let fail!: (error:Error)=>void;
  mocks.url.mockImplementationOnce(()=>new Promise((_resolve,reject)=>{fail=reject;}));await choose(c);
  await choose(a);expect(mocks.source.mock.calls.at(-1)?.[0]).toBe('/video-a.mp4');
  mocks.url.mockClear();mocks.source.mockClear();fail(new Error('old request failed'));await flushPromises();
  await choose(a);expect(mocks.url).not.toHaveBeenCalled();expect(mocks.source).not.toHaveBeenCalled();
 }finally{wrapper.unmount();}
});

it.each([['短视频',ShortVideo],['上下短视频',TopBottom],['拼贴短视频',MosaicVideo]] as const)('%s 普通卡片播放失败也能再次点击重试',async(_,component)=>{
 const wrapper=mount(component,{props:{dataList:[a,b]},global:{plugins:[ElementPlus],components:icons,stubs:{contentRightClickMenu:MenuStub,videoPlay:VideoStub,ElScrollbar:ScrollbarStub,Waterfall:WaterfallStub,layoutCoverPosterMosaic:MosaicStub,ElImage:true,ElBacktop:true}}});
 const click=async()=>{if(component===MosaicVideo)wrapper.findComponent(Mosaic).vm.$emit('playResource',b);else await wrapper.find(component===ShortVideo?'.waterfall-item-1':'.material-item-1').trigger('click');await flushPromises();};
 try{
  await flushPromises();mocks.url.mockClear();mocks.source.mockClear();mocks.url.mockResolvedValueOnce({playUrl:'',playType:'video/mp4'});
  await click();expect(mocks.url).toHaveBeenCalledTimes(1);expect(mocks.source).not.toHaveBeenCalled();
  await click();expect(mocks.url).toHaveBeenCalledTimes(2);expect(mocks.source.mock.calls.at(-1)?.[0]).toBe('/video-b.mp4');
 }finally{wrapper.unmount();}
});
it.each([['短视频',ShortVideo],['上下短视频',TopBottom],['拼贴短视频',MosaicVideo]] as const)('%s 从组视频点击普通无分集卡片，应停止旧视频，后台刷新也不恢复旧选择',async(_,component)=>{
 const empty={...c,dramaSeries:[]};
 const wrapper=mount(component,{props:{dataList:[a,empty]},global:{plugins:[ElementPlus],components:icons,stubs:{contentRightClickMenu:MenuStub,videoPlay:VideoStub,ElScrollbar:ScrollbarStub,Waterfall:WaterfallStub,layoutCoverPosterMosaic:MosaicStub,ElImage:true,ElBacktop:true}}});
 try{
  await flushPromises();if(component===MosaicVideo)wrapper.findComponent(Mosaic).vm.$emit('playResource',b);else groupAction!(b);await flushPromises();
  mocks.pause.mockClear();mocks.url.mockClear();mocks.source.mockClear();
  if(component===MosaicVideo)wrapper.findComponent(Mosaic).vm.$emit('playResource',empty);else await wrapper.find(component===ShortVideo?'.waterfall-item-1':'.material-item-1').trigger('click');await flushPromises();
  expect(mocks.pause).toHaveBeenCalled();expect(mocks.source).not.toHaveBeenCalled();
  await wrapper.setProps({dataList:[{...a,dramaSeries:a.dramaSeries.map(series=>({...series,durationSeconds:80}))},empty]});await flushPromises();
  expect(mocks.url).not.toHaveBeenCalled();expect(mocks.source).not.toHaveBeenCalled();
 }finally{wrapper.unmount();}
});
it.each(['control','modifier'] as const)('短视频播放器不抢占 %s 的方向键操作，普通方向键仍可切换视频',async(kind)=>{
 const wrapper=mount(ShortVideo,{props:{dataList:[a,c]},global:{plugins:[ElementPlus],components:icons,stubs:{contentRightClickMenu:MenuStub,videoPlay:VideoStub,ElScrollbar:ScrollbarStub,Waterfall:WaterfallStub,ElImage:true,ElBacktop:true}}});
 const popover=document.createElement('div');popover.className='el-popper';const button=document.createElement('button');const span=document.createElement('span');button.append(span);popover.append(button);document.body.append(popover);
 try{
  await flushPromises();mocks.url.mockClear();mocks.source.mockClear();
  const event=new KeyboardEvent('keydown',{key:'ArrowDown',bubbles:true,cancelable:true,ctrlKey:kind==='modifier'});
  (kind==='control'?span:window).dispatchEvent(event);await flushPromises();
  expect(event.defaultPrevented).toBe(false);expect(mocks.url).not.toHaveBeenCalled();
  popover.remove();window.dispatchEvent(new KeyboardEvent('keydown',{key:'ArrowDown',bubbles:true,cancelable:true}));await flushPromises();expect(mocks.url).toHaveBeenLastCalledWith('video-c');
 }finally{popover.remove();wrapper.unmount();}
});

it('普通短视频卡片统一播放入口后，翻页仍保留原播放，下一项可从新页开始',async()=>{
 const wrapper=mount(ShortVideo,{props:{dataList:[a,b]},global:{plugins:[ElementPlus],components:icons,stubs:{contentRightClickMenu:MenuStub,videoPlay:VideoStub,ElScrollbar:ScrollbarStub,Waterfall:WaterfallStub,ElImage:true,ElBacktop:true}}});
 try{
  await flushPromises();await wrapper.find('.waterfall-item-1').trigger('click');await flushPromises();mocks.url.mockClear();mocks.source.mockClear();
  await wrapper.setProps({dataList:[c]});await flushPromises();expect(mocks.source).not.toHaveBeenCalled();
  const next=wrapper.findAll('button').find(button=>button.text().includes('下一个'))!;expect(next.attributes('disabled')).toBeUndefined();
  await next.trigger('click');await flushPromises();expect(mocks.url).toHaveBeenLastCalledWith('video-c');
 }finally{wrapper.unmount();}
});

it.each([['短视频',ShortVideo],['拼贴短视频',MosaicVideo]] as const)('%s 通过方向键切换后加载失败，列表刷新不能恢复上次点击的资源',async(_,component)=>{
 const wrapper=mount(component,{props:{dataList:[a,b]},global:{plugins:[ElementPlus],components:icons,stubs:{contentRightClickMenu:MenuStub,videoPlay:VideoStub,ElScrollbar:ScrollbarStub,Waterfall:WaterfallStub,layoutCoverPosterMosaic:MosaicStub,ElImage:true,ElBacktop:true}}});
 try{
  await flushPromises();if(component===MosaicVideo)wrapper.findComponent(Mosaic).vm.$emit('playResource',a);else await wrapper.find('.waterfall-item-0').trigger('click');await flushPromises();
  mocks.url.mockClear();mocks.source.mockClear();mocks.url.mockResolvedValueOnce({playUrl:'',playType:'video/mp4'});
  window.dispatchEvent(new KeyboardEvent('keydown',{key:'ArrowDown',cancelable:true}));await flushPromises();expect(mocks.url).toHaveBeenLastCalledWith('video-b');expect(mocks.source).not.toHaveBeenCalled();
  mocks.url.mockClear();await wrapper.setProps({dataList:[{...a,dramaSeries:a.dramaSeries.map(series=>({...series,durationSeconds:80}))},b]});await flushPromises();
  expect(mocks.url).not.toHaveBeenCalled();expect(mocks.source).not.toHaveBeenCalled();
 }finally{wrapper.unmount();}
});
