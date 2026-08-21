<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

function handleLogout() {
  authStore.clearAuth()
  router.push({ name: 'login' })
}
</script>

<template>
  <el-container class="layout">
    <el-header class="header">
      <div class="header-left">
        <span class="brand">WebShell Manager</span>
        <el-menu mode="horizontal" :default-active="$route.name as string" router class="nav">
          <el-menu-item index="dashboard" :route="{ name: 'dashboard' }">概览</el-menu-item>
          <el-menu-item index="targets" :route="{ name: 'targets' }">目标</el-menu-item>
        </el-menu>
      </div>
      <div class="header-right">
        <span class="username">{{ authStore.user?.username || '用户' }}</span>
        <el-button link type="danger" @click="handleLogout">退出</el-button>
      </div>
    </el-header>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
</template>

<style scoped>
.layout {
  min-height: 100vh;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #e5e7eb;
  padding: 0 24px;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 24px;
}
.brand {
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}
.nav {
  border-bottom: none;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.username {
  color: #374151;
}
</style>
