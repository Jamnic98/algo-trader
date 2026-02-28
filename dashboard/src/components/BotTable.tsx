import { useNavigate } from 'react-router-dom'

import { BotActionButtons } from 'components'
import type { BotData } from 'types'

type BotTableProps = {
  bots: BotData[]
  startBot: (id: string) => void
  stopBot: (id: string) => void
  attachBot: (id: string) => void
  detachBot: (id: string) => void
  deleteBot: (id: string) => void
}

const BotTable = ({ bots, startBot, stopBot, attachBot, detachBot, deleteBot }: BotTableProps) => {
  const navigate = useNavigate()

  return (
    <table className="border-collapse border w-full bg-gray-50">
      <thead>
        <tr className="text-blue-900 font-semibold bg-blue-200">
          <th className="border px-3 py-1 text-left">Id</th>
          <th className="border px-3 py-1 text-left">Symbol</th>
          <th className="border px-3 py-1 text-left">Quantity</th>
          <th className="border px-3 py-1 text-left">Interval</th>
          <th className="border px-3 py-1 text-left">Lookback</th>
          <th className="border px-3 py-1 text-left">Status</th>
          <th className="border px-3 py-1 text-left">Started</th>
          <th className="border px-3 py-1 text-left">Actions</th>
        </tr>
      </thead>

      <tbody className="text-gray-800">
        {bots.map((bot, index) => {
          const rowBg = index % 2 === 0 ? 'bg-white' : 'bg-blue-50'
          return (
            <tr
              key={bot.id}
              className={`group cursor-pointer transition-colors ${rowBg} duration-200 hover:bg-gray-200`}
              onClick={() => navigate(`${bot.id}`)}
            >
              <td className="border px-3 py-1 font-mono text-sm max-w-30 truncate">{bot.id}</td>
              <td className="border px-3 py-1">{bot.symbol}</td>
              <td className="border px-3 py-1">{bot.quantity}</td>
              <td className="border px-3 py-1">{bot.interval}</td>
              <td className="border px-3 py-1">{bot.lookback}</td>
              <td className="border px-3 py-1 capitalize">{bot.status}</td>
              <td className="border px-3 py-1 text-center">
                {bot.started
                  ? new Intl.DateTimeFormat('en-GB', {
                      day: '2-digit',
                      month: '2-digit',
                      year: '2-digit',
                      hour: '2-digit',
                      minute: '2-digit',
                      hour12: false,
                    }).format(new Date(bot.started))
                  : '-'}
              </td>
              <td className="border px-3 py-1">
                <div className="flex justify-center gap-2">
                  <BotActionButtons
                    botId={bot.id}
                    botStatus={bot.status}
                    startBot={startBot}
                    stopBot={stopBot}
                    attachBot={attachBot}
                    detachBot={detachBot}
                    deleteBot={deleteBot}
                  />
                </div>
              </td>
            </tr>
          )
        })}
      </tbody>
    </table>
  )
}

export default BotTable
