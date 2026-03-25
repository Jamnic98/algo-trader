import { api } from 'api'
import type { CreateStrategy, Strategy, StrategySchema } from 'types'

export const getStrategies = async (): Promise<Strategy[]> =>
  (await api.fetchJson<{ strategies: Strategy[] }>('/strategies')).strategies

export const getStrategySchemas = async (): Promise<StrategySchema[]> =>
  api.fetchJson<StrategySchema[]>('/strategies/schemas')

export const createStrategy = async (data: CreateStrategy): Promise<Strategy> =>
  api.fetchJson<Strategy>('/strategies', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })

export const getStrategy = async (id: string): Promise<Strategy> =>
  api.fetchJson<Strategy>(`/strategies/${id}`)
