import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer } from 'recharts'

const fmt = (b: number): string => {
  if (b >= 1_073_741_824) return `${(b / 1_073_741_824).toFixed(1)} GB`
  if (b >= 1_048_576) return `${(b / 1_048_576).toFixed(1)} MB`
  return `${(b / 1024).toFixed(1)} KB`
}

type PieCardProps = {
  rssBytes: number
  heapAllocMB: number
  totalBytes: number
  usedBytes: number
}

const COLORS = ['#10b981', '#60a5fa']

const PieCard = ({ rssBytes, heapAllocMB, totalBytes, usedBytes }: PieCardProps) => {
  const heapBytes = heapAllocMB * 1_048_576
  const rssOverhead = rssBytes - heapBytes

  const slices = [
    { label: 'Heap alloc', bytes: heapBytes, fill: COLORS[0] },
    { label: 'RSS overhead', bytes: rssOverhead, fill: COLORS[1] },
  ]

  const total = slices.reduce((a, s) => a + s.bytes, 0)
  const pct = (b: number) => `${((b / total) * 100).toFixed(1)}%`
  const usedPct = ((usedBytes / totalBytes) * 100).toFixed(1)

  return (
    <div className="space-y-4">
      {/* RSS pie */}
      <div className="flex flex-wrap gap-3 text-xs text-content-secondary">
        {slices.map((s) => (
          <span key={s.label} className="flex items-center gap-1.5">
            <span className="w-2.5 h-2.5 rounded-sm shrink-0" style={{ background: s.fill }} />
            {s.label} {pct(s.bytes)}
          </span>
        ))}
      </div>

      <ResponsiveContainer width="100%" height={200}>
        <PieChart style={{ outline: 'none' }}>
          <Pie
            data={slices}
            dataKey="bytes"
            nameKey="label"
            cx="50%"
            cy="50%"
            innerRadius="55%"
            outerRadius="75%"
            paddingAngle={2}
            style={{ outline: 'none' }}
          >
            {slices.map((s) => (
              <Cell key={s.label} fill={s.fill} style={{ outline: 'none' }} />
            ))}
          </Pie>
          <Tooltip
            formatter={(value, name) => [fmt(value as number), name]}
            contentStyle={{
              background: 'var(--color-surface-primary)',
              border: '1px solid var(--color-border)',
              borderRadius: 6,
              fontSize: 12,
            }}
          />
        </PieChart>
      </ResponsiveContainer>

      <div className="grid grid-cols-2 gap-2">
        {slices.map((s) => (
          <div key={s.label} className="bg-surface-secondary rounded-md p-3">
            <div className="text-xs font-mono text-content-secondary mb-1">{s.label}</div>
            <div className="text-base text-content-primary">{fmt(s.bytes)}</div>
          </div>
        ))}
      </div>

      {/* System RAM bar */}
      <div className="bg-surface-secondary rounded-md p-3 space-y-2">
        <div className="flex justify-between text-xs text-content-secondary font-mono">
          <span>System RAM</span>
          <span>
            {usedPct}% · {fmt(usedBytes)} / {fmt(totalBytes)}
          </span>
        </div>
        <div className="h-1.5 rounded-full bg-surface-tertiary overflow-hidden">
          <div
            className="h-full rounded-full bg-blue-400 transition-all duration-500"
            style={{ width: `${usedPct}%` }}
          />
        </div>
      </div>
    </div>
  )
}

export default PieCard
