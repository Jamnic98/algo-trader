export type BacktestRequest = {
  strategy_slug: string
  symbol: string
  interval: string
  start: string
  end: string
  initial_balance?: string
  fee_rate?: string
  quantity?: string
}

export type BacktestTrade = {
  time: number
  side: 'BUY' | 'SELL'
  price: string
  qty: string
  fee: string
  balance: string
  pnl: string | null
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
  symbol: string
  interval: string
  start: number
  end: number
  stats: BacktestStats
  equity_curve: EquityPoint[]
  trades: BacktestTrade[]
}
