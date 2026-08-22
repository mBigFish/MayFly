<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { deleteFile, listFiles, mkdir, readFile, renameFile, writeFile } from '../api/file'

const route = useRoute()
const targetId = Number(route.params.id)

const currentPath = ref('.')
const fileOutput = ref('')
const loading = ref(false)

// 文件内容编辑。
const editingPath = ref('')
const editingContent = ref('')
const editorVisible = ref(false)

async function loadDir() {
  loading.value = true
  try {
    const res = await listFiles(targetId, currentPath.value)
    fileOutput.value = res.data.data?.output ?? ''
  } catch {
    // 错误已由拦截器提示。
  } finally {
    loading.value = false
  }
}

async function handleOpenFile() {
  try {
    const { value } = await ElMessageBox.prompt('输入要打开的文件路径', '打开文件', {
      inputValue: currentPath.value,
    })
    editingPath.value = value
    const res = await readFile(targetId, value)
    editingContent.value = res.data.data?.output ?? ''
    editorVisible.value = true
  } catch {
    // 用户取消或错误。
  }
}

async function handleSaveFile() {
  try {
    await ElMessageBox.confirm('确认保存对文件的修改？', '保存确认', { type: 'warning' })
  } catch {
    return
  }
  try {
    await writeFile(targetId, editingPath.value, editingContent.value)
    ElMessage.success('保存成功')
    editorVisible.value = false
    loadDir()
  } catch {
    // 错误已由拦截器提示。
  }
}

async function handleRename() {
  try {
    const { value } = await ElMessageBox.prompt('输入新路径/名称', '重命名', {
      inputValue: currentPath.value,
    })
    await renameFile(targetId, currentPath.value, value)
    ElMessage.success('重命名成功')
    currentPath.value = value
    loadDir()
  } catch {
    // 用户取消或错误。
  }
}

async function handleMkdir() {
  try {
    const { value } = await ElMessageBox.prompt('输入目录名', '创建目录')
    await mkdir(targetId, value)
    ElMessage.success('创建成功')
    loadDir()
  } catch {
    // 用户取消或错误。
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm('确认删除当前路径？', '删除确认', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteFile(targetId, currentPath.value)
    ElMessage.success('删除成功')
    loadDir()
  } catch {
    // 错误已由拦截器提示。
  }
}

onMounted(loadDir)
</script>

<template>
  <div class="file-manager">
    <div class="toolbar">
      <h2 class="title">文件管理</h2>
      <div class="actions">
        <el-input v-model="currentPath" placeholder="路径" class="path-input" @keyup.enter="loadDir" />
        <el-button type="primary" @click="loadDir">刷新</el-button>
        <el-button @click="handleOpenFile">打开文件</el-button>
        <el-button @click="handleMkdir">新建目录</el-button>
        <el-button @click="handleRename">重命名</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
      </div>
    </div>

    <el-card v-loading="loading">
      <pre class="file-output">{{ fileOutput || '（空）' }}</pre>
    </el-card>

    <el-dialog v-model="editorVisible" title="编辑文件" width="700px">
      <el-input v-model="editingPath" disabled class="editor-path" />
      <el-input
        v-model="editingContent"
        type="textarea"
        :rows="15"
        class="editor-content"
      />
      <template #footer>
        <el-button @click="editorVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveFile">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.title {
  margin: 0;
}
.actions {
  display: flex;
  gap: 8px;
  align-items: center;
}
.path-input {
  width: 260px;
}
.file-output {
  margin: 0;
  font-family: monospace;
  white-space: pre-wrap;
  word-break: break-all;
}
.editor-path {
  margin-bottom: 8px;
}
</style>
