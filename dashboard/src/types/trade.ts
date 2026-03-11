export type Trade = {
  id: number
  botID: string
  symbol: string
  side: string // BUY / SELL / NONE
  price: number // per unit price
  quantity: number
  fee: number // in fee asset
  feeAsset: string // e.g. "USDT", "BTC"
  exchange: string
  timestamp: Date
  createdAt: Date // GORM convention
}

export type TradeFilters = {
  limit: number
  symbol: string
  side: string
  botId: string
  dateFrom: string
  dateTo: string
}
