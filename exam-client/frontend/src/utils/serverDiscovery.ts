/** exam-server 在 GET /api/ping 返回的标识，用于区分同名端口上的其它服务 */
export const EXAM_SERVICE_ID = 'exam-server'

const PROBE_MS = 900
const LAN_WORKERS = 48

function stripTrailingSlash(u: string): string {
  return u.replace(/\/+$/, '')
}

/**
 * Wails 桌面端由资源服务器同源提供 GET /discover（Go UDP/TCP 发现）。
 * 不使用 127.0.0.1 随机端口，否则 WebView2 可能按本地环回策略拦截 fetch。
 */
async function discoverViaAppDiscoverRoute(): Promise<string | null> {
  const ctrl = new AbortController()
  const t = window.setTimeout(() => ctrl.abort(), 125000)
  try {
    const res = await fetch('/discover', {
      method: 'GET',
      signal: ctrl.signal,
      headers: { Accept: 'application/json' },
    })
    if (!res.ok) return null
    const ct = res.headers.get('content-type') ?? ''
    if (!ct.includes('json')) return null
    const j = (await res.json()) as { url?: string }
    const u = (j.url ?? '').trim()
    return u ? stripTrailingSlash(u) : null
  } catch {
    return null
  } finally {
    window.clearTimeout(t)
  }
}

/** 探测单个 HTTP 根地址（含协议、主机、端口）是否为 exam-server */
export async function probeExamServerBase(baseOrigin: string): Promise<boolean> {
  const root = stripTrailingSlash(baseOrigin)
  const ctrl = new AbortController()
  const t = window.setTimeout(() => ctrl.abort(), PROBE_MS)
  try {
    const res = await fetch(`${root}/api/ping`, {
      method: 'GET',
      signal: ctrl.signal,
    })
    if (!res.ok) return false
    const j = (await res.json()) as { code?: number; data?: { service?: string } }
    return j?.code === 0 && j?.data?.service === EXAM_SERVICE_ID
  } catch {
    return false
  } finally {
    window.clearTimeout(t)
  }
}

async function raceProbe(urls: string[]): Promise<string | null> {
  if (!urls.length) return null
  let i = 0
  async function worker(): Promise<string | null> {
    for (;;) {
      const cur = i++
      if (cur >= urls.length) return null
      const u = urls[cur]
      if (await probeExamServerBase(u)) return u
    }
  }
  const n = Math.min(LAN_WORKERS, urls.length)
  const hits = await Promise.all(Array.from({ length: n }, () => worker()))
  return hits.find(Boolean) ?? null
}

function subnetOf(ipv4: string): string {
  const p = ipv4.split('.')
  return `${p[0]}.${p[1]}.${p[2]}`
}

/** 通过 WebRTC 候选尽量拿到本机局域网 IPv4（同一 Wi‑Fi / 有线网段） */
export function gatherLocalIPv4s(timeoutMs = 2800): Promise<string[]> {
  return new Promise((resolve) => {
    const ips = new Set<string>()
    let settled = false
    const finish = () => {
      if (settled) return
      settled = true
      try {
        pc.close()
      } catch {
        /* ignore */
      }
      resolve([...ips])
    }

    const pc = new RTCPeerConnection({ iceServers: [] })
    pc.createDataChannel('')
    pc
      .createOffer()
      .then((o) => pc.setLocalDescription(o))
      .catch(() => finish())

    const timer = window.setTimeout(finish, timeoutMs)

    pc.onicecandidate = (e) => {
      const c = e.candidate?.candidate
      if (!c) return
      const m = /\b([0-9]{1,3}(?:\.[0-9]{1,3}){3})\b/.exec(c)
      if (!m) return
      const ip = m[1]
      if (
        ip.startsWith('127.') ||
        ip.startsWith('169.254.') ||
        ip === '0.0.0.0'
      ) {
        return
      }
      ips.add(ip)
    }

    pc.onicegatheringstatechange = () => {
      if (pc.iceGatheringState === 'complete') {
        window.clearTimeout(timer)
        finish()
      }
    }
  })
}

/**
 * 桌面端同源 GET /discover（Go 内已用 /api/ping 校验）。
 * 切勿再在前端 probe：WebView2 常拦截页面访问 127.0.0.1/localhost，会导致「同机也找不到」的假阴性。
 */
export async function tryNativeLanDiscover(): Promise<string | null> {
  return discoverViaAppDiscoverRoute()
}

/**
 * 仅在浏览器内可用的扫描（不含同源 /discover）：
 * 1. 本机 loopback，端口 8080、80
 * 2. 局域网 HTTP 扫描
 */
export async function discoverExamServerUrlBrowserOnly(): Promise<string | null> {
  const localPorts = [8080, 80]
  const localUrls: string[] = []
  for (const p of localPorts) {
    localUrls.push(`http://127.0.0.1:${p}`, `http://localhost:${p}`, `http://[::1]:${p}`)
  }
  const localHit = await raceProbe(localUrls)
  if (localHit) return stripTrailingSlash(localHit)

  const prefixes = new Set<string>()
  try {
    const locals = await gatherLocalIPv4s()
    for (const ip of locals) prefixes.add(subnetOf(ip))
  } catch {
    /* ignore */
  }
  // WebRTC 拿不到本机 IP 时（部分 WebView/权限策略），补上家用路由与手机热点常见网段
  if (prefixes.size === 0) {
    for (const pre of [
      '192.168.1',
      '192.168.0',
      '192.168.31',
      '192.168.43', // 多数 Android 热点
      '192.168.137', // Windows「移动热点」常见
      '172.20.10', // 部分 iPhone 个人热点给下游分配的网段
      '10.0.0',
    ]) {
      prefixes.add(pre)
    }
  }

  const lanUrls: string[] = []
  for (const pre of prefixes) {
    for (let n = 1; n <= 254; n++) {
      lanUrls.push(`http://${pre}.${n}:8080`)
    }
  }

  const lanHit = await raceProbe(lanUrls)
  return lanHit ? stripTrailingSlash(lanHit) : null
}

/** 原生发现（桌面）+ 浏览器扫描兜底 */
export async function discoverExamServerUrl(): Promise<string | null> {
  const n = await tryNativeLanDiscover()
  if (n) return n
  return discoverExamServerUrlBrowserOnly()
}
