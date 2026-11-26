import { useState } from 'react'
import { Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'

import { runBacktest } from 'api'
import { SectionLabel, StatCard } from 'components'
import {
  AVAILABLE_EXCHANGES,
  CANDLE_INTERVALS,
  EXCHANGE_CONFIG,
  getExchangeLabel,
} from 'utils'
import type { BacktestResult, Exchange } from 'types'

const inputCls =
  'border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors'

const fmtPct = (v: string) => `${Number(v).toFixed(2)}%`
const fmtNum = (v: string | null) => (v === null ? '—' : Number(v).toFixed(4))

const BacktestPanel = ({ strategySlug }: { strategySlug: string }) => {
  const [exchange, setExchange] = useState<Exchange>('binance')
  const [symbol, setSymbol] = useState(EXCHANGE_CONFIG.binance.defaultSymbol)
  const [interval, setInterval] = useState('1h')
  const [start, setStart] = useState(() => new Date(Date.now() - 30 * 864e5).toISOString().slice(0, 10))
  const [end, setEnd] = useState(() => new Date().toISOString().slice(0, 10))
  const [initialBalance, setInitialBalance] = useState('10000')
  const [feeRate, setFeeRate] = useState('0.001')
  const [quantity, setQuantity] = useState('')

  const [result, setResult] = useState<BacktestResult | null>(null)
  const [running, setRunning] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleRun = async (e: React.FormEvent) => {
    e.preventDefault()
    setRunning(true)
    setError(null)
    try {
      const res = await runBacktest({
        strategy_slug: strategySlug,
        exchange,
        symbol: symbol.toUpperCase(),
        interval,
        start,
        end: `${end}T23:59:59Z`,
        initial_balance: initialBalance,
        fee_rate: feeRate,
        ...(quantity ? { quantity } : {}),
      })
      setResult(res)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Backtest failed')
      setResult(null)
    } finally {
      setRunning(false)
    }
  }

  const equityData = (result?.equity_curve ?? []).map((p) => ({
    time: new Date(p.time).toLocaleString(),
    equity: Number(p.equity),
  }))

  const exchangeConfig = EXCHANGE_CONFIG[exchange]
  const availableIntervals = exchangeConfig?.intervals ?? CANDLE_INTERVALS

  const handleExchangeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const next = e.target.value as Exchange
    setExchange(next)
    const cfg = EXCHANGE_CONFIG[next]
    if (cfg) {
      setSymbol(cfg.defaultSymbol)
      if (!cfg.intervals.includes(interval)) setInterval('1h')
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <form onSubmit={handleRun} className="flex flex-wrap items-end gap-3">
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">Exchange</label>
          <select name="exchange" value={exchange} onChange={handleExchangeChange} className={inputCls}>
            {AVAILABLE_EXCHANGES.map((ex) => (
              <option key={ex.name} value={ex.name}>
                {ex.label}
              </option>
            ))}
          </select>
        </div>
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">Symbol</label>
          <input
            name="symbol"
            value={symbol}
            onChange={(e) => setSymbol(e.target.value)}
            placeholder={exchangeConfig?.defaultSymbol ?? 'BTCUSDT'}
            className={`${inputCls} w-32`}
            required
          />
        </div>
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">Interval</label>
          <select name="interval" value={interval} onChange={(e) => setInterval(e.target.value)} className={inputCls}>
            {availableIntervals.map((i) => (
              <option key={i} value={i}>
                {i}
              </option>
            ))}
          </select>
        </div>
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">Start</label>
          <input
            type="date"
            value={start}
            onChange={(e) => setStart(e.target.value)}
            className={inputCls}
            required
          />
        </div>
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">End</label>
          <input
            type="date"
            value={end}
            onChange={(e) => setEnd(e.target.value)}
            className={inputCls}
            required
          />
        </div>
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">Balance</label>
          <input
            type="number"
            min="0"
            step="any"
            value={initialBalance}
            onChange={(e) => setInitialBalance(e.target.value)}
            className={`${inputCls} w-28`}
          />
        </div>
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">Fee</label>
          <input
            type="number"
            min="0"
            step="any"
            value={feeRate}
            onChange={(e) => setFeeRate(e.target.value)}
            className={`${inputCls} w-24`}
          />
        </div>
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">
            Qty (optional)
          </label>
          <input
            type="number"
            min="0"
            step="any"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            placeholder="all-in"
            className={`${inputCls} w-24`}
          />
        </div>
        <button
          type="submit"
          disabled={running}
          className="px-3 py-1.5 rounded bg-accent text-surface-primary font-semibold text-sm uppercase tracking-wider cursor-pointer hover:opacity-90 active:scale-95 transition-all disabled:opacity-50"
        >
          {running ? 'Running…' : 'Run backtest'}
        </button>
      </form>

      {error && <p className="text-red-400 text-sm">{error}</p>}

      {result && (
        <div className="flex flex-col gap-6">
          <div>
            <SectionLabel>
              Results — {getExchangeLabel((result.exchange as Exchange) ?? exchange)}{' '}
              {result.symbol} {result.interval}
            </SectionLabel>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              <StatCard
                label="Total return"
                value={fmtPct(result.stats.total_return_pct)}
                accent={Number(result.stats.total_return_pct) >= 0}
              />
              <StatCard label="Final balance" value={Number(result.stats.final_balance).toFixed(2)} />
              <StatCard label="Max drawdown" value={fmtPct(result.stats.max_drawdown_pct)} />
              <StatCard label="Trades" value={result.stats.num_trades} />
              <StatCard label="Win rate" value={fmtPct(result.stats.win_rate)} />
              <StatCard label="Profit factor" value={fmtNum(result.stats.profit_factor)} />
              <StatCard label="Sharpe" value={fmtNum(result.stats.sharpe)} />
            </div>
          </div>

          {equityData.length > 1 && (
            <div>
              <SectionLabel>Equity curve</SectionLabel>
              <div className="bg-surface-secondary border border-table-border rounded-lg px-5 py-4">
                <ResponsiveContainer width="100%" height={220}>
                  <LineChart data={equityData} margin={{ top: 4, right: 8, left: -16, bottom: 0 }}>
                    <XAxis dataKey="time" tick={{ fontSize: 10 }} minTickGap={48} />
                    <YAxis domain={['auto', 'auto']} tick={{ fontSize: 10 }} />
                    <Tooltip
                      formatter={(v) => [Number(v).toFixed(2), 'equity']}
                      labelStyle={{ fontSize: 11 }}
                    />
                    <Line type="monotone" dataKey="equity" stroke="#22d3ee" dot={false} strokeWidth={2} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </div>
          )}

          {result.trades.length > 0 && (
            <div>
              <SectionLabel>Trades</SectionLabel>
              <div className="overflow-x-auto border border-table-border rounded-lg">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="text-content-secondary text-xs uppercase tracking-wider text-left border-b border-table-border">
                      <th className="px-3 py-2">Time</th>
                      <th className="px-3 py-2">Side</th>
                      <th className="px-3 py-2">Price</th>
                      <th className="px-3 py-2">Qty</th>
                      <th className="px-3 py-2">Fee</th>
                      <th className="px-3 py-2">PnL</th>
                    </tr>
                  </thead>
                  <tbody>
                    {result.trades.map((t, i) => (
                      <tr key={i} className="border-b border-table-border last:border-0">
                        <td className="px-3 py-2 whitespace-nowrap">
                          {new Date(t.time).toLocaleString()}
                        </td>
                        <td
                          className={`px-3 py-2 font-mono ${t.side === 'BUY' ? 'text-green-400' : 'text-red-400'}`}
                        >
                          {t.side}
                        </td>
                        <td className="px-3 py-2 font-mono">{Number(t.price).toFixed(2)}</td>
                        <td className="px-3 py-2 font-mono">{Number(t.qty).toFixed(6)}</td>
                        <td className="px-3 py-2 font-mono">{Number(t.fee).toFixed(4)}</td>
                        <td className="px-3 py-2 font-mono">
                          {t.pnl === null ? '—' : Number(t.pnl).toFixed(2)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default BacktestPanel
