<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { checkTarget, getTarget } from '../api/target'
import type { CheckResponse, Target } from '../types'

const route = useRoute()
const router = useRouter()

const target = ref<Target | null>(null)
const loading = ref(false)
const checking = ref(false)
const checkResult = ref<CheckResponse | null>(null)

async function loadTarget() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    const res = await getTarget(id)
    target.value = res.data.data ?? null
  } catch {
    // 错误已由拦截器提示。
  } finally {
    loading.value = false
  }
}

async function handleCheck() {
  if (!target.value) return
  checking.value = true
  checkResult.value = null
  try {
    const res = await checkTarget(target.value.id)
    checkResult.value = res.data.data ?? null
  } catch {
    // 错误已由拦截器提示。
  } finally {
    checking.value = false
  }
}

function goBack() {
  router.push({ name: 'targets' })
}

onMounted(loadTarget)
</script>

<template>
  <div v-loading="loading" class="detail">
    <div class="toolbar">
      <el-button @click="goBack">返回列表</el-button>
      <h2 class="title">{{ target?.name || '目标详情' }}</h2>
      <el-button type="primary" :loading="checking" @click="handleCheck">探活检测</el-button>
    </div>

    <template v-if="target">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="ID">{{ target.id }}</el-descriptions-item>
        <el-descriptions-item label="名称">{{ target.name }}</el-descriptions-item>
        <el-descriptions-item label="URL" :span="2">{{ target.url }}</el-descriptions-item>
        <el-descriptions-item label="协议">{{ target.protocol }}</el-descriptions-item>
        <el-descriptions-item label="方法">{{ target.method }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ target.type }}</el-descriptions-item>
        <el-descriptions-item label="超时(秒)">{{ target.timeout }}</el-descriptions-item>
        <el-descriptions-item label="编码">{{ target.encoding }}</el-descriptions-item>
        <el-descriptions-item label="代理">{{ target.proxy || '-' }}</el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ target.remark || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间" :span="2">{{ target.created_at }}</el-descriptions-item>
      </el-descriptions>

      <el-card v-if="checkResult" class="check-result">
        <template #header>探活结果</template>
        <el-alert
          :type="checkResult.success ? 'success' : 'error'"
          :title="checkResult.success ? '探活成功' : '探活失败'"
          :description="checkResult.message"
          show-icon
          :closable="false"
        />
      </el-card>
    </template>

    <el-empty v-else-if="!loading" description="目标不存在" />
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}
.title {
  margin: 0;
  flex: 1;
}
.check-result {
  margin-top: 16px;
}
</style>
