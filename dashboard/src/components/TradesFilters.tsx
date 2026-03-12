import { RefreshCcw } from 'lucide-react'

import type { Bot, TradeFilters } from 'types'

type TradesFiltersProps = {
  bots: Bot[]
  filters: TradeFilters
  onChange: (filters: TradeFilters) => void
  onClear: () => void
  includeDead: boolean
  onIncludeDeadChange: (includeDead: boolean) => void
}

const LIMITS = [15, 25, 50]
const SIDES = ['', 'BUY', 'SELL']

const TradesFilters = ({
  bots,
  filters,
  onChange,
  onClear,
  includeDead,
  onIncludeDeadChange,
}: TradesFiltersProps) => {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    onChange({ ...filters, [e.target.name]: e.target.value })
  }

  const hasFilters = Object.values(filters).some(Boolean)

  return (
    <div className="flex flex-wrap items-end gap-4">
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">Per page</label>
        <select
          name="limit"
          value={filters.limit}
          onChange={handleChange}
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
        >
          {LIMITS.map((l) => (
            <option key={l} value={l}>
              {l}
            </option>
          ))}
        </select>
      </div>

      {/* Bot */}
      <div className="flex flex-col gap-1">
        <div className="flex items-center justify-between gap-3">
          <label className="text-content-secondary text-xs uppercase tracking-wider">Bot</label>
          <div className="flex items-center gap-1.5">
            <label className="text-content-tertiary text-xs">incl. dead</label>
            <div
              onClick={() => onIncludeDeadChange(!includeDead)}
              className={`relative w-7 h-4 rounded-full cursor-pointer transition-colors duration-200 ${
                includeDead ? 'bg-accent' : 'bg-surface-secondary border border-border'
              }`}
            >
              <div
                className={`absolute top-0.5 w-3 h-3 rounded-full bg-surface-primary shadow transition-all duration-200 ${
                  includeDead ? 'left-3.5' : 'left-0.5'
                }`}
              />
            </div>
          </div>
        </div>
        <select
          name="botId"
          value={filters.botId}
          onChange={handleChange}
          className="border border-border bg-surface-secondary text-content-secondary px-3 py-1.5 rounded text-sm font-mono focus:outline-none focus:border-accent transition-colors"
        >
          <option value="">All</option>
          {bots.map((b) => (
            <option key={b.id} value={b.id}>
              {b.id.split('-')[0]}
              {includeDead && b.deletedAt ? ' · dead' : ''}
            </option>
          ))}
        </select>
      </div>

      {/* Side */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">Side</label>
        <select
          name="side"
          value={filters.side}
          onChange={handleChange}
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
        >
          {SIDES.map((s) => (
            <option key={s} value={s}>
              {s || 'All'}
            </option>
          ))}
        </select>
      </div>

      {/* Symbol */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">Symbol</label>
        <input
          name="symbol"
          value={filters.symbol}
          onChange={handleChange}
          placeholder="BTCUSDT"
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded w-28 text-sm uppercase focus:outline-none focus:border-accent transition-colors placeholder:text-content-secondary/10"
        />
      </div>

      {/* Date From */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">From</label>
        <input
          name="dateFrom"
          type="date"
          value={filters.dateFrom}
          onChange={handleChange}
          className={`border border-border bg-surface-secondary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors ${
            filters.dateFrom ? 'text-content-primary' : 'text-content-secondary/10'
          }`}
        />
      </div>

      {/* Date To */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">To</label>
        <input
          name="dateTo"
          type="date"
          value={filters.dateTo}
          onChange={handleChange}
          className={`border border-border bg-surface-secondary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors ${
            filters.dateTo ? 'text-content-primary' : 'text-content-secondary/10'
          }`}
        />
      </div>

      {/* Clear */}
      {hasFilters && (
        <button
          onClick={onClear}
          className="flex items-center gap-1.5 px-2 py-1.5 rounded border border-border text-content-secondary text-sm hover:text-content-primary hover:bg-white/4 transition-all cursor-pointer"
        >
          <RefreshCcw size={14} /> Reset
        </button>
      )}
    </div>
  )
}

export default TradesFilters
