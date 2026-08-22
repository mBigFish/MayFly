import request from './request'

export interface DashboardData {
  targets: number
  targets_online: number
  sessions_active: number
  listeners: number
  tasks: number
  audit_logs: number
  users: number
}

export const dashboardApi = {
  get(): Promise<DashboardData> {
    return request.get('/dashboard')
  },
}
