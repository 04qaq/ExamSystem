/** 后端试卷时间格式 `2006-01-02 15:04:05` → 时间戳（本地解析） */
export function parseBackendDateTime(s: string): number {
  const normalized = s.trim().replace(' ', 'T')
  const t = Date.parse(normalized)
  return Number.isNaN(t) ? Date.now() : t
}

/** DatePicker 时间戳 → ISO8601，供 Gin 绑定 time.Time */
export function toApiDateTime(ms: number): string {
  return new Date(ms).toISOString()
}
