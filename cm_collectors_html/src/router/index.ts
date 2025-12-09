import { createRouter, createWebHistory } from 'vue-router'
import AdminLoginView from '@/views/AdminLoginView.vue'
import IndexView from '../views/IndexView.vue'
import PerformerView from '@/views/performerView.vue'
import PerformerBasesListView from '@/views/performerBasesListView.vue'
import SettingView from '@/views/SettingView.vue'
import playMovies from '@/views/play/playMovies.vue'
import playMoviesMobile from '@/views/play/playMoviesMobile.vue'
import playComic from '@/views/play/playComic.vue'
import playComicMobile from '@/views/play/playComicMobile.vue'
import playAtlas from '@/views/play/playAtlas.vue'
import playAtlasMobile from '@/views/play/playAtlasMobile.vue'

import MobileView from '@/views/MobileView.vue'
import { isMobile } from '@/assets/mobile'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/adminLogin',
      name: 'adminLogin',
      component: AdminLoginView,
      meta: { mobileAccess: false },
    },
    {
      path: '/',
      name: 'index',
      component: IndexView,
      meta: { mobileAccess: false },
    },
    {
      path: '/mobile',
      name: 'mobile',
      component: MobileView,
      meta: { mobileAccess: true },
    },
    {
      path: '/performer/:mainPerformerBasesId',
      name: 'performer',
      props: route => ({ mainPerformerBasesId: route.params.mainPerformerBasesId }),
      component: PerformerView,
      meta: { mobileAccess: false },
    },
    {
      path: '/performer/basesList/:filesBasesId',
      name: 'performerBasesList',
      props: route => ({ filesBasesId: route.params.filesBasesId }),
      component: PerformerBasesListView,
      meta: { mobileAccess: false },
    },
    {
      path: '/setting',
      name: 'setting',
      component: SettingView,
      meta: { mobileAccess: false },
    },
    {
      path: '/play/movies/:resourceId/:dramaSeriesId?',
      name: 'playMovies',
      props: route => ({
        resourceId: route.params.resourceId,
        dramaSeriesId: route.params.dramaSeriesId || ''
      }),
      component: playMovies,
      meta: { mobileAccess: false },
    },
    {
      path: '/play/moviesMobile/:resourceId/:dramaSeriesId?',
      name: 'playMoviesMobile',
      props: route => ({
        resourceId: route.params.resourceId,
        dramaSeriesId: route.params.dramaSeriesId || ''
      }),
      component: playMoviesMobile,
      meta: { mobileAccess: true },
    },
    {
      path: '/play/comic/:resourceId/:dramaSeriesId?',
      name: 'playComic',
      props: route => ({
        resourceId: route.params.resourceId,
        dramaSeriesId: route.params.dramaSeriesId || ''
      }),
      component: playComic,
      meta: { mobileAccess: true },
    },
    {
      path: '/play/comicMobile/:resourceId/:dramaSeriesId?',
      name: 'playComicMobile',
      props: route => ({
        resourceId: route.params.resourceId,
        dramaSeriesId: route.params.dramaSeriesId || ''
      }),
      component: playComicMobile,
      meta: { mobileAccess: true },
    },
    {
      path: '/play/atlas/:resourceId/:dramaSeriesId?',
      name: 'playAtlas',
      props: route => ({
        resourceId: route.params.resourceId,
        dramaSeriesId: route.params.dramaSeriesId || ''
      }),
      component: playAtlas,
      meta: { mobileAccess: true },
    },
    {
      path: '/play/atlasMobile/:resourceId/:dramaSeriesId?',
      name: 'playAtlasMobile',
      props: route => ({
        resourceId: route.params.resourceId,
        dramaSeriesId: route.params.dramaSeriesId || ''
      }),
      component: playAtlasMobile,
      meta: { mobileAccess: true },
    },
  ],
})


// 添加路由守卫来检测设备类型
router.beforeEach((to, from, next) => {
  // 获取目标路由的meta信息
  const routeMeta = to.meta;

  // 如果是移动设备且目标页面不允许移动端访问，且没有desktop查询参数，则重定向到mobile页面
  if (isMobile() && routeMeta.mobileAccess === false && !to.query.desktop) {
    next('/mobile');
    return;
  }

  next();
});

export default router
