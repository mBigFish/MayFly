<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { createSession } from '../api/session'

const route = useRoute()
const targetId = Number(route.params.id)

const terminalEl = ref<HTMLDivElement>()
let term: Terminal | null = null
let fitAddon: FitAddon | null = null
let ws: WebSocket | null = null

onMounted(async () => {
  // 1. 创建会话。
  let sessionId = ''
  try {
    const res = await createSession(targetId)
    sessionId = res.data.data?.id ?? ''
  } catch {
    return
  }
  if (!sessionId) {
    ElMessage.error('创建会话失败')
    return
  }

  // 2. 初始化 xterm。
  term = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    theme: { background: '#1e1e1e', foreground: '#d4d4d4' },
  })
  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(terminalEl.value!)
  fitAddon.fit()

  // 3. 建立 WebSocket。
  const token = localStorage.getItem('token') || ''
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  ws = new WebSocket(`${proto}://${location.host}/ws/v1/session/${sessionId}?token=${token}`)

  ws.onopen = () => {
    term!.writeln('\x1b[32m已连接到目标，输入命令开始。\x1b[0m')
  }
  ws.onmessage = (e) => {
    term!.write(e.data + '\r\n')
  }
  ws.onerror = () => {
    term!.writeln('\x1b[31m连接错误\x1b[0m')
  }
  ws.onclose = () => {
    term!.writeln('\x1b[33m连接已断开\x1b[0m')
  }

  // 4. 键盘输入 → WebSocket。
  term.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(data)
    }
  })

  window.addEventListener('resize', handleResize)
})

function handleResize() {
  fitAddon?.fit()
}

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  ws?.close()
  term?.dispose()
})
</script>

<template>
  <div class="terminal-page">
    <h2 class="title">终端</h2>
    <div ref="terminalEl" class="terminal-container" />
  </div>
</template>

<style scoped>
.terminal-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 100px);
}
.title {
  margin: 0 0 12px;
}
.terminal-container {
  flex: 1;
  background: #1e1e1e;
  border-radius: 4px;
  overflow: hidden;
}
</style>
