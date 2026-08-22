<template>
  <div style="height: 100%; display: flex; flex-direction: column;">
    <div class="page-title">终端</div>
    <div class="glass-card" style="flex: 1; display: flex; flex-direction: column; overflow: hidden;">
      <div class="toolbar" style="padding: 10px 15px; margin-bottom: 0; border-bottom: 1px solid var(--border);">
        <button type="button" class="btn btn-primary" @click="openLocalTerminal">
          <el-icon><Monitor /></el-icon> 本地终端
        </button>
        <select v-model="selectedServerId" class="input" style="width: 200px; margin-left: 10px;">
          <option :value="0">选择 SSH 服务器...</option>
          <option v-for="s in servers" :key="s.id" :value="s.id">{{ s.name }} ({{ s.host }})</option>
        </select>
        <button type="button" class="btn btn-success" @click="openSSHTerminal" :disabled="!selectedServerId">
          <el-icon><Connection /></el-icon> SSH 连接
        </button>
        <button type="button" v-if="term" class="btn btn-danger" style="margin-left: 10px;" @click="closeTerminal">
          断开
        </button>
        <div class="toolbar-spacer" />
        <span v-if="wsStatus" style="color: var(--text-secondary); font-size: 12px;">{{ wsStatus }}</span>
      </div>
      <div ref="terminalContainer" class="terminal-container" style="flex: 1; margin: 10px; position: relative;">
        <div v-if="!term" class="empty-state" style="display: flex; align-items: center; justify-content: center; height: 100%;">
          请选择终端类型
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
export default { name: 'Terminal' }
</script>

<script setup lang="ts">
import { ref, shallowRef, onUnmounted, onMounted, onActivated, nextTick } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { ElMessage } from 'element-plus'
import { serverApi, type Server } from '@/api/server'

const terminalContainer = ref<HTMLElement>()
// 使用 shallowRef 避免 Vue 对 xterm Terminal 实例做深度响应式代理
const term = shallowRef<Terminal | null>(null)
const wsStatus = ref('')
const servers = ref<Server[]>([])
const selectedServerId = ref<number>(0)
let ws: WebSocket | null = null
let fitAddon: FitAddon | null = null
let resizeHandler: (() => void) | null = null

async function loadServers() {
  try {
    servers.value = await serverApi.list() || []
  } catch {}
}

async function createTerminal(): Promise<boolean> {
  // 先清理旧终端
  if (term.value) {
    term.value.dispose()
    term.value = null
  }
  if (resizeHandler) {
    window.removeEventListener('resize', resizeHandler)
    resizeHandler = null
  }

  // 等待 DOM 更新完成
  await nextTick()

  const container = terminalContainer.value
  if (!container) {
    ElMessage.error('终端容器未就绪')
    return false
  }

  // 清空容器内容
  while (container.firstChild) {
    container.removeChild(container.firstChild)
  }

  const t = new Terminal({
    fontSize: 14,
    fontFamily: 'Consolas, "Courier New", monospace',
    theme: {
      background: '#0a0e15',
      foreground: '#e0e6ed',
      cursor: '#00d4ff',
      selectionBackground: 'rgba(0, 212, 255, 0.3)',
    },
    cursorBlink: true,
  })

  fitAddon = new FitAddon()
  t.loadAddon(fitAddon)
  t.open(container)
  fitAddon.fit()

  term.value = t

  // 终端输入发送到 WebSocket
  t.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(data)
    }
  })

  // 窗口大小变化时自适应 + 发送 resize 消息
  resizeHandler = () => {
    fitAddon?.fit()
    sendResize()
  }
  window.addEventListener('resize', resizeHandler)

  return true
}

function sendResize() {
  if (!term.value || !fitAddon) return
  if (ws && ws.readyState === WebSocket.OPEN) {
    const msg = JSON.stringify({
      type: 'resize',
      cols: term.value.cols,
      rows: term.value.rows,
    })
    ws.send(msg)
  }
}

function connectWS(wsUrl: string, onOpen: () => void) {
  wsStatus.value = '连接中...'
  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    wsStatus.value = '已连接'
    onOpen()
    // 连接后发送初始窗口大小
    setTimeout(() => sendResize(), 100)
  }

  ws.onmessage = (event) => {
    term.value?.write(event.data)
  }

  ws.onclose = () => {
    wsStatus.value = '已断开'
    term.value?.writeln('\r\n\x1b[31m[Mayfly] 连接已断开\x1b[0m')
  }

  ws.onerror = () => {
    wsStatus.value = '连接错误'
    ElMessage.error('WebSocket 连接错误')
  }
}

async function openLocalTerminal() {
  const ok = await createTerminal()
  if (!ok) return
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws/terminal?type=local`
  connectWS(wsUrl, () => {
    term.value?.writeln('\x1b[32m[Mayfly] 本地终端已连接\x1b[0m')
  })
}

async function openSSHTerminal() {
  if (!selectedServerId.value) {
    ElMessage.warning('请先选择 SSH 服务器')
    return
  }

  const ok = await createTerminal()
  if (!ok) return
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const cols = term.value?.cols || 80
  const rows = term.value?.rows || 24
  const wsUrl = `${protocol}//${window.location.host}/ws/terminal?type=ssh&server_id=${selectedServerId.value}&cols=${cols}&rows=${rows}`
  connectWS(wsUrl, () => {
    const server = servers.value.find(s => s.id === selectedServerId.value)
    term.value?.writeln(`\x1b[32m[Mayfly] 正在连接 ${server?.name} (${server?.host})...\x1b[0m`)
  })
}

function closeTerminal() {
  if (ws) {
    ws.close()
    ws = null
  }
  if (term.value) {
    term.value.dispose()
    term.value = null
  }
  if (resizeHandler) {
    window.removeEventListener('resize', resizeHandler)
    resizeHandler = null
  }
  wsStatus.value = ''
}

onMounted(() => loadServers())

// keep-alive 重新激活时，重适应终端大小
onActivated(() => {
  if (term.value && fitAddon) {
    nextTick(() => {
      fitAddon?.fit()
      sendResize()
    })
  }
})

onUnmounted(() => {
  if (ws) {
    ws.close()
  }
  if (term.value) {
    term.value.dispose()
  }
  if (resizeHandler) {
    window.removeEventListener('resize', resizeHandler)
  }
})
</script>
