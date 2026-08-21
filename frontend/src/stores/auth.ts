import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '../types'

// 登录态管理。
export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<User | null>(null)

  // 设置登录态。
  function setAuth(newToken: string, newUser: User) {
    token.value = newToken
    user.value = newUser
    localStorage.setItem('token', newToken)
  }

  // 清除登录态。
  function clearAuth() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  // 是否已登录。
  const isLoggedIn = () => !!token.value

  return { token, user, setAuth, clearAuth, isLoggedIn }
})
