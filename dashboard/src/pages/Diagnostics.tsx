import { useEffect, useState, useRef } from 'react'
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'

import { Heading } from 'components'
import type { DiagnosticData } from 'types'
import { useAlert } from 'hooks'

type HistoryPoint = { t: string; cpu: number; mem: number }

const MAX_HISTORY = 30
const SESSION_KEY = 'diagnostics_history'
const WS_URL = `ws://localhost:8080/api/diagnostics/ws?api_key=${import.meta.env.VITE_SERVER_API_KEY}`

function toHHMMSS(secs: number): string {
  const h = Math.floor(secs / 3600)
  const m = Math.floor((secs % 3600) / 60)
  const s = Math.floor(secs % 60)
  return [h, m, s].map((n) => String(n).padStart(2, '0')).join(':')
}

function formatBytes(bytes: number): string {
  if (bytes >= 1_073_741_824) return `${(bytes / 1_073_741_824).toFixed(1)} GB`
  if (bytes >= 1_048_576) return `${(bytes / 1_048_576).toFixed(1)} MB`
  return `${(bytes / 1024).toFixed(1)} KB`
}

function nowLabel(): string {
  return new Date().toLocaleTimeString('en-GB', { hour12: false })
}

type StatCardProps = { label: string; value: string | number; sub?: string; accent?: boolean }
const StatCard = ({ label, value, sub, accent }: StatCardProps) => (
  <div className="bg-surface-secondary border border-table-border rounded-lg px-5 py-4 flex flex-col gap-1">
    <span className="text-content-secondary text-[11px] font-mono uppercase tracking-widest">
      {label}
    </span>
    <span
      className={`text-[22px] font-semibold tracking-tight ${accent ? 'text-accent' : 'text-content-primary'}`}
    >
      {value}
    </span>
    {sub && <span className="text-content-secondary text-xs">{sub}</span>}
  </div>
)

type SectionLabelProps = { children: React.ReactNode }
const SectionLabel = ({ children }: SectionLabelProps) => (
  <p className="text-content-secondary text-[11px] font-mono uppercase tracking-widest mb-2.5 mt-0">
    {children}
  </p>
)

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
          tick={{ fontSize: 10, fill: 'var(--color-content-secondary)', fontFamily: 'monospace' }}
          tickLine={false}
          axisLine={false}
          interval={Math.floor(MAX_HISTORY / 8)}
        />
        <YAxis
          domain={[0, 100]}
          tick={{ fontSize: 10, fill: 'var(--color-content-secondary)', fontFamily: 'monospace' }}
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

type ConnStatus = 'connecting' | 'live' | 'disconnected'

const Diagnostics = () => {
  const { showAlert } = useAlert()
  const [diagnostics, setDiagnostics] = useState<DiagnosticData>()
  const [connStatus, setConnStatus] = useState<ConnStatus>('connecting')
  const stored = sessionStorage.getItem(SESSION_KEY)
  const initial: HistoryPoint[] = stored ? JSON.parse(stored) : []
  const historyRef = useRef<HistoryPoint[]>(initial)
  const [history, setHistory] = useState<HistoryPoint[]>(initial)

  useEffect(() => {
    const ws = new WebSocket(WS_URL)
    ws.onopen = () => setConnStatus('live')
    ws.onmessage = (event) => {
      const data: DiagnosticData = JSON.parse(event.data)
      setDiagnostics(data)
      const point: HistoryPoint = {
        t: nowLabel(),
        cpu: data.cpu.used_percent,
        mem: data.memory.used_percent,
      }
      const next = [...historyRef.current, point].slice(-MAX_HISTORY)
      historyRef.current = next
      sessionStorage.setItem(SESSION_KEY, JSON.stringify(next))
      setHistory(next)
    }
    ws.onerror = () => showAlert({ title: 'Diagnostics connection error', type: 'error' })
    ws.onclose = () => setConnStatus('disconnected')
    return () => ws.close()
  }, [showAlert])

  if (connStatus === 'connecting' && !diagnostics)
    return <div className="text-content-secondary p-8 font-mono text-[13px]">Connecting...</div>

  if (connStatus === 'disconnected' && !diagnostics)
    return (
      <div className="text-content-secondary p-8 font-mono text-[13px]">
        Failed to connect to diagnostics.
      </div>
    )

  if (!diagnostics) return null

  const { cpu, memory, process, go_runtime, process_uptime_secs, stale } = diagnostics
  const isStale = stale || connStatus === 'disconnected'

  return (
    <div className="flex flex-col gap-7">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <Heading title="Diagnostics" />
        <div
          className={`flex items-center gap-2 rounded-full px-3 py-1.5 text-xs font-mono border ${isStale ? 'bg-red-500/10 border-red-500/30 text-red-400' : 'bg-accent-muted border-accent/30 text-accent'}`}
        >
          <span
            className={`w-1.5 h-1.5 rounded-full ${isStale ? 'bg-red-400' : 'bg-accent shadow-[0_0_6px_var(--color-accent)]'}`}
          />
          {connStatus === 'disconnected' ? 'Disconnected' : isStale ? 'Stale' : 'Live'} · uptime{' '}
          {toHHMMSS(process_uptime_secs)}
        </div>
      </div>

      <div>
        <SectionLabel>Usage over time</SectionLabel>
        <div className="space-y-8">
          <ChartCard
            label="CPU Usage"
            dataKey="cpu"
            history={history}
            color="var(--color-accent)"
            unit="%"
            latest={cpu.used_percent}
          />
          <ChartCard
            label="Memory Usage"
            dataKey="mem"
            history={history}
            color="#60a5fa"
            unit="%"
            latest={memory.used_percent}
          />
        </div>
      </div>

      <div>
        <SectionLabel>CPU &amp; Memory</SectionLabel>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-3">
          <StatCard label="CPU Usage" value={`${cpu.used_percent.toFixed(1)}%`} accent />
          <StatCard label="CPU Cores" value={cpu.num_cpu} />
          <StatCard
            label="Mem Used"
            value={formatBytes(memory.used_bytes)}
            sub={`of ${formatBytes(memory.total_bytes)}`}
          />
          <StatCard label="Mem Free" value={formatBytes(memory.free_bytes)} />
        </div>
      </div>

      <div>
        <SectionLabel>Process</SectionLabel>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-3">
          <StatCard label="PID" value={process.pid} />
          <StatCard label="RSS Memory" value={formatBytes(process.rss_bytes)} />
          <StatCard label="Threads" value={process.threads} />
        </div>
      </div>

      <div>
        <SectionLabel>Go Runtime</SectionLabel>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-3">
          <StatCard label="Goroutines" value={go_runtime.goroutines} />
          <StatCard label="Heap Alloc" value={`${go_runtime.heap_alloc_mb.toFixed(2)} MB`} />
          <StatCard label="GC Cycles" value={go_runtime.gc_cycles} />
        </div>
      </div>
    </div>
  )
}

export default Diagnostics
