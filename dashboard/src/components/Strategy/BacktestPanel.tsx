import { useState } from 'react'
import { Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'

import { runBacktest, runSweep } from 'api'
import { SectionLabel, StatCard } from 'components'
import {
  AVAILABLE_EXCHANGES,
  CANDLE_INTERVALS,
  EXCHANGE_CONFIG,
  getExchangeLabel,
} from 'utils'
import type { BacktestResult, Exchange, SweepRequest, SweepResponse } from 'types'

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
  const [slippageBps, setSlippageBps] = useState('0')
  const [quantity, setQuantity] = useState('')

  const [result, setResult] = useState<BacktestResult | null>(null)
  const [running, setRunning] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // sweep state (same market data window as the single run above)
  const [sweepGrid, setSweepGrid] = useState('{\n  "fast": [3, 5],\n  "slow": [10, 20]\n}')
  const [sweepSort, setSweepSort] = useState('total_return_pct')
  const [sweepResult, setSweepResult] = useState<SweepResponse | null>(null)
  const [sweeping, setSweeping] = useState(false)

  // shared request fields; the symbol box accepts a comma-separated list for
  // multi-symbol portfolio runs
  const baseRequest = () => {
    const syms = symbol
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    return {
      strategy_slug: strategySlug,
      exchange,
      interval,
      start,
      end: `${end}T23:59:59Z`,
      initial_balance: initialBalance,
      fee_rate: feeRate,
      ...(slippageBps && slippageBps !== '0' ? { slippage_bps: slippageBps } : {}),
      ...(quantity ? { quantity } : {}),
      ...(syms.length > 1
        ? { symbols: syms.map((s) => s.toUpperCase()) }
        : { symbol: (syms[0] ?? '').toUpperCase() }),
    }
  }

  const handleRun = async (e: React.FormEvent) => {
    e.preventDefault()
    setRunning(true)
    setError(null)
    try {
      const res = await runBacktest(baseRequest())
      setResult(res)
      setSweepResult(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Backtest failed')
      setResult(null)
    } finally {
      setRunning(false)
    }
  }

  const handleSweep = async (e: React.FormEvent) => {
    e.preventDefault()
    setSweeping(true)
    setError(null)
    try {
      const param_grid = JSON.parse(sweepGrid) as Record<string, unknown[]>
      if (typeof param_grid !== 'object' || param_grid === null || Array.isArray(param_grid)) {
        throw new Error('Param grid must be a JSON object of arrays, e.g. {"fast": [3, 5]}')
      }
      for (const [k, v] of Object.entries(param_grid)) {
        if (!Array.isArray(v) || v.length === 0) {
          throw new Error(`param_grid["${k}"] must be a non-empty array of values`)
        }
      }
      const res = await runSweep({
        ...baseRequest(),
        param_grid,
        sort_by: sweepSort as SweepRequest['sort_by'],
      })
      setSweepResult(res)
      setResult(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Sweep failed')
      setSweepResult(null)
    } finally {
      setSweeping(false)
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
          <label className="text-content-secondary text-xs uppercase tracking-wider">Symbol(s)</label>
          <input
            name="symbol"
            value={symbol}
            onChange={(e) => setSymbol(e.target.value)}
            placeholder={`${exchangeConfig?.defaultSymbol ?? 'BTCUSDT'}, ETHUSDT`}
            className={`${inputCls} w-40`}
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
            Slippage bps
          </label>
          <input
            name="slippage_bps"
            type="number"
            min="0"
            step="any"
            value={slippageBps}
            onChange={(e) => setSlippageBps(e.target.value)}
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
              {result.symbols?.length ? result.symbols.join(', ') : result.symbol}{' '}
              {result.interval}
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
                      {(result.symbols?.length ?? 1) > 1 && <th className="px-3 py-2">Symbol</th>}
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
                        {(result.symbols?.length ?? 1) > 1 && (
                          <td className="px-3 py-2 font-mono text-xs">{t.symbol ?? ''}</td>
                        )}
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

      {/* Parameter sweep: cartesian product over the grid, same market window */}
      <div className="flex flex-col gap-3 border-t border-table-border pt-4">
        <SectionLabel>Parameter sweep (same symbol/exchange/window)</SectionLabel>
        <form onSubmit={handleSweep} className="flex flex-wrap items-end gap-3">
          <div className="flex flex-col gap-1">
            <label className="text-content-secondary text-xs uppercase tracking-wider">
              Param grid (JSON)
            </label>
            <textarea
              name="param_grid"
              value={sweepGrid}
              onChange={(e) => setSweepGrid(e.target.value)}
              rows={3}
              spellCheck={false}
              className={`${inputCls} w-72 font-mono text-xs`}
            />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-content-secondary text-xs uppercase tracking-wider">Sort by</label>
            <select
              name="sort_by"
              value={sweepSort}
              onChange={(e) => setSweepSort(e.target.value)}
              className={inputCls}
            >
              <option value="total_return_pct">Return</option>
              <option value="sharpe">Sharpe</option>
              <option value="profit_factor">Profit factor</option>
              <option value="max_drawdown_pct">Max drawdown</option>
              <option value="num_trades">Trades</option>
            </select>
          </div>
          <button
            type="submit"
            disabled={sweeping}
            className="px-3 py-1.5 rounded bg-accent text-surface-primary font-semibold text-sm uppercase tracking-wider cursor-pointer hover:opacity-90 active:scale-95 transition-all disabled:opacity-50"
          >
            {sweeping ? 'Sweeping…' : 'Run sweep'}
          </button>
        </form>

        {sweepResult && (
          <div>
            <SectionLabel>
              Sweep results — {sweepResult.count} combos sorted by {sweepResult.sort_by}
            </SectionLabel>
            <div className="overflow-x-auto border border-table-border rounded-lg">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-content-secondary text-xs uppercase tracking-wider text-left border-b border-table-border">
                    <th className="px-3 py-2">#</th>
                    <th className="px-3 py-2">Params</th>
                    <th className="px-3 py-2">Return</th>
                    <th className="px-3 py-2">Max DD</th>
                    <th className="px-3 py-2">Sharpe</th>
                    <th className="px-3 py-2">PF</th>
                    <th className="px-3 py-2">Trades</th>
                    <th className="px-3 py-2">Win rate</th>
                  </tr>
                </thead>
                <tbody>
                  {sweepResult.results.map((r, i) => (
                    <tr key={i} className="border-b border-table-border last:border-0">
                      <td className="px-3 py-2 font-mono text-content-secondary">{i + 1}</td>
                      <td className="px-3 py-2 font-mono text-xs">{JSON.stringify(r.params)}</td>
                      <td
                        className={`px-3 py-2 font-mono ${Number(r.stats.total_return_pct) >= 0 ? 'text-green-400' : 'text-red-400'}`}
                      >
                        {fmtPct(r.stats.total_return_pct)}
                      </td>
                      <td className="px-3 py-2 font-mono">{fmtPct(r.stats.max_drawdown_pct)}</td>
                      <td className="px-3 py-2 font-mono">{fmtNum(r.stats.sharpe)}</td>
                      <td className="px-3 py-2 font-mono">{fmtNum(r.stats.profit_factor)}</td>
                      <td className="px-3 py-2 font-mono">{r.stats.num_trades}</td>
                      <td className="px-3 py-2 font-mono">{fmtPct(r.stats.win_rate)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default BacktestPanel
