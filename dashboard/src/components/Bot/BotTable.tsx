import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { GlobeIcon, GlobeOff } from 'lucide-react'

import { BotActionButtons, BotStatusPill } from 'components'
import type { Bot, LoadingAction } from 'types'
import { displaySymbol, getExchangeLabel, getStrategyLabel } from 'utils'

type BotStatus = 'running' | 'attached' | 'created'
type BotFilters = { status?: BotStatus }
type ColumnKey =
  | 'id'
  | 'mode'
  | 'exchange'
  | 'symbol'
  | 'strategy'
  | 'quantity'
  | 'interval'
  | 'maxCandles'
  | 'status'
  | 'started'

type BotTableProps = {
  bots: Bot[]
  botFilters?: BotFilters
  botLoadingActions?: Record<string, LoadingAction>
  botActions?: {
    startBot: (id: string) => void
    stopBot: (id: string) => void
    deleteBot: (id: string) => void
  }
  columns?: ColumnKey[]
}

const ALL_COLUMNS: { key: ColumnKey; label: string; className?: string }[] = [
  { key: 'id', label: 'Id', className: 'w-24' },
  { key: 'mode', label: 'Mode', className: 'w-16' },
  { key: 'exchange', label: 'Exchange', className: 'w-16' },
  { key: 'strategy', label: 'Strategy', className: 'w-24' },
  { key: 'symbol', label: 'Symbol', className: 'w-24' },
  { key: 'quantity', label: 'Quantity', className: 'w-24' },
  { key: 'interval', label: 'Interval', className: 'w-20' },
  { key: 'maxCandles', label: 'Max Candles', className: 'w-24' },
  { key: 'status', label: 'Status', className: 'w-24' },
  { key: 'started', label: 'Started', className: 'min-w-36 max-w-36' },
]

const BotTable = ({ bots, botFilters, botActions, botLoadingActions, columns }: BotTableProps) => {
  const navigate = useNavigate()

  const visibleColumns = useMemo(
    () => (columns ? ALL_COLUMNS.filter((col) => columns.includes(col.key)) : ALL_COLUMNS),
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
      case 'mode':
        return bot.mode === 'paper' ? (
          <GlobeOff size={18} className="text-content-secondary" />
        ) : (
          <GlobeIcon size={18} className="text-accent" />
        )
      case 'exchange':
        return getExchangeLabel(bot.exchange)
      case 'status':
        return <BotStatusPill status={bot.status} />
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
      case 'strategy':
        return getStrategyLabel(bot.strategy.name)
      case 'symbol':
        return displaySymbol(bot.assetType, bot.base, bot.quote)
      default:
        return bot[key]
    }
  }

  return (
    <table className="border-collapse w-full select-none text-nowrap">
      <thead>
        <tr className="bg-table-header text-accent font-semibold text-sm uppercase tracking-wider">
          {visibleColumns.map((col) => (
            <th
              key={col.key}
              className={`border-b border-table-border px-4 py-3 text-left ${col.className ?? ''}`}
            >
              {col.label}
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
              <td key={col.key} className={`px-4 text-content-secondary ${col.className ?? ''}`}>
                {renderCell(bot, col.key)}
              </td>
            ))}
            {botActions && (
              <td className="px-4 py-1">
                <div className="flex justify-center gap-2">
                  <BotActionButtons
                    botId={bot.id}
                    botStatus={bot.status}
                    loadingAction={botLoadingActions?.[bot.id] ?? null}
                    {...botActions}
                  />
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
