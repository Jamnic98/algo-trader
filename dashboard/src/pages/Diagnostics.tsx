import { useEffect, useRef, useState } from 'react'

import { BarLoader, ChartCard, Heading, PieCard, SectionLabel, StatCard } from 'components'
import type { ConnStatus, DiagnosticData, HistoryPoint } from 'types'
import { useAlert } from 'hooks'

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

const Diagnostics = () => {
  const { showAlert } = useAlert()
  const [diagnostics, setDiagnostics] = useState<DiagnosticData>()
  const [connStatus, setConnStatus] = useState<ConnStatus>('connecting')
  const [history, setHistory] = useState<HistoryPoint[]>([])
  const [uptime, setUptime] = useState<string | null>(null)

  const serverUptimeBaseRef = useRef<number | null>(null)
  const clientStartRef = useRef<number | null>(null)

  useEffect(() => {
    const interval = setInterval(() => {
      if (serverUptimeBaseRef.current === null || clientStartRef.current === null) return
      const elapsed = Math.floor((performance.now() - clientStartRef.current) / 1000)
      setUptime(toHHMMSS(serverUptimeBaseRef.current + elapsed))
    }, 1000)
    return () => clearInterval(interval)
  }, [])

  useEffect(() => {
    const source = new EventSource(
      `/api/diagnostics/stream?api_key=${encodeURIComponent(import.meta.env.VITE_SERVER_API_KEY)}`
    )

    source.onmessage = (e) => {
      const { stats, history } = JSON.parse(e.data)
      setDiagnostics(stats)
      setHistory(history)
      setConnStatus('live')
      if (serverUptimeBaseRef.current === null) {
        serverUptimeBaseRef.current = stats.process_uptime_secs
        clientStartRef.current = performance.now()
      }
    }

    source.onerror = () => {
      setConnStatus('disconnected')
      const errorMsg = 'Diagnostics connection error'
      console.error(errorMsg)
      showAlert({ type: 'error', title: errorMsg })
      source.close()
    }

    return () => source.close()
  }, [showAlert])

  if (!diagnostics) return null
  const { cpu, memory, process, go_runtime, stale } = diagnostics
  const available = memory.total_bytes - memory.used_bytes
  const isStale = stale || connStatus === 'disconnected'

  if (connStatus === 'disconnected' && !diagnostics)
    return (
      <span className="inline-flex items-center w-16 h-[1em]">
        {uptime ? uptime : <BarLoader width="narrow" />}
      </span>
    )

  return (
    <div className="flex flex-col gap-10">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <Heading title="Diagnostics" />
        <div
          className={`flex items-center gap-2 rounded-full px-3 py-1.5 text-xs font-mono border ${
            isStale
              ? 'bg-red-500/10 border-red-500/30 text-red-400'
              : 'bg-accent-muted border-accent/30 text-accent'
          }`}
        >
          <span
            className={`w-1.5 h-1.5 rounded-full ${
              isStale ? 'bg-red-400' : 'bg-accent shadow-[0_0_6px_var(--color-accent)]'
            }`}
          />
          <span className="inline-flex items-center gap-2 text-nowrap">
            <span>
              {connStatus === 'disconnected' ? 'Disconnected' : isStale ? 'Stale' : 'Live'} ·
              uptime{' '}
            </span>
            <span className="inline-block w-[8ch] text-left">
              {uptime ? uptime : <BarLoader width="uptime" />}
            </span>
          </span>
        </div>
      </div>

      {/* Usage over time */}
      <div className="flex flex-col gap-4">
        <SectionLabel>Usage over time</SectionLabel>
        <div className="flex flex-col gap-6">
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
            latest={(process.rss_bytes / available) * 100}
          />
        </div>
      </div>

      {/* CPU & Memory */}
      <div className="flex flex-col gap-4">
        <SectionLabel>CPU &amp; Memory</SectionLabel>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-3">
          <StatCard label="CPU Usage" value={`${cpu.used_percent.toFixed(1)}%`} accent />
          <StatCard label="CPU Cores" value={cpu.num_cpu} />
        </div>
        <PieCard
          rssBytes={process.rss_bytes}
          heapAllocMB={go_runtime.heap_alloc_mb}
          totalBytes={memory.total_bytes}
          usedBytes={memory.used_bytes}
        />
      </div>

      {/* Process */}
      <div className="flex flex-col gap-4">
        <SectionLabel>Process</SectionLabel>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-3">
          <StatCard label="PID" value={process.pid} />
          <StatCard label="RSS Memory" value={formatBytes(process.rss_bytes)} />
          <StatCard label="Threads" value={process.threads} />
        </div>
      </div>

      {/* Go Runtime */}
      <div className="flex flex-col gap-4">
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
