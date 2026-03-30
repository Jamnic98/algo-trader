import type { OHLCVCandle } from 'types'

export type BotStatus = 'created' | 'trading' | 'dead'
export type BotMode = 'paper' | 'live'
// TODO: Add exchange
export type Exchange = 'binance' //| 'alpaca'
export type AssetType = 'crypto' // | 'stocks'

export interface Bot extends CreateBot {
  id: string
  status: BotStatus
  started?: string
  deletedAt?: string
  candles: OHLCVCandle[]
  strategy_name: string
}

export interface CreateBot {
  mode: BotMode
  exchange: Exchange
  assetType: AssetType
  base: string
  quote?: string
  interval: string
  maxCandles: number
  quantity?: string
  strategy_slug: string
}
