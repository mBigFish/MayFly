<template>
  <div style="height: 100%; display: flex; flex-direction: column;">
    <div class="page-title">请求检查器</div>
    <div class="glass-card" style="flex: 1; display: flex; flex-direction: column; overflow: hidden;">
      <!-- 工具栏 -->
      <div class="toolbar" style="padding: 10px 15px; margin-bottom: 0; border-bottom: 1px solid var(--border);">
        <select v-model="targetId" class="input" style="width: 180px;" @change="loadLogs">
          <option :value="0">全部目标</option>
          <option v-for="t in targets" :key="t.id" :value="t.id">{{ t.name }}</option>
        </select>
        <input v-model="keyword" class="input" style="width: 180px; margin-left: 10px;" placeholder="搜索操作/目标/用户" @keyup.enter="loadLogs" />
        <button type="button" class="btn btn-primary" @click="loadLogs">
          <el-icon><Search /></el-icon> 查询
        </button>
        <div class="toolbar-spacer" />
        <span style="color: var(--text-secondary); font-size: 12px;">共 {{ total }} 条</span>
      </div>

      <div style="flex: 1; display: flex; overflow: hidden;">
        <!-- 请求列表 -->
        <div style="flex: 1; overflow: auto;">
          <table class="log-table" v-if="logs.length > 0">
            <thead>
              <tr>
                <th style="width: 60px;">ID</th>
                <th style="width: 120px;">目标</th>
                <th style="width: 100px;">操作</th>
                <th style="width: 80px;">状态</th>
                <th style="width: 80px;">耗时</th>
                <th style="width: 100px;">用户</th>
                <th style="width: 160px;">时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="log in logs" :key="log.id" @click="selectLog(log)" :class="{ selected: selectedLog?.id === log.id }">
                <td>{{ log.id }}</td>
                <td>{{ log.target_name }}</td>
                <td><span class="op-tag">{{ formatOp(log.operation) }}</span></td>
                <td>
                  <span :class="log.status === 'ok' ? 'status-ok' : 'status-error'">{{ log.status }}</span>
                </td>
                <td>{{ log.duration }}ms</td>
                <td>{{ log.username }}</td>
                <td>{{ formatTime(log.created_at) }}</td>
              </tr>
            </tbody>
          </table>
          <div v-else class="empty-state" style="padding: 40px; text-align: center; color: var(--text-secondary);">
            暂无请求日志
          </div>
        </div>

        <!-- 详情面板 -->
        <div v-if="selectedLog" style="width: 400px; border-left: 1px solid var(--border); overflow: auto; padding: 15px;">
          <div style="margin-bottom: 15px;">
            <div style="color: var(--text-secondary); font-size: 12px; margin-bottom: 4px;">操作类型</div>
            <div>{{ formatOp(selectedLog.operation) }}</div>
          </div>
          <div style="margin-bottom: 15px;">
            <div style="color: var(--text-secondary); font-size: 12px; margin-bottom: 4px;">参数</div>
            <pre class="code-block">{{ selectedLog.params || '-' }}</pre>
          </div>
          <div v-if="selectedLog.error" style="margin-bottom: 15px;">
            <div style="color: var(--text-secondary); font-size: 12px; margin-bottom: 4px;">错误信息</div>
            <pre class="code-block" style="color: var(--danger);">{{ selectedLog.error }}</pre>
          </div>
          <div style="margin-bottom: 15px;">
            <div style="color: var(--text-secondary); font-size: 12px; margin-bottom: 4px;">响应内容</div>
            <pre class="code-block">{{ selectedLog.response || '-' }}</pre>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div style="padding: 10px 15px; border-top: 1px solid var(--border); display: flex; justify-content: center; gap: 8px;">
        <button type="button" class="btn btn-sm" @click="prevPage" :disabled="page <= 1">上一页</button>
        <span style="line-height: 30px; color: var(--text-secondary);">{{ page }} / {{ totalPages }}</span>
        <button type="button" class="btn btn-sm" @click="nextPage" :disabled="page >= totalPages">下一页</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { requestLogApi, type RequestLog } from '@/api/requestLog'
import { targetApi, type Target } from '@/api/target'

const targets = ref<Target[]>([])
const logs = ref<RequestLog[]>([])
const selectedLog = ref<RequestLog | null>(null)
const targetId = ref(0)
const keyword = ref('')
const page = ref(1)
const pageSize = 20
const total = ref(0)

const totalPages = computed(() => Math.ceil(total.value / pageSize) || 1)

async function loadTargets() {
  try {
    targets.value = await targetApi.list() || []
  } catch {}
}

async function loadLogs() {
  try {
    const res = await requestLogApi.list({
      target_id: targetId.value || undefined,
      keyword: keyword.value || undefined,
      page: page.value,
      page_size: pageSize,
    })
    logs.value = res.list || []
    total.value = res.total || 0
    selectedLog.value = null
  } catch {}
}

function selectLog(log: RequestLog) {
  selectedLog.value = log
}

function prevPage() {
  if (page.value > 1) {
    page.value--
    loadLogs()
  }
}

function nextPage() {
  if (page.value < totalPages.value) {
    page.value++
    loadLogs()
  }
}

function formatOp(op: string): string {
  const map: Record<string, string> = {
    command: '执行命令',
    read_file: '读取文件',
    list_dir: '列目录',
    write_file: '写入文件',
    delete_file: '删除文件',
    rename_file: '重命名',
    mkdir: '创建目录',
    system_info: '系统信息',
    db_query: '数据库查询',
  }
  return map[op] || op
}

function formatTime(t: string): string {
  if (!t) return '-'
  return new Date(t).toLocaleString()
}

onMounted(() => {
  loadTargets()
  loadLogs()
})
</script>

<style scoped>
.log-table {
  width: 100%;
  border-collapse: collapse;
}
.log-table th {
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
.log-table td {
  padding: 6px 12px;
  border-bottom: 1px solid var(--border);
  font-size: 13px;
  cursor: pointer;
}
.log-table tr:hover {
  background: rgba(0, 212, 255, 0.05);
}
.log-table tr.selected {
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
.code-block {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 8px;
  font-size: 12px;
  overflow: auto;
  max-height: 200px;
  white-space: pre-wrap;
  word-break: break-all;
}
.btn-sm {
  padding: 4px 12px;
  font-size: 12px;
}
</style>
