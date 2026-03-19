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

// TODO: add more valid intervals
export const candleIntervals = [
  '1m',
  // '3m',
  '5m',
  '15m',
  // '30m',
  '1h',
  // '2h',
  // '4h',
  // '6h',
  // '8h',
  // '12h',
  // '1d',
  // '3d',
  // '1w',
  // '1M'
]

// --- Config ---
// When you add a new exchange, just add it here.
// assetTypes drives what's shown in the Asset Type dropdown.
export const EXCHANGE_CONFIG: Record<Exchange, { assetTypes: AssetType[] }> = {
  binance: {
    assetTypes: ['crypto'],
  },
}

// Hardcoded for now — swap the array for an API/DB fetch later
// keeping the shape the same so the select doesn't need to change
export const AVAILABLE_STRATEGIES: { name: string; label: string }[] = [
  { name: 'simple', label: 'Simple' },
  { name: 'simpleDca', label: 'Simple DCA' },
  // TODO: add smart dca strategy
  // { name: 'smartDca', label: 'Smart DCA' },
]

export const getStrategyLabel = (name: string): string => {
  return AVAILABLE_STRATEGIES.find((s) => s.name === name)?.label ?? name
}
