export * from './account'
export * from './alert'
export * from './bot'
export * from './candle'
export * from './diagnostic'
export * from './tab'
export * from './trade'

export type HistoryPoint = { t: string; cpu: number; mem: number }
export type ConnStatus = 'connecting' | 'live' | 'disconnected'
export type Pagination = {
  page: number
  limit: number
  total: number
  total_pages: number
}

export type LoadingAction = 'attach' | 'detach' | 'start' | 'stop' | 'delete' | null

export type Position = {
  symbol: string
  qty: string
  avg_entry: string
  total_spent: string
  realised_pnl: string
  total_fees: string
}
