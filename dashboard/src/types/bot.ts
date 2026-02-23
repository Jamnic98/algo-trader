type BotStatus = 'created' | 'attached' | 'running'

export type BotCreateData = {
  id?: string
  interval: string
  symbol: string
  lookback: string
}

export interface BotData extends BotCreateData {
  id: string
  status: BotStatus
  started?: string | null
}
