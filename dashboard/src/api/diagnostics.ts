import { api } from 'api'
import type { DiagnosticData } from 'types'

export const getDiagnostics = async (): Promise<DiagnosticData> =>
  await api.fetchJson<DiagnosticData>('/diagnostics')

/* {
    "collected_at": "2026-03-03T12:58:30.247726386Z",
    "process_uptime_secs": 1186.003893417,
    "stale": false,
    "cpu": {
        "used_percent": 7.59748427672956,
        "num_cpu": 20
    },
    "memory": {
        "total_bytes": 29196644352,
        "used_bytes": 7144996864,
        "free_bytes": 17168080896,
        "used_percent": 24.47197964895771
    },
    "process": {
        "pid": 10082,
        "rss_bytes": 28323840,
        "threads": 14
    },
    "go_runtime": {
        "goroutines": 14,
        "heap_alloc_mb": 1.8010482788085938,
        "gc_cycles": 13
    }
} */
