import type { OHLCVCandle, Strategy } from 'types'

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
  strategy: Strategy
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
  strategy_id: number
}

export interface BotStrategy extends CreateBotStrategy {
  id?: string
}

export type CreateBotStrategy = {
  name: string
  displayName?: string
  params?: Record<string, unknown> // strategy-specific config
  makerFee?: number // defaults to 0.001
  takerFee?: number // defaults to 0.001
}
