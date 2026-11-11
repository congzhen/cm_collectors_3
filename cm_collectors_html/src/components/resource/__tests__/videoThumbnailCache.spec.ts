import { shallowMount, flushPromises } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import Preview from '../videoThumbnailPosterDialog.vue'
import { ElMessage } from 'element-plus'
import { ffmpegServer } from '@/server/ffmpeg.server'
vi.mock('@/server/ffmpeg.server', () => ({ ffmpegServer: { getVideoThumbnails: vi.fn() } }))
vi.mock('@/assets/debounce', () => ({ debounceNow: (fn: unknown) => fn }))
const dialog = defineComponent({ setup(_, { expose }) { expose({ open() {}, close() {}, disabledSubmit() {} }) }, template: '<div><slot /></div>' })
const mountPage = () => shallowMount(Preview, { global: { stubs: { dialogCommon: dialog, ElButton: { template: '<button><slot /></button>' } }, directives: { loading: {} } } })
const ok = (data: string[]) => ({ status: true, statusCode: 200, msg: '', data })
beforeEach(() => vi.resetAllMocks())
describe('封面预览缓存请求', () => {
 it('打开使用缓存，刷新强制获取，仍可选择图片提交', async () => {
  vi.mocked(ffmpegServer.getVideoThumbnails).mockResolvedValue(ok(['one','two']))
  const w=mountPage();w.vm.open('video',32);await flushPromises()
  expect(ffmpegServer.getVideoThumbnails).toHaveBeenLastCalledWith('video',32,false)
  await w.find('.keyframe-toolbar button').trigger('click');await flushPromises()
  expect(ffmpegServer.getVideoThumbnails).toHaveBeenLastCalledWith('video',32,true)
  await w.findAll('.keyframe-option')[1].trigger('click');w.findComponent(dialog).vm.$emit('submit')
  expect(w.emitted('selectImage')).toEqual([['two']]);w.unmount()
 })
 it('关闭后换资源，旧请求不能覆盖新图片', async () => {
  let resolve!: (v: ReturnType<typeof ok>) => void
  vi.mocked(ffmpegServer.getVideoThumbnails).mockReturnValueOnce(new Promise(r=>resolve=r)).mockResolvedValueOnce(ok(['new']))
  const w=mountPage();w.vm.open('old',32);w.findComponent(dialog).vm.$emit('closed');w.vm.open('new',32);await flushPromises()
  resolve(ok(['old']));await flushPromises();w.findComponent(dialog).vm.$emit('submit')
  expect(w.emitted('selectImage')).toEqual([['new']]);w.unmount()
 })
 it('加载期间不重复刷新或提交空图片', async () => {
  vi.mocked(ffmpegServer.getVideoThumbnails).mockReturnValue(new Promise(()=>{}))
  const w=mountPage();w.vm.open('video',32);await flushPromises();await w.find('.keyframe-toolbar button').trigger('click');w.findComponent(dialog).vm.$emit('submit')
  expect(ffmpegServer.getVideoThumbnails).toHaveBeenCalledTimes(1);expect(w.emitted('selectImage')).toBeUndefined();w.unmount()
 })
 it('刷新失败保留旧选图，允许继续设置封面', async () => {
  vi.mocked(ffmpegServer.getVideoThumbnails).mockResolvedValueOnce(ok(['one','two'])).mockResolvedValueOnce({...ok([]),status:false,msg:'获取失败'})
  const w=mountPage();w.vm.open('video',32);await flushPromises();await w.findAll('.keyframe-option')[1].trigger('click')
  await w.find('.keyframe-toolbar button').trigger('click');await flushPromises();w.findComponent(dialog).vm.$emit('submit')
  expect(w.emitted('selectImage')).toEqual([['two']]);w.unmount()
 })
 it('窗口销毁后旧请求失败不再弹出错误提示', async () => {
  let reject!: (e: Error) => void
  const error=vi.spyOn(ElMessage,'error')
  vi.mocked(ffmpegServer.getVideoThumbnails).mockReturnValue(new Promise((_,r)=>reject=r))
  const w=mountPage();w.vm.open('video',32);w.unmount();reject(new Error('late response'));await flushPromises()
  expect(error).not.toHaveBeenCalled();error.mockRestore()
 })

 it('旧请求先返回不能解除新资源的加载状态，连续刷新不重复请求', async () => {
  let oldDone!: (v: ReturnType<typeof ok>) => void
  let newDone!: (v: ReturnType<typeof ok>) => void
  vi.mocked(ffmpegServer.getVideoThumbnails).mockReturnValueOnce(new Promise(r=>oldDone=r)).mockReturnValueOnce(new Promise(r=>newDone=r))
  const w=mountPage();w.vm.open('old',32);w.vm.open('new',16)
  oldDone(ok(['old']));await flushPromises()
  await w.find('.keyframe-toolbar button').trigger('click');w.findComponent(dialog).vm.$emit('submit')
  expect(ffmpegServer.getVideoThumbnails).toHaveBeenCalledTimes(2);expect(w.emitted('selectImage')).toBeUndefined();expect(w.findAll('.keyframe-option')).toHaveLength(0)
  newDone(ok(['new']));await flushPromises();w.findComponent(dialog).vm.$emit('submit');expect(w.emitted('selectImage')).toEqual([['new']]);w.unmount()
 })
 it('空截图结果不能设置封面，刷新后恢复正常', async () => {
  vi.mocked(ffmpegServer.getVideoThumbnails).mockResolvedValueOnce(ok([])).mockResolvedValueOnce(ok(['image']))
  const w=mountPage();w.vm.open('video',32);await flushPromises();w.findComponent(dialog).vm.$emit('submit');expect(w.emitted('selectImage')).toBeUndefined()
  await w.find('.keyframe-toolbar button').trigger('click');await flushPromises();w.findComponent(dialog).vm.$emit('submit');expect(w.emitted('selectImage')).toEqual([['image']]);w.unmount()
 })

})
