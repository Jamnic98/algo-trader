import type { AssetType, CreateBotStrategy, Exchange } from 'types'

type StrategyFieldConfig = {
  managesQuantity?: boolean
  managesInterval?: boolean
  managesLookback?: boolean
  defaultLookback?: number
}

export const STRATEGY_CONFIG: Record<string, StrategyFieldConfig> = {
  simple: {
    managesQuantity: false, // bot sets quantity
    managesInterval: false, // bot sets interval
    managesLookback: false, // bot sets lookback
    defaultLookback: 200,
  },
  // future — e.g. a strategy that self-manages position sizing
  // grid: {
  //   managesQuantity: true,  // hide quantity field, strategy handles it
  // }
}

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
export const AVAILABLE_STRATEGIES: { label: string; value: CreateBotStrategy }[] = [
  {
    label: 'Simple',
    value: {
      name: 'simple',
    },
  },
  // add more strategies here, or fetch from DB and map to this shape
]
