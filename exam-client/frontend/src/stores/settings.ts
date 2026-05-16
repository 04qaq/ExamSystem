import { defineStore } from 'pinia'
import {
  discoverExamServerUrl,
  discoverExamServerUrlBrowserOnly,
  tryNativeLanDiscover,
} from '@/utils/serverDiscovery'

const LS_KEY = 'exam-client-api-base-url'

function normalizeBaseUrl(url: string): string {
  return url.trim().replace(/\/+$/, '')
}

function packagedBaseUrl(): string {
  return normalizeBaseUrl((import.meta.env.VITE_API_BASE_URL as string | undefined) || '')
}

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
      const trimmed = normalizeBaseUrl(this.apiBaseUrl)
      if (trimmed) return trimmed
      const env = packagedBaseUrl()
      if (env) return env
      // 生产构建（如 Docker + Nginx 同源反代）未设置 VITE_API_BASE_URL 时使用当前站点根地址
      if (import.meta.env.PROD && typeof window !== 'undefined' && window.location?.origin) {
        return window.location.origin
      }
      return 'http://127.0.0.1:8080'
    },
    defaultBaseUrl(): string {
      const env = packagedBaseUrl()
      if (env) return env
      if (import.meta.env.PROD && typeof window !== 'undefined' && window.location?.origin) {
        return window.location.origin
      }
      return 'http://127.0.0.1:8080'
    },
    isDiscovering(): boolean {
      return this.discoveryPhase === 'running'
    },
  },
  actions: {
    setApiBaseUrl(url: string) {
      this.apiBaseUrl = normalizeBaseUrl(url)
      try {
        localStorage.setItem(LS_KEY, this.apiBaseUrl)
      } catch {
        /* ignore */
      }
    },
    resetApiBaseUrl() {
      this.apiBaseUrl = ''
      try {
        localStorage.removeItem(LS_KEY)
      } catch {
        /* ignore */
      }
    },

    /**
     * 启动时：若已保存 API 地址或构建时配置了 VITE_API_BASE_URL，则不再做局域网发现（避免覆盖云端地址）。
     * 否则依次尝试桌面同源 GET /discover、生产环境直接返回、开发环境再走浏览器 HTTP 扫描。
     */
    async ensureDiscoveredOnStartup() {
      this.discoveryPhase = 'running'
      this.lastDiscoveryMessage = ''
      try {
        const manual = normalizeBaseUrl(this.apiBaseUrl)
        const env = packagedBaseUrl()
        // 已保存地址或构建时写死的服务端：不要用局域网发现覆盖（避免盖掉云端部署地址）
        if (manual || env) {
          if (env && !manual) {
            this.lastDiscoveryMessage = '已使用打包配置的服务端地址'
          }
          return
        }

        const native = await tryNativeLanDiscover()
        if (native) {
          this.setApiBaseUrl(native)
          this.lastDiscoveryMessage = '已自动连接到考试服务'
          return
        }
        // 生产环境默认走同源 /api，不再做局域网扫描（避免部署在服务器上时无意义扫网段）
        if (import.meta.env.PROD) {
          return
        }

        const found = await discoverExamServerUrlBrowserOnly()
        if (found) {
          this.setApiBaseUrl(found)
          this.lastDiscoveryMessage = '已自动连接到考试服务'
        } else {
          this.lastDiscoveryMessage =
            '未找到局域网内的考试服务，将尝试使用本机默认连接；可在设置中手动指定地址'
        }
      } catch {
        this.lastDiscoveryMessage = '自动查找未成功，请确认考试服务已启动，或在设置中填写地址'
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
          this.lastDiscoveryMessage = '已找到考试服务'
          return found
        }
        this.lastDiscoveryMessage =
          '未发现运行中的考试服务，请确认服务已启动且网络与防火墙允许访问'
        return null
      } finally {
        this.discoveryPhase = 'done'
      }
    },
  },
})
