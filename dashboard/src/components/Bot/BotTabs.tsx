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
import { makeBotRunDurationStr } from 'utils'

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

  const { id } = bot
  const tabs: Tab[] = [
    {
      label: 'Info',
      icon: <Info size={13} />,
      content: <BotInfo bot={bot} runningFor={runningFor} />,
    },
    {
      label: 'Stats',
      icon: <BarChart2 size={13} />,
      content: <BotStats bot={bot} tick={positionsTick} />,
    },
    {
      label: 'Trades',
      icon: <ArrowLeftRight size={13} />,
      content: <BotTrades id={id} onNewTrade={onNewTrade} />,
    },
    {
      label: 'Candles',
      icon: <CChart size={13} />,
      content:
        bot?.started && bot.status === 'trading' ? (
          <BotCandles bot={bot} />
        ) : (
          <div className="text-content-secondary text-sm font-mono">
            Start bot to fetch candlestick data.
          </div>
        ),
    },
    {
      label: 'Logs',
      icon: <ScrollText size={13} />,
      content: <BotLogs key={id} id={id} />,
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
