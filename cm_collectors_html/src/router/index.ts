import { createRouter, createWebHistory } from 'vue-router'
import IndexView from '../views/IndexView.vue'
import PerformerView from '@/views/performerView.vue'
import SettingView from '@/views/SettingView.vue'

import playMovies from '@/views/play/playMovies.vue'

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
    }
  ],
})

export default router
