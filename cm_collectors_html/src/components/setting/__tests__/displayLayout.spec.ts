import { mount, shallowMount } from '@vue/test-utils'
import { h, ref, nextTick } from 'vue'
import { createPinia } from 'pinia'
import { ElForm, ElInput } from 'element-plus'
import { expect, it } from 'vitest'
import SettingFieldsScope from '../SettingFieldsScope.vue'
import SharedDisplayFields from '../sharedConfig/SharedDisplayFields.vue'
import { createDefaultConfigApp } from '@/dataType/config.dataType'

it('跟随公共配置仅禁用通用项，本库独立项保持可编辑，退出跟随后恢复编辑', async () => {
  const following = ref(true)
  const w = mount({ render: () => h(ElForm, {}, () => [
    h(SettingFieldsScope, { disabled: following.value }, () => h(ElInput, { modelValue: '公共参数' })),
    h(ElInput, { modelValue: '本库参数' }),
  ]) })
  expect(w.findAll('input')[0].element.disabled).toBe(true)
  expect(w.findAll('input')[1].element.disabled).toBe(false)
  expect(w.findAll('form')).toHaveLength(1)
  following.value = false
  await nextTick()
  expect(w.findAll('input')[0].element.disabled).toBe(false)
  w.unmount()
})

it('独立项插回对应功能分组，公共配置编辑器不出现本库独立项', () => {
  const options = { props: { config: createDefaultConfigApp(), readonly: true }, global: {
    plugins: [createPinia()], renderStubDefaultSlot: true,
  } }
  const w = shallowMount(SharedDisplayFields, { ...options, slots: {
    actors: '<span>LOCAL_ACTORS</span>', tags: '<span>LOCAL_TAGS</span>',
    sample: '<span>LOCAL_SAMPLE</span>', avatar: '<span>LOCAL_AVATAR</span>', poster: '<span>LOCAL_POSTER</span>',
  } })
  const html = w.html()
  for (const [before, local, after] of [
    ['左侧边栏', 'LOCAL_ACTORS', '显示设置'],
    ['封面上显示标签(属性)', 'LOCAL_TAGS', '标签背景色'],
    ['剧照设置', 'LOCAL_SAMPLE', '参数设置'],
    ['演员&amp;导演自定义', 'LOCAL_AVATAR', '插件设置'],
    ['封面海报设置', 'LOCAL_POSTER', '封面海报显示宽度'],
  ]) {
    expect(html.indexOf(before)).toBeGreaterThanOrEqual(0)
    expect(html.indexOf(local)).toBeGreaterThan(html.indexOf(before))
    expect(html.indexOf(after)).toBeGreaterThan(html.indexOf(local))
  }
  expect(w.text()).toContain('通用参数来自公共配置')
  w.unmount()
  const editor = shallowMount(SharedDisplayFields, options)
  expect(editor.html()).not.toContain('LOCAL_')
  editor.unmount()
})
