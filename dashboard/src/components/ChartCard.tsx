import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'

import type { HistoryPoint } from 'types'

type ChartCardProps = {
  label: string
  dataKey: 'cpu' | 'mem'
  history: HistoryPoint[]
  color: string
  unit: string
  latest: number
}
const ChartCard = ({ label, dataKey, history, color, unit, latest }: ChartCardProps) => (
  <div className="bg-surface-secondary border border-table-border rounded-lg px-5 py-4 flex-1 min-w-0">
    <div className="flex justify-between items-start mb-4">
      <span className="text-content-secondary text-[11px] font-mono uppercase tracking-widest">
        {label}
      </span>
      <span className="text-[20px] font-semibold tracking-tight" style={{ color }}>
        {latest.toFixed(1)}
        {unit}
      </span>
    </div>
    <ResponsiveContainer width="100%" height={120}>
      <AreaChart data={history} margin={{ top: 2, right: 0, left: -30, bottom: 0 }}>
        <defs>
          <linearGradient id={`grad-${dataKey}`} x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor={color} stopOpacity={0.25} />
            <stop offset="95%" stopColor={color} stopOpacity={0} />
          </linearGradient>
        </defs>
        <XAxis
          dataKey="t"
          tick={{ fontSize: 8, fill: 'var(--color-content-secondary)', fontFamily: 'monospace' }}
          tickLine={true}
          axisLine={false}
          interval={10}
        />
        <YAxis
          domain={[0, 100]}
          tick={{ fontSize: 8, fill: 'var(--color-content-secondary)', fontFamily: 'monospace' }}
          tickLine={false}
          axisLine={false}
        />
        <Tooltip
          contentStyle={{
            background: 'var(--color-surface-primary)',
            border: '1px solid var(--color-table-border)',
            borderRadius: 6,
            fontSize: 12,
            fontFamily: 'monospace',
            color: 'var(--color-content-primary)',
          }}
          formatter={(v: number | undefined) =>
            v !== undefined ? [`${v.toFixed(1)}${unit}`, label] : ['N/A', label]
          }
          labelStyle={{ color: 'var(--color-content-secondary)' }}
        />
        <Area
          type="monotone"
          dataKey={dataKey}
          stroke={color}
          strokeWidth={1.5}
          fill={`url(#grad-${dataKey})`}
          dot={false}
          isAnimationActive={false}
        />
      </AreaChart>
    </ResponsiveContainer>
  </div>
)

export default ChartCard
