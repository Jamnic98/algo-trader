import { useEffect, useState } from 'react'

import { useAlert } from 'hooks'

type LogEntry = {
  time: string
  level: string
  message: string
}

const BotLogs = ({ id }: { id: string }) => {
  const { showAlert } = useAlert()
  const [logs, setLogs] = useState<LogEntry[]>([])

  useEffect(() => {
    const es = new EventSource(
      `/api/bots/${id}/logs/stream?api_key=${import.meta.env.VITE_SERVER_API_KEY}`
    )

    es.onmessage = (e) => {
      const entry: LogEntry = JSON.parse(e.data)
      setLogs((prev) => [entry, ...prev].slice(0, 200))
    }

    es.onerror = () => {
      const errorMsg = 'Bot logs stream error'
      console.error(errorMsg)
      showAlert({ type: 'error', title: errorMsg })
      es.close()
    }
    return () => es.close()
  }, [id, showAlert])

  if (!logs.length) return <div className="text-gray-500">No logs yet.</div>

  return (
    <div className="font-mono text-xs space-y-1 overflow-y-auto max-h-96">
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
