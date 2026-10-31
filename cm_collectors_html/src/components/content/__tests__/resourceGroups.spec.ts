import { computed, defineComponent, inject, nextTick, ref } from 'vue';
import { flushPromises, mount, shallowMount } from '@vue/test-utils';
import ElementPlus, { ElMessage, ElMessageBox, ElPagination, ElRadioGroup } from 'element-plus';
import * as icons from '@element-plus/icons-vue';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { I_resource } from '@/dataType/resource.dataType';
import { claimResourceGroupFan, releaseResourceGroupFan, mergeResourceGroupPage, resourceGroupContextKey, resourceGroupKeepOpenKey, resourceGroupSelectKey, resourceGroupPageKey } from '../resourceGroups';
import ContentList from '../contentList.vue';
import LayoutPoster from '../layoutCoverPoster.vue';
import ContentMenu from '../contentRightClickMenu.vue';
import GroupManager from '../ResourceGroupManager.vue';
import GroupPanel from '../ResourceGroupPanel.vue';
import GroupFan from '../ResourceGroupFan.vue';
import ModuleContent from '../moduleContent.vue';
import GroupMark from '../ResourceGroupMark.vue';
import WaterfallLayout from '../layoutCoverPosterWaterfall.vue';
import GroupCover from '../ResourceGroupCover.vue';
import { RESOURCE_GROUP_MARK_COLORS, resourceSelectionKey } from '../resourceSelection';

const mocks = vi.hoisted(() => ({ summaries: vi.fn(), info: vi.fn(), mutate: vi.fn(), mobile: false, recommendations: vi.fn() }));
vi.mock('@/main', () => ({ eventBus: { emit: vi.fn() } }));
vi.mock('@/server/resourceGroup.server', () => ({ resourceGroupServer: mocks }));
vi.mock('@/common/play', () => ({ playResource: vi.fn(), playOpenResourceFolder: vi.fn() }));
vi.mock('@/common/resource', () => ({ resourceDelete: vi.fn() }));
vi.mock('@/common/playList', () => ({ playListAdd: vi.fn() }));
vi.mock('@/common/photo', () => ({ getResourceCoverPoster: (resource: { id: string }) => `/cover/${resource.id}.jpg` }));
vi.mock('@/storeData/app.storeData', async () => {
  const { reactive } = await import('vue');
  const store = reactive({ currentConfigApp: { resourcesShowMode: 'coverPoster', mergeResourceGroups: true, contentPadding: 0 }, currentFilesBases: { id: 'lib' }, displayAdminFn: false });
  return { appStoreData: () => store };
});
vi.mock('@/storeData/search.storeData', () => ({ searchStoreData: () => ({ searchData: { sort: '' } }) }));
vi.mock('@/server/resource.server', () => ({ resourceServer: { dataListCasualView: mocks.recommendations } }));
vi.mock('@/server/videoTranscode.server', () => ({ videoTranscodeServer: {} }));
vi.mock('@/server/tvboxRecommend.server', () => ({ tvboxRecommendServer: {} }));
vi.mock('@/assets/mobile', () => ({ isMobile: () => mocks.mobile }));
import { appStoreData } from '@/storeData/app.storeData';
import { playResource } from '@/common/play';

const resource = (id: string, group = '') => ({ id, title: id, filesBases_id: 'lib', associationGroupId: group, tags: [], performers: [], directors: [], dramaSeries: [] } as unknown as I_resource);
const a = resource('a', 'group-1');
const b = resource('b', 'group-1');
const c = resource('c');
const globals = { plugins: [ElementPlus], components: icons, directives: { admin: {} } };
beforeEach(() => {
  vi.clearAllMocks(); mocks.mobile = false;
  appStoreData().currentConfigApp.resourcesShowMode = 'coverPoster';
  appStoreData().currentConfigApp.mergeResourceGroups = true;
  appStoreData().currentConfigApp.resourceGroupExpandStyle = 'fan';
  appStoreData().currentConfigApp.resourceGroupCoverStyle = 'stack';
  appStoreData().currentConfigApp.resourceDetailsShowMode = 'right';
  mocks.summaries.mockResolvedValue({ status: true, data: { dataList: [{ id: 'group-1', count: 2, filesBasesId: 'lib', preview: [a,b] }], total: 1 } });
  mocks.info.mockResolvedValue({ status: true, data: [a,b] });
  mocks.recommendations.mockResolvedValue({status:true,data:[a,b]});
  mocks.mutate.mockResolvedValue({ status: true, data: { id: 'group-1' } });
});
afterEach(() => vi.restoreAllMocks());

it('按本页首次出现的位置合并，关闭时原样返回且不修改原始数组', () => {
  const raw = [c,b,a];
  expect(mergeResourceGroupPage(raw,true)).toEqual([c,b]);
  expect(mergeResourceGroupPage(raw,false)).toBe(raw);
  expect(raw).toEqual([c,b,a]);
});

it('分组标记按本页组号循环使用背景、文字和边框配色', async () => {
  const labels = new Map(Array.from({length: RESOURCE_GROUP_MARK_COLORS.length + 1}, (_, index) => [`g${index}`, `组 ${index + 1}`]));
  const wrapper = mount(GroupMark, {props:{groupId:'g0'},global:{...globals,provide:{[resourceSelectionKey as symbol]:{
    active:computed(()=>true),selectedIds:computed(()=>new Set<string>()),groupLabels:computed(()=>labels),toggle:vi.fn(),
  }}}});
  try {
    const first = wrapper.attributes('style');
    expect(first).toContain(RESOURCE_GROUP_MARK_COLORS[0].background);
    expect(first).toContain(RESOURCE_GROUP_MARK_COLORS[0].text);
    await wrapper.setProps({groupId:'g1'});
    expect(wrapper.attributes('style')).not.toBe(first);
    expect(wrapper.attributes('style')).toContain(RESOURCE_GROUP_MARK_COLORS[1].background);
    expect(wrapper.attributes('style')).toContain(RESOURCE_GROUP_MARK_COLORS[1].text);
    await wrapper.setProps({groupId:`g${RESOURCE_GROUP_MARK_COLORS.length}`});
    expect(wrapper.attributes('style')).toBe(first);
  } finally { wrapper.unmount(); }
});

it('主列表切换合并开关不重新搜索，表格和移动端不合并', async () => {
  const wrapper = shallowMount(ContentList,{ props:{dataList:[b,c,a]}, global: globals });
  try {
    await flushPromises();
    expect(wrapper.findComponent(LayoutPoster).props('dataList')).toEqual([b,c]);
    expect(mocks.summaries).toHaveBeenCalledWith('lib',['group-1']);
    appStoreData().currentConfigApp.mergeResourceGroups = false;
    await nextTick();
    expect(wrapper.findComponent(LayoutPoster).props('dataList')).toEqual([b,c,a]);
    appStoreData().currentConfigApp.mergeResourceGroups = true;
    appStoreData().currentConfigApp.resourcesShowMode = 'table';
    await flushPromises();
    expect(wrapper.findComponent({ name:'layoutTable' }).props('dataList')).toEqual([b,c,a]);
    appStoreData().currentConfigApp.resourcesShowMode = 'coverPoster';
    mocks.mobile = true;
    await wrapper.setProps({dataList:[a,b,c]});
    expect(wrapper.findComponent(LayoutPoster).props('dataList')).toEqual([a,b,c]);
  } finally { wrapper.unmount(); }
});

it('单个命中成员也显示叠牌，推荐模块中的资源不叠牌，选择成员走布局响应', async () => {
  const select = vi.fn(); const layoutAction = vi.fn();
  const enabled = ref(true);
  const wrapper = mount(ContentMenu,{
    props:{resource:a}, slots:{default:'<div class="sample-card" style="width:160px;height:230px">card</div>'},
    global:{...globals, provide:{
      [resourceGroupContextKey as symbol]:{ enabled:computed(()=>enabled.value), pageIds:computed(()=>new Set(['a'])), summaries:ref(new Map()), select },
      [resourceGroupSelectKey as symbol]:layoutAction,
    },stubs:{contentMenu:{template:'<div><slot /></div>'},ResourceGroupPanel:defineComponent({props:['expandStyle'],emits:['select'],template:'<button @click="$emit(\'select\', {})">member</button>'})}},
  });
  try {
    expect(wrapper.classes()).toContain('resource-group-stack');
    expect(wrapper.findAll('.resource-group-back')).toHaveLength(2);
    appStoreData().currentConfigApp.resourceGroupExpandStyle = 'strip';
    await wrapper.trigger('pointerenter');
    await new Promise(resolve => setTimeout(resolve, 200));
    await nextTick();
    expect(wrapper.findComponent(GroupPanel).props('expandStyle')).toBe('strip');
    wrapper.findComponent(GroupPanel).vm.$emit('select',b);
    expect(layoutAction).toHaveBeenCalledWith(b);
    expect(select).not.toHaveBeenCalled();
    await wrapper.setProps({resource:b});
    expect(wrapper.classes()).not.toContain('resource-group-stack');
    await wrapper.setProps({resource:a});
    enabled.value = false; await nextTick();
    expect(wrapper.classes()).not.toContain('resource-group-stack');
  } finally { wrapper.unmount(); }
});

it('原资源与浮层交接不关闭，离开后延迟关闭，可取消未完成的展开并重新进入', async () => {
  vi.useFakeTimers();
  const popover = defineComponent({name:'ElPopover',props:['visible'],template:'<div class="hover-popover"><slot /></div>'});
  const wrapper = mount(ContentMenu, { props:{resource:a},global:{...globals,
    provide:{[resourceGroupContextKey as symbol]:{enabled:computed(()=>true),pageIds:computed(()=>new Set(['a'])),summaries:ref(new Map()),select:vi.fn()}},
    stubs:{ElPopover:popover,contentMenu:{template:'<div><slot /></div>'},ResourceGroupPanel:true},
  }});
  const visible=()=>wrapper.findComponent(popover).props('visible');
  let anchorTop=0;
  vi.spyOn(wrapper.element,'getBoundingClientRect').mockImplementation(()=>({top:anchorTop,left:0,width:0,height:0,right:0,bottom:anchorTop,x:0,y:anchorTop,toJSON(){}}));
  try {
    await wrapper.trigger('pointerenter');
    await vi.advanceTimersByTimeAsync(100);
    await wrapper.trigger('pointerleave');
    await vi.advanceTimersByTimeAsync(200);
    expect(visible()).toBe(false);
    await wrapper.trigger('pointerenter');
    await vi.advanceTimersByTimeAsync(180);
    expect(visible()).toBe(true);
    const surface=wrapper.find('.hover-popover').element.firstElementChild!;
    await wrapper.trigger('pointerleave',{relatedTarget:surface});
    await vi.advanceTimersByTimeAsync(600);
    expect(visible()).toBe(true);
    surface.dispatchEvent(new Event('pointerleave'));
    await vi.advanceTimersByTimeAsync(200);
    surface.dispatchEvent(new Event('pointerenter'));
    await vi.advanceTimersByTimeAsync(400);
    expect(visible()).toBe(true);
    surface.dispatchEvent(new Event('scroll'));
    await nextTick();
    expect(visible()).toBe(true);
    window.dispatchEvent(new Event('scroll'));
    await nextTick();
    expect(visible()).toBe(true);
    anchorTop=40;
    window.dispatchEvent(new Event('scroll'));
    await nextTick();
    expect(visible()).toBe(false);
    wrapper.findComponent(popover).vm.$emit('after-leave');
    await nextTick();
    expect(wrapper.findComponent(GroupPanel).props('displayActive')).toBe(false);
    await wrapper.trigger('pointerenter');
    await vi.advanceTimersByTimeAsync(180);
    expect(wrapper.findComponent(GroupPanel).props('displayActive')).toBe(true);
    surface.dispatchEvent(new Event('pointerleave'));
    await vi.advanceTimersByTimeAsync(400);
    expect(visible()).toBe(false);
    await wrapper.trigger('pointerenter');
    await wrapper.setProps({resource:c});
    await vi.advanceTimersByTimeAsync(500);
    expect(wrapper.findComponent(popover).exists()).toBe(false);
  } finally { wrapper.unmount(); vi.useRealTimers(); }
});
it.each(['spine','magazine'] as const)('%s 贴近首次悬停位置展开，宽卡片不再要求鼠标横跨整张卡，移动后锚点保持稳定',async(style)=>{
  vi.useFakeTimers();appStoreData().currentConfigApp.resourceGroupExpandStyle=style;
  const wrapper=mount(ContentMenu,{props:{resource:a},global:{...globals,
    provide:{[resourceGroupContextKey as symbol]:{enabled:computed(()=>true),pageIds:computed(()=>new Set(['a'])),summaries:ref(new Map())}},
    stubs:{ElPopover:defineComponent({name:'ElPopover',props:['visible','virtualRef','placement','offset'],template:'<div><slot /></div>'}),contentMenu:{template:'<div><slot /></div>'},ResourceGroupPanel:true},
  }});
  try{
    vi.spyOn(wrapper.element,'getBoundingClientRect').mockReturnValue({left:100,top:50,right:900,bottom:350,width:800,height:300,x:100,y:50,toJSON(){}});
    await wrapper.trigger('pointerenter',{clientX:140,clientY:120});
    await vi.advanceTimersByTimeAsync(80);
    await wrapper.trigger('pointermove',{clientX:170,clientY:130});
    await vi.advanceTimersByTimeAsync(100);
    const popover=wrapper.findComponent({name:'ElPopover'});
    const anchor=popover.props('virtualRef') as {getBoundingClientRect:()=>DOMRect};
    expect(anchor.getBoundingClientRect().left).toBe(170);
    expect(anchor.getBoundingClientRect().top).toBe(130);
    expect(popover.props('placement')).toBe('right');
    expect(popover.props('offset')).toBe(-24);
    document.dispatchEvent(new MouseEvent('pointermove',{clientX:200,clientY:160}));
    await nextTick();
    expect(anchor.getBoundingClientRect().left).toBe(170);
    expect(anchor.getBoundingClientRect().top).toBe(130);
    expect(popover.props('visible')).toBe(true);
  }finally{wrapper.unmount();vi.useRealTimers();}
});

it.each([true,false])('组内选择 keepOpen=%s：内嵌播放器保留面板可连续选择，普通详情收起面板',async(keepOpen)=>{
  vi.useFakeTimers();
  const action=vi.fn();
  const popover=defineComponent({name:'ElPopover',props:['visible'],template:'<div><slot /></div>'});
  const wrapper=mount(ContentMenu,{props:{resource:a},global:{...globals,
    provide:{[resourceGroupContextKey as symbol]:{enabled:computed(()=>true),pageIds:computed(()=>new Set(['a'])),summaries:ref(new Map())},[resourceGroupKeepOpenKey as symbol]:keepOpen,[resourceGroupSelectKey as symbol]:action},
    stubs:{ElPopover:popover,contentMenu:{template:'<div><slot /></div>'},ResourceGroupPanel:true},
  }});
  try{
    await wrapper.trigger('pointerenter');await vi.advanceTimersByTimeAsync(180);
    expect(wrapper.findComponent(popover).props('visible')).toBe(true);
    wrapper.findComponent(GroupPanel).vm.$emit('select',b);await nextTick();
    expect(action).toHaveBeenLastCalledWith(b);
    expect(wrapper.findComponent(popover).props('visible')).toBe(keepOpen);
    if(keepOpen){
      wrapper.findComponent(GroupPanel).vm.$emit('select',c);await nextTick();
      expect(action).toHaveBeenLastCalledWith(c);
      expect(wrapper.findComponent(popover).props('visible')).toBe(true);
      await wrapper.trigger('pointerleave');await vi.advanceTimersByTimeAsync(400);
      expect(wrapper.findComponent(popover).props('visible')).toBe(false);
    }
  }finally{wrapper.unmount();vi.useRealTimers();}
});

it.each((['fan','strip','fold','spine','magazine'] as const).flatMap(style=>[
  {style,details:'none',inline:false,plays:true},
  {style,details:'none',inline:true,plays:false},
  {style,details:'right',inline:true,plays:false},
  {style,details:'popup',inline:true,plays:false},
  {style,details:'right',inline:false,plays:true},
  {style,details:'popup',inline:false,plays:true},
]))('$style / $details / 内嵌=$inline：组内点击统一播放，不依赖详情设置',async({style,details,inline,plays})=>{
  vi.useFakeTimers();
  appStoreData().currentConfigApp.resourceDetailsShowMode=details;
  appStoreData().currentConfigApp.resourceGroupExpandStyle=style;
  const action=vi.fn();
  const popover=defineComponent({name:'ElPopover',props:['visible'],template:'<div><slot /></div>'});
  const wrapper=mount(ContentMenu,{props:{resource:a},global:{...globals,
    provide:{[resourceGroupContextKey as symbol]:{enabled:computed(()=>true),pageIds:computed(()=>new Set(['a'])),summaries:ref(new Map())},[resourceGroupKeepOpenKey as symbol]:inline,...(inline?{[resourceGroupSelectKey as symbol]:action}:{})},
    stubs:{ElPopover:popover,contentMenu:{template:'<div><slot /></div>'},ResourceGroupPanel:true},
  }});
  try{
    await wrapper.trigger('pointerenter');await vi.advanceTimersByTimeAsync(180);
    wrapper.findComponent(GroupPanel).vm.$emit('select',b);await nextTick();
    if(plays){expect(playResource).toHaveBeenCalledExactlyOnceWith(b);expect(action).not.toHaveBeenCalled();}
    else{expect(action).toHaveBeenCalledExactlyOnceWith(b);expect(playResource).not.toHaveBeenCalled();}
    expect(wrapper.findComponent(popover).props('visible')).toBe(inline);
  }finally{wrapper.unmount();vi.useRealTimers();}
});

it('首次展开时布局触发离开或滚动通知，指针仍在原资源内就保持展开，真正移出后收起', async () => {
  vi.useFakeTimers();
  let finish!: (value: unknown) => void;
  mocks.info.mockImplementationOnce(()=>new Promise(resolve=>{finish=resolve;}));
  const popover=defineComponent({name:'ElPopover',props:['visible'],template:'<div><slot /></div>'});
  const wrapper=mount(ContentMenu,{props:{resource:a},global:{...globals,
    provide:{[resourceGroupContextKey as symbol]:{enabled:computed(()=>true),pageIds:computed(()=>new Set(['a'])),summaries:ref(new Map()),select:vi.fn()}},
    stubs:{ElPopover:popover,contentMenu:{template:'<div><slot /></div>'}},
  }});
  const visible=()=>wrapper.findComponent(popover).props('visible');
  try {
    vi.spyOn(wrapper.element,'getBoundingClientRect').mockReturnValue({left:50,top:50,right:250,bottom:300,width:200,height:250,x:50,y:50,toJSON(){}});
    await wrapper.trigger('pointerenter',{clientX:100,clientY:120});
    await vi.advanceTimersByTimeAsync(180);
    expect(visible()).toBe(true);
    expect(wrapper.find('.fan-state').exists()).toBe(true);
    await wrapper.trigger('pointerleave',{clientX:100,clientY:120,relatedTarget:document.body});
    window.dispatchEvent(new Event('scroll'));
    await vi.advanceTimersByTimeAsync(800);
    expect(visible()).toBe(true);
    finish({status:true,data:[a,b]});
    await flushPromises();
    expect(wrapper.findAll('.fan-card')).toHaveLength(2);
    expect(visible()).toBe(true);
    document.dispatchEvent(new MouseEvent('pointermove',{clientX:600,clientY:600}));
    await vi.advanceTimersByTimeAsync(400);
    expect(visible()).toBe(false);
  } finally { wrapper.unmount();vi.useRealTimers(); }
});

it('推荐模块即使命中当前页的同一资源，也不继承叠牌展示', async () => {
  const probe = defineComponent({ setup() { return { grouping: inject(resourceGroupContextKey) }; }, template: '<div class="recommendation-probe" :data-grouping="grouping?.enabled.value" />' });
  const wrapper = mount(ModuleContent, { props: { resourcesShowMode:'coverPoster', moduleType:'casualView' }, global:{ ...globals,
    provide:{ [resourceGroupContextKey as symbol]: { enabled:computed(()=>true),pageIds:computed(()=>new Set(['a','b'])),summaries:ref(new Map()),select:vi.fn() } },
    stubs:{contentStyleIndex:probe},
  } });
  try {
    await flushPromises();
    expect(wrapper.findAll('.recommendation-probe')).toHaveLength(2);
    wrapper.findAll('.recommendation-probe').forEach(card=>expect(card.attributes('data-grouping')).toBe('false'));
  } finally { wrapper.unmount(); }
});

const managerButton = (wrapper: ReturnType<typeof mount>, label: string) => wrapper.findAll('button').find(item=>item.text()===label)!;
it('极密集封面带使用小数布局尺寸，末尾资源不会因整数取整而偏到邻牌', async () => {
  const members=Array.from({length:1000},(_,index)=>resource(String(index)));
  const wrapper=mount(GroupFan,{props:{members,loading:false,error:'',expandStyle:'strip'},global:{...globals,stubs:{ElImage:true}}});
  try {
    const deck=wrapper.find('.fan-deck');
    vi.spyOn(deck.element,'getBoundingClientRect').mockReturnValue({left:0,top:0,width:503.5,height:272,right:503.5,bottom:272,x:0,y:0,toJSON(){}});
    Object.defineProperty(deck.element,'clientWidth',{value:504});
    const first=wrapper.find('.fan-card').element as HTMLElement;
    first.style.width='120.72px';
    Object.defineProperty(first,'offsetWidth',{value:121});
    const lastCenter=(503.5+(503.5-120.72-48))/2;
    await deck.trigger('pointermove',{clientX:lastCenter});
    expect(wrapper.find('.fan-card.is-active').attributes('aria-label')).toBe('打开资源：999');
    wrapper.findAll('.fan-card')[999].element.dispatchEvent(new MouseEvent('click',{bubbles:true,detail:1,clientX:lastCenter}));
    expect(wrapper.emitted('select')?.[0]).toEqual([members[999]]);
    const last=wrapper.findAll('.fan-card')[999];
    last.element.dispatchEvent(new MouseEvent('pointerdown',{bubbles:true,cancelable:true,button:0,clientX:lastCenter}));
    last.element.dispatchEvent(new MouseEvent('click',{bubbles:true,detail:1,clientX:Math.floor(lastCenter)}));
    expect(wrapper.emitted('select')?.[1]).toEqual([members[999]]);
  } finally {wrapper.unmount();}
});

it('默认扇形，切换横向封面带后保持直立并使用固定命中区域，点击和键盘响应相同', async () => {
  const wrapper=mount(GroupFan,{props:{members:[a,b,c],loading:false,error:''},global:globals});
  try {
    expect(wrapper.classes()).not.toContain('group-fan--strip');
    await wrapper.setProps({expandStyle:'strip'});
    expect(wrapper.classes()).toContain('group-fan--strip');
    wrapper.findAll('.fan-card').forEach(card=>expect(card.attributes('style')).toContain('--card-angle: 0deg'));
    Object.defineProperty(wrapper.find('.fan-deck').element,'clientWidth',{value:288});
    Object.defineProperty(wrapper.find('.fan-card').element,'offsetWidth',{value:160});
    await wrapper.find('.fan-deck').trigger('pointermove',{clientX:184});
    expect(wrapper.find('.fan-card.is-active').attributes('aria-label')).toBe('打开资源：c');
    await wrapper.find('.fan-card.is-active').trigger('click');
    expect(wrapper.emitted('select')?.[0]).toEqual([c]);
    await wrapper.find('.fan-deck').trigger('keydown',{key:'Home'});
    expect(wrapper.find('.fan-card.is-active').attributes('aria-label')).toBe('打开资源：a');
  } finally {wrapper.unmount();}
});

it('关闭牌组释放封面节点，再展开复用成员数据，不重复加载', async () => {
  const wrapper=mount(GroupPanel,{props:{filesBasesId:'lib',groupId:'group-1',displayActive:true},global:globals});
  try {
    await flushPromises();
    expect(wrapper.findAll('.fan-card')).toHaveLength(2);
    const calls=mocks.info.mock.calls.length;
    await wrapper.setProps({displayActive:false});
    expect(wrapper.findAll('.fan-card')).toHaveLength(0);
    await wrapper.setProps({displayActive:true});
    expect(wrapper.findAll('.fan-card')).toHaveLength(2);
    expect(mocks.info).toHaveBeenCalledTimes(calls);
  } finally { wrapper.unmount(); }
});

it('列表更新期间保持当前牌组稳定，下次展开刷新旧成员缓存', async () => {
  const wrapper=mount(GroupPanel,{props:{filesBasesId:'lib',groupId:'group-1',displayActive:true,sourceResource:a,expectedCount:2},global:globals});
  try {
    await flushPromises();
    const calls=mocks.info.mock.calls.length;
    mocks.info.mockResolvedValueOnce({status:true,data:[{...a,title:'更新后的标题'},b,c]});
    await wrapper.setProps({sourceResource:{...a,title:'更新后的标题'},expectedCount:3});
    expect(mocks.info).toHaveBeenCalledTimes(calls);
    expect(wrapper.findAll('.fan-card')).toHaveLength(2);
    await wrapper.setProps({displayActive:false});
    await wrapper.setProps({displayActive:true});
    await flushPromises();
    expect(mocks.info).toHaveBeenCalledTimes(calls+1);
    expect(wrapper.findAll('.fan-card')).toHaveLength(3);
    expect(wrapper.text()).toContain('更新后的标题');
  } finally {wrapper.unmount();}
});

it('解散确认等待期间切换组或卸载，不对后来切换的组发起解散', async () => {
  let confirm!:()=>void;
  vi.spyOn(ElMessageBox,'confirm').mockImplementation(()=>new Promise<void>(resolve=>{confirm=resolve;}) as unknown as ReturnType<typeof ElMessageBox.confirm>);
  const wrapper=mount(GroupPanel,{props:{filesBasesId:'lib',groupId:'group-1',editable:true},global:globals});
  try {
    await flushPromises();
    await wrapper.findAll('button').find(button=>button.text()==='解散分组')!.trigger('click');
    await wrapper.setProps({groupId:'other-group'});
    confirm();await flushPromises();
    expect(mocks.mutate).not.toHaveBeenCalled();
    await wrapper.findAll('button').find(button=>button.text()==='解散分组')!.trigger('click');
    wrapper.unmount();
    confirm();await flushPromises();
    expect(mocks.mutate).not.toHaveBeenCalled();
  } finally {if(wrapper.exists())wrapper.unmount();}
});

it('浮层在静止指针下展开，首次点击也按点击位置打开，键盘点击仍使用焦点资源', async () => {
  const wrapper=mount(GroupFan,{props:{members:[a,b,c],loading:false,error:''},global:globals});
  try {
    Object.defineProperty(wrapper.find('.fan-deck').element,'clientWidth',{value:500});
    Object.defineProperty(wrapper.find('.fan-card').element,'offsetWidth',{value:100});
    expect(wrapper.find('.fan-card.is-active').attributes('aria-label')).toBe('打开资源：b');
    wrapper.findAll('.fan-card')[0].element.dispatchEvent(new MouseEvent('click',{bubbles:true,detail:1,clientX:176}));
    await nextTick();
    expect(wrapper.emitted('select')?.[0]).toEqual([a]);
    await wrapper.findAll('.fan-card')[2].trigger('click');
    expect(wrapper.emitted('select')?.[1]).toEqual([c]);
  } finally { wrapper.unmount(); }
});

it('切换牌组只保留当前浮层，旧浮层清理不会影响新浮层', () => {
  const first = vi.fn(() => releaseResourceGroupFan(first));
  const second = vi.fn(() => releaseResourceGroupFan(second));
  try {
    claimResourceGroupFan(first);
    claimResourceGroupFan(first);
    expect(first).not.toHaveBeenCalled();
    claimResourceGroupFan(second);
    expect(first).toHaveBeenCalledTimes(1);
    releaseResourceGroupFan(first);
    claimResourceGroupFan(first);
    expect(second).toHaveBeenCalledTimes(1);
  } finally { releaseResourceGroupFan(first); releaseResourceGroupFan(second); }
});

it('密集牌组可用方向键与首尾键逐张切换，仅当前牌进入 Tab 顺序', async () => {
  const wrapper = mount(GroupFan, { props: { members:[a,b,c],loading:false,error:'' }, global:globals,attachTo:document.body });
  try {
    const deck=wrapper.find('.fan-deck');
    expect(wrapper.findAll('[tabindex="0"]')).toHaveLength(1);
    await deck.trigger('keydown',{key:'ArrowRight'});
    expect(wrapper.find('.fan-card.is-active').attributes('aria-label')).toBe('打开资源：c');
    expect(document.activeElement).toBe(wrapper.findAll('.fan-card')[2].element);
    await deck.trigger('keydown',{key:'ArrowRight'});
    expect(wrapper.find('.fan-card.is-active').attributes('aria-label')).toBe('打开资源：c');
    await deck.trigger('keydown',{key:'Home'});
    expect(wrapper.find('.fan-card.is-active').attributes('aria-label')).toBe('打开资源：a');
    await deck.trigger('keydown',{key:'End'});
    await deck.trigger('keydown',{key:'ArrowLeft'});
    expect(wrapper.find('.fan-card.is-active').attributes('aria-label')).toBe('打开资源：b');
    await wrapper.find('.fan-card.is-active').trigger('click');
    expect(wrapper.emitted('select')?.[0]).toEqual([b]);
  } finally { wrapper.unmount(); }
});

it('展开手牌时悬停和键盘焦点置顶对应成员，点击保留原资源响应', async () => {
  const wrapper = mount(GroupFan, { props: { members: [a,b,c], loading: false, error: '', matchedIds: new Set(['b']) }, global: globals });
  try {
    expect(wrapper.find('.fan-card.is-active').attributes('aria-label')).toBe('打开资源：b');
    Object.defineProperty(wrapper.find('.fan-deck').element, 'clientWidth', { value: 500 });
    Object.defineProperty(wrapper.find('.fan-card').element, 'offsetWidth', { value: 100 });
    await wrapper.find('.fan-deck').trigger('pointermove', { clientX: 100 });
    expect(wrapper.find('.fan-card.is-active').attributes('aria-label')).toBe('打开资源：a');
    expect(wrapper.find('.fan-card.is-active').attributes('style')).toContain('z-index: 4');
    await wrapper.findAll('.fan-card')[2].trigger('focus');
    expect(wrapper.find('.fan-card.is-active').attributes('aria-label')).toBe('打开资源：c');
    await wrapper.findAll('.fan-card')[2].trigger('click');
    expect(wrapper.emitted('select')?.[0]).toEqual([c]);
  } finally { wrapper.unmount(); }
});

it('整组始终单行密集展示，不分页不换行，末尾成员可打开，重试加载可用', async () => {
  const members = Array.from({ length: 40 }, (_, index) => resource(String(index)));
  const wrapper = mount(GroupFan, { props: { members, loading: false, error: '' }, global: globals });
  try {
    expect(wrapper.findAll('.fan-card')).toHaveLength(40);
    expect(wrapper.findAll('.fan-deck')).toHaveLength(1);
    expect(wrapper.classes()).toContain('is-dense');
    expect(wrapper.find('nav').exists()).toBe(false);
    expect(wrapper.find('.fan-caption').exists()).toBe(false);
    Object.defineProperty(wrapper.find('.fan-deck').element, 'clientWidth', { value: 720 });
    Object.defineProperty(wrapper.find('.fan-card').element, 'offsetWidth', { value: 132 });
    await wrapper.find('.fan-deck').trigger('pointermove', { clientX: 586 });
    expect(wrapper.find('.fan-card.is-active').attributes('aria-label')).toBe('打开资源：38');
    await wrapper.findAll('.fan-card')[39].trigger('focus');
    expect(wrapper.find('.fan-card.is-active').attributes('aria-label')).toBe('打开资源：39');
    await wrapper.findAll('.fan-card')[39].trigger('click');
    expect(wrapper.emitted('select')?.[0]).toEqual([members[39]]);
    await wrapper.setProps({ members: [a,b] });
    expect(wrapper.find('.fan-card').attributes('aria-label')).toBe('打开资源：a');
    await wrapper.setProps({ error: '加载失败' });
    expect(wrapper.findAll('.fan-card')).toHaveLength(0);
    await wrapper.find('button').trigger('click');
    expect(wrapper.emitted('retry')).toHaveLength(1);
  } finally { wrapper.unmount(); }
});

describe('点击式分组管理',()=>{
  it('两个未分组资源直接创建，不填写名称',async()=>{
    const wrapper=mount(GroupManager,{props:{filesBasesId:'lib',selected:[resource('a'),c]},global:globals});
    try{
      await managerButton(wrapper,'创建关联组').trigger('click');await flushPromises();
      expect(mocks.mutate).toHaveBeenCalledWith('create','lib',['a','c'],'');
      expect(wrapper.emitted('changed')).toHaveLength(1);
    }finally{wrapper.unmount();}
  });
  it('原组成员与新资源一起选中时直接加入，移出只作用于组内成员',async()=>{
    const wrapper=mount(GroupManager,{props:{filesBasesId:'lib',selected:[a,c]},global:globals});
    try{
      await managerButton(wrapper,'加入此组').trigger('click');await flushPromises();
      expect(mocks.mutate).toHaveBeenLastCalledWith('add','lib',['a','c'],'group-1');
      await managerButton(wrapper,'移出分组').trigger('click');await flushPromises();
      expect(mocks.mutate).toHaveBeenLastCalledWith('remove','lib',['a'],'group-1');
    }finally{wrapper.unmount();}
  });
  it('可选择其他页的已有组；不同组成员不会静默合并',async()=>{
    const wrapper=mount(GroupManager,{props:{filesBasesId:'lib',selected:[c],pageResources:[c]},global:globals});
    try{
      await managerButton(wrapper,'加入已有组').trigger('click');await flushPromises();
      expect(mocks.summaries).not.toHaveBeenCalled();
      expect(document.body.textContent).toContain('本页没有匹配的关联组');
      await wrapper.findComponent(ElRadioGroup).get('input[value=all]').setValue(true);
      await flushPromises();
      expect(mocks.summaries).toHaveBeenCalledWith('lib',[],'',1);
      (document.body.querySelector('.group-choice') as HTMLButtonElement).click();await flushPromises();
      expect(mocks.mutate).toHaveBeenLastCalledWith('add','lib',['c'],'group-1');
      await wrapper.setProps({selected:[a,resource('d','other-group')]});
      expect(wrapper.text()).toContain('暂不支持合并多个组');
      expect(wrapper.findAll('button').filter(button=>button.text()==='加入此组')).toHaveLength(0);
    }finally{wrapper.unmount();}
  });
  it('默认仅请求本页出现的组，切换范围可查看同库全部组，重新打开恢复本页',async()=>{
    const wrapper=mount(GroupManager,{props:{filesBasesId:'lib',selected:[c],pageResources:[a,b,c]},global:globals});
    try{
      await managerButton(wrapper,'加入已有组').trigger('click');await flushPromises();
      expect(mocks.summaries).toHaveBeenLastCalledWith('lib',['group-1'],'',1);
      expect(document.body.querySelector('.group-choice__info')?.textContent).toContain('本页');
      await wrapper.findComponent(ElRadioGroup).get('input[value=all]').setValue(true);await flushPromises();
      expect(mocks.summaries).toHaveBeenLastCalledWith('lib',[],'',1);
      (document.body.querySelector('.group-choice') as HTMLButtonElement).click();await flushPromises();
      await managerButton(wrapper,'加入已有组').trigger('click');await flushPromises();
      expect(mocks.summaries).toHaveBeenLastCalledWith('lib',['group-1'],'',1);
    }finally{wrapper.unmount();}
  });
  it('切换本页和全库范围时忽略旧请求，当前页移除全部组后不误查全库',async()=>{
    let oldResult!:(value:unknown)=>void;
    mocks.summaries.mockImplementationOnce(()=>new Promise(resolve=>{oldResult=resolve;}));
    const wrapper=mount(GroupManager,{props:{filesBasesId:'lib',selected:[c],pageResources:[a,b,c]},global:globals});
    try{
      await managerButton(wrapper,'加入已有组').trigger('click');
      mocks.summaries.mockResolvedValueOnce({status:true,data:{dataList:[{id:'other-group',count:2,preview:[resource('x')]}],total:1}});
      await wrapper.findComponent(ElRadioGroup).get('input[value=all]').setValue(true);await flushPromises();
      oldResult({status:true,data:{dataList:[{id:'group-1',count:2,preview:[a,b]}],total:1}});await flushPromises();
      expect(document.body.querySelector('.group-choice__info')?.textContent).toContain('x');
      await wrapper.findComponent(ElRadioGroup).get('input[value=page]').setValue(true);await flushPromises();
      const calls=mocks.summaries.mock.calls.length;
      await wrapper.setProps({pageResources:[c]});await flushPromises();
      expect(mocks.summaries).toHaveBeenCalledTimes(calls);
      expect(document.body.querySelectorAll('.group-choice')).toHaveLength(0);
    }finally{wrapper.unmount();}
  });
  it('整组面板支持排序和移出，剩余不足两项时关闭',async()=>{
    const wrapper=mount(GroupPanel,{props:{filesBasesId:'lib',groupId:'group-1',editable:true},global:globals});
    try{
      await flushPromises();
      await wrapper.get('[aria-label="上移：b"]').trigger('click');await flushPromises();
      expect(mocks.mutate).toHaveBeenLastCalledWith('reorder','lib',['b','a'],'group-1');
      await wrapper.findAll('button').find(button=>button.text()==='移出分组')!.trigger('click');await flushPromises();
      expect(mocks.mutate).toHaveBeenLastCalledWith('remove','lib',['b'],'group-1');
      expect(wrapper.emitted('closed')).toHaveLength(1);
    }finally{wrapper.unmount();}
  });

  it('选组分页加载和失败时清除旧页候选，不能继续选择上一页的组',async()=>{
    mocks.summaries.mockResolvedValue({status:true,data:{dataList:[{id:'group-1',count:2,preview:[a,b]}],total:41}});
    const wrapper=mount(GroupManager,{props:{filesBasesId:'lib',selected:[c],pageResources:[]},global:globals});
    try {
      await managerButton(wrapper,'加入已有组').trigger('click');await flushPromises();
      await wrapper.findComponent(ElRadioGroup).get('input[value=all]').setValue(true);await flushPromises();
      expect(document.body.querySelectorAll('.group-choice')).toHaveLength(1);
      let finish!:(value:unknown)=>void;
      mocks.summaries.mockImplementationOnce(()=>new Promise(resolve=>{finish=resolve;}));
      wrapper.findComponent(ElPagination).vm.$emit('current-change',2);await nextTick();
      expect(document.body.querySelectorAll('.group-choice')).toHaveLength(0);
      expect(wrapper.findComponent(ElPagination).props('disabled')).toBe(true);
      finish({status:false,msg:'第二页加载失败'});await flushPromises();
      expect(document.body.textContent).toContain('第二页加载失败');
      expect(document.body.querySelectorAll('.group-choice')).toHaveLength(0);
      expect(mocks.mutate).not.toHaveBeenCalled();
    } finally {wrapper.unmount();}
  });
});

it.each(['triptych','book','adaptive'] as const)('%s 封面取消原叠牌缩放和背牌，管理模式或未合并时不应用组外观',async(style)=>{
  const enabled=ref(true);
  const wrapper=mount(ContentMenu,{props:{resource:a},slots:{default:'<button class="original-play">播放</button>'},global:{...globals,provide:{[resourceGroupContextKey as symbol]:{enabled:computed(()=>enabled.value),pageIds:computed(()=>new Set(['a'])),summaries:ref(new Map())}},stubs:{contentMenu:{template:'<div><slot/></div>'},ResourceGroupPanel:true}}});
  try{
    expect(wrapper.findAll('.resource-group-back')).toHaveLength(2);
    appStoreData().currentConfigApp.resourceGroupCoverStyle=style;await nextTick();
    expect(wrapper.classes()).toContain('resource-group-composite');
    expect(wrapper.findAll('.resource-group-back')).toHaveLength(0);
    expect(wrapper.findAll('.original-play')).toHaveLength(1);
    enabled.value=false;await nextTick();
    expect(wrapper.classes()).not.toContain('resource-group-composite');
    expect(wrapper.find('.resource-group-front').exists()).toBe(false);
  }finally{wrapper.unmount();}
});

it('预取摘要与当前页匹配时直接使用，不再补发摘要请求；过期预取不影响新页',async()=>{
 const data=[a,b];
 const page={key:resourceGroupPageKey('lib',data),summaries:[{id:'group-1',count:7,filesBasesId:'lib',preview:[a,b]}]};
 const wrapper=shallowMount(ContentList,{props:{dataList:data,groupPage:page},global:globals});
 try{
  await flushPromises();expect(mocks.summaries).not.toHaveBeenCalled();
  await wrapper.setProps({dataList:[resource('next','next-group')]});await flushPromises();
  expect(mocks.summaries).toHaveBeenCalledWith('lib',['next-group']);
 }finally{wrapper.unmount();}
});

it('加入已有组的摘要即使有七张封面，选组缩略图仍只显示三张，不挤占组名',async()=>{
 mocks.summaries.mockResolvedValue({status:true,data:{dataList:[{id:'group-1',count:10,preview:Array.from({length:7},(_,i)=>resource(String(i)))}],total:1}});
 const wrapper=mount(GroupManager,{props:{filesBasesId:'lib',selected:[c],pageResources:[a]},global:globals});
 try{
  await managerButton(wrapper,'加入已有组').trigger('click');await flushPromises();
  expect(document.body.querySelectorAll('.group-choice__covers img')).toHaveLength(3);
  expect(document.body.textContent).toContain('10 个资源');
 }finally{wrapper.unmount();}
});
it('普通瀑布流使用统一组封面组件，保留资源与加载后布局刷新事件',async()=>{
 vi.useFakeTimers();const renderer=vi.fn();
 const wrapper=mount(WaterfallLayout,{props:{dataList:[a]},global:{...globals,stubs:{
  Waterfall:defineComponent({props:['list'],emits:[],template:'<div><slot :item="list[0]" :index="0"/></div>',setup(_props,{expose}){expose({renderer});return {};}}),
  contentRightClickMenu:{template:'<div><slot/></div>'},contentVideoDurationBadge:true,ElImage:true,ElBacktop:true,
 }}});
 try{
  const cover=wrapper.findComponent(GroupCover);expect(cover.exists()).toBe(true);
  expect(cover.props('resource').id).toBe('a');expect(cover.props('src')).toBe('/cover/a.jpg');
  await nextTick();const renderCall=vi.spyOn((wrapper.vm as any).$.setupState.waterfallRef,'renderer');
  cover.findComponent({name:'ElImage'}).vm.$emit('load',new Event('load'));await vi.runOnlyPendingTimersAsync();await nextTick();expect(renderCall).toHaveBeenCalled();
 }finally{wrapper.unmount();vi.useRealTimers();}
});

it.each(['manager','panel'])('%s 的旧库/旧组操作失败后，不在新库/新组弹出错误',async(kind)=>{
 const error=vi.spyOn(ElMessage,'error');let reject!: (reason:Error)=>void;
 mocks.mutate.mockImplementationOnce(()=>new Promise((_resolve,fail)=>{reject=fail;}));
 const wrapper=kind==='manager'
  ?mount(GroupManager,{props:{filesBasesId:'lib',selected:[resource('one'),resource('two')]},global:globals})
  :mount(GroupPanel,{props:{filesBasesId:'lib',groupId:'group-1',editable:true},global:globals});
 try{
  await flushPromises();
  const button=wrapper.findAll('button').find(button=>button.text()===(kind==='manager'?'创建关联组':'移出分组'))!;
  await button.trigger('click');
  await wrapper.setProps(kind==='manager'?{filesBasesId:'other'}:{groupId:'another-group'});
  reject(new Error('old request failed'));await flushPromises();expect(error).not.toHaveBeenCalled();
 }finally{wrapper.unmount();}
});

it('组详情旧请求晚到不能覆盖后来切换的组，下一次展开继续使用当前组缓存',async()=>{
 let finish!: (value:unknown)=>void;
 mocks.info.mockImplementationOnce(()=>new Promise(resolve=>{finish=resolve;}));
 const other=resource('other','group-2');
 mocks.info.mockResolvedValueOnce({status:true,data:[c,other]});
 const wrapper=mount(GroupPanel,{props:{filesBasesId:'lib',groupId:'group-1',displayActive:true},global:globals});
 try{
  await wrapper.setProps({groupId:'group-2'});await flushPromises();
  const names=()=>wrapper.findAll('.fan-card').map(card=>card.attributes('aria-label'));
  expect(names()).toEqual(['打开资源：c','打开资源：other']);
  finish({status:true,data:[a,b]});await flushPromises();expect(names()).toEqual(['打开资源：c','打开资源：other']);
  await wrapper.setProps({displayActive:false});await wrapper.setProps({displayActive:true});await flushPromises();
  expect(names()).toEqual(['打开资源：c','打开资源：other']);expect(mocks.info).toHaveBeenCalledTimes(2);
 }finally{wrapper.unmount();}
});
it('新的预取摘要就绪后，旧摘要请求晚到不能覆盖当前数量或封面',async()=>{
 let finish!: (value:unknown)=>void;let count=()=>0;
 mocks.summaries.mockImplementationOnce(()=>new Promise(resolve=>{finish=resolve;}));
 const layout=defineComponent({setup(){const context=inject(resourceGroupContextKey)!;count=()=>context.summaries.value.get('group-1')?.count||0;return()=>null;}});
 const data=[a,b];
 const wrapper=shallowMount(ContentList,{props:{dataList:data},global:{...globals,stubs:{layoutCoverPoster:layout}}});
 try{
  await flushPromises();expect(mocks.summaries).toHaveBeenCalledTimes(1);
  await wrapper.setProps({groupPage:{key:resourceGroupPageKey('lib',data),summaries:[{id:'group-1',filesBasesId:'lib',count:7,preview:[a,b]}]}});await flushPromises();expect(count()).toBe(7);
  finish({status:true,data:{dataList:[{id:'group-1',count:2,preview:[a,b]}],total:1}});await flushPromises();
  expect(count()).toBe(7);expect(mocks.summaries).toHaveBeenCalledTimes(1);
 }finally{wrapper.unmount();}
});
