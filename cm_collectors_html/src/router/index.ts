import { createRouter, createWebHistory } from 'vue-router'
import IndexView from '../views/IndexView.vue'
import PerformerView from '@/views/performerView.vue'
import SettingView from '@/views/SettingView.vue'

import playMovies from '@/views/play/playMovies.vue'
import playComic from '@/views/play/playComic.vue'
import playAtlas from '@/views/play/playAtlas.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'index',
      component: IndexView,
    },
    {
      path: '/performer/:mainPerformerBasesId',
      name: 'performer',
      props: route => ({ mainPerformerBasesId: route.params.mainPerformerBasesId }),
      component: PerformerView,
    },
    {
      path: '/setting',
      name: 'setting',
      component: SettingView,
    },
    {
      path: '/play/movies/:resourceId/:dramaSeriesId?',
      name: 'playMovies',
      props: route => ({
        resourceId: route.params.resourceId,
        dramaSeriesId: route.params.dramaSeriesId || ''
      }),
      component: playMovies,
    },
    {
      path: '/play/comic/:resourceId/:dramaSeriesId?',
      name: 'playComic',
      props: route => ({
        resourceId: route.params.resourceId,
        dramaSeriesId: route.params.dramaSeriesId || ''
      }),
      component: playComic,
    },
    {
      path: '/play/atlas/:resourceId/:dramaSeriesId?',
      name: 'playAtlas',
      props: route => ({
        resourceId: route.params.resourceId,
        dramaSeriesId: route.params.dramaSeriesId || ''
      }),
      component: playAtlas,
    },
  ],
})

export default router
