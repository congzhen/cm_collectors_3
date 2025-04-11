import { createRouter, createWebHistory } from 'vue-router'
import IndexView from '../views/IndexView.vue'
import performerView from '@/views/performerView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'index',
      component: IndexView,
    },
    {
      path: '/performer',
      name: 'performer',
      component: performerView,
    },
  ],
})

export default router
