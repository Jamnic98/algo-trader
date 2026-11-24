import { beforeEach, describe, expect, it, vi } from 'vitest'

import { runBacktest } from './backtests'
import type { BacktestRequest } from 'types'

const req: BacktestRequest = {
  strategy_slug: 'simple',
  symbol: 'BTCUSDT',
  interval: '1h',
  start: '2024-01-01',
  end: '2024-02-01T23:59:59Z',
  initial_balance: '10000',
  fee_rate: '0.001',
}

describe('runBacktest', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('POSTs to /backtests with auth header and JSON body, returns parsed result', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ strategy_slug: 'simple', stats: { num_trades: 2 } }),
    })
    vi.stubGlobal('fetch', fetchMock)

    const res = await runBacktest(req)

    expect(res.strategy_slug).toBe('simple')
    expect(res.stats.num_trades).toBe(2)

    const [url, options] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toContain('/backtests')
    expect(options.method).toBe('POST')
    expect((options.headers as Record<string, string>).Authorization).toMatch(/^ApiKey /)
    expect((options.headers as Record<string, string>)['Content-Type']).toBe('application/json')
    const body = JSON.parse(options.body as string)
    expect(body.strategy_slug).toBe('simple')
    expect(body.symbol).toBe('BTCUSDT')
    expect(body.interval).toBe('1h')
  })

  it('throws with the HTTP status when the API errors', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({ ok: false, status: 404 })
    )

    await expect(runBacktest(req)).rejects.toThrow('HTTP 404')
  })
})
