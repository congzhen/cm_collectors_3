import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import ElementPlus, { ElInputNumber, ElMessage, ElSelect } from 'element-plus'
import Quick from '../fileDatabaseSetting/CategoryBrowseQuickSettings.vue'
import { appStoreData } from '@/storeData/app.storeData'
import { createDefaultConfigApp } from '@/dataType/config.dataType'
const mocks = vi.hoisted(() => ({ save: vi.fn(), tags: vi.fn(), emit: vi.fn() }))
vi.mock('@/main', () => ({ eventBus: { emit: mocks.emit } }))
vi.mock('@/assets/request', () => ({ default: vi.fn() }))
vi.mock('@/server/tag.server', () => ({ tagServer: { tagClassListByFilesBasesId: mocks.tags } }))
vi.mock('@/server/categoryBrowse.server', async () => {
  const original = await vi.importActual<typeof import('@/server/categoryBrowse.server')>(
    '@/server/categoryBrowse.server',
  )
  return { ...original, categoryBrowseServer: { settings: mocks.save } }
})
vi.mock('@/storeData/app.storeData', async () => {
  const { reactive } = await import('vue')
  const store = reactive({ currentFilesBases: { id: 'A' }, currentConfigApp: {} })
  return { appStoreData: () => store }
})
beforeEach(() => {
  vi.restoreAllMocks()
  vi.clearAllMocks()
  mocks.save.mockReset().mockResolvedValue({ status: true })
  mocks.tags.mockReset().mockResolvedValue({ status: true, data: [] })
  appStoreData().currentFilesBases.id = 'A'
  appStoreData().currentConfigApp = createDefaultConfigApp()
})
const create = () =>
  mount(Quick, { props: { horizontal: true }, global: { plugins: [ElementPlus] } })
it.each(['performer', 'tag:type'])('演员和标签均可保存独立分类排序：%s', async (source) => {
  mocks.tags.mockResolvedValue({ status: true, data: [{ id: 'type', name: '类型', status: true }] })
  Object.assign(appStoreData().currentConfigApp, {
    categoryBrowseEnabled: true,
    categoryBrowseSource: source,
  })
  const w = create()
  try {
    await flushPromises()
    const sort = w.findAll('.category-field')[2]!.findComponent(ElSelect)
    expect(sort.props('modelValue')).toBe('count')
    for (const mode of ['name', 'default', 'count']) {
      sort.vm.$emit('update:modelValue', mode)
      await flushPromises()
      await w.get('.category-apply').trigger('click')
      await flushPromises()
      expect(mocks.save).toHaveBeenLastCalledWith(
        'A', expect.objectContaining({ categoryBrowseSource: source, categoryBrowseSort: mode }),
      )
      expect(appStoreData().currentConfigApp.categoryBrowseSort).toBe(mode)
    }
  } finally {
    w.unmount()
  }
})
it.each([0, 1, 100])('允许保存每组上限 %s，应用后保留原值', async (cap) => {
  const w = create()
  try {
    await flushPromises()
    await w.get('[role="switch"]').trigger('click')
    const input = w.findComponent(ElInputNumber).get('input')
    await input.setValue(String(cap))
    await input.trigger('change')
    await w.get('.category-apply').trigger('click')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledWith(
      'A',
      expect.objectContaining({ categoryBrowseGroupLimit: cap }),
    )
    expect(appStoreData().currentConfigApp.categoryBrowseGroupLimit).toBe(cap)
    expect(w.findComponent(ElInputNumber).props('modelValue')).toBe(cap)
  } finally {
    w.unmount()
  }
})
it('关闭时表单保留禁用，开启后空上限阻止提交，修正后仅应用分类字段', async () => {
  const w = create()
  try {
    await flushPromises()
    expect(w.findAll('.category-field')).toHaveLength(6)
    expect(w.findComponent(ElInputNumber).props('disabled')).toBe(true)
    await w.get('[role="switch"]').trigger('click')
    const number = w.findComponent(ElInputNumber)
    expect(number.get('input').attributes('aria-disabled')).toBe('false')
    await number.get('input').setValue('')
    await number.get('input').trigger('change')
    await w.get('.category-apply').trigger('click')
    await flushPromises()
    expect(mocks.save).not.toHaveBeenCalled()
    expect(w.text()).toContain('每组资源上限请输入')
    await number.get('input').setValue('8')
    await number.get('input').trigger('change')
    await w.get('.category-apply').trigger('click')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledWith(
      'A',
      expect.objectContaining({ categoryBrowseEnabled: true, categoryBrowseGroupLimit: 8 }),
    )
    expect(mocks.save.mock.calls[0][1]).not.toHaveProperty('pageLimit')
    expect(appStoreData().currentConfigApp.categoryBrowseGroupLimit).toBe(8)
  } finally {
    w.unmount()
  }
})
it('旧库保存失败不能提示到新库，也不能解除新库保存状态', async () => {
  let first!: (value: unknown) => void, second!: (value: unknown) => void
  mocks.save
    .mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          first = resolve
        }),
    )
    .mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          second = resolve
        }),
    )
  const error = vi.spyOn(ElMessage, 'error')
  const w = create()
  try {
    await flushPromises()
    await w.get('[role="switch"]').trigger('click')
    await w.get('.category-apply').trigger('click')
    appStoreData().currentFilesBases.id = 'B'
    appStoreData().currentConfigApp = createDefaultConfigApp()
    await flushPromises()
    await w.get('.category-apply').trigger('click')
    first({ status: false, msg: '旧库失败' })
    await flushPromises()
    expect(error).not.toHaveBeenCalled()
    expect(w.find('.category-apply.is-loading').exists()).toBe(true)
    second({ status: true })
    await flushPromises()
    expect(w.find('.category-apply.is-loading').exists()).toBe(false)
    expect(appStoreData().currentConfigApp.categoryBrowseEnabled).toBe(false)
  } finally {
    w.unmount()
  }
})
it('同库配置被等值替换后可重新应用排序，旧保存不影响新请求', async () => {
  let first!: (value: unknown) => void, second!: (value: unknown) => void
  mocks.save
    .mockImplementationOnce(() => new Promise((resolve) => { first = resolve }))
    .mockImplementationOnce(() => new Promise((resolve) => { second = resolve }))
  Object.assign(appStoreData().currentConfigApp, { categoryBrowseEnabled: true })
  const error = vi.spyOn(ElMessage, 'error')
  const w = create()
  try {
    await flushPromises()
    await w.get('.category-apply').trigger('click')
    expect(w.find('.category-apply.is-loading').exists()).toBe(true)
    const replacement = { ...appStoreData().currentConfigApp }
    appStoreData().currentConfigApp = replacement
    await flushPromises()
    expect(w.find('.category-apply.is-loading').exists()).toBe(false)
    w.findAll('.category-field')[2]!.findComponent(ElSelect).vm.$emit('update:modelValue', 'name')
    await flushPromises()
    await w.get('.category-apply').trigger('click')
    expect(mocks.save).toHaveBeenCalledTimes(2)
    first({ status: false, msg: '过期保存失败' })
    await flushPromises()
    expect(error).not.toHaveBeenCalled()
    expect(w.find('.category-apply.is-loading').exists()).toBe(true)
    second({ status: true })
    await flushPromises()
    expect(w.find('.category-apply.is-loading').exists()).toBe(false)
    expect(appStoreData().currentConfigApp.categoryBrowseSort).toBe('name')
  } finally {
    w.unmount()
  }
})
it('标签尚未加载时保留设置并提示等待，加载成功后可以应用', async () => {
  let finish!: (value: unknown) => void
  mocks.tags.mockReturnValueOnce(
    new Promise((resolve) => {
      finish = resolve
    }),
  )
  Object.assign(appStoreData().currentConfigApp, {
    categoryBrowseEnabled: true,
    categoryBrowseSource: 'tag:type',
  })
  const w = create()
  try {
    await w.get('.category-apply').trigger('click')
    await flushPromises()
    expect(w.text()).toContain('标签分类正在加载')
    expect(mocks.save).not.toHaveBeenCalled()
    finish({ status: true, data: [{ id: 'type', name: '类型', status: true }] })
    await flushPromises()
    expect(w.text()).not.toContain('标签分类正在加载')
    await w.get('.category-apply').trigger('click')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledOnce()
  } finally {
    w.unmount()
  }
})
