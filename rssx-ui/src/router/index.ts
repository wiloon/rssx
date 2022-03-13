import Vue from 'vue'
import VueRouter, { RouteConfig } from 'vue-router'
import FeedList from '../views/FeedList.vue'
import FeedNewsList from '../views/FeedNewsList.vue'
import News from '../views/News.vue'
import { getJwtToken } from '@/utils/auth'

Vue.use(VueRouter)

const routes: Array<RouteConfig> = [
  {
    path: '/',
    name: 'FeedList',
    component: FeedList
  },
  {
    path: '/feed-news-list',
    name: 'FeedNewsList',
    component: FeedNewsList
  },
  {
    path: '/news',
    name: 'News',
    component: News
  },
  {
    path: '/about',
    name: 'About',
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () => import(/* webpackChunkName: "about" */ '../views/About.vue')
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import(/* webpackChunkName: "about" */ '../views/Login.vue')
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import(/* webpackChunkName: "about" */ '../views/Register.vue')
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

router.beforeEach((to, from, next) => {
  console.log(`from name: ${from.name}, from path: ${from.fullPath}, to name: ${to.name}, to path: ${to.fullPath}`)
  if (to.fullPath.indexOf('/login') > -1 || to.fullPath.indexOf('/register') > -1) {
    console.log('white list hit, next')
    next()
    return
  }

  const jwtToken = getJwtToken()
  if (jwtToken === undefined || jwtToken === '' || jwtToken === null) {
    next('/login')
  }
  next()
})

export default router
