import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'

import { getBot } from 'api'
import { useAlert } from 'hooks'
import { type BotData } from 'types'
import { Heading } from 'components'

const BotInfo = () => {
  const { id } = useParams()
  const { showAlert } = useAlert()

  const [bot, setBot] = useState<BotData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchBot = async () => {
      try {
        if (!id) throw new Error('No ID provided')
        const botData = await getBot(id)
        if (!botData) throw new Error('Bot not found')
        setBot(botData)
      } catch (err) {
        console.error(err)
        const errorMsg = `Failed to load bot with id: ${id}`
        setError(errorMsg)
        showAlert({
          title: errorMsg,
          type: 'error',
        })
      } finally {
        setLoading(false)
      }
    }

    fetchBot()
  }, [id, showAlert])

  if (loading) return <div>Loading...</div>
  if (error) return <div>{error} 😢</div>

  return (
    <div className="space-y-8">
      <Heading title="Bot Info" />

      <div className="space-y-3 text-gray-500">
        <p>ID: {bot!.id}</p>
        <p>Symbol: {bot!.symbol}</p>
        <p>Interval: {bot!.interval}</p>
        <p>Lookback: {bot!.lookback}</p>
        <p>Quantity: {bot!.quantity}</p>
        <p>Status: {bot!.status}</p>
        {bot?.started && <p>Started: {bot!.started}</p>}
      </div>
    </div>
  )
}

export default BotInfo
