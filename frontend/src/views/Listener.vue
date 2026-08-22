<template>
  <div>
    <div class="page-title">监听器</div>
    <div class="glass-card" style="padding: 20px;">
      <div class="toolbar">
        <button type="button" class="btn btn-primary" @click="showAddDialog = true">
          <el-icon><Plus /></el-icon> 创建监听
        </button>
        <button type="button" class="btn" @click="loadList">
          <el-icon><Refresh /></el-icon> 刷新
        </button>
        <div class="toolbar-spacer" />
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>名称</th>
            <th>监听地址</th>
            <th>端口</th>
            <th>状态</th>
            <th>连接数</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="listeners.length === 0">
            <td colspan="6" class="empty-state">暂无监听器</td>
          </tr>
          <tr v-for="l in listeners" :key="l.id">
            <td>{{ l.name }}</td>
            <td>{{ l.host }}</td>
            <td>{{ l.port }}</td>
            <td>
              <span class="status-dot" :class="l.status === 'running' ? 'status-online' : 'status-offline'"></span>
              {{ l.status === 'running' ? '运行中' : '已停止' }}
            </td>
            <td>{{ l.connections }}</td>
            <td>
              <button type="button" v-if="l.status !== 'running'" class="btn btn-success" style="padding: 2px 8px; font-size: 12px;" @click="startListener(l)">启动</button>
              <button type="button" v-else class="btn" style="padding: 2px 8px; font-size: 12px;" @click="stopListener(l)">停止</button>
              <button type="button" class="btn btn-danger" style="padding: 2px 8px; font-size: 12px; margin-left: 5px;" @click="deleteListener(l)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showAddDialog" class="modal-overlay" @click.self="showAddDialog = false">
      <div class="modal">
        <div class="modal-header">
          <span class="modal-title">创建监听器</span>
          <span style="cursor: pointer;" @click="showAddDialog = false">&times;</span>
        </div>
        <div class="modal-body">
          <div style="margin-bottom: 12px;">
            <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">名称</label>
            <input v-model="newListener.name" class="input" placeholder="监听器名称" />
          </div>
          <div style="margin-bottom: 12px;">
            <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">监听地址</label>
            <input v-model="newListener.host" class="input" placeholder="0.0.0.0" />
          </div>
          <div style="margin-bottom: 12px;">
            <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">端口</label>
            <input v-model.number="newListener.port" type="number" class="input" placeholder="4444" />
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn" @click="showAddDialog = false">取消</button>
          <button type="button" class="btn btn-primary" @click="addListener">创建</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listenerApi, type Listener } from '@/api/listener'

const listeners = ref<Listener[]>([])
const showAddDialog = ref(false)
const newListener = reactive({ name: '', host: '0.0.0.0', port: 4444 })

async function loadList() {
  try {
    listeners.value = await listenerApi.list() || []
  } catch {}
}

async function addListener() {
  if (!newListener.name || !newListener.port) {
    ElMessage.warning('请填写完整')
    return
  }
  try {
    await listenerApi.create({ name: newListener.name, host: newListener.host, port: newListener.port })
    ElMessage.success('创建成功')
    showAddDialog.value = false
    newListener.name = ''
    newListener.host = '0.0.0.0'
    newListener.port = 4444
    await loadList()
  } catch {}
}

async function startListener(l: Listener) {
  try {
    await listenerApi.start(l.id)
    ElMessage.success('已启动')
    await loadList()
  } catch {}
}

async function stopListener(l: Listener) {
  try {
    await listenerApi.stop(l.id)
    ElMessage.success('已停止')
    await loadList()
  } catch {}
}

async function deleteListener(l: Listener) {
  try {
    await ElMessageBox.confirm(`确定删除监听器 "${l.name}" 吗？`, '确认', { type: 'warning' })
    await listenerApi.delete(l.id)
    ElMessage.success('已删除')
    await loadList()
  } catch {}
}

onMounted(() => loadList())
</script>
