import { mount, flushPromises } from '@vue/test-utils'
import { h, reactive } from 'vue'
import { describe, expect, it, vi } from 'vitest'
import ElementPlus, { ElFormItem, ElInput } from 'element-plus'
import DrawerForm from '../../com/dialog/drawer-form.vue'
import DrawerCommon from '../../com/dialog/drawer-common.vue'
vi.mock('@/assets/debounce',()=>({debounceNow:(fn:unknown)=>fn}))
vi.mock('@/language/app.lang',()=>({AppLang:()=>({btn:(key:string)=>key})}))
describe('抽屉必填校验',()=>{
 it.each(['title','name'])('%s 为空仅展示校验提示，补全后提交一次',async field=>{
  const model=reactive<Record<string,string>>({[field]:''})
  const wrapper=mount(DrawerForm,{props:{modelValue:model,rules:{[field]:[{required:true,message:'必填',trigger:'blur'}]}},slots:{default:()=>h(ElFormItem,{prop:field},()=>h(ElInput,{modelValue:model[field]}))},global:{plugins:[ElementPlus]}})
  try{
   wrapper.vm.open();await flushPromises();wrapper.findComponent(DrawerCommon).vm.$emit('submit');await flushPromises()
   expect(wrapper.emitted('submit')).toBeUndefined();expect(document.querySelector('.el-form-item.is-error')).not.toBeNull()
   model[field]='已填写';await flushPromises();wrapper.findComponent(DrawerCommon).vm.$emit('submit');await flushPromises()
   expect(wrapper.emitted('submit')).toHaveLength(1)
  }finally{wrapper.unmount()}
 })
})
