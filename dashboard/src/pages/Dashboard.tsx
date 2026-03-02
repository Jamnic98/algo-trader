import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { getAllBots } from 'api'
import { BotTable, PageTitle } from 'components'
import type { BotData } from 'types'
import { useAlert } from 'hooks'

const Dashboard = () => {
  const { showAlert } = useAlert()
  const navigate = useNavigate()

  const [bots, setBots] = useState<BotData[]>([])
  const [loading, setLoading] = useState(true)

  // Fetch all bots on mount
  useEffect(() => {
    const fetchBots = async () => {
      try {
        const bots = await getAllBots()
        setBots(bots.filter((bot) => bot.status === 'running'))
      } catch {
        showAlert({
          title: 'Failed to load bots',
          type: 'error',
        })
      } finally {
        setLoading(false)
      }
    }
    fetchBots()
  }, [showAlert])

  if (loading) return <div>Loading bots...</div>

  return (
    <>
      <PageTitle title="Dashboard" />
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
    </>
  )
}

export default Dashboard
