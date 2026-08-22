import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      component: () => import('@/views/Layout.vue'),
      redirect: '/home',
      children: [
        {
          path: 'home',
          name: 'Home',
          component: () => import('@/views/Home.vue'),
          meta: { title: '首页' },
        },
        {
          path: 'targets',
          name: 'Targets',
          component: () => import('@/views/Targets.vue'),
          meta: { title: '目标管理' },
        },
        {
          path: 'files',
          name: 'FileManager',
          component: () => import('@/views/FileManager.vue'),
          meta: { title: '文件管理' },
        },
        {
          path: 'servers',
          name: 'Servers',
          component: () => import('@/views/Servers.vue'),
          meta: { title: 'SSH 服务器' },
        },
        {
          path: 'terminal',
          name: 'Terminal',
          component: () => import('@/views/Terminal.vue'),
          meta: { title: '终端' },
        },
        {
          path: 'listener',
          name: 'Listener',
          component: () => import('@/views/Listener.vue'),
          meta: { title: '监听器' },
        },
        {
          path: 'payloads',
          name: 'Payloads',
          component: () => import('@/views/Payloads.vue'),
          meta: { title: '脚本生成' },
        },
        {
          path: 'tasks',
          name: 'Tasks',
          component: () => import('@/views/Tasks.vue'),
          meta: { title: '任务中心' },
        },
        {
          path: 'audit',
          name: 'Audit',
          component: () => import('@/views/Audit.vue'),
          meta: { title: '审计日志' },
        },
        {
          path: 'request-inspector',
          name: 'RequestInspector',
          component: () => import('@/views/RequestInspector.vue'),
          meta: { title: '请求检查器' },
        },
        {
          path: 'plugins',
          name: 'Plugins',
          component: () => import('@/views/Plugins.vue'),
          meta: { title: '插件管理' },
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('@/views/Settings.vue'),
          meta: { title: '系统设置' },
        },
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    next()
  } else if (!auth.token) {
    next('/login')
  } else {
    next()
  }
})

export default router
