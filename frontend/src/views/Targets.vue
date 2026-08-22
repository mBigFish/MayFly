<template>
  <div>
    <div class="page-title">目标管理</div>
    <div class="glass-card" style="padding: 20px;">
      <div class="toolbar">
        <button class="btn btn-primary" @click="showAddDialog = true">
          <el-icon><Plus /></el-icon> 添加目标
        </button>
        <button class="btn" @click="loadList">
          <el-icon><Refresh /></el-icon> 刷新
        </button>
        <div class="toolbar-spacer" />
        <input v-model="search" class="input" style="width: 200px;" placeholder="搜索..." />
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>名称</th>
            <th>URL</th>
            <th>类型</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="targets.length === 0">
            <td colspan="5" class="empty-state">暂无目标，请添加</td>
          </tr>
          <tr v-for="t in filteredTargets" :key="t.id">
            <td>{{ t.name }}</td>
            <td style="max-width: 300px; overflow: hidden; text-overflow: ellipsis;">{{ t.url }}</td>
            <td><span class="tag" :class="'tag-' + t.type">{{ t.type }}</span></td>
            <td>
              <span class="status-dot" :class="t.status === 'online' ? 'status-online' : 'status-offline'"></span>
              {{ t.status === 'online' ? '在线' : '离线' }}
            </td>
            <td>
              <button type="button" class="btn" style="padding: 2px 8px; font-size: 12px;" @click="testConnection(t)">测试</button>
              <button type="button" class="btn btn-primary" style="padding: 2px 8px; font-size: 12px; margin-left: 5px;" @click="openCommandDialog(t)">命令</button>
              <button type="button" class="btn btn-danger" style="padding: 2px 8px; font-size: 12px; margin-left: 5px;" @click="deleteTarget(t)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 命令执行对话框 -->
    <div v-if="showCmdDialog" class="modal-overlay" @click.self="showCmdDialog = false">
      <div class="modal" style="min-width: 600px;">
        <div class="modal-header">
          <span class="modal-title">命令执行 - {{ cmdTarget?.name }}</span>
          <span style="cursor: pointer;" @click="showCmdDialog = false">&times;</span>
        </div>
        <div class="modal-body">
          <div style="display: flex; gap: 10px; margin-bottom: 10px;">
            <input v-model="cmdInput" class="input" placeholder="输入命令..." @keyup.enter="executeCommand" />
            <button type="button" class="btn btn-primary" @click="executeCommand" :disabled="cmdLoading">
              {{ cmdLoading ? '执行中...' : '执行' }}
            </button>
          </div>
          <pre v-if="cmdOutput" style="background: #0a0e15; border: 1px solid var(--border); border-radius: 4px; padding: 12px; overflow: auto; color: var(--text-primary); font-size: 13px; max-height: 400px; white-space: pre-wrap; word-break: break-all;">{{ cmdOutput }}</pre>
        </div>
      </div>
    </div>

    <!-- 添加目标对话框 -->
    <div v-if="showAddDialog" class="modal-overlay" @click.self="showAddDialog = false">
      <div class="modal">
        <div class="modal-header">
          <span class="modal-title">添加目标</span>
          <span style="cursor: pointer;" @click="showAddDialog = false">&times;</span>
        </div>
        <div class="modal-body">
          <div style="margin-bottom: 12px;">
            <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">名称</label>
            <input v-model="newTarget.name" class="input" placeholder="目标名称" />
          </div>
          <div style="margin-bottom: 12px;">
            <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">URL</label>
            <input v-model="newTarget.url" class="input" placeholder="http://target.com/shell.php" />
          </div>
          <div style="margin-bottom: 12px;">
            <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">类型</label>
            <select v-model="newTarget.type" class="input">
              <option value="php">PHP</option>
              <option value="jsp">JSP</option>
              <option value="asp">ASP</option>
              <option value="aspx">ASPX</option>
            </select>
          </div>
          <div style="margin-bottom: 12px;">
            <label style="display: block; margin-bottom: 5px; color: var(--text-secondary);">密码</label>
            <input v-model="newTarget.password" class="input" placeholder="WebShell 密码" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="showAddDialog = false">取消</button>
          <button class="btn btn-primary" @click="addTarget">确定</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { targetApi, type Target } from '@/api/target'

const targets = ref<Target[]>([])
const search = ref('')
const showAddDialog = ref(false)
const newTarget = reactive({ name: '', url: '', type: 'php', password: '' })
const showCmdDialog = ref(false)
const cmdTarget = ref<Target | null>(null)
const cmdInput = ref('')
const cmdOutput = ref('')
const cmdLoading = ref(false)

const filteredTargets = computed(() => {
  if (!search.value) return targets.value
  const s = search.value.toLowerCase()
  return targets.value.filter((t) => t.name.toLowerCase().includes(s) || t.url.toLowerCase().includes(s))
})

async function loadList() {
  try {
    const data = await targetApi.list()
    targets.value = data || []
  } catch {
    // 错误已在拦截器中处理
  }
}

async function addTarget() {
  if (!newTarget.name || !newTarget.url) {
    ElMessage.warning('请填写名称和 URL')
    return
  }
  try {
    await targetApi.create({
      name: newTarget.name,
      url: newTarget.url,
      type: newTarget.type,
      password: newTarget.password,
    })
    ElMessage.success('创建成功')
    showAddDialog.value = false
    newTarget.name = ''
    newTarget.url = ''
    newTarget.type = 'php'
    newTarget.password = ''
    await loadList()
  } catch {
    // 错误已在拦截器中处理
  }
}

async function testConnection(t: Target) {
  try {
    ElMessage.info('正在测试连接...')
    const res = await targetApi.check(t.id)
    ElMessage.success('连接成功')
    await loadList()
  } catch {
    // 错误已在拦截器中处理
  }
}

function openCommandDialog(t: Target) {
  cmdTarget.value = t
  cmdInput.value = ''
  cmdOutput.value = ''
  showCmdDialog.value = true
}

async function executeCommand() {
  if (!cmdTarget.value || !cmdInput.value) return
  cmdLoading.value = true
  cmdOutput.value = ''
  try {
    const res = await targetApi.execute(cmdTarget.value.id, cmdInput.value)
    cmdOutput.value = res.output || '(无输出)'
  } catch {
    // 错误已在拦截器中处理
  } finally {
    cmdLoading.value = false
  }
}

async function deleteTarget(t: Target) {
  try {
    await ElMessageBox.confirm(`确定删除目标 "${t.name}" 吗？`, '确认', {
      type: 'warning',
    })
    await targetApi.delete(t.id)
    ElMessage.success('删除成功')
    await loadList()
  } catch {
    // 用户取消或错误
  }
}

onMounted(() => loadList())
</script>
