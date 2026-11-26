import { api } from 'api'
import type { BacktestRequest, BacktestResult, SweepRequest, SweepResponse } from 'types'

export const runBacktest = async (data: BacktestRequest): Promise<BacktestResult> =>
  api.fetchJson<BacktestResult>('/backtests', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })

export const runSweep = async (data: SweepRequest): Promise<SweepResponse> =>
  api.fetchJson<SweepResponse>('/backtests/sweep', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
