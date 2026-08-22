import http from './http'
import type { ApiResponse } from '../types'

// 文件操作结果。
export interface FileResult {
  success: boolean
  output: string
  error?: string
}

// 列出目录。
export function listFiles(targetId: number, path = '.') {
  return http.get<ApiResponse<FileResult>>(`/targets/${targetId}/files`, { params: { path } })
}

// 读取文件。
export function readFile(targetId: number, path: string) {
  return http.post<ApiResponse<FileResult>>(`/targets/${targetId}/files/read`, { path })
}

// 写入文件。
export function writeFile(targetId: number, path: string, content: string) {
  return http.post<ApiResponse<FileResult>>(`/targets/${targetId}/files/write`, { path, content })
}

// 重命名。
export function renameFile(targetId: number, from: string, to: string) {
  return http.post<ApiResponse<FileResult>>(`/targets/${targetId}/files/rename`, { from, to })
}

// 创建目录。
export function mkdir(targetId: number, path: string) {
  return http.post<ApiResponse<FileResult>>(`/targets/${targetId}/files/mkdir`, { path })
}

// 删除文件/目录。
export function deleteFile(targetId: number, path: string) {
  return http.post<ApiResponse<FileResult>>(`/targets/${targetId}/files/delete`, { path })
}
