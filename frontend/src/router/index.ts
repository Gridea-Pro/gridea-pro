import { createRouter, createWebHashHistory, RouteRecordRaw } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'
import Articles from '@/views/article/Articles.vue'
import Menu from '@/views/menu/Index.vue'
import Tags from '@/views/tags/Index.vue'
import Categories from '@/views/categories/Index.vue'
import Links from '@/views/links/Index.vue'
import Theme from '@/views/theme/Index.vue'
import Setting from '@/views/setting/Index.vue'
import Comment from '@/views/comment/Index.vue'
import Loading from '@/views/loading/Index.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'main',
    component: MainLayout,
    redirect: '/articles',
    children: [
      {
        path: '/articles',
        name: 'articles',
        component: Articles,
      },
      {
        path: '/comments',
        name: 'comments',
        component: Comment,
      },
      {
        path: '/menu',
        name: 'menu',
        component: Menu,
      },
      {
        path: '/tags',
        name: 'tags',
        component: Tags,
      },
      {
        path: '/categories',
        name: 'categories',
        component: Categories,
      },
      {
        path: '/links',
        name: 'links',
        component: Links,
      },
      {
        path: '/theme',
        name: 'theme',
        component: Theme,
      },
      {
        path: '/setting',
        name: 'setting',
        component: Setting,
      },
      {
        path: '/loading',
        name: 'loading',
        component: Loading,
      },
      {
        path: '/:pathMatch(.*)*',
        redirect: '/articles',
      },
    ],
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
