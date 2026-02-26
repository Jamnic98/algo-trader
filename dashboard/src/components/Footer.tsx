import { useEffect, useState } from 'react'

import { getHealthStatus } from 'api'

const StatusIndicator = ({ status, label }: { status: string; label: string }) => {
  return (
    <div className="flex flex-row justify-center items-center space-x-1">
      <span>{label}</span>
      <span
        className={`${status === 'ok' ? 'bg-green-500' : 'bg-red-500'} rounded-full w-3 h-3 flex justify-center items-center`}
      />
    </div>
  )
}

const Footer = () => {
  const [healthStatus, setHealthStatus] = useState('')

  useEffect(() => {
    const fetchHealthStatus = async () => {
      try {
        const status = await getHealthStatus()
        setHealthStatus(status)
      } catch {
        setHealthStatus('error')
      }
    }

    // run immediately
    fetchHealthStatus()

    // 1min interval to check API server health
    const intervalId = setInterval(fetchHealthStatus, 60 * 1000)

    return () => {
      clearInterval(intervalId)
    }
  }, [])

  return (
    <footer
      id="footer"
      className="select-none flex items-center h-20 border-t-2"
      aria-label="footer"
    >
      <StatusIndicator label="api status" status={healthStatus} />
    </footer>
  )
}

export default Footer
