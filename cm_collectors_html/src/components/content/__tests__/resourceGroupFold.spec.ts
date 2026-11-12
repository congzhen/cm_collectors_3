import { mount } from '@vue/test-utils';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import ElementPlus from 'element-plus';
import Fold from '../ResourceGroupFold.vue';
import type { I_resource } from '@/dataType/resource.dataType';
vi.mock('@/common/photo',()=>({getResourceCoverPoster:(r:I_resource)=>`/cover/${r.id}`}));
const resources=(count:number)=>Array.from({length:count},(_,i)=>({id:String(i),title:`资源 ${i}`} as I_resource));
beforeEach(()=>vi.stubGlobal('ResizeObserver',class{observe(){}disconnect(){}}));
afterEach(()=>vi.unstubAllGlobals());
const create=(count=5)=>mount(Fold,{props:{members:resources(count),loading:false,error:''},global:{plugins:[ElementPlus],stubs:{ElImage:true}}});
const active=(w:ReturnType<typeof create>)=>w.find('.fold-slice.is-active').attributes('aria-label');
const bounds=(w:ReturnType<typeof create>)=>vi.spyOn(w.find('.fold-screen').element,'getBoundingClientRect').mockReturnValue({left:100,top:0,width:480,height:236,right:580,bottom:236,x:100,y:0,toJSON(){}});
it('实际来源封面控制自适应折屏大小，关闭自适应后释放宽度覆盖', async () => {
  const source = document.createElement('div');
  let resizeSource: (() => void) | undefined;
  let sourceHeight = 180;
  vi.spyOn(source, 'getBoundingClientRect').mockImplementation(() => ({height:sourceHeight} as DOMRect));
  vi.stubGlobal('ResizeObserver', class {
    constructor(private callback: () => void) {}
    observe(element: Element) { if (element === source) resizeSource = this.callback; }
    disconnect() {}
  });
  const w = create(1);
  try {
    await w.setProps({sourceElement:source,adaptCoverRatio:true,members:[{...resources(1)[0],coverPosterWidth:900,coverPosterHeight:1600}]});
    expect((w.get('.fold-frame').element as HTMLElement).style.height).toBe('162px');
    expect(w.emitted('layoutWidth')?.at(-1)).toEqual([162 * 900 / 1600 + 28]);
    sourceHeight = 300;
    resizeSource?.();
    await w.vm.$nextTick();
    expect((w.get('.fold-frame').element as HTMLElement).style.height).toBe('270px');
    expect(w.emitted('layoutWidth')?.at(-1)).toEqual([270 * 900 / 1600 + 28]);
    await w.setProps({adaptCoverRatio:false});
    expect(w.emitted('layoutWidth')?.at(-1)).toEqual([undefined]);
    expect((w.get('.fold-frame').element as HTMLElement).style.height).toBe('');
  } finally { w.unmount(); }
});
it('自适应折屏随选中封面缩短可见高度，横竖封面均不留下空白', async () => {
  const rect = vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({left:0,top:0,width:292,height:238,right:292,bottom:238,x:0,y:0,toJSON(){}});
  const w = create(2);
  try {
    await w.setProps({adaptCoverRatio:true,members:resources(2).map((r,i)=>({...r,coverPosterWidth:i ? 900 : 1600,coverPosterHeight:i ? 1600 : 900}))});
    for (const [key, ratio] of [['Home',1600/900],['End',900/1600]] as const) {
      await w.find('.fold-screen').trigger('keydown',{key});
      const width = parseFloat((w.find('.fold-slice.is-active').element as HTMLElement).style.width);
      const height = parseFloat((w.find('.fold-screen').element as HTMLElement).style.height);
      expect(width / height).toBeCloseTo(ratio);
      expect(height).toBeLessThanOrEqual(238);
      expect(w.find('.fold-frame').attributes('style')).toBeUndefined();
    }
    await w.setProps({adaptCoverRatio:false});
    expect((w.find('.fold-screen').element as HTMLElement).style.height).toBe('');
  } finally { w.unmount(); rect.mockRestore(); }
});
it('纵向书脊遇到超宽封面时同步缩短高度，避免窗口限宽后留下空白', async () => {
  vi.stubGlobal('innerWidth', 390);
  const source = document.createElement('div');
  vi.spyOn(source, 'getBoundingClientRect').mockReturnValue({height:300} as DOMRect);
  const w = create(2);
  try {
    await w.setProps({sourceElement:source,adaptCoverRatio:true,vertical:true,members:resources(2).map(r=>({...r,coverPosterWidth:4000,coverPosterHeight:600}))});
    const frameHeight = parseFloat((w.get('.fold-frame').element as HTMLElement).style.height);
    expect(frameHeight).toBeCloseTo(((390 - 52) / (4000 / 600)) / .66 + 2);
    expect(w.emitted('layoutWidth')?.at(-1)).toEqual([366]);
  } finally { w.unmount(); }
});
it('折屏默认展开原资源，首尾切片完整展开时仍在整体边界内',async()=>{
  const w=create();
  try{
    await w.setProps({sourceResource:w.props('members')[0]});
    expect(active(w)).toBe('打开资源：资源 0');
    expect(parseFloat((w.find('.fold-slice.is-active').element as HTMLElement).style.left)).toBe(0);
    await w.find('.fold-screen').trigger('keydown',{key:'End'});
    const last=w.find('.fold-slice.is-active').element as HTMLElement;
    expect(active(w)).toBe('打开资源：资源 4');
    expect(parseFloat(last.style.left)+parseFloat(last.style.width)).toBeCloseTo(480);
  }finally{w.unmount();}
});
it('固定命中区不随展开动画移动，按下后锁定资源再点击',async()=>{
  const w=create();
  try{
    bounds(w);
    await w.find('.fold-screen').trigger('pointermove',{clientX:148,clientY:90});
    expect(active(w)).toBe('打开资源：资源 0');
    await w.find('.fold-screen').trigger('pointermove',{clientX:532,clientY:90});
    expect(active(w)).toBe('打开资源：资源 4');
    const last=w.findAll('.fold-slice')[4];
    last.element.dispatchEvent(new MouseEvent('pointerdown',{bubbles:true,cancelable:true,button:0,clientX:532,clientY:90}));
    last.element.dispatchEvent(new MouseEvent('click',{bubbles:true,detail:1,clientX:531,clientY:90}));
    expect(w.emitted('select')?.[0]).toEqual([w.props('members')[4]]);
  }finally{w.unmount();}
});
it('密集组全部保留，滚轮与键盘切换后静止鼠标不抢回选择',async()=>{
  const w=create(1000);
  try{
    bounds(w);
    expect(w.findAll('.fold-slice')).toHaveLength(1000);
    const screen=w.find('.fold-screen');
    await screen.trigger('pointermove',{clientX:340,clientY:90});
    await screen.trigger('wheel',{deltaY:40});
    expect(active(w)).toBe('打开资源：资源 501');
    await screen.trigger('pointermove',{clientX:340,clientY:90});
    expect(active(w)).toBe('打开资源：资源 501');
    await screen.trigger('keydown',{key:'Home'});
    await screen.trigger('pointermove',{clientX:340,clientY:90});
    expect(active(w)).toBe('打开资源：资源 0');
    await w.find('.fold-slice.is-active').trigger('click');
    expect(w.emitted('select')?.[0]).toEqual([w.props('members')[0]]);
  }finally{w.unmount();}
});
it('加载失败可以重试，空组和加载期间不显示旧切片',async()=>{
  const w=create();
  try{
    await w.setProps({loading:true});expect(w.findAll('.fold-slice')).toHaveLength(0);
    await w.setProps({loading:false,error:'加载失败'});
    await w.find('button').trigger('click');expect(w.emitted('retry')).toHaveLength(1);
    await w.setProps({error:'',members:[]});expect(w.text()).toContain('暂无关联资源');
  }finally{w.unmount();}
});
it('切片动画中点击落到容器或细缝上，仍打开按下时选中的资源',async()=>{
  const w=create();
  try{
    bounds(w);
    const screen=w.find('.fold-screen');
    await screen.trigger('pointermove',{clientX:340,clientY:90});
    await screen.trigger('wheel',{deltaY:40});
    expect(active(w)).toBe('打开资源：资源 3');
    screen.element.dispatchEvent(new MouseEvent('pointerdown',{bubbles:true,cancelable:true,button:0,clientX:340,clientY:90}));
    screen.element.dispatchEvent(new MouseEvent('click',{bubbles:true,detail:1,clientX:340,clientY:90}));
    expect(w.emitted('select')?.[0]).toEqual([w.props('members')[3]]);
  }finally{w.unmount();}
});
it('键盘焦点进入折屏后，滚轮和鼠标切换同步焦点，Enter 激活当前预览',async()=>{
  const w=mount(Fold,{props:{members:resources(5),loading:false,error:''},attachTo:document.body,global:{plugins:[ElementPlus],stubs:{ElImage:true}}});
  try{
    bounds(w);
    (w.findAll('.fold-slice')[2].element as HTMLElement).focus();
    const screen=w.find('.fold-screen');
    await screen.trigger('wheel',{deltaY:40});
    expect(active(w)).toBe('打开资源：资源 3');
    expect(document.activeElement).toBe(w.findAll('.fold-slice')[3].element);
    (document.activeElement as HTMLElement).click();
    expect(w.emitted('select')?.[0]).toEqual([w.props('members')[3]]);
    await screen.trigger('pointermove',{clientX:148,clientY:90});
    expect(active(w)).toBe('打开资源：资源 0');
    expect(document.activeElement).toBe(w.findAll('.fold-slice')[0].element);
  }finally{w.unmount();}
});
it('折屏方向键不冒泡到播放器快捷键',async()=>{
  const w=create();const listener=vi.fn();
  w.element.addEventListener('keydown',listener);
  try{
    await w.find('.fold-screen').trigger('keydown',{key:'ArrowRight'});
    expect(listener).not.toHaveBeenCalled();
  }finally{w.unmount();}
});
it('纵向书脊按上下位置选择，首尾展开仍在纵向边界内，上下键切换',async()=>{
  const w=create();
  try{
    await w.setProps({vertical:true});
    vi.spyOn(w.find('.fold-screen').element,'getBoundingClientRect').mockReturnValue({left:100,top:20,width:292,height:480,right:392,bottom:500,x:100,y:20,toJSON(){}});
    expect(w.classes()).toContain('group-fold--spine');
    const screen=w.find('.fold-screen');
    await screen.trigger('pointermove',{clientX:180,clientY:68});
    expect(active(w)).toBe('打开资源：资源 0');
    expect(parseFloat((w.find('.fold-slice.is-active').element as HTMLElement).style.top)).toBe(0);
    await screen.trigger('pointermove',{clientX:180,clientY:452});
    expect(active(w)).toBe('打开资源：资源 4');
    const last=w.find('.fold-slice.is-active').element as HTMLElement;
    expect(parseFloat(last.style.top)+parseFloat(last.style.height)).toBeCloseTo(480);
    await screen.trigger('keydown',{key:'ArrowUp'});expect(active(w)).toBe('打开资源：资源 3');
    await screen.trigger('keydown',{key:'ArrowDown'});expect(active(w)).toBe('打开资源：资源 4');
    await w.find('.fold-slice.is-active').trigger('click');
    expect(w.emitted('select')?.[0]).toEqual([w.props('members')[4]]);
  }finally{w.unmount();}
});
it('纵向书脊滚轮选择后指针不动仍保持预览，切换加载状态不遗留切片',async()=>{
  const w=create(40);
  try{
    await w.setProps({vertical:true});
    vi.spyOn(w.find('.fold-screen').element,'getBoundingClientRect').mockReturnValue({left:100,top:20,width:292,height:480,right:392,bottom:500,x:100,y:20,toJSON(){}});
    const screen=w.find('.fold-screen');
    await screen.trigger('pointermove',{clientX:180,clientY:266});
    await screen.trigger('wheel',{deltaY:40});
    expect(active(w)).toBe('打开资源：资源 21');
    await screen.trigger('pointermove',{clientX:180,clientY:266});
    expect(active(w)).toBe('打开资源：资源 21');
    screen.element.dispatchEvent(new MouseEvent('pointerdown',{bubbles:true,cancelable:true,button:0,clientX:180,clientY:266}));
    screen.element.dispatchEvent(new MouseEvent('click',{bubbles:true,detail:1,clientX:180,clientY:266}));
    expect(w.emitted('select')?.[0]).toEqual([w.props('members')[21]]);
    await w.setProps({loading:true,expectedCount:40});expect(w.findAll('.fold-slice')).toHaveLength(0);
  }finally{w.unmount();}
});
