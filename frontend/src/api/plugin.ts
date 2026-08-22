import request from './request'

export interface Plugin {
  name: string
  version: string
  description: string
}

export interface PluginResult {
  status: string
  data: any
  message: string
}

export const pluginApi = {
  list(): Promise<Plugin[]> {
    return request.get('/plugins')
  },

  execute(name: string, targetId: number, params?: Record<string, string>): Promise<PluginResult> {
    return request.post(`/plugins/${name}/execute`, { target_id: targetId, params: params || {} })
  },
}
