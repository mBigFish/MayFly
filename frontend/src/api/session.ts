import request from './request'

export interface Session {
  id: number
  target_id: number
  target_name: string
  user_id: number
  username: string
  type: string
  status: string
  created_at: string
  last_active?: string
}

export const sessionApi = {
  list(type?: string): Promise<Session[]> {
    return request.get('/sessions', { params: type ? { type } : {} })
  },
  get(id: number): Promise<Session> {
    return request.get(`/sessions/${id}`)
  },
  close(id: number): Promise<void> {
    return request.post(`/sessions/${id}/close`)
  },
  delete(id: number): Promise<void> {
    return request.delete(`/sessions/${id}`)
  },
}
