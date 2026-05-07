import { defineStore } from 'pinia'

const LS_KEY = 'exam-client-api-base-url'

function readStored(): string {
  try {
    return localStorage.getItem(LS_KEY) || ''
  } catch {
    return ''
  }
}

export const useSettingsStore = defineStore('settings', {
  state: () => ({
    /** 用户自定义服务端根地址，不含末尾 `/`，可为空则回落到环境变量 */
    apiBaseUrl: typeof localStorage !== 'undefined' ? readStored() : '',
  }),
  getters: {
    effectiveBaseUrl(): string {
      const trimmed = this.apiBaseUrl.trim().replace(/\/+$/, '')
      if (trimmed) return trimmed
      const env = import.meta.env.VITE_API_BASE_URL as string | undefined
      return (env || 'http://127.0.0.1:8080').replace(/\/+$/, '')
    },
  },
  actions: {
    setApiBaseUrl(url: string) {
      this.apiBaseUrl = url.trim()
      try {
        localStorage.setItem(LS_KEY, this.apiBaseUrl)
      } catch {
        /* ignore */
      }
    },
  },
})
