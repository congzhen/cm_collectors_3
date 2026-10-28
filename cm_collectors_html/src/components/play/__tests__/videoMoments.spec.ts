import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, reactive, ref } from 'vue'
import MomentTools from '../MomentTools.vue'
import MomentCover from '../MomentCover.vue'
import MomentCollection from '../MomentCollection.vue'
import VideoPreviewSetting from '../../setting/VideoPreviewSetting.vue'
import VideoTimeline from '../VideoTimeline.vue'
import { useVideoPreview } from '../useVideoPreview'
import { createPlaybackSeek } from '@/common/playbackSeek'
import {
  previewImage,
  momentTime,
  momentDuration,
  notifyPreviewChanged,
  createPlaybackPreviewTrigger,
  type VideoPreview,
} from '@/server/videoMoments.server'

const mocks = vi.hoisted(() => ({
  push: vi.fn(),
  list: vi.fn(),
  save: vi.fn(),
  remove: vi.fn(),
  preview: vi.fn(),
  queue: vi.fn(),
  request: vi.fn(),
  confirm: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  success: vi.fn(),
  noticeClose: vi.fn(),
}))
const route = reactive({ params: { dramaSeriesId: 'd' }, query: {} as Record<string, string> })
vi.mock('vue-router', () => ({ useRoute: () => route, useRouter: () => ({ push: mocks.push }) }))
vi.mock('@/assets/request', () => ({ default: (...args: unknown[]) => mocks.request(...args) }))
// Node exposes this CommonJS package through a default wrapper. Keep the real
// validator constructor, matching the browser bundle rather than bypassing rules.
vi.mock('async-validator', async (importOriginal) => {
  const actual = await importOriginal<{ default: { default?: unknown } }>()
  return { ...actual, default: actual.default.default || actual.default }
})
vi.mock('element-plus', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  ElMessage: {
    error: (...args: unknown[]) => mocks.error(...args),
    warning: (...args: unknown[]) => mocks.warning(...args),
    success: (...args: unknown[]) => mocks.success(...args),
  },
  ElMessageBox: { confirm: (...args: unknown[]) => mocks.confirm(...args) },
}))
vi.mock('@/server/videoMoments.server', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  momentApi: {
    list: (...args: unknown[]) => mocks.list(...args),
    save: (...args: unknown[]) => mocks.save(...args),
    remove: (...args: unknown[]) => mocks.remove(...args),
    preview: (...args: unknown[]) => mocks.preview(...args),
    queue: (...args: unknown[]) => mocks.queue(...args),
  },
}))
const button = defineComponent({
  props: ['disabled', 'nativeType'],
  setup:
    (props, { slots }) =>
    () =>
      h(
        'button',
        { disabled: props.disabled, type: props.nativeType || 'button' },
        slots.default?.(),
      ),
})
const global = { stubs: { 'el-button': button, VideoTimeline: true } }
const readyPreview: VideoPreview = {
  dramaSeriesId: 'd',
  sourceKey: 'old',
  status: 'ready',
  error: '',
  interval: 30,
  count: 3,
  completed: 3,
  duration: 90,
}

describe('片段收藏和时间轴', () => {
  it('片段总秒数去除浮点误差，时间点不显示时长', () => {
    expect(momentDuration(129, 136)).toBe('共 7 秒')
    expect(momentDuration(4.9, 186.4)).toBe('共 181.5 秒')
    expect(momentDuration(0.1, 0.3)).toBe('共 0.2 秒')
    expect(momentDuration(12, 0)).toBe('')
  })
  beforeEach(() => {
    vi.useFakeTimers()
    vi.clearAllMocks()
    mocks.success.mockReturnValue({ close: mocks.noticeClose })
    route.query = {}
    mocks.list.mockResolvedValue({ status: true, data: [] })
    mocks.preview.mockResolvedValue({
      status: true,
      data: { ...readyPreview, status: 'none', completed: 0 },
    })
    mocks.save.mockResolvedValue({ status: true, data: {} })
    mocks.request.mockResolvedValue({ status: true })
    mocks.confirm.mockResolvedValue('confirm')
  })
  afterEach(() => {
    vi.useRealTimers()
  })
  it.each(['', '   '])('空名称保存使用表单字段提示，修正后可正常提交：%j', async (title) => {
    const wrapper = mount(MomentTools, {
      props: { dramaSeriesId: 'd', getTime: () => 1, getDuration: () => 90, seek: vi.fn(), play: vi.fn() }, global,
    })
    await wrapper.get('.moment-open').trigger('click')
    await flushPromises()
    await wrapper.get('.moment-manage').trigger('click')
    await wrapper.findAll('button').find(b => b.text() === '收藏当前时刻')!.trigger('click')
    await wrapper.get('input[maxlength="200"]').setValue(title)
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    await vi.advanceTimersByTimeAsync(120)
    expect(mocks.save).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('请输入片段名称')
    await wrapper.get('input[maxlength="200"]').setValue('有效名称')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledOnce()
    expect(mocks.save).toHaveBeenCalledWith(expect.objectContaining({ title: '有效名称' }))
    wrapper.unmount()
  })
  it('标记起止位置并保存，同分集数据在切换后清除', async () => {
    let time = 12
    const wrapper = mount(MomentTools, {
      props: {
        dramaSeriesId: 'd',
        getTime: () => time,
        getDuration: () => 120,
        seek: vi.fn(),
        play: vi.fn(),
      },
      global,
    })
    await wrapper.get('.moment-open').trigger('click')
    await flushPromises()
    const click = async (text: string) =>
      wrapper
        .findAll('button')
        .find((b) => b.text() === text)!
        .trigger('click')
    await click('管理')
    await click('标记起点')
    time = 25
    await click('标记终点')
    await wrapper.get('input[maxlength="200"]').setValue('收藏片段')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledWith(
      expect.objectContaining({ dramaSeriesId: 'd', start: 12, end: 25, title: '收藏片段' }),
    )
    await wrapper.setProps({ dramaSeriesId: 'other' })
    expect(wrapper.find('form').exists()).toBe(false)
    expect(wrapper.find('.moment-panel').exists()).toBe(false)
    wrapper.unmount()
  })
  it('等视频就绪后应用链接时间并循环，退出后不再自动跳转', async () => {
    route.query = { t: '10', end: '20', momentTitle: '循环测试' }
    let duration = 0,
      time = 0
    const seek = vi.fn((value: number) => {
        time = value
      }),
      play = vi.fn()
    const wrapper = mount(MomentTools, {
      props: { dramaSeriesId: 'd', getTime: () => time, getDuration: () => duration, seek, play },
      global,
    })
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).not.toHaveBeenCalled()
    duration = 90
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).toHaveBeenLastCalledWith(10)
    expect(wrapper.find('.loop-status').text()).toContain('循环测试')
    time = 20
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).toHaveBeenLastCalledWith(10)
    expect(play).toHaveBeenCalledOnce()
    await wrapper.get('.moment-open').trigger('click')
    await flushPromises()
    await wrapper
      .findAll('button')
      .find((b) => b.text() === '管理')!
      .trigger('click')
    await wrapper
      .findAll('button')
      .find((b) => b.text() === '退出循环')!
      .trigger('click')
    seek.mockClear()
    time = 22
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).not.toHaveBeenCalled()
    expect(wrapper.find('.loop-status').exists()).toBe(false)
    expect(wrapper.emitted('loop-change')?.at(-1)).toEqual([undefined])
    wrapper.unmount()
  })
  it('每次回到片段起点重新提示，等待跳转时不重复弹出，退出后停止', async () => {
    route.query = { t: '10', end: '20' }
    let time = 10
    const wrapper = mount(MomentTools, { props: {
      dramaSeriesId: 'd', getTime: () => time, getDuration: () => 90,
      seek: vi.fn(), play: vi.fn(),
    }, global })
    await vi.advanceTimersByTimeAsync(160)
    expect(mocks.success).toHaveBeenCalledTimes(1)
    time = 20
    await vi.advanceTimersByTimeAsync(160)
    expect(mocks.success).toHaveBeenCalledTimes(2)
    await vi.advanceTimersByTimeAsync(900)
    expect(mocks.success).toHaveBeenCalledTimes(2)
    time = 10
    await vi.advanceTimersByTimeAsync(160)
    time = 20
    await vi.advanceTimersByTimeAsync(160)
    expect(mocks.success).toHaveBeenCalledTimes(3)
    expect(mocks.success).toHaveBeenLastCalledWith(expect.objectContaining({ message: '已进入片段循环', duration: 2000 }))
    expect(mocks.noticeClose).toHaveBeenCalledTimes(2)
    await wrapper.get('.moment-exit').trigger('click')
    expect(mocks.noticeClose).toHaveBeenCalledTimes(3)
    await vi.advanceTimersByTimeAsync(1000)
    expect(mocks.success).toHaveBeenCalledTimes(3)
    wrapper.unmount()
  })
  it('弹出播放器使用传入片段，不依赖播放页路由，加载后定位并开始循环', async () => {
    let duration = 0, time = 0
    const seek = vi.fn((value: number) => { time = value })
    const play = vi.fn()
    route.query = {t:'80',end:'90'}
    const wrapper = mount(MomentTools, {props:{dramaSeriesId:'d',initialMoment:{start:10,end:20,title:'弹窗片段'},getTime:()=>time,getDuration:()=>duration,seek,play},global})
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).not.toHaveBeenCalled()
    duration = 100
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).toHaveBeenLastCalledWith(10)
    expect(play).toHaveBeenCalledOnce()
    expect(wrapper.find('.loop-status').text()).toContain('弹窗片段')
    time = 20
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).toHaveBeenLastCalledWith(10)
    expect(play).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })
  it('放映室预载片段定位到起点并设置循环，但保持待播放状态', async () => {
    let time = 0
    const seek = vi.fn((value: number) => { time = value })
    const play = vi.fn()
    const wrapper = mount(MomentTools, { props: {
      dramaSeriesId: 'd', initialMoment: { start: 10, end: 20, title: '首个片段', autoplay: false },
      getTime: () => time, getDuration: () => 90, isPlaying: () => false, seek, play,
    }, global })
    await vi.advanceTimersByTimeAsync(600)
    expect(seek).toHaveBeenCalledWith(10)
    expect(wrapper.find('.loop-status').text()).toContain('首个片段')
    expect(play).not.toHaveBeenCalled()
    wrapper.unmount()
  })
  it.each([{start:10,end:100}, {start:100,end:0}])('视频替换变短后拒绝越界片段并提示，保留收藏数据 %j', async (initialMoment) => {
    const seek = vi.fn(), play = vi.fn()
    const wrapper = mount(MomentTools, { props: {
      dramaSeriesId:'d', initialMoment:{...initialMoment,title:'原片段'},
      getTime:()=>0, getDuration:()=>90, seek, play,
    }, global })
    await vi.advanceTimersByTimeAsync(1000)
    expect(seek).not.toHaveBeenCalled()
    expect(play).not.toHaveBeenCalled()
    expect(mocks.warning).toHaveBeenCalledOnce()
    expect(mocks.warning).toHaveBeenCalledWith('收藏时间超出当前视频范围，请在管理中调整片段')
    expect(mocks.remove).not.toHaveBeenCalled()
    expect(wrapper.find('.moment-exit').exists()).toBe(false)
    wrapper.unmount()
  })
  it('弹窗循环期间主页面地址参数变化，不会重新定位或重置片段', async () => {
    let time = 10
    const seek = vi.fn((value:number)=>{time=value})
    const wrapper = mount(MomentTools, { props: {
      dramaSeriesId:'d', initialMoment:{start:10,end:20,title:'弹窗片段'},
      getTime:()=>time, getDuration:()=>90, seek, play:vi.fn(),
    }, global })
    await vi.advanceTimersByTimeAsync(160)
    seek.mockClear()
    time=15
    route.query={t:'30',end:'40'}
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).not.toHaveBeenCalled()
    expect(wrapper.get('.loop-status').text()).toContain('弹窗片段')
    expect(mocks.success).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })
  it('播放器列表点击越界收藏时保持管理入口，避免进入无法完成的循环', async () => {
    mocks.list.mockResolvedValue({status:true,data:[{id:'old',title:'旧片段',start:10,end:120,available:true}]})
    const seek=vi.fn(),play=vi.fn()
    const wrapper=mount(MomentTools,{props:{dramaSeriesId:'d',getTime:()=>0,getDuration:()=>90,seek,play},global})
    await wrapper.get('.moment-open').trigger('click')
    await flushPromises()
    await wrapper.get('.moment-play').trigger('click')
    expect(seek).not.toHaveBeenCalled()
    expect(play).not.toHaveBeenCalled()
    expect(wrapper.find('.moment-panel').exists()).toBe(true)
    expect(mocks.warning).toHaveBeenCalledOnce()
    wrapper.unmount()
  })
  it.each([true, false])('循环中快捷键或手势越过起点会被限制，保持播放状态 %s', async (playing) => {
    route.query = { t: '10', end: '20' }
    let time = 10
    const seek = vi.fn((value: number) => { time = value })
    const play = vi.fn()
    const wrapper = mount(MomentTools, { props: {
      dramaSeriesId: 'd', getTime: () => time, getDuration: () => 90,
      isPlaying: () => playing, seek, play,
    }, global })
    await vi.advanceTimersByTimeAsync(160)
    seek.mockClear()
    time = 5
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).toHaveBeenLastCalledWith(10)
    expect(play).not.toHaveBeenCalled()
    expect(mocks.success).toHaveBeenCalledTimes(1)
    await wrapper.get('.moment-exit').trigger('click')
    seek.mockClear()
    time = 5
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).not.toHaveBeenCalled()
    wrapper.unmount()
  })
  it('现有预览不会因改变密度重建，只有确认删除后允许生成', async () => {
    mocks.preview.mockResolvedValue({ status: true, data: readyPreview })
    const wrapper = mount(VideoTimeline, {
      props: { dramaSeriesId: 'd' },
      global: { stubs: { 'el-button': button } },
    })
    await flushPromises()
    await wrapper.get('select').setValue('10')
    const generate = wrapper.findAll('button').find((b) => b.text() === '生成预览')!
    expect(generate.attributes('disabled')).toBeDefined()
    expect(mocks.queue).not.toHaveBeenCalled()
    await wrapper
      .findAll('button')
      .find((b) => b.text() === '删除预览')!
      .trigger('click')
    await flushPromises()
    expect(mocks.confirm).toHaveBeenCalledOnce()
    expect(mocks.request).toHaveBeenCalledWith({ url: '/videoPreviews/d', method: 'delete' })
    await wrapper.findAll('.timeline-frames button')[1].trigger('click')
    expect(wrapper.emitted('seek')?.[0]).toEqual([30])
    wrapper.unmount()
  })
  it('未生成的时间点不显示错误画面，视频末尾使用最后一帧', () => {
    expect(previewImage(readyPreview, 50)).toContain('/image/1?')
    expect(previewImage({ ...readyPreview, completed: 1 }, 89)).toBe('')
    expect(previewImage(readyPreview, 90)).toContain('/image/2?')
    expect(previewImage(readyPreview, Number.NaN)).toBe('')
    expect(momentTime(3661)).toBe('01:01:01')
  })
  it.each([false, true])('慢速定位只发起一次循环跳转，暂停时不自动播放（提供 seeking：%s）', async (withSeeking) => {
    route.query = { t: '10', end: '20' }
    let time = 10,
      playing = false
    const seek = vi.fn(),
      play = vi.fn()
    const wrapper = mount(MomentTools, {
      props: {
        dramaSeriesId: 'd',
        getTime: () => time,
        getDuration: () => 90,
        isPlaying: () => playing,
        isSeeking: withSeeking ? () => false : undefined,
        seek,
        play,
      },
      global,
    })
    await vi.advanceTimersByTimeAsync(160)
    seek.mockClear()
    time = 20
    await vi.advanceTimersByTimeAsync(600)
    expect(seek).not.toHaveBeenCalled()
    playing = true
    await vi.advanceTimersByTimeAsync(900)
    expect(seek).toHaveBeenCalledTimes(1)
    expect(play).toHaveBeenCalledTimes(1)
    time = 10
    await vi.advanceTimersByTimeAsync(160)
    time = 20
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })
  it('连续用户定位尚未完成时不被循环回跳覆盖，完成后恢复循环', async () => {
    let time = 10, seeking = false
    const seek = vi.fn(), play = vi.fn()
    const wrapper = mount(MomentTools, { props: {
      dramaSeriesId: 'd', initialMoment: { start: 10, end: 20, title: '片段', autoplay: false },
      getTime: () => time, getDuration: () => 90, isSeeking: () => seeking,
      isPlaying: () => true, seek, play,
    }, global })
    await vi.advanceTimersByTimeAsync(160)
    seek.mockClear()
    seeking = true
    time = 20
    await vi.advanceTimersByTimeAsync(600)
    expect(seek).not.toHaveBeenCalled()
    time = 5
    await vi.advanceTimersByTimeAsync(600)
    expect(seek).not.toHaveBeenCalled()
    time = 15
    seeking = false
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).not.toHaveBeenCalled()
    time = 20
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).toHaveBeenCalledExactlyOnceWith(10)
    wrapper.unmount()
  })
  it('通过 seeked 完成事件恢复循环，不依赖定时器碰巧观察到 seeking 或起点', async () => {
    let time = 30, revision = 0
    const seek = vi.fn()
    const wrapper = mount(MomentTools, { props: {
      dramaSeriesId: 'd', initialMoment: { start: 10, end: 20, title: '片段', autoplay: false },
      getTime: () => time, getDuration: () => 90, isSeeking: () => false,
      getSeekRevision: () => revision, isPlaying: () => true, seek, play: vi.fn(),
    }, global })
    await vi.advanceTimersByTimeAsync(600)
    expect(seek).toHaveBeenCalledExactlyOnceWith(10)
    // A fast later user seek completed at the end between timer ticks.
    time = 20
    revision++
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).toHaveBeenCalledTimes(2)
    await vi.advanceTimersByTimeAsync(900)
    expect(seek).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })
  it('退出循环会取消尚未提交的自动回跳，继续停留在退出时的位置', async () => {
    let time = 0
    const seek = createPlaybackSeek(value => { time = value })
    const wrapper = mount(MomentTools, { props: {
      dramaSeriesId: 'd', initialMoment: { start: 10, end: 20, title: '片段', autoplay: false },
      getTime: () => time, getDuration: () => 90, isPlaying: () => true,
      isSeeking: seek.pending, seek: seek.request, cancelPendingSeek: seek.cancel, play: vi.fn(),
    }, global })
    await vi.advanceTimersByTimeAsync(270)
    seek.request(20)
    expect(time).toBe(20)
    await vi.advanceTimersByTimeAsync(40)
    expect(seek.pending()).toBe(true)
    await wrapper.get('.moment-exit').trigger('click')
    await vi.advanceTimersByTimeAsync(600)
    expect(time).toBe(20)
    expect(seek.pending()).toBe(false)
    wrapper.unmount()
  })
  it('加载更多失败后重试同一页', async () => {
    mocks.list
      .mockResolvedValueOnce({
        status: true,
        data: Array.from({ length: 50 }, (_, i) => ({
          id: String(i),
          title: '收藏',
          start: 1,
          end: 0,
          available: true,
        })),
      })
      .mockResolvedValueOnce({ status: false, msg: '网络错误' })
      .mockResolvedValueOnce({ status: true, data: [] })
    const wrapper = mount(MomentTools, {
      props: {
        dramaSeriesId: 'd',
        getTime: () => 0,
        getDuration: () => 90,
        seek: vi.fn(),
        play: vi.fn(),
      },
      global,
    })
    await wrapper.get('.moment-open').trigger('click')
    await flushPromises()
    const more = () => wrapper.findAll('button').find((b) => b.text() === '加载更多')!
    await more().trigger('click')
    await flushPromises()
    await more().trigger('click')
    await flushPromises()
    expect(mocks.list.mock.calls.map((call) => call[1])).toEqual([1, 2, 2])
    wrapper.unmount()
  })
  it('删除确认期间切换分集仍只删除原分集预览', async () => {
    mocks.preview.mockResolvedValue({ status: true, data: readyPreview })
    let confirm!: (value: string) => void
    mocks.confirm.mockReturnValueOnce(
      new Promise((resolve) => {
        confirm = resolve
      }),
    )
    const wrapper = mount(VideoTimeline, {
      props: { dramaSeriesId: 'd' },
      global: { stubs: { 'el-button': button } },
    })
    await flushPromises()
    await wrapper
      .findAll('button')
      .find((b) => b.text() === '删除预览')!
      .trigger('click')
    await wrapper.setProps({ dramaSeriesId: 'other' })
    confirm('confirm')
    await flushPromises()
    expect(mocks.request).toHaveBeenCalledWith({ url: '/videoPreviews/d', method: 'delete' })
    wrapper.unmount()
  })
  it('预览变更同步到其它展示组件，卸载后停止订阅', async () => {
    const wrapper = mount(
      defineComponent({
        setup() {
          const { preview } = useVideoPreview(ref('d'))
          return () => h('span', preview.value?.status)
        },
      }),
    )
    await flushPromises()
    expect(wrapper.text()).toBe('none')
    mocks.preview.mockResolvedValue({ status: true, data: readyPreview })
    notifyPreviewChanged('other')
    await flushPromises()
    expect(wrapper.text()).toBe('none')
    notifyPreviewChanged('d')
    await flushPromises()
    expect(wrapper.text()).toBe('ready')
    wrapper.unmount()
    mocks.preview.mockClear()
    notifyPreviewChanged()
    expect(mocks.preview).not.toHaveBeenCalled()
  })
  it.each(['播放器', '收藏页'])('%s刷新失败保留已加载页码', async (location) => {
    const batch = (offset: number) =>
      Array.from({ length: 50 }, (_, i) => ({
        id: String(offset + i),
        title: '收藏',
        start: 1,
        end: 0,
        available: true,
        dramaSeriesId: 'd',
        resourceId: 'r',
      }))
    mocks.list
      .mockResolvedValueOnce({ status: true, data: batch(0) })
      .mockResolvedValueOnce({ status: true, data: batch(50) })
      .mockResolvedValueOnce({ status: false, msg: '刷新失败' })
      .mockResolvedValueOnce({ status: true, data: [] })
    const wrapper =
      location === '播放器'
        ? mount(MomentTools, {
            props: {
              dramaSeriesId: 'd',
              getTime: () => 1,
              getDuration: () => 90,
              seek: vi.fn(),
              play: vi.fn(),
            },
            global,
          })
        : mount(MomentCollection, {
            global: { stubs: { 'el-button': button, MomentCover: true, 'el-dialog': true } },
          })
    if (location === '播放器') await wrapper.get('.moment-open').trigger('click')
    await flushPromises()
    const click = async (label: string) => {
      await wrapper
        .findAll('button')
        .find((b) => b.text() === label)!
        .trigger('click')
      await flushPromises()
    }
    await click('加载更多')
    if (location === '播放器') {
      await wrapper.get('.moment-open').trigger('click')
      await wrapper.get('.moment-open').trigger('click')
      await flushPromises()
    } else await click('刷新')
    await click('加载更多')
    expect(mocks.list.mock.calls.map((call) => call[1])).toEqual([1, 2, 1, 3])
    wrapper.unmount()
  })
  it('旧保存请求完成后不关闭新建的编辑草稿', async () => {
    let resolveSave!: (value: unknown) => void
    mocks.save.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveSave = resolve
      }),
    )
    const wrapper = mount(MomentTools, {
      props: {
        dramaSeriesId: 'd',
        getTime: () => 1,
        getDuration: () => 90,
        seek: vi.fn(),
        play: vi.fn(),
      },
      global,
    })
    await wrapper.get('.moment-open').trigger('click')
    await flushPromises()
    await wrapper
      .findAll('button')
      .find((b) => b.text() === '管理')!
      .trigger('click')
    const capture = () =>
      wrapper
        .findAll('button')
        .find((b) => b.text() === '收藏当前时刻')!
        .trigger('click')
    await capture()
    await wrapper.get('input[maxlength="200"]').setValue('旧草稿')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    await capture()
    await wrapper.get('input[maxlength="200"]').setValue('新草稿')
    resolveSave({ status: true, data: {} })
    await flushPromises()
    expect((wrapper.get('input[maxlength="200"]').element as HTMLInputElement).value).toBe('新草稿')
    expect(mocks.save).toHaveBeenCalledWith(expect.objectContaining({ title: '旧草稿' }))
    wrapper.unmount()
  })
  it('切换分集后可立即保存，旧保存响应不影响新请求和草稿', async () => {
    let finishOld!: (value: unknown) => void
    let finishNew!: (value: unknown) => void
    mocks.save
      .mockImplementationOnce(() => new Promise(resolve => { finishOld = resolve }))
      .mockImplementationOnce(() => new Promise(resolve => { finishNew = resolve }))
    const wrapper = mount(MomentTools, {
      props: { dramaSeriesId: 'd', getTime: () => 1, getDuration: () => 90, seek: vi.fn(), play: vi.fn() },
      global,
    })
    const editAndSave = async (title: string) => {
      await wrapper.get('.moment-open').trigger('click')
      await flushPromises()
      await wrapper.get('.moment-manage').trigger('click')
      await wrapper.findAll('button').find(b => b.text() === '收藏当前时刻')!.trigger('click')
      await wrapper.get('input[maxlength="200"]').setValue(title)
      await wrapper.get('form').trigger('submit')
      await flushPromises()
    }
    await editAndSave('旧草稿')
    await wrapper.setProps({ dramaSeriesId: 'new' })
    await editAndSave('新草稿')
    expect(mocks.save).toHaveBeenCalledTimes(2)
    finishOld({ status: false, msg: '旧保存错误' })
    await flushPromises()
    expect(mocks.error).not.toHaveBeenCalled()
    expect((wrapper.get('input[maxlength="200"]').element as HTMLInputElement).value).toBe('新草稿')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledTimes(2)
    finishNew({ status: true, data: {} })
    await flushPromises()
    expect(wrapper.find('form').exists()).toBe(false)
    wrapper.unmount()
  })
  it.each([true, false])('播放器卸载后删除响应不再刷新列表或显示错误：%s', async (success) => {
    let finish!: (value: unknown) => void
    mocks.list.mockResolvedValue({ status: true, data: [
      { id: 'm', dramaSeriesId: 'd', title: '收藏', start: 1, end: 2, available: true },
    ] })
    mocks.remove.mockImplementationOnce(() => new Promise(resolve => { finish = resolve }))
    const wrapper = mount(MomentTools, {
      props: { dramaSeriesId: 'd', getTime: () => 1, getDuration: () => 90, seek: vi.fn(), play: vi.fn() },
      global,
    })
    await wrapper.get('.moment-open').trigger('click')
    await flushPromises()
    await wrapper.get('.moment-manage').trigger('click')
    await wrapper.get('.moment-delete').trigger('click')
    await flushPromises()
    expect(mocks.remove).toHaveBeenCalledWith('m')
    const calls = mocks.list.mock.calls.length
    wrapper.unmount()
    finish({ status: success, msg: '旧删除错误' })
    await flushPromises()
    expect(mocks.error).not.toHaveBeenCalled()
    expect(mocks.list).toHaveBeenCalledTimes(calls)
  })
  it('预览查询短暂失败后自动恢复，连续失败限制重试次数', async () => {
    mocks.preview
      .mockResolvedValueOnce({ status: false, msg: '网络错误' })
      .mockResolvedValueOnce({ status: true, data: readyPreview })
    const wrapper = mount(
      defineComponent({
        setup() {
          const { preview, error } = useVideoPreview(ref('d'))
          return () => h('span', error.value || preview.value?.status)
        },
      }),
    )
    await flushPromises()
    expect(wrapper.text()).toBe('网络错误')
    await vi.advanceTimersByTimeAsync(5000)
    expect(wrapper.text()).toBe('ready')
    mocks.preview.mockResolvedValue({ status: false, msg: '仍未连接' })
    mocks.preview.mockClear()
    notifyPreviewChanged('d')
    await flushPromises()
    await vi.advanceTimersByTimeAsync(100000)
    expect(mocks.preview).toHaveBeenCalledTimes(4)
    wrapper.unmount()
    await vi.advanceTimersByTimeAsync(100000)
    expect(mocks.preview).toHaveBeenCalledTimes(4)
  })
  it('切换分集忽略旧预览响应并清除旧错误', async () => {
    const id = ref('old')
    let resolveOld!: (value: unknown) => void
    mocks.preview.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveOld = resolve
      }),
    )
    const wrapper = mount(
      defineComponent({
        setup() {
          const { preview, error } = useVideoPreview(id)
          return () => h('span', error.value || preview.value?.dramaSeriesId)
        },
      }),
    )
    mocks.preview.mockResolvedValue({
      status: true,
      data: { ...readyPreview, dramaSeriesId: 'new' },
    })
    id.value = 'new'
    await flushPromises()
    resolveOld({ status: false, msg: '旧分集错误' })
    await flushPromises()
    expect(wrapper.text()).toBe('new')
    id.value = ''
    await flushPromises()
    expect(wrapper.text()).toBe('')
    wrapper.unmount()
  })
  it('设置保存期间的新输入不会被覆盖，重复提交被阻止', async () => {
    let resolveSave!: (value: unknown) => void
    mocks.request.mockImplementation((options: { url: string; method?: string }) => {
      if (options.method === 'put')
        return new Promise((resolve) => {
          resolveSave = resolve
        })
      return Promise.resolve({
        status: true,
        data:
          options.url === '/videoPreviewSetting'
            ? { width: 240, maxMB: 1024 }
            : { bytes: 0, queued: 0, paused: false, running: false },
      })
    })
    const passthrough = defineComponent({
      setup:
        (_, { slots }) =>
        () =>
          h('div', slots.default?.()),
    })
    const select = defineComponent({
      name: 'SettingSelect',
      props: ['modelValue'],
      emits: ['update:modelValue'],
      setup: () => () => h('div'),
    })
    const wrapper = mount(VideoPreviewSetting, {
      global: {
        directives: { loading: () => {} },
        stubs: {
          'el-form': passthrough,
          'el-form-item': passthrough,
          'el-select': select,
          'el-option': true,
          'el-input-number': true,
          'el-button': button,
          'el-divider': true,
        },
      },
    })
    await flushPromises()
    const width = wrapper.findComponent(select)
    width.vm.$emit('update:modelValue', 320)
    await wrapper
      .findAll('button')
      .find((b) => b.text() === '保存设置')!
      .trigger('click')
    width.vm.$emit('update:modelValue', 480)
    await wrapper
      .findAll('button')
      .find((b) => b.text() === '保存设置')!
      .trigger('click')
    expect(mocks.request.mock.calls.filter(([r]) => r.method === 'put')).toHaveLength(1)
    expect(mocks.request.mock.calls.find(([r]) => r.method === 'put')?.[0].data.width).toBe(320)
    resolveSave({ status: true })
    await flushPromises()
    expect(width.props('modelValue')).toBe(480)
    expect(
      mocks.request.mock.calls.filter(([r]) => r.url === '/videoPreviewSetting' && !r.method),
    ).toHaveLength(1)
    wrapper.unmount()
  })
  it('片段入口空闲后隐藏，操作播放器唤出，面板打开时保持显示', async () => {
    const surface = document.createElement('div')
    document.body.appendChild(surface)
    const wrapper = mount(MomentTools, {
      attachTo: surface,
      props: {
        activityRoot: surface,
        dramaSeriesId: 'd',
        getTime: () => 1,
        getDuration: () => 90,
        seek: vi.fn(),
        play: vi.fn(),
      },
      global,
    })
    const entry = () => wrapper.get('.moment-open')
    await vi.advanceTimersByTimeAsync(3001)
    expect(entry().classes()).toContain('moment-open--hidden')
    surface.dispatchEvent(new MouseEvent('mousemove', { bubbles: true }))
    await flushPromises()
    expect(entry().classes()).not.toContain('moment-open--hidden')
    await entry().trigger('click')
    await flushPromises()
    await vi.advanceTimersByTimeAsync(6000)
    expect(entry().classes()).not.toContain('moment-open--hidden')
    expect(wrapper.find('.moment-panel').exists()).toBe(true)
    await wrapper
      .findAll('button')
      .find((b) => b.text() === '关闭')!
      .trigger('click')
    await vi.advanceTimersByTimeAsync(3001)
    expect(entry().classes()).toContain('moment-open--hidden')
    surface.dispatchEvent(new Event('touchstart', { bubbles: true }))
    await flushPromises()
    expect(entry().classes()).not.toContain('moment-open--hidden')
    ;(entry().element as HTMLButtonElement).focus()
    await vi.advanceTimersByTimeAsync(6000)
    expect(entry().classes()).not.toContain('moment-open--hidden')
    ;(entry().element as HTMLButtonElement).blur()
    await vi.advanceTimersByTimeAsync(3001)
    expect(entry().classes()).toContain('moment-open--hidden')
    wrapper.unmount()
    surface.remove()
    expect(vi.getTimerCount()).toBe(0)
  })
  it('在入口旁直接退出片段播放，保留当前播放位置且不再循环', async () => {
    route.query = { t: '10', end: '20' }
    let time = 10
    const seek = vi.fn(),
      play = vi.fn()
    const wrapper = mount(MomentTools, {
      props: { dramaSeriesId: 'd', getTime: () => time, getDuration: () => 90, seek, play },
      global,
    })
    await vi.advanceTimersByTimeAsync(160)
    expect(wrapper.find('.moment-panel').exists()).toBe(false)
    expect(wrapper.get('.moment-exit').text()).toBe('')
    expect(wrapper.get('.moment-exit').attributes('aria-label')).toBe('退出片段播放')
    expect(wrapper.find('.moment-exit svg').exists()).toBe(true)
    await vi.advanceTimersByTimeAsync(3001)
    expect(wrapper.get('.moment-exit').classes()).toContain('moment-open--hidden')
    seek.mockClear()
    time = 18
    await wrapper.get('.moment-exit').trigger('click')
    expect(wrapper.find('.moment-exit').exists()).toBe(false)
    expect(wrapper.get('.moment-open').text()).toBe('片段 / 预览')
    expect(time).toBe(18)
    time = 22
    await vi.advanceTimersByTimeAsync(1000)
    expect(seek).not.toHaveBeenCalled()
    expect(play).not.toHaveBeenCalled()
    expect(wrapper.find('.moment-panel').exists()).toBe(false)
    wrapper.unmount()
  })
  it.each([false, true])('点击播放器外关闭，视频和进度条操作不关闭，页面浮层=%s', async (pageOverlay) => {
    const host = document.createElement('div')
    host.className = 'el-overlay'
    document.body.appendChild(host)
    const player = document.createElement('div')
    host.appendChild(player)
    const wrapper = mount(MomentTools, {
      attachTo: player,
      props: { dramaSeriesId: 'd', pageOverlay, activityRoot: player, getTime: () => 0, getDuration: () => 90, seek: vi.fn(), play: vi.fn() },
      global,
    })
    const outside = document.createElement('button')
    const outsideAction = vi.fn()
    outside.addEventListener('click', outsideAction)
    host.appendChild(outside)
    const dialog = document.createElement('div')
    dialog.className = 'el-message-box__wrapper'
    document.body.appendChild(dialog)
    try {
      await wrapper.get('.moment-open').trigger('click')
      await flushPromises()
      const panel = document.querySelector('.moment-panel')!
      panel.dispatchEvent(new MouseEvent('mouseleave'))
      panel.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      await flushPromises()
      expect(document.querySelector('.moment-panel')).toBe(panel)
      for (const tag of ['video', 'input']) {
        const control = document.createElement(tag)
        player.appendChild(control)
        control.click()
        await flushPromises()
        expect(document.querySelector('.moment-panel')).toBe(panel)
      }
      dialog.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      await flushPromises()
      expect(document.querySelector('.moment-panel')).toBe(panel)
      outside.click()
      await flushPromises()
      expect(document.querySelector('.moment-panel')).toBeNull()
      expect(outsideAction).toHaveBeenCalledOnce()
      await wrapper.get('.moment-open').trigger('click')
      await flushPromises()
      expect(document.querySelector('.moment-panel')).not.toBeNull()
      await wrapper.get('.moment-open').trigger('click')
      expect(document.querySelector('.moment-panel')).toBeNull()
    } finally {
      wrapper.unmount()
      host.remove()
      outside.remove()
      dialog.remove()
    }
  })
  it('默认预览片段，播放入口明确，管理操作按需展开', async () => {
    mocks.list.mockResolvedValue({
      status: true,
      data: [{ id: 'moment', title: '片段一', start: 12, end: 20, available: true }],
    })
    let time = 0
    const seek = vi.fn((value: number) => {
        time = value
      }),
      play = vi.fn()
    const wrapper = mount(MomentTools, {
      props: { dramaSeriesId: 'd', getTime: () => time, getDuration: () => 90, seek, play },
      global,
    })
    await wrapper.get('.moment-open').trigger('click')
    await flushPromises()
    expect(wrapper.find('.moment-actions').exists()).toBe(false)
    expect(wrapper.find('.moment-edit').exists()).toBe(false)
    expect(wrapper.get('.moment-play').text()).toBe('播放')
    expect(wrapper.findComponent({ name: 'VideoTimeline' }).attributes('manage')).toBe('false')
    await wrapper.get('.moment-manage').trigger('click')
    expect(wrapper.find('.moment-actions').exists()).toBe(true)
    expect(wrapper.find('.moment-edit').exists()).toBe(true)
    await wrapper.get('.moment-manage').trigger('click')
    expect(wrapper.find('.moment-edit').exists()).toBe(false)
    await wrapper.get('.moment-play').trigger('click')
    expect(seek).toHaveBeenLastCalledWith(12)
    expect(play).toHaveBeenCalledOnce()
    expect(wrapper.find('.moment-panel').exists()).toBe(false)
    time = 20
    await vi.advanceTimersByTimeAsync(160)
    expect(seek).toHaveBeenLastCalledWith(12)
    await wrapper.get('.moment-open').trigger('click')
    await flushPromises()
    expect(wrapper.get('.moment-manage').text()).toBe('管理')
    wrapper.unmount()
  })
  it('预览模式下时间轴不显示生成与删除操作', async () => {
    mocks.preview.mockResolvedValue({ status: true, data: readyPreview })
    const wrapper = mount(VideoTimeline, {
      props: { dramaSeriesId: 'd', manage: false },
      global: { stubs: { 'el-button': button } },
    })
    await flushPromises()
    expect(wrapper.find('select').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('删除预览')
    expect(wrapper.text()).not.toContain('生成预览')
    await wrapper.findAll('.timeline-frames button')[1].trigger('click')
    expect(wrapper.emitted('seek')?.[0]).toEqual([30])
    await wrapper.setProps({ manage: true })
    expect(wrapper.find('select').exists()).toBe(true)
    expect(wrapper.text()).toContain('删除预览')
    wrapper.unmount()
  })
  it('播放触发生成请求按分集去重，失败不影响播放且下次可重试', async () => {
    const trigger = createPlaybackPreviewTrigger()
    mocks.request.mockResolvedValue({ status: true, data: false })
    await trigger('')
    await Promise.all([trigger('d'), trigger('d')])
    expect(mocks.request).toHaveBeenCalledTimes(1)
    expect(mocks.request).toHaveBeenLastCalledWith({
      url: '/videoPreviews/d/playback',
      method: 'post',
    })
    await trigger('other')
    expect(mocks.request).toHaveBeenCalledTimes(2)
    mocks.request.mockResolvedValueOnce({ status: false, msg: '离线' })
    await trigger('retry')
    await trigger('retry')
    expect(mocks.request).toHaveBeenCalledTimes(4)
  })
  it('手动生成默认使用本库间隔，支持120秒且刷新不覆盖用户选择', async () => {
    mocks.preview.mockResolvedValue({
      status: true,
      data: { ...readyPreview, status: 'none', completed: 0, interval: 120 },
    })
    mocks.queue.mockResolvedValue({ status: true, data: {} })
    const wrapper = mount(VideoTimeline, {
      props: { dramaSeriesId: 'd' },
      global: { stubs: { 'el-button': button } },
    })
    await flushPromises()
    expect((wrapper.get('select').element as HTMLSelectElement).value).toBe('120')
    await wrapper.get('select').setValue('60')
    notifyPreviewChanged('d')
    await flushPromises()
    expect((wrapper.get('select').element as HTMLSelectElement).value).toBe('60')
    await wrapper
      .findAll('button')
      .find((b) => b.text() === '生成预览')!
      .trigger('click')
    await flushPromises()
    expect(mocks.queue).toHaveBeenCalledWith('d', 60)
    await wrapper.setProps({ dramaSeriesId: 'new' })
    await flushPromises()
    expect((wrapper.get('select').element as HTMLSelectElement).value).toBe('120')
    wrapper.unmount()
  })
  it('资源片段卡片只提供播放，隐藏编辑删除和重复资源标题', async () => {
    mocks.list.mockResolvedValue({
      status: true,
      data: [
        {
          id: 'moment',
          resourceId: 'r',
          dramaSeriesId: 'd',
          title: '片段名称',
          resourceTitle: '重复的视频标题',
          start: 12,
          end: 20,
          note: '喜欢的镜头',
          available: true,
        },
      ],
    })
    const wrapper = mount(MomentCollection, {
      props: { resourceId: 'r' },
      global: { stubs: { 'el-button': button, MomentCover: true } },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('片段名称')
    expect(wrapper.text()).toContain('喜欢的镜头')
    expect(wrapper.get('.duration-badge').text()).toBe('8s')
    expect(wrapper.get('.card-time').text()).not.toContain('秒')
    expect(wrapper.text()).not.toContain('重复的视频标题')
    expect(wrapper.text()).not.toContain('编辑')
    expect(wrapper.text()).not.toContain('删除')
    expect(wrapper.find('form').exists()).toBe(false)
    await wrapper.get('.collection-play').trigger('click')
    expect(mocks.push).toHaveBeenCalledWith({
      path: '/play/movies/r/d',
      query: { t: '12', end: '20', momentTitle: '片段名称' },
    })
    expect(wrapper.emitted('play')?.[0]).toEqual([expect.objectContaining({ id: 'moment' })])
    await wrapper.setProps({ resourceId: '' })
    await flushPromises()
    expect(wrapper.text()).toContain('重复的视频标题')
    wrapper.unmount()
  })
  it('收藏组件可用列表布局隐藏标题、展示来源并交给父组件播放', async () => {
    const item = { id: 'm', resourceId: 'r', dramaSeriesId: 'd', title: '收藏', resourceTitle: '来源资源', start: 1, end: 2, available: true }
    mocks.list.mockResolvedValue({ status: true, data: [item] })
    const wrapper = mount(MomentCollection, {
      props: { resourceId: 'r', playMode: 'emit', layout: 'list', showHeading: false, showResourceTitle: true },
      global: { stubs: { 'el-button': button, MomentCover: true } },
    })
    await flushPromises()
    expect(wrapper.classes()).toContain('moment-collection--list')
    expect(wrapper.find('.collection-heading').exists()).toBe(false)
    expect(wrapper.text()).toContain('来源资源')
    await wrapper.get('.collection-play').trigger('click')
    expect(wrapper.emitted('play')).toEqual([[item]])
    expect(mocks.push).not.toHaveBeenCalled()
    await wrapper.setProps({ disabled: true })
    expect(wrapper.get('.collection-play').attributes('disabled')).toBeDefined()
    wrapper.unmount()
  })
  it('收藏首图使用独立接口，读取失败时显示占位而非破图', async () => {
    const wrapper = mount(MomentCover, { props: { momentId: 'bookmark', time: 12 } })
    expect(wrapper.get('img').attributes('src')).toBe('/api/videoMoments/bookmark/cover?t=12')
    expect(wrapper.get('img').attributes('src')).not.toContain('videoPreviews')
    await wrapper.get('img').trigger('error')
    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.text()).toBe('暂无画面')
    await wrapper.setProps({ time: 20 })
    expect(wrapper.get('img').attributes('src')).toContain('t=20')
    wrapper.unmount()
  })
  it.each(['收藏列表', '播放器'])('%s 刷新列表后重试失败首图，正常首图不重建', async (location) => {
    mocks.list.mockResolvedValue({ status: true, data: [
      { id: 'm', dramaSeriesId: 'd', title: '收藏', start: 1, end: 2, available: true },
      { id: 'ok', dramaSeriesId: 'd', title: '正常首图', start: 2, end: 3, available: true },
    ] })
    const wrapper = location === '播放器'
      ? mount(MomentTools, {
          props: { dramaSeriesId: 'd', getTime: () => 1, getDuration: () => 90, seek: vi.fn(), play: vi.fn() }, global,
        })
      : mount(MomentCollection, { props: { resourceId: 'r' }, global: { stubs: { 'el-button': button } } })
    if (location === '播放器') await wrapper.get('.moment-open').trigger('click')
    await flushPromises()
    const healthy = wrapper.findAll('img')[1].element
    await wrapper.findAll('img')[0].trigger('error')
    expect(wrapper.findAll('img')).toHaveLength(1)
    expect(wrapper.text()).toContain('暂无画面')
    if (location === '播放器') {
      await wrapper.get('.moment-manage').trigger('click')
      await wrapper.findAll('.moment-edit')[0].trigger('click')
      await wrapper.get('input[maxlength="200"]').setValue('更新名称')
      await wrapper.get('form').trigger('submit')
    } else await wrapper.get('.collection-refresh').trigger('click')
    await flushPromises()
    expect(wrapper.findAll('img')).toHaveLength(2)
    expect(wrapper.text()).not.toContain('暂无画面')
    expect(wrapper.findAll('img')[1].element).toBe(healthy)
    wrapper.unmount()
  })
  it('时间轴图片丢失后刷新为缺失状态，重新生成按钮可用', async () => {
    mocks.preview
      .mockResolvedValueOnce({ status: true, data: readyPreview })
      .mockResolvedValueOnce({
        status: true,
        data: { ...readyPreview, status: 'missing', completed: 0 },
      })
      .mockResolvedValue({
        status: true,
        data: { ...readyPreview, status: 'queued', completed: 0 },
      })
    mocks.queue.mockResolvedValue({ status: true, data: {} })
    const wrapper = mount(VideoTimeline, {
      props: { dramaSeriesId: 'd' },
      global: { stubs: { 'el-button': button } },
    })
    await flushPromises()
    expect(wrapper.findAll('.timeline-frames img')).toHaveLength(3)
    const images = wrapper.findAll('.timeline-frames img')
    await Promise.all(images.map((img) => img.trigger('error')))
    await flushPromises()
    expect(mocks.preview).toHaveBeenCalledTimes(2)
    expect(wrapper.findAll('.timeline-frames img')).toHaveLength(0)
    const regenerate = wrapper.findAll('button').find((b) => b.text() === '重新生成')!
    expect(regenerate.attributes('disabled')).toBeUndefined()
    await regenerate.trigger('click')
    await flushPromises()
    expect(mocks.queue).toHaveBeenCalledWith('d', 30)
    wrapper.unmount()
  })
  it('切换分集后图片错误独立检查缓存，不等待旧分集的请求', async () => {
    let finishOld!: (value: unknown) => void
    mocks.preview.mockResolvedValue({ status: true, data: readyPreview })
    const wrapper = mount(VideoTimeline, {
      props: { dramaSeriesId: 'd' }, global: { stubs: { 'el-button': button } },
    })
    await flushPromises()
    mocks.preview.mockImplementationOnce(() => new Promise(resolve => { finishOld = resolve }))
    await wrapper.get('.timeline-frames img').trigger('error')
    await wrapper.setProps({ dramaSeriesId: 'new' })
    await flushPromises()
    mocks.preview.mockResolvedValueOnce({
      status: true, data: { ...readyPreview, status: 'missing', completed: 0 },
    })
    await wrapper.get('.timeline-frames img').trigger('error')
    await flushPromises()
    expect(mocks.preview).toHaveBeenLastCalledWith('new')
    expect(mocks.preview).toHaveBeenCalledTimes(4)
    expect(wrapper.findAll('.timeline-frames img')).toHaveLength(0)
    finishOld({ status: true, data: readyPreview })
    await flushPromises()
    expect(wrapper.text()).toContain('时间轴缓存已清理或不完整')
    wrapper.unmount()
  })
  it('切换分集后可以立即生成，旧请求不能清除新分集忙碌状态或报错', async () => {
    let finishOld!: (value: unknown) => void
    let finishNew!: (value: unknown) => void
    mocks.queue
      .mockImplementationOnce(() => new Promise(resolve => { finishOld = resolve }))
      .mockImplementationOnce(() => new Promise(resolve => { finishNew = resolve }))
    const wrapper = mount(VideoTimeline, {
      props: { dramaSeriesId: 'd' }, global: { stubs: { 'el-button': button } },
    })
    await flushPromises()
    const generate = () => wrapper.findAll('button').find(item => item.text() === '生成预览')!
    await generate().trigger('click')
    await wrapper.setProps({ dramaSeriesId: 'new' })
    await flushPromises()
    await generate().trigger('click')
    expect(mocks.queue).toHaveBeenCalledTimes(2)
    finishOld({ status: false, msg: '旧分集失败' })
    await flushPromises()
    expect(mocks.error).not.toHaveBeenCalled()
    await generate().trigger('click')
    expect(mocks.queue).toHaveBeenCalledTimes(2)
    finishNew({ status: true })
    await flushPromises()
    expect(mocks.preview).toHaveBeenLastCalledWith('new')
    wrapper.unmount()
  })
  it('时间轴图片临时请求失败会重试一次，加载成功恢复显示，手动刷新可再次加载', async () => {
    mocks.preview.mockResolvedValue({status:true,data:readyPreview})
    const wrapper = mount(VideoTimeline,{props:{dramaSeriesId:'d'},global:{stubs:{'el-button':button}}})
    await flushPromises()
    const img=wrapper.findAll('.timeline-frames img')[0]
    const secondImg=wrapper.findAll('.timeline-frames img')[1]
    const before=mocks.preview.mock.calls.length
    await Promise.all([img.trigger('error'),secondImg.trigger('error')])
    await flushPromises()
    expect(mocks.preview).toHaveBeenCalledTimes(before+1)
    expect(img.attributes('src')).toContain('&retry=1')
    expect(secondImg.attributes('src')).toContain('&retry=1')
    await img.trigger('load')
    expect((img.element as HTMLImageElement).style.visibility).toBe('')
    const calls=mocks.preview.mock.calls.length
    await img.trigger('error')
    await flushPromises()
    expect(mocks.preview).toHaveBeenCalledTimes(calls)
    await wrapper.findAll('button').find(item=>item.text()==='刷新状态')!.trigger('click')
    await flushPromises()
    expect(wrapper.findAll('.timeline-frames img')[0].element).not.toBe(img.element)
    expect(wrapper.findAll('.timeline-frames img')[0].attributes('src')).not.toContain('&retry=1')
    wrapper.unmount()
  })
})
