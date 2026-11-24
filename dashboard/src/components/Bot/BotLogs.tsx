import { useEffect, useRef, useState } from 'react'

import { BarLoader } from 'components'
import { useAlert } from 'hooks'

type LogEntry = {
  time: string
  level: string
  message: string
}

const BotLogs = ({ id }: { id: string }) => {
  const { showAlert } = useAlert()
  const showAlertRef = useRef(showAlert)
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const es = new EventSource(
      `/api/bots/${id}/logs/stream?api_key=${encodeURIComponent(import.meta.env.VITE_SERVER_API_KEY)}`
    )

    // connection established, stop loading
    es.onopen = () => {
      setLoading(false)
    }

    es.onmessage = (e) => {
      setLoading(false)
      const entry: LogEntry = JSON.parse(e.data)
      setLogs((prev) => [entry, ...prev].slice(0, 200))
    }

    es.onerror = () => {
      setLoading(false)
      const errorMsg = 'Bot logs stream error'
      console.error(errorMsg)
      showAlertRef.current({ type: 'error', title: errorMsg })
      es.close()
    }

    return () => es.close()
  }, [id])

  if (loading) return <BarLoader />

  if (!logs.length)
    return <div className="text-content-secondary text-sm font-mono">No logs yet.</div>

  return (
    <div className="font-mono text-xs space-y-1 overflow-y-auto max-h-128 pr-2 overflow-x-auto [&::-webkit-scrollbar]:w-1 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-300 [&::-webkit-scrollbar-thumb]:rounded-full hover:[&::-webkit-scrollbar-thumb]:bg-gray-400">
      {logs.map((log, i) => (
        <div key={i} className="flex gap-3">
          <span className="text-content-secondary/50 shrink-0">
            {new Date(log.time).toLocaleTimeString('en-GB', { hour12: false })}
          </span>
          <span className={log.level === 'ERROR' ? 'text-red-400' : 'text-content-secondary'}>
            {log.message}
          </span>
        </div>
      ))}
    </div>
  )
}

export default BotLogs
