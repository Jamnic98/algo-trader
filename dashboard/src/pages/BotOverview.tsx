import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { getBot, startBot, stopBot, deleteBot } from 'api'
import { BarLoader, BotActionButtons, BotStatusPill, BotTabs, Heading } from 'components'
import { useAlert } from 'hooks'
import type { Bot, LoadingAction } from 'types'

const BotOverview = () => {
  const { id } = useParams()
  const { showAlert } = useAlert()
  const navigate = useNavigate()

  const [bot, setBot] = useState<Bot | null>(null)
  const [loadingAction, setLoadingAction] = useState<LoadingAction>(null)
  const [positionsTick, setPositionsTick] = useState(0)
  const onNewTrade = () => setPositionsTick((t) => t + 1)

  const [loading, setLoading] = useState(true)

  // fetch bot data on page load
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
        showAlert({ title: errorMsg, type: 'error' })
      } finally {
        setLoading(false)
      }
    }

    fetchBot()
  }, [id, showAlert])

  const handleStartBot = async (botId: string) => {
    try {
      setLoadingAction('start')
      const updatedBot = await startBot(botId)
      setBot(updatedBot)
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to start bot with id: ${botId}`
      showAlert({ title: errorMsg, type: 'error' })
    } finally {
      setLoadingAction(null)
    }
  }

  const handleStopBot = async (botId: string) => {
    try {
      if (confirm(`Stop bot ${botId}?`) === true) {
        setLoadingAction('stop')
        const updatedBot = await stopBot(botId)
        setBot(updatedBot)
      }
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to stop bot with id: ${botId}`
      showAlert({ title: errorMsg, type: 'error' })
    } finally {
      setLoadingAction(null)
    }
  }

  const handleDeleteBot = async (botId: string) => {
    try {
      if (confirm(`Delete bot ${botId}?`) === true) {
        setLoadingAction('delete')
        await deleteBot(botId)
        navigate('/bots')
      }
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to delete bot with id: ${botId}`
      showAlert({ title: errorMsg, type: 'error' })
    } finally {
      setLoadingAction(null)
    }
  }

  if (loading) return <BarLoader fullscreen />

  return (
    <div className="space-y-8">
      <Heading title="Bot Overview" />
      {bot && (
        <>
          <div className="w-fit">
            <div className="flex flex-row justify-between items-center gap-4 select-none mb-8 pr-4">
              <div className="space-y-1.5 text-sm text-content-secondary font-mono">
                <p>{`${bot.base}/${bot.quote}`}</p>
                <div className="flex items-center gap-2">
                  <BotStatusPill status={bot.status} />
                </div>
              </div>
              <div className="flex gap-2">
                <BotActionButtons
                  botId={bot.id}
                  botStatus={bot.status}
                  loadingAction={loadingAction}
                  startBot={handleStartBot}
                  stopBot={handleStopBot}
                  deleteBot={handleDeleteBot}
                />
              </div>
            </div>
          </div>
          <BotTabs bot={bot} onNewTrade={onNewTrade} positionsTick={positionsTick} />
        </>
      )}
    </div>
  )
}

export default BotOverview
