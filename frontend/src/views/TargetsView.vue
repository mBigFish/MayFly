<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { checkTarget, createTarget, deleteTarget, listTargets, updateTarget } from '../api/target'
import type { CreateTargetRequest, Target } from '../types'

const router = useRouter()

const targets = ref<Target[]>([])
const total = ref(0)
const loading = ref(false)
const checking = ref<number | null>(null)

const query = reactive({ offset: 0, limit: 20 })

// 弹窗状态。
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const form = reactive<CreateTargetRequest>({
  name: '',
  url: '',
  type: 'webshell',
  protocol: 'php',
  method: 'POST',
  headers: '',
  cookies: '',
  timeout: 30,
  proxy: '',
  encoding: 'utf-8',
  remark: '',
})
const editingId = ref<number | null>(null)

const currentPage = computed({
  get: () => query.offset / query.limit + 1,
  set: () => {},
})

async function loadTargets() {
  loading.value = true
  try {
    const res = await listTargets(query.offset, query.limit)
    targets.value = res.data.data?.items ?? []
    total.value = res.data.data?.total ?? 0
  } catch {
    // 错误已由拦截器提示。
  } finally {
    loading.value = false
  }
}

function onPageChange(page: number) {
  query.offset = (page - 1) * query.limit
  loadTargets()
}

function resetForm() {
  form.name = ''
  form.url = ''
  form.type = 'webshell'
  form.protocol = 'php'
  form.method = 'POST'
  form.headers = ''
  form.cookies = ''
  form.timeout = 30
  form.proxy = ''
  form.encoding = 'utf-8'
  form.remark = ''
}

function openCreate() {
  dialogMode.value = 'create'
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

function openEdit(t: Target) {
  dialogMode.value = 'edit'
  editingId.value = t.id
  form.name = t.name
  form.url = t.url
  form.type = t.type
  form.protocol = t.protocol
  form.method = t.method
  form.headers = t.headers
  form.cookies = t.cookies
  form.timeout = t.timeout
  form.proxy = t.proxy
  form.encoding = t.encoding
  form.remark = t.remark
  dialogVisible.value = true
}

async function handleSubmit() {
  try {
    if (dialogMode.value === 'create') {
      await createTarget({ ...form })
      ElMessage.success('创建成功')
    } else if (editingId.value != null) {
      await updateTarget(editingId.value, { ...form })
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    loadTargets()
  } catch {
    // 错误已由拦截器提示。
  }
}

async function handleDelete(t: Target) {
  try {
    await ElMessageBox.confirm(`确认删除目标「${t.name}」？`, '删除确认', { type: 'warning' })
  } catch {
    return // 用户取消。
  }
  try {
    await deleteTarget(t.id)
    ElMessage.success('删除成功')
    loadTargets()
  } catch {
    // 错误已由拦截器提示。
  }
}

async function handleCheck(t: Target) {
  checking.value = t.id
  try {
    const res = await checkTarget(t.id)
    const data = res.data.data
    if (data?.success) {
      ElMessage.success('探活成功')
    } else {
      ElMessage.warning(`探活失败：${data?.message || '未知原因'}`)
    }
  } catch {
    // 错误已由拦截器提示。
  } finally {
    checking.value = null
  }
}

function goDetail(t: Target) {
  router.push({ name: 'target-detail', params: { id: t.id } })
}

onMounted(loadTargets)
</script>

<template>
  <div class="targets">
    <div class="toolbar">
      <h2 class="title">目标管理</h2>
      <el-button type="primary" @click="openCreate">新增目标</el-button>
    </div>

    <el-table :data="targets" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="名称" min-width="140">
        <template #default="{ row }">
          <el-link type="primary" @click="goDetail(row)">{{ row.name }}</el-link>
        </template>
      </el-table-column>
      <el-table-column prop="url" label="URL" min-width="220" show-overflow-tooltip />
      <el-table-column prop="protocol" label="协议" width="90" />
      <el-table-column prop="method" label="方法" width="80" />
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button size="small" :loading="checking === row.id" @click="handleCheck(row)">
            探活
          </el-button>
          <el-button size="small" type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        layout="total, prev, pager, next"
        :total="total"
        :page-size="query.limit"
        @current-change="onPageChange"
      />
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增目标' : '编辑目标'"
      width="560px"
    >
      <el-form label-width="90px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="目标名称" />
        </el-form-item>
        <el-form-item label="URL" required>
          <el-input v-model="form.url" placeholder="http://host/shell.php" />
        </el-form-item>
        <el-form-item label="协议">
          <el-select v-model="form.protocol">
            <el-option label="PHP" value="php" />
            <el-option label="JSP" value="jsp" />
            <el-option label="ASPX" value="aspx" />
          </el-select>
        </el-form-item>
        <el-form-item label="方法">
          <el-select v-model="form.method">
            <el-option label="POST" value="POST" />
            <el-option label="GET" value="GET" />
          </el-select>
        </el-form-item>
        <el-form-item label="超时(秒)">
          <el-input-number v-model="form.timeout" :min="1" :max="300" />
        </el-form-item>
        <el-form-item label="Cookies">
          <el-input v-model="form.cookies" type="textarea" :rows="2" placeholder="PHPSESSID=xxx" />
        </el-form-item>
        <el-form-item label="Headers">
          <el-input v-model="form.headers" type="textarea" :rows="2" placeholder="Authorization: Bearer xxx" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
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
.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
