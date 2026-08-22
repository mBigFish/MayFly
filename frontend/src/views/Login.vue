<template>
  <div class="login-container">
    <div class="login-box glass-card">
      <div class="login-header">
        <div class="login-logo">Mayfly</div>
        <div class="login-subtitle">WebShell 管理平台</div>
      </div>
      <div class="login-form">
        <div class="form-group">
          <input
            v-model="form.username"
            type="text"
            class="input login-input"
            placeholder="用户名"
            @keyup.enter="handleLogin"
          />
        </div>
        <div class="form-group">
          <input
            v-model="form.password"
            type="password"
            class="input login-input"
            placeholder="密码"
            @keyup.enter="handleLogin"
          />
        </div>
        <button type="button" class="btn btn-primary login-btn" :disabled="loading" @click="handleLogin">
          {{ loading ? '登录中...' : '登 录' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const form = reactive({ username: '', password: '' })
const loading = ref(false)

async function handleLogin() {
  if (!form.username || !form.password) return
  loading.value = true
  try {
    await auth.login(form.username, form.password)
    router.push('/')
  } catch {
    // 错误已在拦截器中处理
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}

.login-box {
  width: 380px;
  padding: 40px 30px;
  text-align: center;
}

.login-header {
  margin-bottom: 30px;
}

.login-logo {
  font-size: 32px;
  font-weight: bold;
  color: var(--accent);
  text-shadow: 0 0 20px var(--accent-glow);
  margin-bottom: 8px;
}

.login-subtitle {
  color: var(--text-secondary);
  font-size: 14px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.login-input {
  height: 42px;
  font-size: 14px;
}

.login-btn {
  height: 42px;
  font-size: 15px;
  justify-content: center;
  margin-top: 5px;
}
</style>
