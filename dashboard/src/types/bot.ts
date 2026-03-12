import type { OHLCVCandle } from 'types'

type BotStatus = 'created' | 'attached' | 'running'

export type BotMode = 'paper' | 'live'

export type BotCreateData = {
  id?: string
  interval: string
  base: string
  mode: BotMode
  quote: string
  lookback: string
  quantity: string
}

export interface Bot extends BotCreateData {
  id: string
  status: BotStatus
  started?: string
  candles?: OHLCVCandle[]
  deletedAt?: string
}
