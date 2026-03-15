import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { attachBot, deleteBot, detachBot, getBot, startBot, stopBot } from 'api'
import { BarLoader, BotActionButtons, BotTabs, Heading } from 'components'
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

  const handleAttachBot = async (botId: string) => {
    try {
      setLoadingAction('attach')
      const updatedBot = await attachBot(botId)
      setBot(updatedBot)
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to attach bot with id: ${botId}`
      showAlert({ title: errorMsg, type: 'error' })
    } finally {
      setLoadingAction(null)
    }
  }

  const handleDetachBot = async (botId: string) => {
    try {
      setLoadingAction('detach')
      const updatedBot = await detachBot(botId)
      setBot(updatedBot)
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to detach bot with id: ${botId}`
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
    <div className="space-y-4">
      <Heading title="Bot Overview" />
      {bot && (
        <>
          <div className="max-w-80">
            <div className="flex flex-row justify-between gap-4 select-none">
              <div className="space-y-1.5 text-sm text-content-secondary font-mono">
                <p>{`${bot.base}/${bot.quote}`}</p>
                <div className="flex items-center gap-2">
                  <span>Status:</span>
                  <span
                    className={`text-xs px-2 py-0.5 rounded-full ${
                      bot.status === 'running'
                        ? 'bg-green-500/10 text-green-400'
                        : bot.status === 'attached'
                          ? 'bg-yellow-500/10 text-yellow-400'
                          : 'bg-surface-secondary text-content-tertiary'
                    }`}
                  >
                    {bot.status}
                  </span>
                </div>
              </div>

              <div className="flex gap-2">
                <BotActionButtons
                  botId={bot.id}
                  botStatus={bot.status}
                  loadingAction={loadingAction}
                  startBot={handleStartBot}
                  stopBot={handleStopBot}
                  attachBot={handleAttachBot}
                  detachBot={handleDetachBot}
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
