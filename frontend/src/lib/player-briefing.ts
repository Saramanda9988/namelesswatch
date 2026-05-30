export const PLAYER_BRIEFING_FILE = 'briefing.json'

export type PlayerBriefingItem = {
  id: string
  text: string
  detail?: string
}

export type PlayerBriefing = {
  title: string
  description?: string
  items: PlayerBriefingItem[]
  confirmText: string
}

const fallbackTitle = '你需要记住的规则'
const fallbackConfirmText = '我已记住'

export function parsePlayerBriefing(files?: Record<string, string>): PlayerBriefing | undefined {
  const raw = files?.[PLAYER_BRIEFING_FILE]?.trim()
  if (!raw) {
    return undefined
  }

  try {
    const source = JSON.parse(raw) as unknown
    if (!isRecord(source)) {
      return undefined
    }

    const title = stringValue(source.title) || fallbackTitle
    const description = stringValue(source.description)
    const confirmText = stringValue(source.confirmText) || fallbackConfirmText
    const rawItems = Array.isArray(source.items)
      ? source.items
      : Array.isArray(source.rules)
        ? source.rules
        : []
    const items = rawItems
      .map((item, index) => parseBriefingItem(item, index))
      .filter((item): item is PlayerBriefingItem => Boolean(item))

    if (items.length === 0 && !description) {
      return undefined
    }

    return {
      title,
      description,
      items,
      confirmText,
    }
  }
  catch {
    return undefined
  }
}

function parseBriefingItem(value: unknown, index: number): PlayerBriefingItem | undefined {
  if (typeof value === 'string') {
    const text = value.trim()
    return text ? { id: `rule-${index + 1}`, text } : undefined
  }
  if (!isRecord(value)) {
    return undefined
  }

  const text = stringValue(value.text) || stringValue(value.label) || stringValue(value.content)
  if (!text) {
    return undefined
  }

  return {
    id: stringValue(value.id) || `rule-${index + 1}`,
    text,
    detail: stringValue(value.detail),
  }
}

function stringValue(value: unknown) {
  return typeof value === 'string' ? value.trim() : ''
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value && typeof value === 'object' && !Array.isArray(value))
}
