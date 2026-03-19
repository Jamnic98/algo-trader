import { useAlert } from 'hooks'
import { Copy } from 'lucide-react'

import type { Bot } from 'types'
import { getStrategyLabel } from 'utils'

type BotInfoProps = {
  bot: Bot
  runningFor: string
}

const BotInfo = ({ bot, runningFor }: BotInfoProps) => {
  const { showAlert } = useAlert()

  return (
    <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1.5 text-sm font-mono select-none min-w-0 max-w-full">
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
      {/* Mode */}
      <span className="text-content-tertiary">Mode</span>
      <span className="text-content-secondary capitalize">{bot.mode}</span>

      {/* Strategy` */}
      <span className="text-content-tertiary">Strategy</span>
      <span className="text-content-secondary">{getStrategyLabel(bot.strategy.name)}</span>
      {bot?.quantity && (
        <>
          <span className="text-content-tertiary">Quantity</span>
          <span className="text-content-secondary">{bot.quantity}</span>
        </>
      )}

      {/* Number of candles */}
      {bot.candles && (
        <>
          <span className="text-content-tertiary">Candles</span>
          <span className="text-content-secondary">{bot.candles.length}</span>
        </>
      )}
      {/* Interval duration */}
      <span className="text-content-tertiary">Interval</span>
      <span className="text-content-secondary">{bot.interval}</span>
      {/* Lookback duration */}
      <span className="text-content-tertiary">Lookback</span>
      <span className="text-content-secondary">{bot.lookback}</span>
      {/* Start time */}
      <span className="text-content-tertiary">Started</span>
      <span className="text-content-secondary">
        {bot.started
          ? new Date(Date.parse(bot.started)).toLocaleString('en-GB', { timeZone: 'UTC' })
          : '-'}
      </span>
      <span className="text-content-tertiary">Run-time</span>
      <span className="text-content-secondary">{bot.started ? runningFor : '-'}</span>
    </div>
  )
}

export default BotInfo
