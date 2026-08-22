<template>
  <div>
    <div class="page-title">SSH 服务器</div>
    <div class="glass-card" style="padding: 20px;">
      <div class="toolbar">
        <button type="button" class="btn btn-primary" @click="showAddDialog = true">
          <el-icon><Plus /></el-icon> 添加服务器
        </button>
        <button type="button" class="btn" @click="loadList">
          <el-icon><Refresh /></el-icon> 刷新
        </button>
        <div class="toolbar-spacer" />
        <input v-model="search" class="input" style="width: 200px;" placeholder="搜索..." @keyup.enter="loadList" />
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>名称</th>
            <th>主机</th>
            <th>端口</th>
            <th>用户名</th>
            <th>分组</th>
            <th>测试状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="servers.length === 0">
            <td colspan="7" class="empty-state">暂无服务器</td>
          </tr>
          <tr v-for="s in servers" :key="s.id">
            <td>{{ s.name }}</td>
            <td>{{ s.host }}</td>
            <td>{{ s.port }}</td>
            <td>{{ s.username }}</td>
            <td>{{ s.group || '-' }}</td>
            <td>
              <span v-if="s.last_test_status" class="status-dot" :class="s.last_test_status === 'ok' ? 'status-online' : 'status-offline'"></span>
              {{ s.last_test_status === 'ok' ? '正常' : s.last_test_status === 'fail' ? '失败' : '未测试' }}
            </td>
            <td>
              <button type="button" class="btn" style="padding: 2px 8px; font-size: 12px;" @click="testServer(s)">测试</button>
              <button type="button" class="btn btn-danger" style="padding: 2px 8px; font-size: 12px; margin-left: 5px;" @click="deleteServer(s)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showAddDialog" class="modal-overlay" @click.self="showAddDialog = false">
      <div class="modal">
        <div class="modal-header">
          <span class="modal-title">添加 SSH 服务器</span>
          <span style="cursor: pointer;" @click="showAddDialog = false">&times;</span>
        </div>
        <div class="modal-body">
          <div style="margin-bottom: 12px;">
            <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">名称</label>
            <input v-model="newServer.name" class="input" placeholder="服务器名称" />
          </div>
          <div style="display: flex; gap: 10px; margin-bottom: 12px;">
            <div style="flex: 3;">
              <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">主机</label>
              <input v-model="newServer.host" class="input" placeholder="192.168.1.1" />
            </div>
            <div style="flex: 1;">
              <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">端口</label>
              <input v-model.number="newServer.port" type="number" class="input" placeholder="22" />
            </div>
          </div>
          <div style="margin-bottom: 12px;">
            <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">用户名</label>
            <input v-model="newServer.username" class="input" placeholder="root" />
          </div>
          <div style="margin-bottom: 12px;">
            <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">密码</label>
            <input v-model="newServer.password" type="password" class="input" placeholder="密码（可选，有私钥可留空）" />
          </div>
          <div style="margin-bottom: 12px;">
            <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">私钥</label>
            <textarea v-model="newServer.private_key" class="input" style="min-height: 80px; resize: vertical; font-family: monospace; font-size: 12px;" placeholder="-----BEGIN RSA PRIVATE KEY-----&#10;...&#10;-----END RSA PRIVATE KEY-----"></textarea>
          </div>
          <div style="margin-bottom: 12px;">
            <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">分组</label>
            <input v-model="newServer.group" class="input" placeholder="分组名称（可选）" />
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn" @click="showAddDialog = false">取消</button>
          <button type="button" class="btn btn-primary" @click="addServer">确定</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { serverApi, type Server } from '@/api/server'

const servers = ref<Server[]>([])
const search = ref('')
const showAddDialog = ref(false)
const newServer = reactive({ name: '', host: '', port: 22, username: 'root', password: '', private_key: '', group: '' })

async function loadList() {
  try {
    servers.value = await serverApi.list(search.value) || []
  } catch {}
}

async function addServer() {
  if (!newServer.name || !newServer.host) {
    ElMessage.warning('请填写名称和主机')
    return
  }
  try {
    await serverApi.create({
      name: newServer.name,
      host: newServer.host,
      port: newServer.port,
      username: newServer.username,
      password: newServer.password,
      private_key: newServer.private_key,
      group: newServer.group,
    })
    ElMessage.success('创建成功')
    showAddDialog.value = false
    Object.assign(newServer, { name: '', host: '', port: 22, username: 'root', password: '', private_key: '', group: '' })
    await loadList()
  } catch {}
}

async function testServer(s: Server) {
  try {
    ElMessage.info('正在测试连接...')
    await serverApi.test(s.id)
    ElMessage.success('连接成功')
    await loadList()
  } catch {}
}

async function deleteServer(s: Server) {
  try {
    await ElMessageBox.confirm(`确定删除服务器 "${s.name}" 吗？`, '确认', { type: 'warning' })
    await serverApi.delete(s.id)
    ElMessage.success('已删除')
    await loadList()
  } catch {}
}

onMounted(() => loadList())
</script>
