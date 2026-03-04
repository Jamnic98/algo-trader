import { useEffect, useState, useRef } from 'react'

import { ChartCard, Heading, SectionLabel, StatCard } from 'components'
import type { ConnStatus, DiagnosticData, HistoryPoint } from 'types'
import { useAlert } from 'hooks'

const MAX_HISTORY = 30
const SESSION_KEY = 'diagnostics_history'

const apiKey = import.meta.env.VITE_SERVER_API_KEY
const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
const wsUrl = `${wsProtocol}//${window.location.host}/api/diagnostics/ws?api_key=${apiKey}`

const toHHMMSS = (secs: number): string => {
  const h = Math.floor(secs / 3600)
  const m = Math.floor((secs % 3600) / 60)
  const s = Math.floor(secs % 60)
  return [h, m, s].map((n) => String(n).padStart(2, '0')).join(':')
}

const formatBytes = (bytes: number): string => {
  if (bytes >= 1_073_741_824) return `${(bytes / 1_073_741_824).toFixed(1)} GB`
  if (bytes >= 1_048_576) return `${(bytes / 1_048_576).toFixed(1)} MB`
  return `${(bytes / 1024).toFixed(1)} KB`
}

const nowLabel = (): string => {
  return new Date().toLocaleTimeString('en-GB', { hour12: false })
}

const Diagnostics = () => {
  const { showAlert } = useAlert()
  const [diagnostics, setDiagnostics] = useState<DiagnosticData>()
  const [connStatus, setConnStatus] = useState<ConnStatus>('connecting')
  const stored = sessionStorage.getItem(SESSION_KEY)
  const initial: HistoryPoint[] = stored ? JSON.parse(stored) : []
  const historyRef = useRef<HistoryPoint[]>(initial)
  const [history, setHistory] = useState<HistoryPoint[]>(initial)

  useEffect(() => {
    const ws = new WebSocket(wsUrl)
    ws.onopen = () => setConnStatus('live')
    ws.onmessage = (event) => {
      const data: DiagnosticData = JSON.parse(event.data)
      setDiagnostics(data)
      const point: HistoryPoint = {
        t: nowLabel(),
        cpu: data.cpu.used_percent,
        mem: (data.process.rss_bytes / (data.process.rss_bytes + data.memory.free_bytes)) * 100,
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
            label="Process Mem Usage"
            dataKey="mem"
            history={history}
            color="#60a5fa"
            unit="%"
            latest={(process.rss_bytes / (process.rss_bytes + memory.free_bytes)) * 100}
          />
        </div>
      </div>

      <div>
        <SectionLabel>CPU &amp; Memory</SectionLabel>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-3">
          <StatCard label="CPU Usage" value={`${cpu.used_percent.toFixed(1)}%`} accent />
          <StatCard label="CPU Cores" value={cpu.num_cpu} />
          <StatCard
            label="Memory Free"
            value={formatBytes(memory.free_bytes)}
            sub={`of ${formatBytes(memory.total_bytes)}`}
          />
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
