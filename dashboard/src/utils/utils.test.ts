import { describe, expect, it } from 'vitest'

import { deriveSymbol, displaySymbol } from './bot'
import { lookbackToString, minutesToDurationString } from './index'

describe('deriveSymbol', () => {
  it('joins base and quote uppercase for binance crypto', () => {
    expect(deriveSymbol('binance', 'crypto', 'btc', 'usdt')).toBe('BTCUSDT')
  })

  it('defaults quote to empty string', () => {
    expect(deriveSymbol('binance', 'crypto', 'btc')).toBe('BTC')
  })
})

describe('displaySymbol', () => {
  it('formats crypto pairs with a slash', () => {
    expect(displaySymbol('crypto', 'btc', 'usdt')).toBe('btc/usdt')
  })

  it('returns the base when no quote', () => {
    expect(displaySymbol('crypto', 'btc')).toBe('btc')
  })
})

describe('minutesToDurationString', () => {
  it('handles zero', () => {
    expect(minutesToDurationString(0)).toBe('0m')
  })

  it('formats hours', () => {
    expect(minutesToDurationString(60)).toBe('1h')
  })

  it('shows the two largest units', () => {
    expect(minutesToDurationString(1500)).toBe('1d 1h') // 25h
    expect(minutesToDurationString(10080)).toBe('1w') // 7d
  })
})

describe('lookbackToString', () => {
  it('converts candles x interval to a duration', () => {
    expect(lookbackToString(200, '1h')).toBe('1w 1d') // 200h = 8d 8h -> weeks win
    expect(lookbackToString(1440, '1m')).toBe('1d')
  })
})
