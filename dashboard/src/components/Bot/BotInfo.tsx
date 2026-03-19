import { Copy } from 'lucide-react'

import { useAlert } from 'hooks'
import { getStrategyLabel } from 'utils'
import type { Bot } from 'types'

type BotInfoProps = {
  bot: Bot
  runningFor: string
}

const BotInfo = ({ bot, runningFor }: BotInfoProps) => {
  const { showAlert } = useAlert()

  return (
    <div className="flex flex-col gap-4 text-sm font-mono select-none min-w-0 max-w-full">
      {/* Static config — always present */}
      <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1.5">
        {/* Bot Id */}
        <span className="text-content-tertiary">Id</span>
        <span className="flex items-center gap-2 text-content-secondary min-w-0">
          <span className="truncate min-w-0">{bot.id}</span>
          <Copy
            size={12}
            className="cursor-pointer shrink-0 hover:text-content-primary transition-colors"
            onClick={() => {
              navigator.clipboard.writeText(bot.id)
              showAlert({ type: 'info', message: 'Bot id copied' })
            }}
          />
        </span>
        {/* Bot Trade Mode */}
        <span className="text-content-tertiary">Mode</span>
        <span className="text-content-secondary capitalize">{bot.mode}</span>
        {/* Exchange */}
        <span className="text-content-tertiary">Exchange</span>
        <span className="text-content-secondary capitalize">{bot.exchange}</span>
        {/* Strategy */}
        <span className="text-content-tertiary">Strategy</span>
        <span className="text-content-secondary">{getStrategyLabel(bot.strategy.name)}</span>
        {/* Interval */}
        <span className="text-content-tertiary">Interval</span>
        <span className="text-content-secondary">{bot.interval}</span>
        {/* Max Candles */}
        <span className="text-content-tertiary">Max Candles</span>
        <span className="text-content-secondary">{bot.maxCandles}</span> {/* Quantity */}
        {bot?.quantity && (
          <>
            <span className="text-content-tertiary">Quantity</span>
            <span className="text-content-secondary">{bot.quantity}</span>
          </>
        )}
      </div>

      {/* Runtime info — only after bot has started */}
      {bot.started && (
        <>
          {/* Divider */}
          <div className="h-px opacity-50 bg-linear-to-r from-accent via-border to-transparent" />
          <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1.5">
            <span className="text-content-tertiary">Started</span>
            <span className="text-content-secondary">
              {new Date(Date.parse(bot.started)).toLocaleString('en-GB', { timeZone: 'UTC' })}
            </span>

            <span className="text-content-tertiary">Run-time</span>
            <span className="text-content-secondary">{runningFor}</span>
          </div>
        </>
      )}
    </div>
  )
}

export default BotInfo
