import http from './http'
import type { ApiResponse, LoginResponse } from '../types'

// 登录。
export function login(username: string, password: string) {
  return http.post<ApiResponse<LoginResponse>>('/auth/login', { username, password })
}
