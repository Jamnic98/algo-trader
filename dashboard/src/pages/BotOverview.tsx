import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { attachBot, deleteBot, detachBot, getBot, startBot, stopBot } from 'api'
import { BotActionButtons, Heading, Tabs } from 'components'
import { useAlert } from 'hooks'
import type { BotData, Tab } from 'types'

const BotOverview = () => {
  const { id } = useParams()
  const { showAlert } = useAlert()
  const navigate = useNavigate()

  const [bot, setBot] = useState<BotData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const tabs: Tab[] | null = bot && [
    {
      label: 'Info',
      content: (
        <div className="space-y-3 text-gray-500">
          <p>ID: {bot!.id}</p>
          <p>Symbol: {bot!.symbol}</p>
          <p>Interval: {bot!.interval}</p>
          <p>Lookback: {bot!.lookback}</p>
          <p>Quantity: {bot!.quantity}</p>
          <p>Status: {bot!.status}</p>
          {bot?.started && <p>Started: {bot!.started}</p>}
        </div>
      ),
    },
    {
      label: 'Trades',
      content: (
        <div className="space-y-3 text-gray-500">
          <p>No trades yet.</p>
        </div>
      ),
    },
    {
      label: 'Stats',
      content: (
        <div className="space-y-3 text-gray-500">
          <p>No performance data yet.</p>
        </div>
      ),
    },
    {
      label: 'Logs',
      content: (
        <div className="space-y-3 text-gray-500">
          <p>No logs yet.</p>
        </div>
      ),
    },
  ]

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

  const handleStartBot = async (botId: string) => {
    try {
      const updatedBot = await startBot(botId)
      setBot(updatedBot)
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to start bot with id: ${botId}`
      setError(errorMsg)
      showAlert({ title: errorMsg, type: 'error' })
    }
  }

  const handleStopBot = async (botId: string) => {
    try {
      if (confirm(`Stop bot ${botId}?`) === true) {
        const updatedBot = await stopBot(botId)
        setBot(updatedBot)
      }
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to stop bot with id: ${botId}`
      setError(errorMsg)
      showAlert({ title: errorMsg, type: 'error' })
    }
  }

  const handleAttachBot = async (botId: string) => {
    try {
      const updatedBot = await attachBot(botId)
      setBot(updatedBot)
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to attach bot with id: ${botId}`
      setError(errorMsg)
      showAlert({ title: errorMsg, type: 'error' })
    }
  }

  const handleDetachBot = async (botId: string) => {
    try {
      const updatedBot = await detachBot(botId)
      setBot(updatedBot)
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to detach bot with id: ${botId}`
      setError(errorMsg)
      showAlert({ title: errorMsg, type: 'error' })
    }
  }

  const handleDeleteBot = async (botId: string) => {
    try {
      if (confirm(`Delete bot ${botId}?`) === true) {
        await deleteBot(botId)
        navigate('/bots')
      }
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to delete bot with id: ${botId}`
      setError(errorMsg)
      showAlert({ title: errorMsg, type: 'error' })
    }
  }

  if (loading) return <div>Loading...</div>
  if (error) return <div>{error} 😢</div>

  return (
    <div className="space-y-4">
      <Heading title="Bot Overview" />
      {bot && (
        <div className="flex gap-4">
          <BotActionButtons
            botId={bot.id}
            botStatus={bot.status}
            startBot={handleStartBot}
            stopBot={handleStopBot}
            attachBot={handleAttachBot}
            detachBot={handleDetachBot}
            deleteBot={handleDeleteBot}
          />
        </div>
      )}
      {tabs && <Tabs tabs={tabs} />}
    </div>
  )
}

export default BotOverview
