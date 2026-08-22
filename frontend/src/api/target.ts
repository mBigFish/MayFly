import request from './request'

export interface Target {
  id: number
  name: string
  url: string
  type: string
  encoding: string
  remark: string
  status: string
  created_at: string
  updated_at: string
}

export interface CreateTargetReq {
  name: string
  url: string
  type: string
  password?: string
  encoding?: string
  remark?: string
}

export const targetApi = {
  list(keyword?: string): Promise<Target[]> {
    return request.get('/targets', { params: keyword ? { keyword } : {} })
  },

  get(id: number): Promise<Target> {
    return request.get(`/targets/${id}`)
  },

  create(data: CreateTargetReq): Promise<Target> {
    return request.post('/targets', data)
  },

  update(id: number, data: Partial<CreateTargetReq>): Promise<Target> {
    return request.put(`/targets/${id}`, data)
  },

  delete(id: number): Promise<void> {
    return request.delete(`/targets/${id}`)
  },

  check(id: number): Promise<{ status: string }> {
    return request.post(`/targets/${id}/check`)
  },

  execute(id: number, command: string): Promise<{ output: string; duration: string }> {
    return request.post(`/targets/${id}/execute`, { command })
  },

  listFiles(id: number, path: string): Promise<{ path: string; raw?: string; entries?: any[] }> {
    return request.post(`/targets/${id}/files`, { path })
  },

  readFile(id: number, path: string): Promise<{ path: string; content: string }> {
    return request.post(`/targets/${id}/files/read`, { path })
  },

  writeFile(id: number, path: string, content: string): Promise<void> {
    return request.post(`/targets/${id}/files/write`, { path, content })
  },

  deleteFile(id: number, path: string): Promise<void> {
    return request.post(`/targets/${id}/files/delete`, { path })
  },

  renameFile(id: number, oldPath: string, newPath: string): Promise<void> {
    return request.post(`/targets/${id}/files/rename`, { old_path: oldPath, new_path: newPath })
  },

  mkdir(id: number, path: string): Promise<void> {
    return request.post(`/targets/${id}/files/mkdir`, { path })
  },

  downloadFile(id: number, path: string): Promise<Blob> {
    return request.post(`/targets/${id}/files/download`, { path }, { responseType: 'blob' })
  },

  getInfo(id: number): Promise<{ info: string }> {
    return request.get(`/targets/${id}/info`)
  },

  batchCheck(ids: number[]): Promise<any[]> {
    return request.post('/targets/batch-check', { ids })
  },
}
