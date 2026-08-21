import http from './http'
import type { ApiResponse, CheckResponse, CreateTargetRequest, PageData, Target, UpdateTargetRequest } from '../types'

// 目标列表。
export function listTargets(offset = 0, limit = 20) {
  return http.get<ApiResponse<PageData<Target>>>('/targets', { params: { offset, limit } })
}

// 目标详情。
export function getTarget(id: number) {
  return http.get<ApiResponse<Target>>(`/targets/${id}`)
}

// 创建目标。
export function createTarget(data: CreateTargetRequest) {
  return http.post<ApiResponse<Target>>('/targets', data)
}

// 更新目标。
export function updateTarget(id: number, data: UpdateTargetRequest) {
  return http.put<ApiResponse<Target>>(`/targets/${id}`, data)
}

// 删除目标。
export function deleteTarget(id: number) {
  return http.delete<ApiResponse<null>>(`/targets/${id}`)
}

// 目标探活。
export function checkTarget(id: number) {
  return http.post<ApiResponse<CheckResponse>>(`/targets/${id}/check`)
}
