import type { AssetType, Exchange } from 'types'

type StrategyFieldConfig = {
  managesQuantity?: boolean
  managesInterval?: boolean
  managesLookback?: boolean
  defaultLookback?: number
}

export const STRATEGY_CONFIG: Record<string, StrategyFieldConfig> = {
  simple: {
    managesQuantity: false,
    managesInterval: false,
    managesLookback: false,
    defaultLookback: 200,
  },
  simpleDca: {
    managesQuantity: false,
    managesInterval: false,
    managesLookback: true,
  },
  // TODO: implement smart dca
  // smartDca: {
  //   managesQuantity: true,
  //   managesInterval: true,
  //   managesLookback: true,
  // },
}

export const candleIntervals = [
  '1m',
  '3m',
  '5m',
  '15m',
  '30m',
  '1h',
  '2h',
  '4h',
  '6h',
  '8h',
  '12h',
  '1d',
  '3d',
  // TODO: include these
  // '1w',
  // '1M',
]

// --- Config ---
// When you add a new exchange, just add it here.
// assetTypes drives what's shown in the Asset Type dropdown.
export const EXCHANGE_CONFIG: Record<Exchange, { assetTypes: AssetType[] }> = {
  binance: {
    assetTypes: ['crypto'],
  },
  // alpaca: {
  //   assetTypes: ['crypto', 'stocks'],
  // },
}

// Hardcoded for now — swap the array for an API/DB fetch later
// keeping the shape the same so the select doesn't need to change
export const AVAILABLE_STRATEGIES: { name: string; label: string }[] = [
  { name: 'simple', label: 'Simple' },
  { name: 'simpleDca', label: 'Simple DCA' },
  // TODO: add smart dca strategy
  // { name: 'smartDca', label: 'Smart DCA' },
]

export const AVAILABLE_EXCHANGES: { name: Exchange; label: string }[] = [
  { name: 'binance', label: 'Binance' },
  // { name: 'alpaca', label: 'Alpaca' },
]

export const getStrategyLabel = (name: string): string => {
  return AVAILABLE_STRATEGIES.find((s) => s.name === name)?.label ?? name
}

export const getExchangeLabel = (name: string): string => {
  return AVAILABLE_EXCHANGES.find((s) => s.name === name)?.label ?? name
}

export const MAX_CANDLES: Record<string, number> = {
  '1m': 500,
  '3m': 500,
  '5m': 500,
  '15m': 500,
  '30m': 500,
  '1h': 500,
  '2h': 500,
  '4h': 400,
  '6h': 300,
  '8h': 300,
  '12h': 200,
  '1d': 365,
  '3d': 200,
  '1w': 104,
  '1M': 100,
}
