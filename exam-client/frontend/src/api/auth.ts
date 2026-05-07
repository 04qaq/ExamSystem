import { http } from './http'
import type { LoginBody, LoginResult, RegisterBody, UserInfo } from './types'

export async function loginApi(body: LoginBody): Promise<LoginResult> {
  const res = await http.post<LoginResult>('/api/auth/login', body)
  return res.data
}

export async function registerApi(body: RegisterBody): Promise<UserInfo> {
  const res = await http.post<UserInfo>('/api/auth/register', body)
  return res.data
}
