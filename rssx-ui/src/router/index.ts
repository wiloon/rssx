import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { getJwtToken } from '@/utils/auth'

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'Reader', component: () => import('@/views/Reader.vue') },
  { path: '/login', name: 'Login', component: () => import('@/views/Login.vue') },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/Register.vue')
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

const publicRoutes = new Set(['Login', 'Register'])

router.beforeEach((to) => {
  if (publicRoutes.has(to.name as string)) return true
  if (!getJwtToken()) return { name: 'Login' }
  return true
})

export default router
