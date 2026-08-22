<template>
  <div style="height: 100%; display: flex; flex-direction: column;">
    <div class="page-title">任务中心</div>
    <div class="glass-card" style="flex: 1; display: flex; flex-direction: column; overflow: hidden;">
      <div class="toolbar" style="padding: 10px 15px; margin-bottom: 0; border-bottom: 1px solid var(--border);">
        <button type="button" class="btn btn-primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon> 创建任务
        </button>
        <button type="button" class="btn" @click="loadList">
          <el-icon><Refresh /></el-icon> 刷新
        </button>
        <div class="toolbar-spacer" />
      </div>

      <div style="flex: 1; overflow: auto;">
        <table class="data-table" v-if="tasks.length > 0">
          <thead>
            <tr>
              <th style="width: 40px;">ID</th>
              <th>名称</th>
              <th style="width: 120px;">类型</th>
              <th style="width: 100px;">状态</th>
              <th style="width: 100px;">进度</th>
              <th style="width: 160px;">创建时间</th>
              <th style="width: 180px;">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in tasks" :key="t.id" @click="selectTask(t)" :class="{ selected: selectedTask?.id === t.id }" style="cursor: pointer;">
              <td>{{ t.id }}</td>
              <td>{{ t.name }}</td>
              <td><span class="op-tag">{{ formatType(t.type) }}</span></td>
              <td>
                <span :class="t.status === 'completed' ? 'status-ok' : t.status === 'running' ? 'status-running' : t.status === 'failed' ? 'status-error' : t.status === 'cancelled' ? 'status-cancelled' : ''">{{ t.status }}</span>
              </td>
              <td>
                <div class="progress-bar">
                  <div class="progress-fill" :style="{ width: t.total > 0 ? (t.done / t.total * 100) + '%' : '0%' }"></div>
                  <span class="progress-text">{{ t.done }}/{{ t.total }}</span>
                </div>
              </td>
              <td>{{ formatTime(t.created_at) }}</td>
              <td>
                <button type="button" v-if="t.status === 'running' || t.status === 'pending'" class="btn btn-sm" @click.stop="cancelTask(t)">取消</button>
                <button type="button" class="btn btn-sm" @click.stop="selectTask(t)">详情</button>
                <button type="button" class="btn btn-sm btn-danger" @click.stop="deleteTask(t)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-state" style="padding: 40px; text-align: center; color: var(--text-secondary);">
          暂无任务
        </div>
      </div>
    </div>

    <!-- 创建任务弹窗 -->
    <div v-if="showCreateDialog" class="modal-overlay" @click.self="showCreateDialog = false">
      <div class="modal-content" style="width: 500px;">
        <div class="modal-header">
          <span>创建任务</span>
          <button type="button" class="modal-close" @click="showCreateDialog = false">&times;</button>
        </div>
        <div style="padding: 15px;">
          <div style="margin-bottom: 12px;">
            <label class="form-label">任务名称</label>
            <input v-model="newTask.name" class="input" style="width: 100%;" placeholder="任务名称" />
          </div>
          <div style="margin-bottom: 12px;">
            <label class="form-label">任务类型</label>
            <select v-model="newTask.type" class="input" style="width: 100%;">
              <option value="batch_check">批量连接测试</option>
              <option value="batch_command">批量命令执行</option>
            </select>
          </div>
          <div v-if="newTask.type === 'batch_command'" style="margin-bottom: 12px;">
            <label class="form-label">执行命令</label>
            <input v-model="newTask.command" class="input" style="width: 100%;" placeholder="如: whoami" />
          </div>
          <div style="margin-bottom: 12px;">
            <label class="form-label">选择目标</label>
            <div class="target-list">
              <label v-for="t in targets" :key="t.id" class="target-item">
                <input type="checkbox" :value="t.id" v-model="newTask.targetIds" />
                <span>{{ t.name }} ({{ t.type }})</span>
              </label>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn" @click="showCreateDialog = false">取消</button>
          <button type="button" class="btn btn-primary" @click="createTask">创建并执行</button>
        </div>
      </div>
    </div>

    <!-- 任务结果弹窗 -->
    <div v-if="selectedTask" class="modal-overlay" @click.self="selectedTask = null">
      <div class="modal-content" style="width: 700px; max-height: 80%; display: flex; flex-direction: column;">
        <div class="modal-header">
          <span>任务详情: {{ selectedTask.name }}</span>
          <button type="button" class="modal-close" @click="selectedTask = null">&times;</button>
        </div>
        <div style="padding: 15px; overflow: auto; flex: 1;">
          <div style="display: flex; gap: 20px; margin-bottom: 15px;">
            <div><span class="form-label">类型:</span> {{ formatType(selectedTask.type) }}</div>
            <div><span class="form-label">状态:</span> {{ selectedTask.status }}</div>
            <div><span class="form-label">进度:</span> {{ selectedTask.done }}/{{ selectedTask.total }}</div>
          </div>
          <div v-if="taskResults.length > 0">
            <div class="form-label" style="margin-bottom: 8px;">执行结果</div>
            <table class="result-table">
              <thead>
                <tr>
                  <th>目标</th>
                  <th style="width: 80px;">状态</th>
                  <th>输出</th>
                  <th style="width: 80px;">耗时</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="r in taskResults" :key="r.target_id">
                  <td>{{ r.target_name }}</td>
                  <td><span :class="r.status === 'ok' ? 'status-ok' : 'status-error'">{{ r.status }}</span></td>
                  <td><pre class="result-output">{{ r.output || r.error || '-' }}</pre></td>
                  <td>{{ r.duration || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else style="color: var(--text-secondary); text-align: center; padding: 20px;">
            暂无结果数据
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { taskApi, type Task } from '@/api/task'
import { targetApi, type Target } from '@/api/target'

const tasks = ref<Task[]>([])
const targets = ref<Target[]>([])
const selectedTask = ref<Task | null>(null)
const showCreateDialog = ref(false)
const newTask = reactive({
  name: '',
  type: 'batch_check',
  command: '',
  targetIds: [] as number[],
})

let pollTimer: number | null = null

const taskResults = computed(() => {
  if (!selectedTask.value?.result) return []
  try {
    const parsed = JSON.parse(selectedTask.value.result)
    return parsed.results || []
  } catch {
    return []
  }
})

async function loadList() {
  try {
    tasks.value = await taskApi.list() || []
    // 如果有选中的任务，更新选中任务的数据
    if (selectedTask.value) {
      const updated = tasks.value.find(t => t.id === selectedTask.value!.id)
      if (updated) selectedTask.value = updated
    }
  } catch {}
}

async function loadTargets() {
  try {
    targets.value = await targetApi.list() || []
  } catch {}
}

function openCreateDialog() {
  newTask.name = ''
  newTask.type = 'batch_check'
  newTask.command = ''
  newTask.targetIds = []
  showCreateDialog.value = true
  loadTargets()
}

async function createTask() {
  if (!newTask.name) {
    ElMessage.warning('请填写任务名称')
    return
  }
  if (newTask.targetIds.length === 0) {
    ElMessage.warning('请选择至少一个目标')
    return
  }
  if (newTask.type === 'batch_command' && !newTask.command) {
    ElMessage.warning('请填写执行命令')
    return
  }

  const payload = JSON.stringify({
    target_ids: newTask.targetIds,
    command: newTask.command,
  })

  try {
    await taskApi.create({ name: newTask.name, type: newTask.type, payload })
    ElMessage.success('任务已创建并开始执行')
    showCreateDialog.value = false
    await loadList()
  } catch {}
}

function selectTask(t: Task) {
  selectedTask.value = t
}

async function cancelTask(t: Task) {
  try {
    await taskApi.cancel(t.id)
    ElMessage.success('已取消')
    await loadList()
  } catch {}
}

async function deleteTask(t: Task) {
  try {
    await ElMessageBox.confirm(`确定删除任务 "${t.name}" 吗？`, '确认', { type: 'warning' })
    await taskApi.delete(t.id)
    ElMessage.success('已删除')
    if (selectedTask.value?.id === t.id) selectedTask.value = null
    await loadList()
  } catch {}
}

function formatType(type: string): string {
  const map: Record<string, string> = {
    batch_check: '批量测试',
    batch_command: '批量执行',
    custom: '自定义',
  }
  return map[type] || type
}

function formatTime(t: string): string {
  if (!t) return ''
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(() => {
  loadList()
  // 自动刷新（有运行中任务时）
  pollTimer = window.setInterval(() => {
    if (tasks.value.some(t => t.status === 'running' || t.status === 'pending')) {
      loadList()
    }
  }, 3000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
.data-table {
  width: 100%;
  border-collapse: collapse;
}
.data-table th {
  text-align: left;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: normal;
  position: sticky;
  top: 0;
  background: var(--bg-secondary);
  z-index: 1;
}
.data-table td {
  padding: 6px 12px;
  border-bottom: 1px solid var(--border);
  font-size: 13px;
}
.data-table tr:hover {
  background: rgba(0, 212, 255, 0.05);
}
.data-table tr.selected {
  background: rgba(0, 212, 255, 0.1);
}
.op-tag {
  background: rgba(0, 212, 255, 0.15);
  color: var(--accent);
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
}
.status-ok { color: var(--success); }
.status-error { color: var(--danger); }
.status-running { color: var(--warning); }
.status-cancelled { color: var(--text-secondary); }
.progress-bar {
  position: relative;
  width: 80px;
  height: 20px;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 3px;
  overflow: hidden;
}
.progress-fill {
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, var(--accent), rgba(0, 212, 255, 0.5));
  transition: width 0.3s;
}
.progress-text {
  position: relative;
  font-size: 11px;
  line-height: 20px;
  text-align: center;
  display: block;
}
.btn-sm {
  padding: 2px 8px;
  font-size: 12px;
  margin-right: 4px;
}
.form-label {
  display: block;
  margin-bottom: 5px;
  color: var(--text-secondary);
  font-size: 12px;
}
.target-list {
  max-height: 200px;
  overflow: auto;
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 8px;
}
.target-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0;
  font-size: 13px;
  cursor: pointer;
}
.target-item:hover {
  color: var(--accent);
}
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}
.modal-content {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.modal-header {
  padding: 12px 15px;
  border-bottom: 1px solid var(--border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;
}
.modal-close {
  background: none;
  border: none;
  color: var(--text-secondary);
  font-size: 20px;
  cursor: pointer;
}
.modal-footer {
  padding: 10px 15px;
  border-top: 1px solid var(--border);
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.result-table {
  width: 100%;
  border-collapse: collapse;
}
.result-table th {
  text-align: left;
  padding: 6px 10px;
  border-bottom: 1px solid var(--border);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: normal;
}
.result-table td {
  padding: 6px 10px;
  border-bottom: 1px solid var(--border);
  font-size: 12px;
  vertical-align: top;
}
.result-output {
  background: rgba(0, 0, 0, 0.3);
  padding: 4px 6px;
  border-radius: 3px;
  font-size: 11px;
  max-height: 100px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}
</style>
