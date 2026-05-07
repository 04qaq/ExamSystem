import axios, { type AxiosInstance } from 'axios'
import type { ApiEnvelope } from './types'
import { useSettingsStore } from '@/stores/settings'
import { useAuthStore } from '@/stores/auth'

async function goLogin() {
  const { default: router } = await import('@/router')
  if (router.currentRoute.value.name !== 'login') {
    router.push({
      name: 'login',
      query: { redirect: router.currentRoute.value.fullPath },
    })
  }
}

function unwrap<T>(data: unknown): T {
  const body = data as ApiEnvelope<T>
  if (body && typeof body.code === 'number') {
    if (body.code !== 0) {
      throw new Error(body.message || '请求失败')
    }
    return body.data as T
  }
  return data as T
}

export const http: AxiosInstance = axios.create({
  timeout: 60000,
})

http.interceptors.request.use((config) => {
  const settings = useSettingsStore()
  config.baseURL = settings.effectiveBaseUrl
  const auth = useAuthStore()
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  return config
})

http.interceptors.response.use(
  (res) => {
    const data = res.data
    if (
      data &&
      typeof data === 'object' &&
      !(data instanceof Blob) &&
      'code' in data &&
      typeof (data as ApiEnvelope).code === 'number'
    ) {
      res.data = unwrap(data)
    }
    return res
  },
  (err) => {
    const status = err.response?.status
    const data = err.response?.data as ApiEnvelope | undefined
    if (status === 401) {
      useAuthStore().logout()
      void goLogin()
    }
    const msg =
      (data && typeof data.message === 'string' && data.message) ||
      err.message ||
      '网络错误'
    return Promise.reject(new Error(msg))
  }
)

/** 下载二进制（不走 JSON 解包） */
export async function getBlob(path: string): Promise<{ blob: Blob; filename: string }> {
  const settings = useSettingsStore()
  const auth = useAuthStore()
  const url = `${settings.effectiveBaseUrl}${path.startsWith('/') ? path : `/${path}`}`
  const res = await fetch(url, {
    headers: auth.token ? { Authorization: `Bearer ${auth.token}` } : {},
  })
  if (res.status === 401) {
    useAuthStore().logout()
    void goLogin()
    throw new Error('登录已过期')
  }
  if (!res.ok) {
    try {
      const j = (await res.json()) as ApiEnvelope
      throw new Error(j.message || `下载失败 (${res.status})`)
    } catch {
      throw new Error(`下载失败 (${res.status})`)
    }
  }
  const cd = res.headers.get('Content-Disposition')
  let filename = 'export.xlsx'
  if (cd) {
    const m = /filename\*?=(?:UTF-8'')?["']?([^;"']+)/i.exec(cd)
    if (m?.[1]) filename = decodeURIComponent(m[1])
  }
  const blob = await res.blob()
  return { blob, filename }
}
