import { ChevronLeft, ChevronRight } from 'lucide-react'

import { SignalLabel } from 'components'
import type { Fill, Pagination } from 'types'

const COLUMNS = [
  { key: 'botID', label: 'Bot' },
  { key: 'side', label: 'Side' },
  { key: 'symbol', label: 'Symbol' },
  { key: 'quantity', label: 'Quantity' },
  { key: 'price', label: 'Price' },
  { key: 'fee', label: 'Fee' },
  { key: 'feeAsset', label: 'Fee Asset' },
  { key: 'timestamp', label: 'Timestamp' },
]

const renderCell = (trade: Fill, key: string) => {
  switch (key) {
    case 'side':
      return <SignalLabel signal={trade.side} />
    case 'timestamp':
      return new Intl.DateTimeFormat('en-GB', {
        day: '2-digit',
        month: '2-digit',
        year: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false,
      }).format(new Date(trade.timestamp))
    case 'botID':
      return <span className="font-mono text-sm truncate">{trade.botID.split('-')[0]}</span>
    default:
      return trade[key as keyof Fill] as string
  }
}

type TradesTableProps = {
  fills: Fill[]
  pagination: Pagination | null
  page: number
  onPageChange: (page: number) => void
}

const FillsTable = ({ fills, pagination, page, onPageChange }: TradesTableProps) => {
  if (!fills.length) {
    return <div className="text-content-secondary text-sm font-mono">No fills yet.</div>
  }

  return (
    <div>
      <div className="overflow-x-auto [&::-webkit-scrollbar]:h-0.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-300 [&::-webkit-scrollbar-thumb]:rounded-full hover:[&::-webkit-scrollbar-thumb]:bg-gray-400">
        <table className="border-collapse w-full select-none">
          <thead>
            <tr className="bg-table-header text-accent font-semibold text-sm uppercase tracking-wider text-nowrap">
              {COLUMNS.map((col) => (
                <th key={col.key} className="border-b border-table-border px-4 py-2 text-left">
                  {col.label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {fills.map((trade) => (
              <tr
                key={trade.id}
                className="bg-table-row border-b border-table-border h-8 transition-colors duration-150 text-nowrap"
              >
                {COLUMNS.map((col) => (
                  <td key={col.key} className="px-4 text-content-secondary">
                    {renderCell(trade, col.key)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {pagination && pagination.total_pages > 1 && (
        <div className="flex items-center justify-between mt-4 text-xs font-mono text-content-secondary">
          <span className="text-content-secondary/50">{pagination.total} fills</span>
          <div className="flex items-center gap-1">
            <button
              onClick={() => onPageChange(page - 1)}
              disabled={page === 1}
              className="p-1.5 rounded-md border border-border hover:text-content-primary hover:bg-white/4 disabled:opacity-30 disabled:cursor-not-allowed transition-all cursor-pointer"
            >
              <ChevronLeft size={14} />
            </button>
            {Array.from({ length: pagination.total_pages }, (_, i) => i + 1)
              .filter((p) => p === 1 || p === pagination.total_pages || Math.abs(p - page) <= 1)
              .reduce<(number | '...')[]>((acc, p, i, arr) => {
                if (i > 0 && p - (arr[i - 1] as number) > 1) acc.push('...')
                acc.push(p)
                return acc
              }, [])
              .map((p, i) =>
                p === '...' ? (
                  <span key={`ellipsis-${i}`} className="px-1 text-content-secondary/50">
                    ...
                  </span>
                ) : (
                  <button
                    key={p}
                    onClick={() => onPageChange(p as number)}
                    className={`px-3 py-1 rounded-md border transition-all cursor-pointer ${
                      p === page
                        ? 'bg-accent-muted text-accent border-accent/25'
                        : 'border-border text-content-secondary hover:text-content-primary hover:bg-white/4'
                    }`}
                  >
                    {p}
                  </button>
                )
              )}
            <button
              onClick={() => onPageChange(page + 1)}
              disabled={page === pagination.total_pages}
              className="p-1.5 rounded-md border border-border hover:text-content-primary hover:bg-white/4 disabled:opacity-30 disabled:cursor-not-allowed transition-all cursor-pointer"
            >
              <ChevronRight size={14} />
            </button>
          </div>
          <span className="text-content-secondary/50">
            {page} / {pagination.total_pages}
          </span>
        </div>
      )}
    </div>
  )
}

export default FillsTable
