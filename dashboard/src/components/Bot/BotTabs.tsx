import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import {
  ArrowLeftRight,
  BarChart2,
  CandlestickChart as CChart,
  Info,
  ScrollText,
} from 'lucide-react'

import { BotCandles, BotInfo, BotLogs, BotStats, BotTrades, Tabs } from 'components'
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
      content: <BotInfo bot={bot} runningFor={runningFor} />,
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
      content: bot.started ? (
        <BotCandles bot={bot} />
      ) : (
        <div className="text-gray-500">Bot not attached.</div>
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
