import http from './http'
import type { ApiResponse } from '../types'

// 会话实体。
export interface Session {
  id: string
  target_id: number
  user_id: number
  created_at: string
  last_seen: string
}

// 创建会话。
export function createSession(targetId: number) {
  return http.post<ApiResponse<Session>>('/sessions', { target_id: targetId })
}

// 会话列表。
export function listSessions() {
  return http.get<ApiResponse<Session[]>>('/sessions')
}

// 关闭会话。
export function closeSession(id: string) {
  return http.delete<ApiResponse<null>>(`/sessions/${id}`)
}
