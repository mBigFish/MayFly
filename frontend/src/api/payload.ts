import request from './request'

export interface Payload {
  type: string
  label: string
  command: string
}

export const payloadApi = {
  reverse(host: string, port: number): Promise<Payload[]> {
    return request.get('/payloads/reverse', { params: { host, port } })
  },
  shell(type: string, password?: string): Promise<{ type: string; script: string }> {
    return request.get('/payloads/shell', { params: { type, password } })
  },
}
