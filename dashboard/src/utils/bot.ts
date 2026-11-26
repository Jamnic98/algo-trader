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
      break
    case 'alpaca':
      // crypto pairs are "BTC/USD", stocks just the ticker
      return assetType === 'crypto'
        ? `${base}/${quote ?? 'USD'}`.toUpperCase()
        : base.toUpperCase()
  }
  return `${base}${quote ?? ''}`.toUpperCase() // fallback
}

export const displaySymbol = (assetType: string, base: string, quote?: string): string => {
  if (assetType === 'crypto' && quote) return `${base}/${quote}`
  return base
}
