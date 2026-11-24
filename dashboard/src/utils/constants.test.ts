import { describe, expect, it } from 'vitest'

import { AVAILABLE_EXCHANGES, EXCHANGE_CONFIG } from './constants'

describe('EXCHANGE_CONFIG', () => {
  it('includes binance and alpaca', () => {
    expect(AVAILABLE_EXCHANGES.map((e) => e.name)).toEqual(['binance', 'alpaca'])
    expect(Object.keys(EXCHANGE_CONFIG).sort()).toEqual(['alpaca', 'binance'])
  })

  it('alpaca supports stocks and crypto with USD quote and a subset of intervals', () => {
    const alpaca = EXCHANGE_CONFIG.alpaca
    expect(alpaca.assetTypes).toEqual(['crypto', 'stocks'])
    expect(alpaca.defaultQuote).toBe('USD')
    expect(alpaca.intervals).toContain('1h')
    expect(alpaca.intervals).not.toContain('3m') // not offered by Alpaca
  })

  it('binance keeps the full interval list and USDT quote', () => {
    const binance = EXCHANGE_CONFIG.binance
    expect(binance.assetTypes).toEqual(['crypto'])
    expect(binance.defaultQuote).toBe('USDT')
    expect(binance.intervals).toContain('3m')
  })
})
