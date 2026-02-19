import { apiFetch } from 'api'

const tradesEndpoint = '/api/trades'

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

export const getAllTrades = async () => (await apiFetch<{ trades: Trade[] }>(tradesEndpoint)).trades
