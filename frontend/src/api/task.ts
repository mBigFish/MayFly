import request from './request'

export interface Task {
  id: number
  name: string
  type: string
  status: string
  payload: string
  result: string
  total: number
  done: number
  created_at: string
  started_at?: string
  done_at?: string
}

export const taskApi = {
  list(keyword?: string): Promise<Task[]> {
    return request.get('/tasks', { params: keyword ? { keyword } : {} })
  },
  get(id: number): Promise<Task> {
    return request.get(`/tasks/${id}`)
  },
  create(data: { name: string; type: string; payload?: string }): Promise<Task> {
    return request.post('/tasks', data)
  },
  cancel(id: number): Promise<void> {
    return request.post(`/tasks/${id}/cancel`)
  },
  delete(id: number): Promise<void> {
    return request.delete(`/tasks/${id}`)
  },
}
