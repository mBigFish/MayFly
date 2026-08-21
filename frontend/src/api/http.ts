import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'

// 统一 axios 实例，指向后端 API。
const http = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

// 请求拦截器：自动附带 JWT。
http.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器：统一处理错误。
http.interceptors.response.use(
  (response) => {
    // 后端统一返回 { code, message, data }。
    const body = response.data
    if (body && typeof body === 'object' && 'code' in body && body.code !== 0) {
      ElMessage.error(body.message || '请求失败')
      return Promise.reject(new Error(body.message || '请求失败'))
    }
    return response
  },
  (error) => {
    const status = error.response?.status
    const message = error.response?.data?.message || error.message || '网络错误'

    if (status === 401) {
      // 未认证或 token 失效，清除并跳转登录页。
      localStorage.removeItem('token')
      ElMessage.error('登录已失效，请重新登录')
      router.push({ name: 'login' })
    } else if (status === 403) {
      ElMessage.error('权限不足')
    } else {
      ElMessage.error(message)
    }
    return Promise.reject(error)
  },
)

export default http
