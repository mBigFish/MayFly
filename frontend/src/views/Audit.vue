<template>
  <div>
    <div class="page-title">审计日志</div>
    <div class="glass-card" style="padding: 20px;">
      <div class="toolbar">
        <input v-model="keyword" class="input" style="width: 200px;" placeholder="搜索用户/操作/资源..." @keyup.enter="loadList" />
        <button type="button" class="btn btn-primary" @click="loadList"><el-icon><Search /></el-icon> 搜索</button>
        <div class="toolbar-spacer" />
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>时间</th>
            <th>用户</th>
            <th>操作</th>
            <th>资源</th>
            <th>详情</th>
            <th>IP</th>
            <th>状态</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="logs.length === 0">
            <td colspan="7" class="empty-state">暂无日志</td>
          </tr>
          <tr v-for="log in logs" :key="log.id">
            <td style="white-space: nowrap;">{{ formatTime(log.created_at) }}</td>
            <td>{{ log.username }}</td>
            <td><span class="tag" :class="'tag-' + log.action.slice(0,3)">{{ log.action }}</span></td>
            <td>{{ log.resource }}{{ log.resource_id ? '#' + log.resource_id : '' }}</td>
            <td style="max-width: 300px; overflow: hidden; text-overflow: ellipsis;">{{ log.detail }}</td>
            <td>{{ log.ip }}</td>
            <td>
              <span class="status-dot" :class="log.status === 'success' ? 'status-online' : 'status-offline'"></span>
              {{ log.status }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { auditApi, type AuditLog } from '@/api/audit'

const logs = ref<AuditLog[]>([])
const keyword = ref('')

async function loadList() {
  try {
    const res = await auditApi.list({ keyword: keyword.value, page: 1, per_page: 50 })
    logs.value = res.list || []
  } catch {}
}

function formatTime(t: string): string {
  if (!t) return ''
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(() => loadList())
</script>
