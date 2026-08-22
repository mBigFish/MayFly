<template>
  <div>
    <div class="page-title">脚本生成</div>
    <div class="glass-card" style="padding: 20px; margin-bottom: 15px;">
      <div style="color: var(--text-secondary); margin-bottom: 10px;">WebShell 脚本</div>
      <div class="toolbar">
        <select v-model="shellType" class="input" style="width: 120px;">
          <option value="php">PHP</option>
          <option value="jsp">JSP</option>
          <option value="asp">ASP</option>
          <option value="aspx">ASPX</option>
        </select>
        <input v-model="shellPassword" class="input" style="width: 180px; margin-left: 10px;" placeholder="连接密码（默认 mayfly）" />
        <button type="button" class="btn btn-primary" @click="generateShell">
          <el-icon><MagicStick /></el-icon> 生成脚本
        </button>
        <button type="button" class="btn" @click="copy(shellScript)" :disabled="!shellScript">
          <el-icon><CopyDocument /></el-icon> 复制
        </button>
        <div class="toolbar-spacer" />
      </div>
      <pre v-if="shellScript" style="margin-top: 15px; background: #0a0e15; border: 1px solid var(--border); border-radius: 4px; padding: 15px; overflow: auto; color: var(--text-primary); font-size: 13px; max-height: 400px;">{{ shellScript }}</pre>
    </div>

    <div class="glass-card" style="padding: 20px;">
      <div style="color: var(--text-secondary); margin-bottom: 10px;">反向 Shell Payload</div>
      <div class="toolbar">
        <input v-model="reverseHost" class="input" style="width: 180px;" placeholder="监听 IP" />
        <input v-model.number="reversePort" type="number" class="input" style="width: 100px;" placeholder="端口" />
        <button type="button" class="btn btn-primary" @click="generateReverse">
          <el-icon><MagicStick /></el-icon> 生成
        </button>
        <div class="toolbar-spacer" />
      </div>
      <table v-if="payloads.length > 0" class="data-table" style="margin-top: 10px;">
        <thead>
          <tr>
            <th style="width: 100px;">类型</th>
            <th>命令</th>
            <th style="width: 80px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in payloads" :key="p.type">
            <td><span class="tag tag-aspx">{{ p.label }}</span></td>
            <td style="font-family: monospace; font-size: 12px; word-break: break-all;">{{ p.command }}</td>
            <td>
              <button type="button" class="btn" style="padding: 2px 8px; font-size: 12px;" @click="copy(p.command)">复制</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state" style="margin-top: 15px;">
        填写监听 IP 和端口，点击生成
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { payloadApi, type Payload } from '@/api/payload'

const shellType = ref('php')
const shellPassword = ref('')
const shellScript = ref('')
const reverseHost = ref('')
const reversePort = ref(4444)
const payloads = ref<Payload[]>([])

async function generateShell() {
  try {
    const res = await payloadApi.shell(shellType.value, shellPassword.value || undefined)
    shellScript.value = res.script
  } catch {}
}

async function generateReverse() {
  if (!reverseHost.value) {
    ElMessage.warning('请填写监听 IP')
    return
  }
  try {
    payloads.value = await payloadApi.reverse(reverseHost.value, reversePort.value)
  } catch {}
}

async function copy(text: string) {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制')
  } catch {
    ElMessage.error('复制失败')
  }
}
</script>
