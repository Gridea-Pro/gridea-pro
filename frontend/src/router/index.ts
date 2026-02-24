import { createRouter, createWebHashHistory, RouteRecordRaw } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'
import Articles from '@/views/articles/index.vue'
import Menu from '@/views/menu/index.vue'
import Tags from '@/views/tags/index.vue'
import Categories from '@/views/categories/index.vue'
import Links from '@/views/links/index.vue'
import Theme from '@/views/theme/index.vue'
import Setting from '@/views/settings/index.vue'
import Comment from '@/views/comments/index.vue'
import Memo from '@/views/memos/index.vue'

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
        path: '/memos',
        name: 'memos',
        component: Memo,
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
        path: '/settings',
        name: 'settings',
        component: Setting,
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
 
