/** 解析题目 options 字段（后端存 JSON 数组字符串或原始片段） */
export function parseOptionList(raw: string): string[] {
  if (!raw?.trim()) return []
  try {
    const j = JSON.parse(raw) as unknown
    if (Array.isArray(j)) return j.map((x) => String(x))
    if (j && typeof j === 'object') return Object.values(j as Record<string, unknown>).map(String)
  } catch {
    /* fallthrough */
  }
  return raw
    .split(/\r?\n/)
    .map((s) => s.trim())
    .filter(Boolean)
}

/**
 * 教师保存题目时，`exam-server` 的 optionsToJSON 按「换行分隔」接收选项文本，再打成 JSON 入库。
 * 若误粘贴 JSON 数组字符串，先转成多行再提交，避免整段被当成一个选项。
 */
export function normalizeQuestionOptionsInput(raw: string): string {
  const t = raw.replace(/\r\n/g, '\n').trim()
  if (!t) return ''
  try {
    const j = JSON.parse(t) as unknown
    if (Array.isArray(j)) return j.map(String).join('\n')
  } catch {
    /* 非 JSON，当作多行原文 */
  }
  return raw.replace(/\r\n/g, '\n').trim()
}
