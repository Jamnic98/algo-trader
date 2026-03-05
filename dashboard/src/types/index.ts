export * from './account'
export * from './alert'
export * from './bot'
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
