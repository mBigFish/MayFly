import request from './request'

export interface LoginResult {
  token: string
  username: string
  role: string
}

export const authApi = {
  login(username: string, password: string): Promise<LoginResult> {
    return request.post('/auth/login', { username, password })
  },

  logout(): Promise<void> {
    return request.post('/auth/logout')
  },

  changePassword(oldPassword: string, newPassword: string): Promise<void> {
    return request.post('/auth/change-password', {
      old_password: oldPassword,
      new_password: newPassword,
    })
  },
}
