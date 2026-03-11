import { X } from 'lucide-react'
import type { TradeFilters } from 'types'

type TradesFiltersProps = {
  filters: TradeFilters
  onChange: (filters: TradeFilters) => void
  onClear: () => void
}

const LIMITS = [5, 15, 25, 50]
const SIDES = ['', 'BUY', 'SELL']

const TradesFilters = ({ filters, onChange, onClear }: TradesFiltersProps) => {
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

      {/* Symbol */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">Symbol</label>
        <input
          name="symbol"
          value={filters.symbol}
          onChange={handleChange}
          placeholder="SOLUSDT"
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded w-28 text-sm uppercase focus:outline-none focus:border-accent transition-colors placeholder:text-content-secondary/40"
        />
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

      {/* Bot ID */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">Bot ID</label>
        <input
          name="botId"
          value={filters.botId}
          onChange={handleChange}
          placeholder="abc123"
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded w-28 text-sm font-mono focus:outline-none focus:border-accent transition-colors placeholder:text-content-secondary/40"
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
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
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
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
        />
      </div>

      {/* Clear */}
      {hasFilters && (
        <button
          onClick={onClear}
          className="flex items-center gap-1.5 px-2 py-1.5 rounded border border-border text-content-secondary text-sm hover:text-content-primary hover:bg-white/4 transition-all cursor-pointer"
        >
          <X size={14} /> Clear
        </button>
      )}
    </div>
  )
}

export default TradesFilters
