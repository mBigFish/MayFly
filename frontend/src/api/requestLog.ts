import request from './request'

export interface RequestLog {
  id: number
  target_id: number
  target_name: string
  username: string
  operation: string
  params: string
  request: string
  response: string
  status: string
  duration: number
  error: string
  created_at: string
}

export const requestLogApi = {
  list(params: { target_id?: number; keyword?: string; page?: number; page_size?: number }): Promise<{ list: RequestLog[]; total: number; page: number; per_page: number }> {
    return request.get('/request-logs', { params })
  },

  get(id: number): Promise<RequestLog> {
    return request.get(`/request-logs/${id}`)
  },
}
