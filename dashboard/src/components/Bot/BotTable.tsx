import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'

import { BotActionButtons } from 'components'
import type { Bot } from 'types'

type ColumnKey = 'id' | 'symbol' | 'quantity' | 'interval' | 'lookback' | 'status' | 'started'
type BotStatus = 'running' | 'attached' | 'created'
type BotFilters = { status?: BotStatus }

type BotTableProps = {
  bots: Bot[]
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

const ALL_COLUMNS: ColumnKey[] = [
  'id',
  'symbol',
  'quantity',
  'interval',
  'lookback',
  'status',
  'started',
]

const BotTable = ({ bots, botFilters, botActions, columns }: BotTableProps) => {
  const navigate = useNavigate()

  const visibleColumns = useMemo(
    () => (columns ? ALL_COLUMNS.filter((col) => columns.includes(col)) : ALL_COLUMNS),
    [columns]
  )

  const filteredBots = useMemo(
    () => (botFilters?.status ? bots.filter((bot) => bot.status === botFilters.status) : bots),
    [bots, botFilters]
  )

  const renderCell = (bot: Bot, key: ColumnKey) => {
    switch (key) {
      case 'id':
        return <span className="font-mono text-sm truncate">{bot.id.split('-')[0]}</span>
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
    <table className="border-collapse w-full select-none">
      <thead>
        <tr className="bg-table-header text-accent font-semibold text-sm uppercase tracking-wider">
          {visibleColumns.map((col) => (
            <th key={col} className="border-b border-table-border px-4 py-3 text-left">
              {col}
            </th>
          ))}
          {botActions && (
            <th className="border-b border-table-border px-4 py-3 text-left">Actions</th>
          )}
        </tr>
      </thead>

      <tbody>
        {filteredBots.map((bot) => (
          <tr
            key={bot.id}
            className="bg-table-row border-b border-table-border h-12 cursor-pointer transition-colors duration-150 hover:bg-table-row-hover"
            onClick={() => navigate(`/bots/${bot.id}`)}
          >
            {visibleColumns.map((col) => (
              <td key={col} className="px-4 text-content-secondary">
                {renderCell(bot, col)}
              </td>
            ))}
            {botActions && (
              <td className="px-4 py-1">
                <div className="flex justify-center gap-2">
                  <BotActionButtons botId={bot.id} botStatus={bot.status} {...botActions} />
                </div>
              </td>
            )}
          </tr>
        ))}
      </tbody>
    </table>
  )
}

export default BotTable
