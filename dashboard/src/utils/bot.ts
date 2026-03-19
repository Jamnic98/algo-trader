const INTERVAL_TO_HOURS: Record<string, number> = {
  '1m': 1 / 60,
  '5m': 5 / 60,
  '15m': 15 / 60,
  '1h': 1,
  '4h': 4,
  '1d': 24,
}

export const lookbackToString = (candles: number, interval: string): string => {
  const hours = (INTERVAL_TO_HOURS[interval] ?? 1) * candles
  return `${Math.round(hours)}h`
}

export const deriveSymbol = (
  exchange: string,
  assetType: string,
  base: string,
  quote?: string
): string => {
  switch (exchange) {
    case 'binance':
      switch (assetType) {
        case 'crypto':
          return `${base}${quote ?? ''}`.toUpperCase()
        case 'stocks':
          return base.toUpperCase()
      }
  }
  return `${base}${quote ?? ''}`.toUpperCase() // fallback
}

export const displaySymbol = (assetType: string, base: string, quote?: string): string => {
  if (assetType === 'crypto' && quote) return `${base}/${quote}`
  return base
}
