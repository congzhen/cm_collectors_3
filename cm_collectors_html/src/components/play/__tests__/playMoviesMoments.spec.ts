import { flushPromises, shallowMount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import PlayMovies from '@/views/play/playMovies.vue'
import PlayMoviesMobile from '@/views/play/playMoviesMobile.vue'
import MomentCollection from '../MomentCollection.vue'
import VideoDetailsExtras from '@/components/details/VideoDetailsExtras.vue'

vi.mock('@/server/resource.server', () => ({
  resourceServer: {
    info: vi.fn(async () => ({
      status: true,
      data: {
        id: 'resource',
        mode: 'movies',
        title: '测试视频',
        country: '',
        definition: '',
        abstract: '',
        stars: 0,
        score: 0,
        tags: [],
        directors: [],
        performers: [{ id: 'actor', name: '演员' }],
        dramaSeries: [{ id: 'episode', src: 'video.mp4' }],
      },
    })),
  },
}))
vi.mock('@/storeData/app.storeData', () => ({
  appStoreData: () => ({
    appConfig: {},
    currentFilesBasesAppConfig: { sampleStatus: true, detailsDramaSeriesMode: 'fileName' },
  }),
}))
vi.mock('@/language/app.lang', () => ({
  AppLang: () => ({ performer: () => '演员', director: () => '导演', country: () => '' }),
}))
vi.mock('@/assets/request', () => ({ default: vi.fn(async () => ({ status: true, data: [] })) }))
vi.mock('@/common/photo', () => ({ getResourceCoverPoster: () => '' }))
vi.mock('@/common/play', () => ({ playUpdate: vi.fn() }))
vi.mock('vue-router', () => ({
  useRouter: () => ({
    currentRoute: { value: { path: '/play/moviesMobile/resource/episode' } },
    replace: vi.fn(),
  }),
}))
vi.mock('@/views/HeaderView.vue', () => ({ default: { template: '<header />' } }))
vi.mock('@/views/MobileHeaderView.vue', () => ({ default: { template: '<header />' } }))
vi.mock('@/components/play/resourceVideoPlayer.vue', () => ({ default: { template: '<div />' } }))
vi.mock('@/components/play/mobileVideoPlayer.vue', () => ({ default: { template: '<div />' } }))
vi.mock('@/components/resource/resourceDramaSeriesList.vue', () => ({
  default: { template: '<div />' },
}))
vi.mock('@/components/performer/performerDetails.vue', () => ({ default: { template: '<div />' } }))
vi.mock('@/components/performer/performerPhoto.vue', () => ({ default: { template: '<div />' } }))
vi.mock('@/components/details/detailsSampleImages.vue', () => ({
  default: { name: 'DetailsSampleImages', template: '<div />' },
}))

describe('独立播放页片段收藏入口', () => {
  it.each(['桌面', '手机'])('%s播放页在演员和剧照之间挂载当前资源收藏', async (mode) => {
    const component = mode === '桌面' ? PlayMovies : PlayMoviesMobile
    const wrapper = shallowMount(component, {
      props: { resourceId: 'resource' },
      global: {
        directives: { loading: {} },
        stubs: { VideoDetailsExtras: false, 'el-rate': true, 'el-alert': true, 'el-tag': true, 'el-image': true, 'el-button': true },
      },
    })
    await flushPromises()
    const extras = wrapper.findComponent(VideoDetailsExtras)
    expect(extras.exists()).toBe(true)
    expect(extras.findComponent(MomentCollection).props('resourceId')).toBe('resource')
    const html = wrapper.html()
    const actorPosition = html.indexOf('class="performer-list"')
    const momentsPosition = html.indexOf('class="video-extras"')
    const samplesPosition = html.indexOf('<details-sample-images-stub')
    expect(actorPosition).toBeGreaterThan(-1)
    expect(momentsPosition).toBeGreaterThan(actorPosition)
    expect(samplesPosition).toBeGreaterThan(momentsPosition)
    wrapper.unmount()
  })
})
