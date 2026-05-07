import { http } from './http'
import type { UserRow } from './types'

export interface UserListResult {
  total: number
  items: UserRow[]
}

export async function adminUsers(page = 1, pageSize = 20): Promise<UserListResult> {
  const res = await http.get<UserListResult>('/api/admin/users', {
    params: { page, page_size: pageSize },
  })
  return res.data
}

export async function adminCreateUser(body: {
  username: string
  password: string
  role: number
  real_name?: string
}) {
  await http.post('/api/admin/users', body)
}

export async function adminUpdateUser(
  id: number,
  body: {
    real_name?: string
    role?: number
    status?: number
    password?: string
  }
) {
  await http.put(`/api/admin/users/${id}`, body)
}

export interface LogRow {
  id: number
  user_id: number
  username: string
  action: string
  target: string
  detail: string
  ip: string
  created_at: string
}

export interface LogListResult {
  total: number
  items: LogRow[]
}

export async function adminLogs(page = 1, pageSize = 20, action?: string): Promise<LogListResult> {
  const res = await http.get<LogListResult>('/api/admin/logs', {
    params: { page, page_size: pageSize, action: action || undefined },
  })
  return res.data
}
