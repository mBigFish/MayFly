<template>
  <div style="height: 100%; display: flex; flex-direction: column;">
    <div class="page-title">文件管理</div>
    <div class="glass-card" style="flex: 1; display: flex; flex-direction: column; overflow: hidden;">
      <!-- 工具栏 -->
      <div class="toolbar" style="padding: 10px 15px; margin-bottom: 0; border-bottom: 1px solid var(--border);">
        <select v-model="selectedTargetId" class="input" style="width: 200px;" @change="onTargetChange">
          <option :value="0">选择目标...</option>
          <option v-for="t in targets" :key="t.id" :value="t.id">{{ t.name }} ({{ t.type }})</option>
        </select>
        <button type="button" class="btn btn-primary" @click="loadFiles(currentPath)" :disabled="!selectedTargetId">
          <el-icon><Refresh /></el-icon> 刷新
        </button>
        <button type="button" class="btn" @click="showMkdirDialog = true" :disabled="!selectedTargetId">
          <el-icon><FolderAdd /></el-icon> 新建目录
        </button>
        <button type="button" class="btn" @click="triggerUpload" :disabled="!selectedTargetId">
          <el-icon><Upload /></el-icon> 上传
        </button>
        <input ref="uploadInput" type="file" style="display: none;" @change="handleUpload" />
        <div class="toolbar-spacer" />
      </div>

      <!-- 面包屑导航 -->
      <div style="padding: 8px 15px; border-bottom: 1px solid var(--border); display: flex; align-items: center; gap: 4px; flex-wrap: wrap;">
        <el-icon style="cursor: pointer;" @click="loadFiles('/')"><HomeFilled /></el-icon>
        <template v-for="(seg, i) in pathSegments" :key="i">
          <span style="color: var(--text-secondary);">/</span>
          <span style="cursor: pointer; color: var(--accent);" @click="loadFiles(seg.path)">{{ seg.name }}</span>
        </template>
      </div>

      <!-- 文件列表 -->
      <div style="flex: 1; overflow: auto;">
        <table class="file-table" v-if="fileList.length > 0">
          <thead>
            <tr>
              <th>名称</th>
              <th style="width: 80px;">类型</th>
              <th style="width: 100px;">大小</th>
              <th style="width: 160px;">修改时间</th>
              <th style="width: 200px;">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="f in fileList" :key="f.name" @dblclick="onFileDblClick(f)">
              <td>
                <el-icon style="vertical-align: middle; margin-right: 6px;">
                  <Folder v-if="f.type === 'd'" /><Document v-else />
                </el-icon>
                {{ f.name }}
              </td>
              <td>{{ f.type === 'd' ? '目录' : '文件' }}</td>
              <td>{{ f.type === 'd' ? '-' : formatSize(f.size) }}</td>
              <td>{{ formatTime(f.mtime) }}</td>
              <td>
                <button v-if="f.type === 'f'" type="button" class="btn btn-sm" @click="openEditor(f)">编辑</button>
                <button v-if="f.type === 'f'" type="button" class="btn btn-sm" @click="downloadFile(f)">下载</button>
                <button type="button" class="btn btn-sm" @click="startRename(f)">重命名</button>
                <button type="button" class="btn btn-sm btn-danger" @click="deleteFile(f)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-else-if="selectedTargetId" class="empty-state" style="padding: 40px; text-align: center; color: var(--text-secondary);">
          目录为空或未加载
        </div>
        <div v-else class="empty-state" style="padding: 40px; text-align: center; color: var(--text-secondary);">
          请先选择目标
        </div>
      </div>
    </div>

    <!-- Monaco 编辑器弹窗 -->
    <div v-if="editorVisible" class="modal-overlay" @click.self="closeEditor">
      <div class="modal-content" style="width: 80%; height: 80%; display: flex; flex-direction: column;">
        <div class="modal-header">
          <span>{{ editorTitle }}</span>
          <button type="button" class="modal-close" @click="closeEditor">&times;</button>
        </div>
        <div style="flex: 1; position: relative;">
          <div ref="monacoContainer" style="position: absolute; inset: 0;"></div>
        </div>
        <div class="modal-footer">
          <span v-if="editorDirty" style="color: var(--warning); margin-right: 10px;">未保存</span>
          <button type="button" class="btn" @click="closeEditor">取消</button>
          <button type="button" class="btn btn-primary" @click="saveFile" :disabled="!editorDirty">保存</button>
        </div>
      </div>
    </div>

    <!-- 新建目录弹窗 -->
    <div v-if="showMkdirDialog" class="modal-overlay" @click.self="showMkdirDialog = false">
      <div class="modal-content" style="width: 400px;">
        <div class="modal-header">
          <span>新建目录</span>
          <button type="button" class="modal-close" @click="showMkdirDialog = false">&times;</button>
        </div>
        <div style="padding: 15px;">
          <input v-model="mkdirName" class="input" style="width: 100%;" placeholder="目录名" @keyup.enter="confirmMkdir" />
        </div>
        <div class="modal-footer">
          <button type="button" class="btn" @click="showMkdirDialog = false">取消</button>
          <button type="button" class="btn btn-primary" @click="confirmMkdir">创建</button>
        </div>
      </div>
    </div>

    <!-- 重命名弹窗 -->
    <div v-if="renameVisible" class="modal-overlay" @click.self="renameVisible = false">
      <div class="modal-content" style="width: 400px;">
        <div class="modal-header">
          <span>重命名</span>
          <button type="button" class="modal-close" @click="renameVisible = false">&times;</button>
        </div>
        <div style="padding: 15px;">
          <input v-model="renameNewName" class="input" style="width: 100%;" placeholder="新名称" @keyup.enter="confirmRename" />
        </div>
        <div class="modal-footer">
          <button type="button" class="btn" @click="renameVisible = false">取消</button>
          <button type="button" class="btn btn-primary" @click="confirmRename">确定</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted, shallowRef, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import loader from '@monaco-editor/loader'
import { targetApi, type Target } from '@/api/target'

interface FileEntry {
  name: string
  type: string
  size: number
  mtime: number
}

const targets = ref<Target[]>([])
const selectedTargetId = ref(0)
const currentPath = ref('/')
const fileList = ref<FileEntry[]>([])
const uploadInput = ref<HTMLInputElement>()

// Monaco Editor
const editorVisible = ref(false)
const editorTitle = ref('')
const editorDirty = ref(false)
const editorPath = ref('')
const monacoContainer = ref<HTMLElement>()
let monacoEditor: any = null
let monacoReady = false

// 新建目录
const showMkdirDialog = ref(false)
const mkdirName = ref('')

// 重命名
const renameVisible = ref(false)
const renameNewName = ref('')
const renameOldPath = ref('')

// 面包屑
const pathSegments = computed(() => {
  const segs = currentPath.value.split('/').filter(s => s)
  const result: { name: string; path: string }[] = []
  let p = ''
  for (const s of segs) {
    p += '/' + s
    result.push({ name: s, path: p })
  }
  return result
})

async function loadTargets() {
  try {
    targets.value = await targetApi.list() || []
  } catch {}
}

function onTargetChange() {
  currentPath.value = '/'
  if (selectedTargetId.value) {
    loadFiles('/')
  }
}

async function loadFiles(path: string) {
  if (!selectedTargetId.value) return
  currentPath.value = path
  try {
    const res = await targetApi.listFiles(selectedTargetId.value, path)
    if (res.entries) {
      // 排序：目录在前
      fileList.value = res.entries.sort((a: FileEntry, b: FileEntry) => {
        if (a.type !== b.type) return a.type === 'd' ? -1 : 1
        return a.name.localeCompare(b.name)
      })
    } else if (res.raw) {
      // 纯文本格式，尝试按行解析
      fileList.value = res.raw.split('\n').filter(Boolean).map((name: string) => ({
        name: name.trim(),
        type: 'f',
        size: 0,
        mtime: 0,
      }))
    } else {
      fileList.value = []
    }
  } catch (e: any) {
    ElMessage.error('加载失败: ' + (e.message || e))
  }
}

function onFileDblClick(f: FileEntry) {
  if (f.type === 'd') {
    const newPath = currentPath.value === '/' ? '/' + f.name : currentPath.value + '/' + f.name
    loadFiles(newPath)
  } else {
    openEditor(f)
  }
}

function formatSize(size: number): string {
  if (size < 1024) return size + ' B'
  if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' KB'
  if (size < 1024 * 1024 * 1024) return (size / 1024 / 1024).toFixed(1) + ' MB'
  return (size / 1024 / 1024 / 1024).toFixed(1) + ' GB'
}

function formatTime(t: number): string {
  if (!t) return '-'
  return new Date(t * 1000).toLocaleString()
}

// ===== Monaco Editor =====
async function openEditor(f: FileEntry) {
  editorPath.value = currentPath.value === '/' ? '/' + f.name : currentPath.value + '/' + f.name
  editorTitle.value = '编辑: ' + editorPath.value
  editorDirty.value = false
  editorVisible.value = true

  try {
    const res = await targetApi.readFile(selectedTargetId.value, editorPath.value)
    await initMonaco(res.content || '')
  } catch (e: any) {
    ElMessage.error('读取文件失败: ' + (e.message || e))
    editorVisible.value = false
  }
}

async function initMonaco(content: string) {
  await nextTick()
  if (!monacoContainer.value) return

  if (!monacoReady) {
    loader.config({ paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.52.2/min/vs' } })
    monacoReady = true
  }

  const monaco = await loader.load()
  if (monacoEditor) {
    monacoEditor.dispose()
  }

  monacoEditor = monaco.editor.create(monacoContainer.value, {
    value: content,
    language: 'plaintext',
    theme: 'vs-dark',
    fontSize: 14,
    automaticLayout: true,
    minimap: { enabled: false },
    scrollBeyondLastLine: false,
  })

  monacoEditor.onDidChangeModelContent(() => {
    editorDirty.value = true
  })
}

async function saveFile() {
  if (!monacoEditor) return
  const content = monacoEditor.getValue()
  try {
    await targetApi.writeFile(selectedTargetId.value, editorPath.value, content)
    ElMessage.success('保存成功')
    editorDirty.value = false
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e.message || e))
  }
}

function closeEditor() {
  if (monacoEditor) {
    monacoEditor.dispose()
    monacoEditor = null
  }
  editorVisible.value = false
  editorDirty.value = false
}

// ===== 新建目录 =====
async function confirmMkdir() {
  if (!mkdirName.value) {
    ElMessage.warning('请输入目录名')
    return
  }
  const path = currentPath.value === '/' ? '/' + mkdirName.value : currentPath.value + '/' + mkdirName.value
  try {
    await targetApi.mkdir(selectedTargetId.value, path)
    ElMessage.success('创建成功')
    showMkdirDialog.value = false
    mkdirName.value = ''
    loadFiles(currentPath.value)
  } catch (e: any) {
    ElMessage.error('创建失败: ' + (e.message || e))
  }
}

// ===== 重命名 =====
function startRename(f: FileEntry) {
  renameOldPath.value = currentPath.value === '/' ? '/' + f.name : currentPath.value + '/' + f.name
  renameNewName.value = f.name
  renameVisible.value = true
}

async function confirmRename() {
  if (!renameNewName.value) {
    ElMessage.warning('请输入新名称')
    return
  }
  const newPath = currentPath.value === '/' ? '/' + renameNewName.value : currentPath.value + '/' + renameNewName.value
  try {
    await targetApi.renameFile(selectedTargetId.value, renameOldPath.value, newPath)
    ElMessage.success('重命名成功')
    renameVisible.value = false
    loadFiles(currentPath.value)
  } catch (e: any) {
    ElMessage.error('重命名失败: ' + (e.message || e))
  }
}

// ===== 删除 =====
async function deleteFile(f: FileEntry) {
  try {
    await ElMessageBox.confirm(`确定删除 "${f.name}" 吗？`, '确认', { type: 'warning' })
    const path = currentPath.value === '/' ? '/' + f.name : currentPath.value + '/' + f.name
    await targetApi.deleteFile(selectedTargetId.value, path)
    ElMessage.success('删除成功')
    loadFiles(currentPath.value)
  } catch {}
}

// ===== 下载 =====
async function downloadFile(f: FileEntry) {
  const path = currentPath.value === '/' ? '/' + f.name : currentPath.value + '/' + f.name
  try {
    const blob = await targetApi.downloadFile(selectedTargetId.value, path)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = f.name
    a.click()
    URL.revokeObjectURL(url)
  } catch (e: any) {
    ElMessage.error('下载失败: ' + (e.message || e))
  }
}

// ===== 上传 =====
function triggerUpload() {
  uploadInput.value?.click()
}

async function handleUpload(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files || input.files.length === 0) return
  const file = input.files[0]
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.warning('文件不能超过 5MB')
    return
  }
  const text = await file.text()
  const path = currentPath.value === '/' ? '/' + file.name : currentPath.value + '/' + file.name
  try {
    await targetApi.writeFile(selectedTargetId.value, path, text)
    ElMessage.success('上传成功')
    loadFiles(currentPath.value)
  } catch (err: any) {
    ElMessage.error('上传失败: ' + (err.message || err))
  }
  input.value = ''
}

loadTargets()

onUnmounted(() => {
  if (monacoEditor) {
    monacoEditor.dispose()
  }
})
</script>

<style scoped>
.file-table {
  width: 100%;
  border-collapse: collapse;
}
.file-table th {
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
.file-table td {
  padding: 6px 12px;
  border-bottom: 1px solid var(--border);
  font-size: 13px;
}
.file-table tr:hover {
  background: rgba(0, 212, 255, 0.05);
}
.btn-sm {
  padding: 2px 8px;
  font-size: 12px;
  margin-right: 4px;
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
</style>
