export type DiagnosticData = {
  collected_at: string
  process_uptime_secs: number
  stale: boolean
  cpu: {
    used_percent: number
    num_cpu: number
  }
  memory: {
    total_bytes: number
    used_bytes: number
    free_bytes: number
    used_percent: number
  }
  process: {
    pid: number
    rss_bytes: number
    threads: number
  }
  go_runtime: {
    goroutines: number
    heap_alloc_mb: number
    gc_cycles: number
  }
}
