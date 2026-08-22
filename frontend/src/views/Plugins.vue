<template>
  <div style="height: 100%; display: flex; flex-direction: column;">
    <div class="page-title">插件管理</div>
    <div class="glass-card" style="flex: 1; display: flex; flex-direction: column; overflow: hidden;">
      <div style="flex: 1; overflow: auto; padding: 15px;">
        <div v-if="plugins.length === 0" class="empty-state" style="padding: 40px; text-align: center; color: var(--text-secondary);">
          暂无插件
        </div>
        <div v-else style="display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 15px;">
          <div v-for="p in plugins" :key="p.name" class="plugin-card">
            <div style="display: flex; justify-content: space-between; align-items: start; margin-bottom: 8px;">
              <span style="font-weight: bold; color: var(--accent);">{{ p.name }}</span>
              <span style="font-size: 11px; color: var(--text-secondary);">v{{ p.version }}</span>
            </div>
            <div style="font-size: 12px; color: var(--text-secondary); margin-bottom: 12px; min-height: 36px;">
              {{ p.description }}
            </div>
            <div class="toolbar" style="margin-bottom: 0;">
              <select v-model="pluginTargetId[p.name]" class="input" style="width: 140px; font-size: 12px;">
                <option :value="0">选择目标...</option>
                <option v-for="t in targets" :key="t.id" :value="t.id">{{ t.name }}</option>
              </select>
              <button type="button" class="btn btn-sm btn-primary" @click="executePlugin(p)" :disabled="!pluginTargetId[p.name] || executing[p.name]">
                {{ executing[p.name] ? '执行中...' : '执行' }}
              </button>
            </div>
            <div v-if="pluginResults[p.name]" style="margin-top: 10px;">
              <div style="font-size: 11px; color: var(--text-secondary); margin-bottom: 4px;">结果:</div>
              <pre class="result-block">{{ pluginResults[p.name] }}</pre>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { pluginApi, type Plugin } from '@/api/plugin'
import { targetApi, type Target } from '@/api/target'

const plugins = ref<Plugin[]>([])
const targets = ref<Target[]>([])
const pluginTargetId = reactive<Record<string, number>>({})
const pluginResults = reactive<Record<string, string>>({})
const executing = reactive<Record<string, boolean>>({})

async function loadPlugins() {
  try {
    plugins.value = await pluginApi.list() || []
  } catch {}
}

async function loadTargets() {
  try {
    targets.value = await targetApi.list() || []
  } catch {}
}

async function executePlugin(p: Plugin) {
  const tid = pluginTargetId[p.name]
  if (!tid) {
    ElMessage.warning('请先选择目标')
    return
  }
  executing[p.name] = true
  pluginResults[p.name] = ''
  try {
    const res = await pluginApi.execute(p.name, tid)
    if (typeof res.data === 'object') {
      pluginResults[p.name] = JSON.stringify(res.data, null, 2)
    } else {
      pluginResults[p.name] = String(res.data || '')
    }
    if (res.status !== 'ok') {
      ElMessage.warning(res.message || '执行失败')
    }
  } catch (e: any) {
    pluginResults[p.name] = '错误: ' + (e.message || e)
    ElMessage.error('执行失败')
  } finally {
    executing[p.name] = false
  }
}

onMounted(() => {
  loadPlugins()
  loadTargets()
})
</script>

<style scoped>
.plugin-card {
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 15px;
}
.result-block {
  background: rgba(0, 0, 0, 0.4);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 8px;
  font-size: 11px;
  overflow: auto;
  max-height: 200px;
  white-space: pre-wrap;
  word-break: break-all;
}
.btn-sm {
  padding: 4px 10px;
  font-size: 12px;
}
</style>
