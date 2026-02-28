import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'

import { BotActionButtons } from 'components'
import type { BotData } from 'types'

type BotStatus = 'running' | 'attached' | 'created'

type BotFilters = { status?: BotStatus }

type ColumnKey = 'id' | 'symbol' | 'quantity' | 'interval' | 'lookback' | 'status' | 'started'

type BotTableProps = {
  bots: BotData[]
  botFilters?: BotFilters
  botActions?: {
    startBot: (id: string) => void
    stopBot: (id: string) => void
    attachBot: (id: string) => void
    detachBot: (id: string) => void
    deleteBot: (id: string) => void
  }
  columns?: ColumnKey[]
}

const ALL_COLUMNS: { key: ColumnKey; label: string }[] = [
  { key: 'id', label: 'Id' },
  { key: 'symbol', label: 'Symbol' },
  { key: 'quantity', label: 'Quantity' },
  { key: 'interval', label: 'Interval' },
  { key: 'lookback', label: 'Lookback' },
  { key: 'status', label: 'Status' },
  { key: 'started', label: 'Started' },
]

const BotTable = ({ bots, botFilters, botActions, columns }: BotTableProps) => {
  const navigate = useNavigate()

  const visibleColumns = useMemo(() => {
    if (!columns) return ALL_COLUMNS
    return ALL_COLUMNS.filter((col) => columns.includes(col.key))
  }, [columns])

  const filteredBots = useMemo(() => {
    if (!botFilters?.status) return bots
    return bots.filter((bot) => bot.status === botFilters.status)
  }, [bots, botFilters])

  const renderCell = (bot: BotData, key: ColumnKey) => {
    switch (key) {
      case 'id':
        return <span className="font-mono text-sm truncate">{bot.id}</span>
      case 'status':
        return <span className="capitalize">{bot.status}</span>
      case 'started':
        return bot.started
          ? new Intl.DateTimeFormat('en-GB', {
              day: '2-digit',
              month: '2-digit',
              year: '2-digit',
              hour: '2-digit',
              minute: '2-digit',
              hour12: false,
            }).format(new Date(bot.started))
          : '-'
      default:
        return bot[key] as string
    }
  }

  return (
    <table className="border-collapse border w-full bg-gray-50">
      <thead>
        <tr className="text-blue-900 font-semibold bg-blue-200">
          {visibleColumns.map((col) => (
            <th key={col.key} className="border px-3 py-1 text-left">
              {col.label}
            </th>
          ))}
          {botActions && <th className="border px-3 py-1 text-left">Actions</th>}
        </tr>
      </thead>

      <tbody className="text-gray-800">
        {filteredBots.map((bot, index) => {
          const rowBg = index % 2 === 0 ? 'bg-white' : 'bg-blue-50'
          return (
            <tr
              key={bot.id}
              className={`group cursor-pointer transition-colors ${rowBg} duration-200 hover:bg-gray-200`}
              onClick={() => navigate(`/bots/${bot.id}`)}
            >
              {visibleColumns.map((col) => (
                <td key={col.key} className="border px-3 py-1">
                  {renderCell(bot, col.key)}
                </td>
              ))}
              {botActions && (
                <td className="border px-3 py-1">
                  <div className="flex justify-center gap-2">
                    <BotActionButtons botId={bot.id} botStatus={bot.status} {...botActions} />
                  </div>
                </td>
              )}
            </tr>
          )
        })}
      </tbody>
    </table>
  )
}

export default BotTable
