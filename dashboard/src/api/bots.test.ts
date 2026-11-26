import { beforeEach, describe, expect, it, vi } from 'vitest'

import { createBot } from './bots'
import type { CreateBot } from 'types'

const baseReq: CreateBot = {
  mode: 'paper',
  account: 'default',
  exchange: 'binance',
  assetType: 'crypto',
  base: 'BTC',
  quote: 'USDT',
  interval: '1h',
  maxCandles: 200,
  quantity: '0.001',
  strategy_slug: 'simple',
}

describe('createBot', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('POSTs the bot including the risk limits as JSON', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ bot: { id: 'b1' } }),
    })
    vi.stubGlobal('fetch', fetchMock)

    const req: CreateBot = {
      ...baseReq,
      risk: {
        max_position_pct: '25',
        max_daily_loss_pct: '5',
        max_drawdown_pct: '10',
        cooldown_losses: 3,
        cooldown_minutes: 60,
      },
    }
    const bot = await createBot(req)

    expect(bot.id).toBe('b1')
    const [url, options] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toContain('/bots')
    expect(options.method).toBe('POST')
    expect((options.headers as Record<string, string>).Authorization).toMatch(/^ApiKey /)
    const body = JSON.parse(options.body as string)
    expect(body.strategy_slug).toBe('simple')
    expect(body.risk).toEqual({
      max_position_pct: '25',
      max_daily_loss_pct: '5',
      max_drawdown_pct: '10',
      cooldown_losses: 3,
      cooldown_minutes: 60,
    })
  })

  it('omits risk when not set (backend applies defaults)', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ bot: { id: 'b2' } }),
    })
    vi.stubGlobal('fetch', fetchMock)

    await createBot(baseReq)

    const [, options] = fetchMock.mock.calls[0] as [string, RequestInit]
    const body = JSON.parse(options.body as string)
    expect('risk' in body).toBe(false)
  })
})
