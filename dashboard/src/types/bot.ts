import type { OHLCVCandle } from 'types'

export type BotStatus = 'created' | 'trading' | 'dead'
export type BotMode = 'paper' | 'live'
export type Exchange = 'binance' | 'alpaca'
export type AssetType = 'crypto' | 'stocks'

export interface Bot extends CreateBot {
  id: string
  status: BotStatus
  started?: string
  deletedAt?: string
  candles: OHLCVCandle[]
  strategy_name: string
}

export interface RiskLimits {
  max_position_pct?: string // cap a single BUY notional as % of balance, "100" = no cap
  max_daily_loss_pct?: string // kill switch vs day-start equity, "0" = off
  max_drawdown_pct?: string // kill switch vs running peak equity, "0" = off
  cooldown_losses?: number // consecutive losing sells before buys pause, 0 = off
  cooldown_minutes?: number // pause length
}

export interface CreateBot {
  mode: BotMode
  account?: string // paper account name; 'live' when mode is live
  exchange: Exchange
  assetType: AssetType
  base: string
  quote?: string
  interval: string
  maxCandles: number
  quantity?: string
  strategy_slug: string
  risk?: RiskLimits
}
