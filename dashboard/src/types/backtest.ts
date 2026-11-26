import type { Exchange } from './bot'

export type BacktestRequest = {
  strategy_slug: string
  exchange?: Exchange // venue for market data, default binance
  symbol?: string // single venue-native symbol (BTCUSDT, BTC/USD, AAPL)
  symbols?: string[] // multi-symbol portfolio run (overrides symbol)
  interval: string
  start: string
  end: string
  initial_balance?: string
  fee_rate?: string
  slippage_bps?: string // adverse slippage on fills, in basis points
  quantity?: string
}

export type SweepRequest = BacktestRequest & {
  param_grid: Record<string, unknown[]> // param key -> candidate values (cartesian product)
  max_combos?: number
  include_trades?: boolean
  sort_by?: 'total_return_pct' | 'sharpe' | 'profit_factor' | 'max_drawdown_pct' | 'num_trades'
}

export type SweepResultItem = {
  params: Record<string, unknown>
  stats: BacktestStats
  equity_curve: EquityPoint[]
  trades: BacktestTrade[]
}

export type SweepResponse = {
  strategy_slug: string
  exchange?: string
  symbols: string[]
  interval: string
  start: number
  end: number
  count: number
  sort_by: string
  results: SweepResultItem[]
}

export type BacktestTrade = {
  time: number
  side: 'BUY' | 'SELL'
  price: string
  qty: string
  fee: string
  balance: string
  pnl: string | null
  symbol?: string | null
}

export type BacktestStats = {
  initial_balance: string
  final_balance: string
  total_return_pct: string
  max_drawdown_pct: string
  num_trades: number
  win_rate: string
  profit_factor: string | null
  sharpe: string | null
}

export type EquityPoint = {
  time: number
  equity: string
}

export type BacktestResult = {
  strategy_slug: string
  exchange?: string
  symbol: string // first symbol (back-compat)
  symbols?: string[]
  interval: string
  start: number
  end: number
  stats: BacktestStats
  equity_curve: EquityPoint[]
  trades: BacktestTrade[]
}
