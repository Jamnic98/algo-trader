import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { BotTable } from 'components'
import { getAllBots } from 'api'
import type { BotData } from 'types'

const Dashboard = () => {
  const navigate = useNavigate()

  const [bots, setBots] = useState<BotData[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Fetch all bots on mount
  useEffect(() => {
    const fetchBots = async () => {
      try {
        const bots = await getAllBots()
        setBots(bots.filter((bot) => bot.status === 'running'))
      } catch {
        setError('Failed to load bots')
      } finally {
        setLoading(false)
      }
    }

    fetchBots()
  }, [])

  if (loading) return <div>Loading bots...</div>
  if (error) return <div>{error}</div>

  return (
    <div>
      <h1>Dashboard</h1>
      <div className="flex flex-col">
        <div className="border-2">
          <h2>Running Bots</h2>
          <BotTable
            bots={bots}
            columns={['symbol', 'quantity', 'interval', 'lookback', 'started']}
          />
          <button onClick={() => navigate('bots')} className="cursor-pointer">
            all bots
          </button>
        </div>
      </div>

      <br />
      <br />
      <button onClick={() => navigate('trades')} className="cursor-pointer">
        trades page
      </button>
    </div>
  )
}

export default Dashboard
