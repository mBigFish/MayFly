<template>
  <div>
    <div class="page-title">系统概览</div>
    <div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 15px; margin-bottom: 20px;">
      <div class="glass-card" style="padding: 20px;">
        <div style="color: var(--text-secondary); font-size: 13px;">目标总数</div>
        <div style="font-size: 28px; font-weight: bold; color: var(--accent); margin-top: 8px;">{{ data?.targets || 0 }}</div>
      </div>
      <div class="glass-card" style="padding: 20px;">
        <div style="color: var(--text-secondary); font-size: 13px;">在线目标</div>
        <div style="font-size: 28px; font-weight: bold; color: var(--success); margin-top: 8px;">{{ data?.targets_online || 0 }}</div>
      </div>
      <div class="glass-card" style="padding: 20px;">
        <div style="color: var(--text-secondary); font-size: 13px;">活跃会话</div>
        <div style="font-size: 28px; font-weight: bold; color: var(--warning); margin-top: 8px;">{{ data?.sessions_active || 0 }}</div>
      </div>
      <div class="glass-card" style="padding: 20px;">
        <div style="color: var(--text-secondary); font-size: 13px;">监听器</div>
        <div style="font-size: 28px; font-weight: bold; color: var(--danger); margin-top: 8px;">{{ data?.listeners || 0 }}</div>
      </div>
      <div class="glass-card" style="padding: 20px;">
        <div style="color: var(--text-secondary); font-size: 13px;">任务总数</div>
        <div style="font-size: 28px; font-weight: bold; color: var(--accent); margin-top: 8px;">{{ data?.tasks || 0 }}</div>
      </div>
      <div class="glass-card" style="padding: 20px;">
        <div style="color: var(--text-secondary); font-size: 13px;">审计日志</div>
        <div style="font-size: 28px; font-weight: bold; color: var(--text-secondary); margin-top: 8px;">{{ data?.audit_logs || 0 }}</div>
      </div>
    </div>
    <div class="glass-card" style="padding: 20px;">
      <div style="color: var(--text-secondary); margin-bottom: 10px;">欢迎使用 Mayfly WebShell 管理平台</div>
      <div style="color: var(--text-muted); font-size: 13px; line-height: 1.8;">
        Mayfly 是一个轻量级的 WebShell 管理工具，支持 PHP/JSP/ASP/ASPX 多种脚本类型。<br />
        提供命令执行、文件管理、数据库操作、虚拟终端、反向 Shell 等功能。
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { dashboardApi, type DashboardData } from '@/api/dashboard'

const data = ref<DashboardData | null>(null)

async function loadData() {
  try {
    data.value = await dashboardApi.get()
  } catch {
    // 错误已在拦截器中处理
  }
}

onMounted(() => loadData())
</script>
