import { useState, useEffect } from 'react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
  CartesianGrid,
  type TooltipContentProps,
} from 'recharts'
import type { NameType, ValueType } from 'recharts/types/component/DefaultTooltipContent'

import { getBotTrades, getPrice } from 'api'
import type { Bot, Trade } from 'types'

// ─── helpers ────────────────────────────────────────────────────────────────

const fmt = (v: number, dp = 4) =>
  new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: dp,
  }).format(v)

const fmtPct = (v: number) => `${v >= 0 ? '+' : ''}${v.toFixed(2)}%`

const fmtTime = (d: Date) =>
  new Intl.DateTimeFormat('en-GB', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(d))

// ─── derived stats from trades ───────────────────────────────────────────────

type Stats = {
  realisedPnl: number
  totalFees: number
  netPnl: number
  totalInvested: number
  currentQty: number
  avgEntry: number
  unrealisedPnl: number
  unrealisedPct: number
  chartData: { time: string; pnl: number }[]
}

const deriveStats = (trades: Trade[], currentPrice: number): Stats => {
  const sorted = [...trades].sort(
    (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  )

  let realisedPnl = 0
  let totalFees = 0
  let currentQty = 0
  let totalCost = 0
  let runningPnl = 0

  const chartData: { time: string; pnl: number }[] = []

  for (const t of sorted) {
    const fee = parseFloat(String(t.fee))
    const price = parseFloat(String(t.price))
    const quantity = parseFloat(String(t.quantity))

    totalFees += fee

    if (t.side === 'BUY') {
      totalCost += price * quantity
      currentQty += quantity
    } else if (t.side === 'SELL') {
      const avgEntry = currentQty > 0 ? totalCost / currentQty : 0
      const tradePnl = (price - avgEntry) * quantity
      realisedPnl += tradePnl
      totalCost = Math.max(0, totalCost - avgEntry * quantity)
      currentQty = Math.max(0, currentQty - quantity)
    }

    runningPnl = realisedPnl - totalFees
    chartData.push({ time: fmtTime(t.timestamp), pnl: parseFloat(runningPnl.toFixed(4)) })
  }

  const avgEntry = currentQty > 0 ? totalCost / currentQty : 0
  const totalInvested = totalCost
  const unrealisedPnl = currentQty > 0 ? (currentPrice - avgEntry) * currentQty - totalFees : 0
  const unrealisedPct = avgEntry > 0 ? ((currentPrice - avgEntry) / avgEntry) * 100 : 0
  const netPnl = realisedPnl - totalFees

  return {
    realisedPnl,
    totalFees,
    netPnl,
    totalInvested,
    currentQty,
    avgEntry,
    unrealisedPnl,
    unrealisedPct,
    chartData,
  }
}

// ─── stat card ───────────────────────────────────────────────────────────────

const Card = ({
  label,
  value,
  sub,
  color = 'neutral',
}: {
  label: string
  value: string
  sub?: string
  color?: 'green' | 'red' | 'neutral'
}) => (
  <div className="bg-table-row border border-table-border rounded-md px-4 py-3 flex flex-col gap-1">
    <span className="text-content-secondary text-xs uppercase tracking-wider font-semibold">
      {label}
    </span>
    <span
      className={`font-mono text-base font-semibold ${
        color === 'green'
          ? 'text-green-400'
          : color === 'red'
            ? 'text-red-400'
            : 'text-content-primary'
      }`}
    >
      {value}
    </span>
    {sub && <span className="text-content-secondary text-xs font-mono">{sub}</span>}
  </div>
)

// ─── chart tooltip ───────────────────────────────────────────────────────────

const ChartTooltip = ({ active, payload, label }: TooltipContentProps<ValueType, NameType>) => {
  if (!active || !payload?.length) return null
  const val = (payload[0].value ?? 0) as number
  return (
    <div className="bg-table-header border border-table-border rounded px-3 py-2 text-xs font-mono shadow-lg">
      <div className="text-content-secondary mb-1">{label as string}</div>
      <div className={val >= 0 ? 'text-green-400' : 'text-red-400'}>
        {val >= 0 ? '+' : ''}
        {fmt(val)} USDT
      </div>
    </div>
  )
}

// ─── main ────────────────────────────────────────────────────────────────────

const BotStats = ({ bot, tick }: { bot: Bot; tick: number }) => {
  const [trades, setTrades] = useState<Trade[]>([])
  const [currentPrice, setCurrentPrice] = useState<number>(0)
  const [loading, setLoading] = useState(true)

  const symbol = `${bot.base}${bot.quote ?? 'USDT'}`

  useEffect(() => {
    let cancelled = false

    Promise.all([getBotTrades(bot.id, 1, 500), getPrice(symbol)])
      .then(([tradesRes, price]) => {
        if (cancelled) return
        console.log('trades:', tradesRes)
        console.log('price:', price)
        setTrades(tradesRes.data)
        setCurrentPrice(price)
        setLoading(false)
      })
      .catch((err) => {
        console.error('BotStats fetch failed:', err)
        setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [bot.id, tick, symbol])

  if (loading)
    return <div className="text-content-secondary text-sm font-mono animate-pulse">Loading...</div>

  if (!trades.length) {
    return <div className="text-content-secondary text-sm font-mono">No trades yet.</div>
  }

  const s = deriveStats(trades, currentPrice)
  const quote = bot.quote ?? 'USDT'

  const unrealisedColor = s.unrealisedPnl > 0 ? 'green' : s.unrealisedPnl < 0 ? 'red' : 'neutral'
  const realisedColor = s.netPnl > 0 ? 'green' : s.netPnl < 0 ? 'red' : 'neutral'
  const chartColor = s.netPnl >= 0 ? '#4ade80' : '#f87171'

  if (!trades.length)
    return <div className="text-content-secondary text-sm font-mono">No trades yet.</div>

  return (
    <div className="flex flex-col gap-4 w-full select-none">
      {/* KPI cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-2">
        <Card
          label="Unrealised PnL"
          value={`${s.unrealisedPnl >= 0 ? '+' : ''}${fmt(s.unrealisedPnl)} ${quote}`}
          sub={fmtPct(s.unrealisedPct) + ' vs avg entry'}
          color={unrealisedColor}
        />
        <Card
          label="Net Realised PnL"
          value={`${s.netPnl >= 0 ? '+' : ''}${fmt(s.netPnl)} ${quote}`}
          sub={`Gross: ${fmt(s.realisedPnl)} — Fees: -${fmt(s.totalFees)}`}
          color={realisedColor}
        />
        <Card
          label="Position"
          value={`${fmt(s.currentQty, 6)} ${bot.base}`}
          sub={`Avg entry: ${fmt(s.avgEntry)} ${quote}`}
          color="neutral"
        />
        <Card
          label="Current Price"
          value={`${fmt(currentPrice)} ${quote}`}
          sub={`Invested: ${fmt(s.totalInvested)} ${quote}`}
          color="neutral"
        />
      </div>

      {/* Cumulative PnL chart */}
      {s.chartData.length > 1 && (
        <div className="bg-table-row border border-table-border rounded-md px-4 pt-3 pb-2">
          <span className="text-content-secondary text-xs uppercase tracking-wider font-semibold">
            Cumulative Net PnL
          </span>
          <ResponsiveContainer width="100%" height={180}>
            <LineChart data={s.chartData} margin={{ top: 12, right: 8, left: 0, bottom: 0 }}>
              <CartesianGrid
                strokeDasharray="3 3"
                stroke="rgba(255,255,255,0.04)"
                vertical={false}
              />
              <XAxis
                dataKey="time"
                tick={{ fill: 'var(--color-content-tertiary, #666)', fontSize: 10 }}
                tickLine={false}
                axisLine={false}
                interval="preserveStartEnd"
              />
              <YAxis
                tick={{ fill: 'var(--color-content-tertiary, #666)', fontSize: 10 }}
                tickLine={false}
                axisLine={false}
                tickFormatter={(v) => fmt(v)}
                width={64}
              />
              <ReferenceLine y={0} stroke="rgba(255,255,255,0.12)" strokeDasharray="4 4" />
              <Tooltip<ValueType, NameType>
                content={(props) => <ChartTooltip {...props} />}
                cursor={{ stroke: 'rgba(255,255,255,0.1)' }}
              />
              <Line
                type="monotone"
                dataKey="pnl"
                stroke={chartColor}
                strokeWidth={2}
                dot={false}
                activeDot={{ r: 4, fill: chartColor, strokeWidth: 0 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  )
}

export default BotStats
