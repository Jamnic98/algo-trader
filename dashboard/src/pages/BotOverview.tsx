import { useEffect, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'

import { attachBot, deleteBot, detachBot, getBot, startBot, stopBot } from 'api'
import { BarLoader, BotActionButtons, BotTrades, CandlestickChart, Heading, Tabs } from 'components'
import { useAlert } from 'hooks'
import type { BotData, Tab } from 'types'

const BotOverview = () => {
  const { id } = useParams()
  const { showAlert } = useAlert()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()

  const [bot, setBot] = useState<BotData | null>(null)
  const [runningFor, setRunningFor] = useState<string>('-')

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // TODO: move to separate component
  const tabs: Tab[] | null = bot && [
    {
      label: 'Stats',
      content: (
        <div className="space-y-3 text-gray-500">
          <p>No performance data yet.</p>
        </div>
      ),
    },
    {
      label: 'Trades',
      content: (
        <div className="space-y-3 text-gray-500">
          <BotTrades id={bot!.id} />
        </div>
      ),
    },
    {
      label: 'Candles',
      content: (
        <>
          {bot.candles ? (
            <CandlestickChart data={bot.candles} symbol="BTC/USDT" height={400} />
          ) : (
            <div className="space-y-3 text-gray-500">
              <p>No candles yet.</p>
            </div>
          )}
        </>
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

  const tabLabels = tabs ? tabs.map((t) => t.label.toLowerCase()) : []

  const activeTab = (() => {
    const tabParam = searchParams.get('tab')
    const index = tabLabels.indexOf(tabParam ?? '')
    return index >= 0 ? index : 0
  })()

  const handleTabChange = (index: number) => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        next.set('tab', tabLabels[index])
        return next
      },
      { replace: false }
    )
  }

  useEffect(() => {
    if (!bot?.started) return

    const calc = () => {
      const seconds = Math.floor((Date.now() - Date.parse(bot.started!)) / 1000)
      const y = Math.floor(seconds / 31536000)
      const mo = Math.floor((seconds % 31536000) / 2592000)
      const d = Math.floor((seconds % 2592000) / 86400)
      const h = Math.floor((seconds % 86400) / 3600)
      const m = Math.floor((seconds % 3600) / 60)
      const s = seconds % 60
      return [y && `${y}y`, mo && `${mo}mo`, d && `${d}d`, h && `${h}h`, m && `${m}m`, `${s}s`]
        .filter(Boolean)
        .join(' ')
    }

    setRunningFor(calc())
    const interval = setInterval(() => setRunningFor(calc()), 1000)
    return () => clearInterval(interval)
  }, [bot?.started])

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

  if (loading) return <BarLoader fullscreen />
  if (error) return <div>{error} 😢</div>

  return (
    <div className="space-y-4">
      <Heading title="Bot Overview" />
      {bot && (
        <>
          <div className="flex flex-row justify-between gap-4">
            <div className="space-y-1 text-sm text-content-secondary font-mono">
              <p>Status: {bot.status}</p>
              <p>
                Started:{' '}
                {bot?.started
                  ? new Date(Date.parse(bot.started)).toLocaleString('en-GB', { timeZone: 'UTC' })
                  : '-'}
              </p>
              {bot.started ? <p>Running for: {runningFor}</p> : null}
            </div>

            <div className="flex gap-2">
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
          </div>
          <div className="space-y-1 text-sm text-content-secondary font-mono">
            <p>Symbol: {bot.symbol}</p>
            <p>Interval: {bot.interval}</p>
            <p>Lookback: {bot.lookback}</p>
            <p>Quantity: {bot.quantity}</p>
          </div>
        </>
      )}

      {tabs && <Tabs tabs={tabs} activeTab={activeTab} onTabChange={handleTabChange} />}
    </div>
  )
}

export default BotOverview
