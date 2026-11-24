import { api } from 'api'
import type { BacktestRequest, BacktestResult } from 'types'

export const runBacktest = async (data: BacktestRequest): Promise<BacktestResult> =>
  api.fetchJson<BacktestResult>('/backtests', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
