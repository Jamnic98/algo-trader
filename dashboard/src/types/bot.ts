import type { OHLCVCandle, Strategy } from 'types'

export type BotStatus = 'created' | 'trading' | 'dead'
export type BotMode = 'paper' | 'live'
export type Exchange = 'binance'
export type AssetType = 'crypto'

export type Bot = {
  id: string
  mode: BotMode
  exchange: Exchange
  strategy_id: string
  strategy?: Strategy
  assetType: AssetType
  base: string
  quote?: string
  interval: string
  maxCandles: number
  quantity: string
  status: BotStatus
  started?: string
  candles?: OHLCVCandle[]
  deletedAt?: string
}

export type CreateBot = {
  mode: BotMode
  exchange: Exchange
  assetType: AssetType
  base: string
  quote?: string
  interval: string
  maxCandles: number
  quantity?: string
  strategy_id: string
}

export interface BotStrategy extends CreateBotStrategy {
  id?: string
}

export type CreateBotStrategy = {
  name: string
  displayName?: string
  params?: Record<string, unknown> // strategy-specific config
  makerFee?: number
  takerFee?: number
}
