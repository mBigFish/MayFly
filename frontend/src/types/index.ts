// 全局类型定义，对应后端 DTO。

// 统一 API 响应结构。
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

// 分页数据。
export interface PageData<T> {
  items: T[]
  total: number
}

// 用户信息。
export interface User {
  id: number
  username: string
  roles: string[]
}

// 登录响应。
export interface LoginResponse {
  token: string
  user: User
}

// 目标实体。
export interface Target {
  id: number
  name: string
  url: string
  type: string
  protocol: string
  method: string
  headers: string
  cookies: string
  timeout: number
  proxy: string
  encoding: string
  remark: string
  group_id: number
  created_at: string
  updated_at: string
}

// 创建目标请求。
export interface CreateTargetRequest {
  name: string
  url: string
  type?: string
  protocol?: string
  method?: string
  headers?: string
  cookies?: string
  timeout?: number
  proxy?: string
  encoding?: string
  remark?: string
  group_id?: number
}

// 更新目标请求。
export type UpdateTargetRequest = Partial<CreateTargetRequest>

// 探活响应。
export interface CheckResponse {
  target_id: number
  success: boolean
  message: string
}
