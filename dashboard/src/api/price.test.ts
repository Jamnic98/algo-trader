import { beforeEach, describe, expect, it, vi } from 'vitest'

import { getPrice } from './price'

describe('getPrice', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('requests /price with symbol and default binance exchange', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ price: '50123.50', exchange: 'binance' }),
    })
    vi.stubGlobal('fetch', fetchMock)

    const price = await getPrice('BTCUSDT')

    expect(price).toBe(50123.5)
    const [url, options] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toContain('/price?symbol=BTCUSDT&exchange=binance')
    expect((options.headers as Record<string, string>).Authorization).toMatch(/^ApiKey /)
  })

  it('passes the exchange through and encodes the symbol', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ price: '50123.50', exchange: 'alpaca' }),
    })
    vi.stubGlobal('fetch', fetchMock)

    const price = await getPrice('BTC/USD', 'alpaca')

    expect(price).toBe(50123.5)
    const [url] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toContain('symbol=BTC%2FUSD&exchange=alpaca')
  })

  it('throws with the HTTP status when the API errors', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({ ok: false, status: 400 })
    )

    await expect(getPrice('BTCUSD', 'alpaca')).rejects.toThrow('HTTP 400')
  })
})
