import { useEffect, useState } from 'react'

import { getAllBots } from 'api'
import { BotTable, Heading } from 'components'
import { useAlert } from 'hooks'
import type { BotData } from 'types'

const Dashboard = () => {
  const { showAlert } = useAlert()

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
    <div className="space-y-8">
      <Heading title="Dashboard" />

      <div className="space-y-3">
        <Heading title="Running Bots" as={6} size={3} />
        {bots.length > 0 ? (
          <div className="overflow-x-auto [&::-webkit-scrollbar]:h-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-300 [&::-webkit-scrollbar-thumb]:rounded-full hover:[&::-webkit-scrollbar-thumb]:bg-gray-400">
            <BotTable
              bots={bots}
              columns={['symbol', 'quantity', 'interval', 'lookback', 'started']}
            />
          </div>
        ) : (
          <div className="text-gray-500 font-semibold text-center">- No running bots -</div>
        )}
      </div>
    </div>
  )
}

export default Dashboard
