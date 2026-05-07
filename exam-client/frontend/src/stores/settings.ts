import { defineStore } from 'pinia'
import {
  discoverExamServerUrl,
  discoverExamServerUrlBrowserOnly,
  tryNativeLanDiscover,
} from '@/utils/serverDiscovery'

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
    /** 用户自定义服务端根地址，不含末尾 `/`，可为空则回落到环境变量或自动发现结果 */
    apiBaseUrl: typeof localStorage !== 'undefined' ? readStored() : '',
    /** idle | running | done */
    discoveryPhase: 'idle' as 'idle' | 'running' | 'done',
    lastDiscoveryMessage: '' as string,
  }),
  getters: {
    effectiveBaseUrl(): string {
      const trimmed = this.apiBaseUrl.trim().replace(/\/+$/, '')
      if (trimmed) return trimmed
      const env = import.meta.env.VITE_API_BASE_URL as string | undefined
      return (env || 'http://127.0.0.1:8080').replace(/\/+$/, '')
    },
    isDiscovering(): boolean {
      return this.discoveryPhase === 'running'
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

    /**
     * 启动时优先走桌面同源 GET /discover（Go UDP/TCP），与是否已保存 API 地址无关；
     * 否则用户一旦保存过错误地址将永远无法触发 Go 发现。
     * 原生未发现且未保存、未配置 env 时，再走浏览器 HTTP 扫描。
     */
    async ensureDiscoveredOnStartup() {
      this.discoveryPhase = 'running'
      this.lastDiscoveryMessage = ''
      try {
        const native = await tryNativeLanDiscover()
        if (native) {
          this.setApiBaseUrl(native)
          this.lastDiscoveryMessage = `已自动连接 ${native}`
          return
        }

        const manual = this.apiBaseUrl.trim()
        const env = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.trim()
        if (manual || env) {
          return
        }

        const found = await discoverExamServerUrlBrowserOnly()
        if (found) {
          this.setApiBaseUrl(found)
          this.lastDiscoveryMessage = `已自动连接 ${found}`
        } else {
          this.lastDiscoveryMessage =
            '未在局域网找到 exam-server，将使用默认 http://127.0.0.1:8080，可在下方手动配置'
        }
      } catch {
        this.lastDiscoveryMessage = '自动查找失败，请检查服务端是否启动或手动填写地址'
      } finally {
        this.discoveryPhase = 'done'
      }
    },

    /** 设置页「重新查找」 */
    async discoverNow(): Promise<string | null> {
      this.discoveryPhase = 'running'
      this.lastDiscoveryMessage = ''
      try {
        const found = await discoverExamServerUrl()
        if (found) {
          this.setApiBaseUrl(found)
          this.lastDiscoveryMessage = `已找到服务端 ${found}`
          return found
        }
        this.lastDiscoveryMessage =
          '未发现运行中的 exam-server，请确认服务端已启动且防火墙放行 8080 端口'
        return null
      } finally {
        this.discoveryPhase = 'done'
      }
    },
  },
})
