import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/tasks'
  },
  {
    path: '/tasks',
    name: 'TaskList',
    component: () => import('@/views/TaskList.vue')
  },
  {
    path: '/tasks/create',
    name: 'TaskCreate',
    component: () => import('@/views/TaskCreate.vue')
  },
  {
    path: '/tasks/:id',
    name: 'TaskDetail',
    component: () => import('@/views/TaskDetail.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
