import request from './request'

export interface Listener {
  id: number
  name: string
  host: string
  port: number
  protocol: string
  status: string
  connections: number
}

export const listenerApi = {
  list(): Promise<Listener[]> {
    return request.get('/listeners')
  },
  create(data: { name: string; host: string; port: number; protocol?: string }): Promise<Listener> {
    return request.post('/listeners', data)
  },
  start(id: number): Promise<void> {
    return request.post(`/listeners/${id}/start`)
  },
  stop(id: number): Promise<void> {
    return request.post(`/listeners/${id}/stop`)
  },
  delete(id: number): Promise<void> {
    return request.delete(`/listeners/${id}`)
  },
}
