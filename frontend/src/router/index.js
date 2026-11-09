import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/store/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { title: '登录', requiresAuth: false },
  },
  {
    path: '/',
    component: () => import('@/layout/index.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '仪表盘', icon: 'Odometer' },
      },
      {
        path: 'projects',
        name: 'Projects',
        component: () => import('@/views/project/list.vue'),
        meta: { title: '项目列表', icon: 'Folder' },
      },
      {
        path: 'project/:id',
        name: 'ProjectDetail',
        component: () => import('@/views/project/detail.vue'),
        meta: { title: '项目面板', requiresAuth: true },
      },
      {
        path: 'tasks',
        name: 'Tasks',
        component: () => import('@/views/task/list.vue'),
        meta: { title: '任务列表', icon: 'List' },
      },
      {
        path: 'kanban',
        name: 'Kanban',
        component: () => import('@/views/kanban/index.vue'),
        meta: { title: '看板', icon: 'Grid' },
      },
      {
        path: 'gantt',
        name: 'Gantt',
        component: () => import('@/views/gantt/index.vue'),
        meta: { title: '甘特图', icon: 'DataLine' },
      },
      {
        path: 'blocked',
        name: 'Blocked',
        component: () => import('@/views/blocked/index.vue'),
        meta: { title: '阻塞地图', icon: 'Warning' },
      },
      {
        path: 'reports',
        name: 'Reports',
        component: () => import('@/views/report/index.vue'),
        meta: { title: '周报', icon: 'Document' },
      },
      {
        path: 'team',
        name: 'Team',
        component: () => import('@/views/team/index.vue'),
        meta: { title: '团队管理', icon: 'User' },
      },
      {
        path: 'notifications',
        name: 'Notifications',
        component: () => import('@/views/notification/index.vue'),
        meta: { title: '通知中心', icon: 'Bell' },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const { isLoggedIn } = useUserStore()
  // 设置页面标题
  document.title = to.meta.title ? `${to.meta.title} - 破局 Breakout` : '破局 Breakout'
  // 登录页已登录则跳转首页
  if (to.path === '/login') {
    if (isLoggedIn()) {
      next('/dashboard')
    } else {
      next()
    }
    return
  }
  // 需要登录的页面未登录则跳转登录页
  if (to.meta.requiresAuth !== false && !isLoggedIn()) {
    next('/login')
    return
  }
  next()
})

export default router