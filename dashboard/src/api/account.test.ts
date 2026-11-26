import { beforeEach, describe, expect, it, vi } from 'vitest'

import { getLiveAccount, syncLiveAccount } from './account'

const okJson = (body: unknown) => ({ ok: true, status: 200, json: async () => body })

describe('account live endpoints', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('getLiveAccount defaults to no exchange param (server default = binance)', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      okJson({ account: { name: 'live:binance' }, exchange_balances: null })
    )
    vi.stubGlobal('fetch', fetchMock)

    const res = await getLiveAccount()

    expect(res.account.name).toBe('live:binance')
    const [url] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toContain('/account/live')
    expect(url).not.toContain('exchange=')
  })

  it('getLiveAccount passes ?exchange= through', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      okJson({
        account: { name: 'live:alpaca' },
        exchange: 'alpaca',
        exchange_balances: { USD: '100', AAPL: '3' },
      })
    )
    vi.stubGlobal('fetch', fetchMock)

    const res = await getLiveAccount('alpaca')

    expect(res.exchange_balances?.USD).toBe('100')
    const [url] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toContain('/account/live?exchange=alpaca')
  })

  it('syncLiveAccount without a param syncs all exchanges (POST, no query)', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      okJson({ accounts: { binance: { name: 'live:binance' } }, errors: {} })
    )
    vi.stubGlobal('fetch', fetchMock)

    const res = await syncLiveAccount()

    expect('accounts' in res && res.accounts.binance.name).toBe('live:binance')
    const [url, options] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toContain('/account/live/sync')
    expect(url).not.toContain('exchange=')
    expect(options.method).toBe('POST')
  })

  it('syncLiveAccount with an exchange passes ?exchange=', async () => {
    const fetchMock = vi.fn().mockResolvedValue(okJson({ account: { name: 'live:alpaca' } }))
    vi.stubGlobal('fetch', fetchMock)

    const res = await syncLiveAccount('alpaca')

    expect('account' in res && res.account.name).toBe('live:alpaca')
    const [url, options] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toContain('/account/live/sync?exchange=alpaca')
    expect(options.method).toBe('POST')
  })
})
