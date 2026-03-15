import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import {
  ArrowLeftRight,
  BarChart2,
  Copy,
  CandlestickChart as CChart,
  Info,
  ScrollText,
} from 'lucide-react'

import { BotCandleChart, BotLogs, BotStats, BotTrades, Tabs } from 'components'
import { useAlert } from 'hooks'
import type { Bot, Tab } from 'types'

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

interface BotTabsProps {
  bot: Bot
  onNewTrade: () => void
  positionsTick: number
}

const BotTabs = ({ bot, onNewTrade, positionsTick }: BotTabsProps) => {
  const { showAlert } = useAlert()
  const [searchParams, setSearchParams] = useSearchParams()
  const [runningFor, setRunningFor] = useState<string>(
    bot.started ? makeBotRunDurationStr(bot.started) : '-'
  )

  useEffect(() => {
    if (!bot.started) return
    const interval = setInterval(() => setRunningFor(makeBotRunDurationStr(bot.started!)), 1000)
    return () => clearInterval(interval)
  }, [bot.started])

  const tabs: Tab[] = [
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
      content: <BotStats id={bot.id} tick={positionsTick} />,
    },
    {
      label: 'Trades',
      icon: <ArrowLeftRight size={13} />,
      content: (
        <div className="space-y-3 text-gray-500">
          <BotTrades id={bot.id} onNewTrade={onNewTrade} />
        </div>
      ),
    },
    {
      label: 'Candles',
      icon: <CChart size={13} />,
      content: (
        <BotCandleChart id={bot.id} symbol={`${bot.base}/${bot.quote}`} status={bot.status} />
      ),
    },
    {
      label: 'Logs',
      icon: <ScrollText size={13} />,
      content: <BotLogs id={bot.id} />,
    },
  ]

  const tabLabels = tabs.map((t) => t.label.toLowerCase())

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

  return <Tabs tabs={tabs} activeTab={activeTab} onTabChange={handleTabChange} />
}

export default BotTabs
