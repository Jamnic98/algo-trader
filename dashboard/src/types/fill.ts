export type OrderSignal = 'BUY' | 'SELL' | 'HOLD'

export type Fill = {
  id: number
  botID: string
  symbol: string
  side: OrderSignal // BUY / SELL / NONE
  price: number // per unit price
  quantity: number
  fee: number // in fee asset
  feeAsset: string // e.g. "USDT", "BTC"
  exchange: string
  timestamp: Date
  createdAt: Date // GORM convention
}

export type FillFilters = {
  limit: number
  symbol: string
  side: string
  botId: string
  dateFrom: string
  dateTo: string
}
