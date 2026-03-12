import { useEffect, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { ArrowLeftRight, BarChart2, Copy, CandlestickChart as CChart, Info } from 'lucide-react'

import { attachBot, deleteBot, detachBot, getBot, startBot, stopBot } from 'api'
import { BarLoader, BotActionButtons, BotTrades, CandlestickChart, Heading, Tabs } from 'components'
import { useAlert } from 'hooks'
import type { Bot, LoadingAction, Tab } from 'types'

const makeBotRunDurationStr = (botStart: string): string => {
  const seconds = Math.floor((Date.now() - Date.parse(botStart)) / 1000)
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

const BotOverview = () => {
  const { id } = useParams()
  const { showAlert } = useAlert()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()

  const [bot, setBot] = useState<Bot | null>(null)
  const [runningFor, setRunningFor] = useState<string>('-')
  const [loadingAction, setLoadingAction] = useState<LoadingAction>(null)

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const tabs: Tab[] | null = bot && [
    {
      label: 'Info',
      icon: <Info size={13} />,
      content: (
        <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1.5 text-sm font-mono select-none">
          <span className="text-content-tertiary">Id</span>
          <span className="flex items-center gap-2 text-content-secondary">
            <span className="truncate">{bot.id}</span>
            <Copy
              size={12}
              className="cursor-pointer shrink-0 hover:text-content-primary transition-colors"
              onClick={() => {
                navigator.clipboard.writeText(bot.id)
                showAlert({ type: 'info', message: 'Bot id copied' })
              }}
            />
          </span>

          <span className="text-content-tertiary">Mode</span>
          <span className="text-content-secondary">{bot.mode}</span>

          <span className="text-content-tertiary">Interval</span>
          <span className="text-content-secondary">{bot.interval}</span>

          <span className="text-content-tertiary">Lookback</span>
          <span className="text-content-secondary">{bot.lookback}</span>

          {bot.quantity && (
            <>
              <span className="text-content-tertiary">Quantity</span>
              <span className="text-content-secondary">{bot.quantity}</span>
            </>
          )}

          {bot.candles && (
            <>
              <span className="text-content-tertiary">Candles</span>
              <span className="text-content-secondary">{bot.candles.length}</span>
            </>
          )}

          <span className="text-content-tertiary">Started</span>
          <span className="text-content-secondary">
            {bot.started
              ? new Date(Date.parse(bot.started)).toLocaleString('en-GB', { timeZone: 'UTC' })
              : '-'}
          </span>

          <span className="text-content-tertiary">Running for</span>
          <span className="text-content-secondary">{bot.started ? runningFor : '-'}</span>
        </div>
      ),
    },
    {
      label: 'Stats',
      icon: <BarChart2 size={13} />,
      content: (
        <div className="space-y-3 text-gray-500">
          <p>No stats yet.</p>
        </div>
      ),
    },
    {
      label: 'Trades',
      icon: <ArrowLeftRight size={13} />,
      content: (
        <div className="space-y-3 text-gray-500">
          <BotTrades id={bot!.id} />
        </div>
      ),
    },
    {
      label: 'Candles',
      icon: <CChart size={13} />,
      content: (
        <>
          {bot.candles ? (
            <CandlestickChart data={bot.candles} symbol={`${bot.base}/${bot.quote}`} height={400} />
          ) : (
            <div className="space-y-3 text-gray-500">
              <p>No candles yet.</p>
            </div>
          )}
        </>
      ),
    },
    // {
    //   label: 'Config',
    //   icon: <Settings size={13} />,
    //   content: (
    //     <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-3 text-sm font-mono select-none max-w-64">
    //       {/* Mode */}
    //       <span className="text-content-tertiary self-center">Mode</span>
    //       <div className="flex items-center py-1">
    //         <div
    //           onClick={() => handleUpdateBot({ mode: bot.mode === 'live' ? 'paper' : 'live' })}
    //           className={`relative w-11 h-6 rounded-full cursor-pointer transition-colors duration-200 ${
    //             bot.mode === 'live' ? 'bg-accent' : 'bg-surface-secondary border border-border'
    //           }`}
    //         >
    //           <div
    //             className={`absolute top-0.5 w-5 h-5 rounded-full bg-surface-primary shadow transition-all duration-200 ${
    //               bot.mode === 'live' ? 'left-5.5' : 'left-0.5'
    //             }`}
    //           />
    //         </div>
    //       </div>

    //       {/* Strategy */}
    //       <span className="text-content-tertiary self-center">Strategy</span>
    //       <select
    //         defaultValue={bot.strategy}
    //         disabled={bot.status === 'running'}
    //         onChange={(e) => handleUpdateBot({ strategy: e.target.value })}
    //         className="border border-border bg-surface-secondary text-content-secondary px-2 py-1 rounded text-xs focus:outline-none focus:border-accent transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
    //       >
    //         {strategies.map((s) => (
    //           <option key={s} value={s}>
    //             {s}
    //           </option>
    //         ))}
    //       </select>

    //       {/* Quantity */}
    //       <span className="text-content-tertiary self-center">Quantity</span>
    //       <input
    //         type="number"
    //         defaultValue={bot.quantity}
    //         disabled={bot.status === 'running'}
    //         onBlur={(e) => handleUpdateBot({ quantity: e.target.value })}
    //         className="border border-border bg-surface-secondary text-content-secondary px-2 py-1 rounded text-xs w-24 focus:outline-none focus:border-accent transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
    //       />

    //       {/* Interval */}
    //       <span className="text-content-tertiary self-center">Interval</span>
    //       <select
    //         defaultValue={bot.interval}
    //         disabled={bot.status === 'running'}
    //         onChange={(e) => handleUpdateBot({ interval: e.target.value })}
    //         className="border border-border bg-surface-secondary text-content-secondary px-2 py-1 rounded text-xs focus:outline-none focus:border-accent transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
    //       >
    //         {candleIntervals.map((i) => (
    //           <option key={i} value={i}>
    //             {i}
    //           </option>
    //         ))}
    //       </select>

    //       {/* Lookback */}
    //       <span className="text-content-tertiary self-center">Lookback</span>
    //       <input
    //         defaultValue={bot.lookback}
    //         disabled={bot.status === 'running'}
    //         onBlur={(e) => handleUpdateBot({ lookback: e.target.value })}
    //         className="border border-border bg-surface-secondary text-content-secondary px-2 py-1 rounded text-xs w-24 focus:outline-none focus:border-accent transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
    //       />
    //     </div>
    //   ),
    // },
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
    setRunningFor(makeBotRunDurationStr(bot.started))
    const interval = setInterval(() => setRunningFor(makeBotRunDurationStr(bot.started!)), 1000)
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
      setError(errorMsg)
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
      setError(errorMsg)
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
      setError(errorMsg)
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
      setError(errorMsg)
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
      setError(errorMsg)
      showAlert({ title: errorMsg, type: 'error' })
    } finally {
      setLoadingAction(null)
    }
  }

  if (loading) return <BarLoader fullscreen />
  if (error) return <div>{error} 😢</div>

  return (
    <div className="space-y-4">
      <Heading title="Bot Overview" />
      {bot && (
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
      )}

      {tabs && <Tabs tabs={tabs} activeTab={activeTab} onTabChange={handleTabChange} />}
    </div>
  )
}

export default BotOverview
