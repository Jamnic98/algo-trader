import { BotCandleChart } from 'components'
import type { Bot } from 'types'
import { candleRangeToString } from 'utils'

type BotCandlesProps = {
  bot: Bot
}

const BotCandles = ({ bot }: BotCandlesProps) => {
  return (
    <div className="space-y-3 bg-content-secondary/10 rounded-xl p-2">
      <BotCandleChart bot={bot} />
      <div className="flex justify-between text-xs font-mono">
        <div className="flex flex-row gap-6">
          <div className="flex flex-col gap-0.5">
            <span className="text-content-secondary tracking-wider">Lookback</span>
            <span className="text-accent">
              {bot.candles?.length ? candleRangeToString(bot.candles) : '-'}
            </span>
          </div>

          <div className="flex flex-col gap-0.5">
            <span className="text-content-secondary tracking-wider">Interval</span>
            <span className="text-accent">{bot.interval ?? 0}</span>
          </div>
        </div>

        <div className="flex flex-col gap-0.5">
          <span className="text-content-secondary tracking-wider">Candles</span>
          <span className="text-content-secondary">
            {bot.candles?.length ?? 0}
            <span className="text-content-secondary"> / </span>
            <span className="text-accent">{bot.maxCandles}</span>
          </span>
        </div>
      </div>
    </div>
  )
}

export default BotCandles
