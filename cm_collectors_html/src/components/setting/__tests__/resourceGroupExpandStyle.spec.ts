import { flushPromises, shallowMount } from '@vue/test-utils';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { computed } from 'vue';
import ElementPlus, { ElMessage } from 'element-plus';
import CoverAdjuster from '../fileDatabaseSetting/coverAdjuster.vue';
import SelectCoverStyle from '@/components/com/form/selectResourceGroupCoverStyle.vue';
import SelectStyle from '@/components/com/form/selectResourceGroupExpandStyle.vue';
import { createDefaultConfigApp } from '@/dataType/config.dataType';
import { appStoreData } from '@/storeData/app.storeData';
const mocks=vi.hoisted(()=>({save:vi.fn(),init:vi.fn()}));
vi.mock('@/server/filesBases.server',()=>({filesBasesServer:{setFilesBasesConfigById:mocks.save}}));
vi.mock('@/main',()=>({eventBus:{emit:vi.fn()}}));
vi.mock('@/assets/debounce',()=>({debounceNow:(fn:unknown)=>fn}));
vi.mock('@/common/theme',()=>({useTheme:()=>({isBrightTheme:computed(()=>true)})}));
vi.mock('@/storeData/app.storeData',async()=>{
  const {reactive}=await import('vue');
  const store=reactive({currentFilesBases:{id:'lib-a'},currentConfigApp:{} as Record<string,unknown>,appConfig:{appearanceStyle:'modern'},initCurrentFilesBases:mocks.init});
  return {appStoreData:()=>store};
});
afterEach(()=>vi.restoreAllMocks());
beforeEach(()=>{vi.clearAllMocks();mocks.save.mockResolvedValue({status:true});appStoreData().currentConfigApp=createDefaultConfigApp();appStoreData().currentFilesBases.id='lib-a';});
it.each(['strip','fold','spine','magazine'] as const)('显示设置默认扇形，切换 %s 后随本库配置保存，切库读取各自的样式',async(style)=>{
  const wrapper=shallowMount(CoverAdjuster,{global:{plugins:[ElementPlus],renderStubDefaultSlot:true,mocks:{$t:(key:string)=>key},stubs:{ElPopover:{template:'<div><slot name="reference"/><slot/></div>'}}}});
  try{
    expect(wrapper.findComponent(SelectStyle).props('modelValue')).toBe('fan');
    wrapper.findComponent(SelectStyle).vm.$emit('update:modelValue',style);
    await flushPromises();
    expect(appStoreData().currentConfigApp.resourceGroupExpandStyle).toBe(style);
    wrapper.findAllComponents({name:'ElButton'}).find(button=>button.text()==='保存')!.vm.$emit('click');
    await flushPromises();
    expect(mocks.save).toHaveBeenCalledWith('lib-a',expect.objectContaining({resourceGroupExpandStyle:style}),expect.objectContaining({localOnly:true}));
    appStoreData().currentFilesBases.id='lib-b';
    appStoreData().currentConfigApp=createDefaultConfigApp();
    await flushPromises();
    expect(wrapper.findComponent(SelectStyle).props('modelValue')).toBe('fan');
    appStoreData().currentFilesBases.id='lib-a';
    appStoreData().currentConfigApp={...createDefaultConfigApp(),resourceGroupExpandStyle:style};
    await flushPromises();
    expect(wrapper.findComponent(SelectStyle).props('modelValue')).toBe(style);
  }finally{wrapper.unmount();}
});

it.each(['triptych','book','adaptive'] as const)('封面样式默认叠牌，切换 %s 后随本库配置保存，切库读取各自的样式',async(style)=>{
  const wrapper=shallowMount(CoverAdjuster,{global:{plugins:[ElementPlus],renderStubDefaultSlot:true,mocks:{$t:(key:string)=>key},stubs:{ElPopover:{template:'<div><slot name="reference"/><slot/></div>'}}}});
  try{
    expect(wrapper.findComponent(SelectCoverStyle).props('modelValue')).toBe('stack');
    wrapper.findComponent(SelectCoverStyle).vm.$emit('update:modelValue',style);
    await flushPromises();
    expect(appStoreData().currentConfigApp.resourceGroupCoverStyle).toBe(style);
    wrapper.findAllComponents({name:'ElButton'}).find(button=>button.text()==='保存')!.vm.$emit('click');
    await flushPromises();
    expect(mocks.save).toHaveBeenCalledWith('lib-a',expect.objectContaining({resourceGroupCoverStyle:style}),expect.objectContaining({localOnly:true}));
    appStoreData().currentFilesBases.id='lib-b';
    appStoreData().currentConfigApp=createDefaultConfigApp();
    await flushPromises();
    expect(wrapper.findComponent(SelectCoverStyle).props('modelValue')).toBe('stack');
    appStoreData().currentFilesBases.id='lib-a';
    appStoreData().currentConfigApp={...createDefaultConfigApp(),resourceGroupCoverStyle:style};
    await flushPromises();
    expect(wrapper.findComponent(SelectCoverStyle).props('modelValue')).toBe(style);
  }finally{wrapper.unmount();}
});

it.each([true,false])('旧库显示设置保存返回 status=%s，不提示或重载新库配置',async(status)=>{
 let finish!: (result:unknown)=>void;mocks.save.mockImplementationOnce(()=>new Promise(resolve=>{finish=resolve;}));
 const success=vi.spyOn(ElMessage,'success'),error=vi.spyOn(ElMessage,'error');
 const wrapper=shallowMount(CoverAdjuster,{global:{plugins:[ElementPlus],renderStubDefaultSlot:true,mocks:{$t:(key:string)=>key},stubs:{ElPopover:{template:'<div><slot name="reference"/><slot/></div>'}}}});
 try{
  wrapper.findAllComponents({name:'ElButton'}).find(button=>button.text()==='保存')!.vm.$emit('click');await flushPromises();
  appStoreData().currentFilesBases.id='lib-b';appStoreData().currentConfigApp={...createDefaultConfigApp(),resourceGroupCoverStyle:'adaptive'};
  finish({status,msg:'old save failed'});await flushPromises();
  expect(mocks.init).not.toHaveBeenCalled();expect(success).not.toHaveBeenCalled();expect(error).not.toHaveBeenCalled();
  expect(appStoreData().currentConfigApp.resourceGroupCoverStyle).toBe('adaptive');
 }finally{wrapper.unmount();}
});
it('保存失败保留当前封面选择，网络异常同样可以再次保存',async()=>{
 mocks.save.mockResolvedValueOnce({status:false,msg:'failed'});
 const wrapper=shallowMount(CoverAdjuster,{global:{plugins:[ElementPlus],renderStubDefaultSlot:true,mocks:{$t:(key:string)=>key},stubs:{ElPopover:{template:'<div><slot name="reference"/><slot/></div>'}}}});
 try{
  wrapper.findComponent(SelectCoverStyle).vm.$emit('update:modelValue','adaptive');await flushPromises();
  const save=wrapper.findAllComponents({name:'ElButton'}).find(button=>button.text()==='保存')!;
  save.vm.$emit('click');await flushPromises();expect(mocks.init).not.toHaveBeenCalled();expect(appStoreData().currentConfigApp.resourceGroupCoverStyle).toBe('adaptive');
  mocks.save.mockRejectedValueOnce(new Error('offline'));save.vm.$emit('click');await flushPromises();
  expect(appStoreData().currentConfigApp.resourceGroupCoverStyle).toBe('adaptive');
  save.vm.$emit('click');await flushPromises();expect(mocks.save).toHaveBeenCalledTimes(3);
 }finally{wrapper.unmount();}
});

it('保存期间不重复提交，发送提交时的快照，后续编辑保留并提示再次保存',async()=>{
 let finish!: (result:unknown)=>void;mocks.save.mockImplementationOnce(()=>new Promise(resolve=>{finish=resolve;}));const success=vi.spyOn(ElMessage,'success');
 const wrapper=shallowMount(CoverAdjuster,{global:{plugins:[ElementPlus],renderStubDefaultSlot:true,mocks:{$t:(key:string)=>key},stubs:{ElPopover:{template:'<div><slot name="reference"/><slot/></div>'}}}});
 try{
  wrapper.findComponent(SelectCoverStyle).vm.$emit('update:modelValue','adaptive');await flushPromises();
  const save=wrapper.findAllComponents({name:'ElButton'}).find(button=>button.text()==='保存')!;
  save.vm.$emit('click');save.vm.$emit('click');await flushPromises();expect(mocks.save).toHaveBeenCalledTimes(1);expect(save.props('loading')).toBe(true);
  wrapper.findComponent(SelectCoverStyle).vm.$emit('update:modelValue','book');await flushPromises();
  expect(mocks.save.mock.calls[0][1].resourceGroupCoverStyle).toBe('adaptive');
  finish({status:true});await flushPromises();expect(appStoreData().currentConfigApp.resourceGroupCoverStyle).toBe('book');expect(save.props('loading')).toBe(false);
  expect(success).toHaveBeenCalledWith('已保存提交的设置，后续修改请再次保存');
 }finally{wrapper.unmount();}
});

it('切库后可以立即保存新库，旧保存完成不能解除新保存的等待状态',async()=>{
 let first!: (result:unknown)=>void,second!: (result:unknown)=>void;
 mocks.save.mockImplementationOnce(()=>new Promise(resolve=>{first=resolve;})).mockImplementationOnce(()=>new Promise(resolve=>{second=resolve;}));
 const error=vi.spyOn(ElMessage,'error');
 const wrapper=shallowMount(CoverAdjuster,{global:{plugins:[ElementPlus],renderStubDefaultSlot:true,mocks:{$t:(key:string)=>key},stubs:{ElPopover:{template:'<div><slot name="reference"/><slot/></div>'}}}});
 try{
  const save=wrapper.findAllComponents({name:'ElButton'}).find(button=>button.text()==='保存')!;save.vm.$emit('click');await flushPromises();
  appStoreData().currentFilesBases.id='lib-b';appStoreData().currentConfigApp={...createDefaultConfigApp(),resourceGroupCoverStyle:'book'};await flushPromises();
  expect(save.props('loading')).toBe(false);save.vm.$emit('click');await flushPromises();
  expect(mocks.save.mock.calls.map(args=>args[0])).toEqual(['lib-a','lib-b']);
  first({status:false,msg:'old save failed'});await flushPromises();expect(save.props('loading')).toBe(true);expect(error).not.toHaveBeenCalled();
  second({status:true});await flushPromises();expect(save.props('loading')).toBe(false);
 }finally{wrapper.unmount();}
});

it('底栏分组保存时拒绝通用保存，恢复后对外同步保存状态',async()=>{
 let finish!:(value:unknown)=>void;
 mocks.save.mockReturnValueOnce(new Promise(resolve=>{finish=resolve;}));
 const w=shallowMount(CoverAdjuster,{props:{disabled:true},global:{plugins:[ElementPlus],renderStubDefaultSlot:true,mocks:{$t:(key:string)=>key},stubs:{ElPopover:{template:'<div><slot name="reference"/><slot/></div>'}}}});
 try{
  const save=w.findAllComponents({name:'ElButton'}).find(b=>b.text()==='保存')!;
  save.vm.$emit('click');await flushPromises();expect(mocks.save).not.toHaveBeenCalled();
  await w.setProps({disabled:false});save.vm.$emit('click');await flushPromises();
  expect(w.emitted('saving')?.at(-1)).toEqual([true]);
  finish({status:true});await flushPromises();expect(w.emitted('saving')?.at(-1)).toEqual([false]);
 }finally{w.unmount();}
});
it('同库配置被替换后可立即保存新配置，旧请求不能解除新保存状态',async()=>{
 let first!:(value:unknown)=>void,second!:(value:unknown)=>void;
 mocks.save.mockImplementationOnce(()=>new Promise(resolve=>{first=resolve;})).mockImplementationOnce(()=>new Promise(resolve=>{second=resolve;}));
 const w=shallowMount(CoverAdjuster,{global:{plugins:[ElementPlus],renderStubDefaultSlot:true,mocks:{$t:(key:string)=>key},stubs:{ElPopover:{template:'<div><slot name="reference"/><slot/></div>'}}}});
 try{
  const save=w.findAllComponents({name:'ElButton'}).find(b=>b.text()==='保存')!;
  save.vm.$emit('click');await flushPromises();
  appStoreData().currentConfigApp={...createDefaultConfigApp(),resourceGroupCoverStyle:'book'};await flushPromises();
  save.vm.$emit('click');await flushPromises();expect(mocks.save).toHaveBeenCalledTimes(2);
  first({status:true});await flushPromises();expect(save.props('loading')).toBe(true);
  second({status:true});await flushPromises();expect(save.props('loading')).toBe(false);
 }finally{w.unmount();}
});

it('分类预览随底部保存一起持久化，预览本身不触发保存', async () => {
  const wrapper=shallowMount(CoverAdjuster,{global:{plugins:[ElementPlus],renderStubDefaultSlot:true,mocks:{$t:(key:string)=>key},stubs:{ElPopover:{template:'<div><slot name="reference"/><slot/></div>'}}}});
  try {
    Object.assign(appStoreData().currentConfigApp, { categoryBrowseEnabled: true, categoryBrowseSort: 'name', categoryBrowseAdaptCoverRatio: true });
    await flushPromises();
    expect(mocks.save).not.toHaveBeenCalled();
    wrapper.findAllComponents({name:'ElButton'}).find(button=>button.text()==='保存')!.vm.$emit('click');
    await flushPromises();
    expect(mocks.save).toHaveBeenCalledWith('lib-a', expect.objectContaining({ categoryBrowseEnabled: true, categoryBrowseSort: 'name', categoryBrowseAdaptCoverRatio: true }), expect.objectContaining({ localOnly: true }));
  } finally { wrapper.unmount(); }
});

it('分类输入校验失败时底部保存不会提交旧值', async () => {
  const validate = vi.fn(() => false);
  const wrapper=shallowMount(CoverAdjuster,{global:{plugins:[ElementPlus],renderStubDefaultSlot:true,mocks:{$t:(key:string)=>key},stubs:{CategoryBrowseQuickSettings:{template:'<div/>',methods:{validate}},ElPopover:{template:'<div><slot name="reference"/><slot/></div>'}}}});
  try {
    wrapper.findAllComponents({name:'ElButton'}).find(button=>button.text()==='保存')!.vm.$emit('click');
    await flushPromises();
    expect(validate).toHaveBeenCalled();
    expect(mocks.save).not.toHaveBeenCalled();
  } finally { wrapper.unmount(); }
});
