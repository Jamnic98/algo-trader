import type { OHLCVCandle } from 'types'

export * from './bot'
export * from './constants'

export const INTERVAL_TO_HOURS: Record<string, number> = {
  '1m': 1 / 60,
  '3m': 3 / 60,
  '5m': 5 / 60,
  '15m': 15 / 60,
  '30m': 30 / 60,
  '1h': 1,
  '2h': 2,
  '4h': 4,
  '6h': 6,
  '8h': 8,
  '12h': 12,
  '1d': 24,
  '3d': 72,
  '1w': 168,
  '1M': 720, // 30 days
}

// used in form — theoretical, before bot runs
export const lookbackToString = (candles: number, interval: string): string => {
  const totalMinutes = Math.round((INTERVAL_TO_HOURS[interval] ?? 1) * candles * 60)
  return minutesToDurationString(totalMinutes)
}

// used in candles tab — actual range from real data
export const candleRangeToString = (candles: OHLCVCandle[]): string => {
  if (candles.length < 2) return '-'
  const totalMinutes = Math.round(
    (candles[candles.length - 1].openTime - candles[0].openTime) / 60000
  )
  return minutesToDurationString(totalMinutes)
}

// shared logic
export const minutesToDurationString = (totalMinutes: number): string => {
  const years = Math.floor(totalMinutes / 525960)
  const months = Math.floor((totalMinutes % 525960) / 43830)
  const weeks = Math.floor((totalMinutes % 43830) / 10080)
  const days = Math.floor((totalMinutes % 10080) / 1440)
  const hours = Math.floor((totalMinutes % 1440) / 60)
  const minutes = totalMinutes % 60

  const parts = [
    years && `${years}y`,
    months && `${months}mo`,
    weeks && `${weeks}w`,
    days && `${days}d`,
    hours && `${hours}h`,
    minutes && `${minutes}m`,
  ].filter(Boolean)

  return parts.slice(0, 2).join(' ') || '0m'
}

export const makeBotRunDurationStr = (botStart: string): string => {
  const seconds = Math.floor((Date.now() - Date.parse(botStart)) / 1000)
  const y = Math.floor(seconds / 31536000)
  const mo = Math.floor((seconds % 31536000) / 2592000)
  const d = Math.floor((seconds % 2592000) / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  return [y && `${y}y`, mo && `${mo}mo`, d && `${d}d`, h && `${h}h`, m && `${m}m`, `${s}s`]
    .filter(Boolean)
    .join(' ')
}
