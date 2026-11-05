import { flushPromises, shallowMount } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import ElementPlus, { ElAlert } from 'element-plus'
import { createPinia } from 'pinia'
import FileSettings from '../fileDatabaseSetting/fileSettingData.vue'
import Fields from '../fileDatabaseSetting/CategoryBrowseFields.vue'
const mocks = vi.hoisted(() => ({ info: vi.fn(), save: vi.fn(), tags: vi.fn() }))
vi.mock('@/main', () => ({ eventBus: { on: vi.fn(), off: vi.fn(), emit: vi.fn() } }))
vi.mock('@/assets/debounce', () => ({ debounceNow: (fn: unknown) => fn }))
vi.mock('@/server/filesBases.server', () => ({
  filesBasesServer: { infoById: mocks.info, setData: mocks.save },
}))
vi.mock('@/server/tag.server', () => ({ tagServer: { tagClassListByFilesBasesId: mocks.tags } }))
vi.mock('@/storeData/filesBases.storeData', () => ({
  filesBasesStoreData: () => ({ getMainPerformerBasesIdByFilesBasesId: () => 'actors' }),
}))
vi.mock('@/storeData/performerBases.storeData', () => ({
  performerBasesStoreData: () => ({ listByIds: () => [], performerBases: [] }),
}))
beforeEach(() => {
  vi.clearAllMocks()
  mocks.info.mockResolvedValue({
    status: true,
    data: {
      id: 'lib',
      name: '本库',
      status: true,
      filesRelatedPerformerBases: [{ performerBases_id: 'actors', main: true }],
      filesBasesSetting: {
        config_json_data: JSON.stringify({
          categoryBrowseEnabled: true,
          categoryBrowseSource: 'performer',
          categoryBrowseGroupLimit: 12,
        }),
      },
    },
  })
  mocks.save.mockResolvedValue({ status: true })
  mocks.tags.mockResolvedValue({ status: true, data: [{ id: 'type', status: true, name: '类型' }] })
})
const create = () =>
  shallowMount(FileSettings, {
    props: { filesBasesId: 'lib' },
    global: {
      plugins: [createPinia(), ElementPlus],
      renderStubDefaultSlot: true,
      stubs: {
        ElAlert: false,
        SharedDisplayFields: { template: '<div><slot name="resourceGroups" /></div>' },
        CategoryBrowseFields: false,
      },
    },
  })
const save = (w: ReturnType<typeof create>) =>
  w
    .findAllComponents({ name: 'ElButton' })
    .find((b) => b.text() === '保存本库设置')!
    .vm.$emit('click')
it.each([0, 1, 100])('完整文件库表单保存上限 %s 时保留真实值', async (cap) => {
  const w = create()
  try {
    await flushPromises()
    w.findComponent(Fields).props('config').categoryBrowseGroupLimit = cap
    save(w)
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledWith(
      'lib',
      expect.anything(),
      expect.objectContaining({ categoryBrowseGroupLimit: cap }),
      'actors',
      ['actors'],
    )
    expect(w.emitted('setSuccess')).toEqual([['lib']])
  } finally {
    w.unmount()
  }
})
it.each([undefined, -1, 101, 1.5])('完整文件库表单拒绝无效上限 %s，保留修改供修正', async (cap) => {
  const w = create()
  try {
    await flushPromises()
    const config = w.findComponent(Fields).props('config')
    const scroll = vi.fn()
    Object.defineProperty(w.findComponent(Fields).element, 'scrollIntoView', {
      value: scroll,
      configurable: true,
    })
    Object.assign(config, { categoryBrowseGroupLimit: cap })
    save(w)
    await flushPromises()
    expect(mocks.save).not.toHaveBeenCalled()
    expect(w.findComponent(Fields).text()).toContain('0～100')
    expect(config.categoryBrowseGroupLimit).toBe(cap)
    expect(scroll).toHaveBeenCalledWith({ block: 'nearest' })
  } finally {
    w.unmount()
  }
})
it('无效标签分类不能保存，但仍可关闭分类浏览并保存', async () => {
  const w = create()
  try {
    await flushPromises()
    const config = w.findComponent(Fields).props('config')
    config.categoryBrowseSource = 'tag:deleted'
    save(w)
    await flushPromises()
    expect(mocks.save).not.toHaveBeenCalled()
    config.categoryBrowseEnabled = false
    save(w)
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledOnce()
  } finally {
    w.unmount()
  }
})
it.each(['status', 'network', 'json', 'array', 'null', 'string'])(
  '加载 %s 失败后可重试，加载成功前不会保存空配置',
  async (kind) => {
    if (kind === 'status') mocks.info.mockResolvedValueOnce({ status: false, msg: '获取失败' })
    if (kind === 'network') mocks.info.mockRejectedValueOnce(new Error('offline'))
    if (['json', 'array', 'null', 'string'].includes(kind))
      mocks.info.mockResolvedValueOnce({
        status: true,
        data: {
          id: 'lib',
          filesRelatedPerformerBases: [],
          filesBasesSetting: {
            config_json_data: (
              { json: '{invalid', array: '[]', null: 'null', string: '"invalid"' } as Record<
                string,
                string
              >
            )[kind],
          },
        },
      })
    const w = create()
    try {
      await flushPromises()
      expect(w.findComponent(ElAlert).exists()).toBe(true)
      expect(w.get('[role="alert"]').text()).toContain('文件库设置加载失败')
      expect(w.findAllComponents({name:'ElButton'}).find(b=>b.text()==='重试')!.props('loading')).toBe(false)
      expect(w.findComponent(Fields).exists()).toBe(false)
      const button = w
        .findAllComponents({ name: 'ElButton' })
        .find((b) => b.text() === '保存本库设置')!
      expect(button.props('disabled')).toBe(true)
      save(w)
      await flushPromises()
      expect(mocks.save).not.toHaveBeenCalled()
      w.findAllComponents({ name: 'ElButton' })
        .find((b) => b.text() === '重试')!
        .vm.$emit('click')
      await flushPromises()
      expect(w.findComponent(ElAlert).exists()).toBe(false)
      expect(w.findComponent(Fields).exists()).toBe(true)
      expect(button.props('disabled')).toBeFalsy()
      save(w)
      await flushPromises()
      expect(mocks.save).toHaveBeenCalledOnce()
    } finally {
      w.unmount()
    }
  },
)
