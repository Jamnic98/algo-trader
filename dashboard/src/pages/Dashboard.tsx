import { useEffect, useState } from 'react'

import { getAllBots } from 'api'
import { BarLoader, BotTable, Heading } from 'components'
import { useAlert } from 'hooks'
import type { Bot } from 'types'

const Dashboard = () => {
  const { showAlert } = useAlert()

  const [bots, setBots] = useState<Bot[]>([])
  const [loading, setLoading] = useState(true)

  // Fetch all bots on mount
  useEffect(() => {
    const fetchBots = async () => {
      try {
        const bots = await getAllBots()
        setBots(bots.filter((bot) => bot.status === 'trading'))
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

  if (loading) return <BarLoader fullscreen />

  return (
    <div className="space-y-8">
      <Heading title="Dashboard" />

      <div className="space-y-3">
        <Heading title="Running Bots" as={6} size={3} />
        {bots.length > 0 ? (
          <div className="overflow-x-auto [&::-webkit-scrollbar]:h-1 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-300 [&::-webkit-scrollbar-thumb]:rounded-full hover:[&::-webkit-scrollbar-thumb]:bg-gray-400">
            <BotTable bots={bots} columns={['mode', 'symbol', 'quantity', 'interval', 'started']} />
          </div>
        ) : (
          <div className="text-gray-500">No bots trading.</div>
        )}
      </div>
    </div>
  )
}

export default Dashboard
