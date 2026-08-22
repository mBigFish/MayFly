<template>
  <div>
    <div class="page-title">系统设置</div>
    <div class="glass-card" style="padding: 20px; max-width: 600px;">
      <div style="margin-bottom: 20px;">
        <div style="color: var(--text-secondary); margin-bottom: 8px;">当前用户</div>
        <div style="font-size: 16px;">{{ auth.username }} <span class="tag tag-aspx" style="margin-left: 8px;">{{ auth.role }}</span></div>
      </div>
      <div style="margin-bottom: 20px;">
        <div style="color: var(--text-secondary); margin-bottom: 8px;">修改密码</div>
        <div style="display: flex; flex-direction: column; gap: 10px; max-width: 300px;">
          <input v-model="pwdForm.oldPassword" type="password" class="input" placeholder="旧密码" />
          <input v-model="pwdForm.newPassword" type="password" class="input" placeholder="新密码" />
          <input v-model="pwdForm.confirmPassword" type="password" class="input" placeholder="确认新密码" />
          <button class="btn btn-primary" @click="changePassword">修改密码</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/auth'
import { ElMessage } from 'element-plus'

const auth = useAuthStore()
const pwdForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

async function changePassword() {
  if (!pwdForm.oldPassword || !pwdForm.newPassword) {
    ElMessage.warning('请填写完整')
    return
  }
  if (pwdForm.newPassword !== pwdForm.confirmPassword) {
    ElMessage.error('两次密码不一致')
    return
  }
  try {
    await authApi.changePassword(pwdForm.oldPassword, pwdForm.newPassword)
    ElMessage.success('密码修改成功')
    pwdForm.oldPassword = ''
    pwdForm.newPassword = ''
    pwdForm.confirmPassword = ''
  } catch {
    // 错误已在拦截器中处理
  }
}
</script>
