import request from './request'

export interface Server {
  id: number
  name: string
  host: string
  port: number
  username: string
  group: string
  last_test_status: string
  last_test_time?: string
  last_test_message: string
  created_at: string
}

export interface CreateServerReq {
  name: string
  host: string
  port: number
  username: string
  password?: string
  private_key?: string
  group?: string
}

export const serverApi = {
  list(keyword?: string, group?: string): Promise<Server[]> {
    return request.get('/servers', { params: { keyword, group } })
  },
  get(id: number): Promise<Server> {
    return request.get(`/servers/${id}`)
  },
  create(data: CreateServerReq): Promise<Server> {
    return request.post('/servers', data)
  },
  update(id: number, data: Partial<CreateServerReq>): Promise<Server> {
    return request.put(`/servers/${id}`, data)
  },
  delete(id: number): Promise<void> {
    return request.delete(`/servers/${id}`)
  },
  test(id: number): Promise<void> {
    return request.post(`/servers/${id}/test`)
  },
}
