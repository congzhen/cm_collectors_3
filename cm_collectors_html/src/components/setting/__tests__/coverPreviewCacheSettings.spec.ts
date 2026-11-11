import { shallowMount, flushPromises } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Settings from '../CoverPreviewCacheSettings.vue'
import request from '@/assets/request'
import { ElMessageBox } from 'element-plus'
vi.mock('@/assets/request', () => ({ default: vi.fn() }))
vi.mock('element-plus', () => ({ElMessage: {success: vi.fn(), error: vi.fn(), warning: vi.fn()}, ElMessageBox: {confirm: vi.fn()}}))
const mountPage=()=>shallowMount(Settings,{global:{directives:{loading:{}},stubs:{ElInputNumber:{name:'ElInputNumber',props:['modelValue'],template:'<input />'},ElForm:{template:'<div><slot /></div>'},ElFormItem:{template:'<div><slot /></div>'},ElButton:{template:'<button><slot /></button>'}}}})
const ok=(data: unknown)=>({status:true,statusCode:200,msg:'',data})
beforeEach(()=>{vi.resetAllMocks();vi.mocked(request).mockImplementation(async (config)=>ok(config.url?.endsWith('/settings')?{maxResources:5}:{count:3,bytes:1024}) as never)})
describe('封面预览缓存设置',()=>{
 it('读取设置并单独保存资源数，刷新状态保留未保存输入',async()=>{
  const w=mountPage();await flushPromises();const input=w.findComponent({name:'ElInputNumber'});input.vm.$emit('update:modelValue',8);await flushPromises()
  await w.findAll('button')[1].trigger('click');await flushPromises();expect(input.props('modelValue')).toBe(8)
  await w.findAll('button')[0].trigger('click');await flushPromises()
  expect(request).toHaveBeenCalledWith({url:'/coverPreviewCache/settings',method:'put',data:{maxResources:8}});w.unmount()
 })
 it('取消清空不请求，确认仅调用封面缓存清理',async()=>{
  const w=mountPage();await flushPromises();vi.mocked(ElMessageBox.confirm).mockRejectedValueOnce('cancel')
  await w.findAll('button')[2].trigger('click');await flushPromises()
  expect(vi.mocked(request).mock.calls.some(([c])=>c.method==='post')).toBe(false)
  vi.mocked(ElMessageBox.confirm).mockResolvedValue('confirm' as Awaited<ReturnType<typeof ElMessageBox.confirm>>)
  await w.findAll('button')[2].trigger('click');await flushPromises()
  expect(request).toHaveBeenCalledWith({url:'/coverPreviewCache/clear',method:'post'});w.unmount()
 })
 it('设置加载失败不能将默认值误保存',async()=>{
  vi.mocked(request).mockResolvedValue({status:false,statusCode:500,msg:'读取失败',data:null})
  const w=mountPage();await flushPromises();await w.findAll('button')[0].trigger('click');await flushPromises()
  expect(vi.mocked(request).mock.calls.some(([c])=>c.method==='put')).toBe(false);expect(w.text()).toContain('读取失败');w.unmount()
 })
 it('保存失败保留输入，解除忙碌状态后可重试',async()=>{
  const w=mountPage();await flushPromises();const input=w.findComponent({name:'ElInputNumber'});input.vm.$emit('update:modelValue',9);await flushPromises()
  vi.mocked(request).mockResolvedValueOnce({status:false,statusCode:500,msg:'保存失败',data:null})
  await w.findAll('button')[0].trigger('click');await flushPromises();expect(input.props('modelValue')).toBe(9)
  await w.findAll('button')[0].trigger('click');await flushPromises()
  expect(vi.mocked(request).mock.calls.filter(([c])=>c.method==='put')).toHaveLength(2);expect(input.props('modelValue')).toBe(9);w.unmount()
 })

})
