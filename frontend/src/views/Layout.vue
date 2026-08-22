<template>
  <div class="app-layout">
    <!-- 侧边栏 -->
    <aside class="app-sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-logo">
        {{ sidebarCollapsed ? 'M' : 'Mayfly' }}
      </div>
      <nav class="sidebar-nav">
        <div
          v-for="item in menuItems"
          :key="item.path"
          class="sidebar-nav-item"
          :class="{ active: isActive(item.path) }"
          @click="navigate(item.path)"
        >
          <el-icon><component :is="item.icon" /></el-icon>
          <span v-if="!sidebarCollapsed">{{ item.label }}</span>
        </div>
      </nav>
      <div class="sidebar-footer" style="padding: 10px 20px; border-top: 1px solid var(--border);">
        <div class="sidebar-nav-item" @click="toggleSidebar">
          <el-icon><Fold v-if="!sidebarCollapsed" /><Expand v-else /></el-icon>
          <span v-if="!sidebarCollapsed">收起菜单</span>
        </div>
      </div>
    </aside>

    <!-- 主区域 -->
    <div class="app-main">
      <header class="app-header">
        <div class="app-header-title">{{ currentTitle }}</div>
        <div class="app-header-right">
          <el-dropdown @command="handleCommand">
            <span style="cursor: pointer; display: flex; align-items: center; gap: 6px;">
              <el-icon><User /></el-icon>
              {{ auth.username }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="settings">系统设置</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>
      <main class="app-content">
        <router-view v-slot="{ Component }">
          <keep-alive include="Terminal">
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const sidebarCollapsed = ref(false)

const menuItems = [
  { path: '/home', label: '首页', icon: 'HomeFilled' },
  { path: '/targets', label: '目标管理', icon: 'Aim' },
  { path: '/files', label: '文件管理', icon: 'FolderOpened' },
  { path: '/servers', label: 'SSH 服务器', icon: 'Platform' },
  { path: '/terminal', label: '终端', icon: 'Monitor' },
  { path: '/listener', label: '监听器', icon: 'Connection' },
  { path: '/payloads', label: '脚本生成', icon: 'Document' },
  { path: '/tasks', label: '任务中心', icon: 'List' },
  { path: '/request-inspector', label: '请求检查器', icon: 'View' },
  { path: '/plugins', label: '插件管理', icon: 'Cpu' },
  { path: '/audit', label: '审计日志', icon: 'Notebook' },
  { path: '/settings', label: '系统设置', icon: 'Setting' },
]

const currentTitle = computed(() => {
  const item = menuItems.find((m) => m.path === route.path)
  return item?.label || 'Mayfly'
})

function isActive(path: string) {
  return route.path === path
}

function navigate(path: string) {
  router.push(path)
}

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

function handleCommand(cmd: string) {
  if (cmd === 'logout') {
    auth.logout()
    router.push('/login')
  } else if (cmd === 'settings') {
    router.push('/settings')
  }
}
</script>
