import request from './request'

export interface AuditLog {
  id: number
  user_id: number
  username: string
  action: string
  resource: string
  resource_id?: number
  detail: string
  ip: string
  status: string
  created_at: string
}

export const auditApi = {
  list(params?: { keyword?: string; page?: number; per_page?: number }): Promise<{ list: AuditLog[]; total: number; page: number; per_page: number }> {
    return request.get('/audit', { params })
  },
}
