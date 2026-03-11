import type { OHLCVCandle } from 'types'

type BotStatus = 'created' | 'attached' | 'running'

export type BotCreateData = {
  id?: string
  interval: string
  base: string
  quote: string
  lookback: string
  quantity: string
}

export interface Bot extends BotCreateData {
  id: string
  status: BotStatus
  started?: string

  candles?: OHLCVCandle[] | null
}
