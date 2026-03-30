import { useEffect, useState } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'

import { getBotFills } from 'api'
import { BarLoader } from 'components'
import { useAlert } from 'hooks'
import type { Fill, Pagination } from 'types'

type BotTradesProps = {
  id: string
  onNewTrade: () => void
}

const TRADES_LIMIT = 15

const COLUMNS = [
  // { key: 'botID', label: 'Bot ID' },
  // { key: 'symbol', label: 'Symbol' },
  { key: 'side', label: 'Side' },
  { key: 'quantity', label: 'Quantity' },
  { key: 'price', label: 'Price' },
  { key: 'fee', label: 'Fee' },
  { key: 'timestamp', label: 'Timestamp' },
  // { key: 'feeAsset', label: 'Fee Asset' },
]

const renderCell = (trade: Fill, key: string) => {
  switch (key) {
    case 'side':
      return (
        <span className={trade.side === 'BUY' ? 'text-green-400' : 'text-red-400'}>
          {trade.side}
        </span>
      )
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

const BotTrades = ({ id, onNewTrade }: BotTradesProps) => {
  const { showAlert } = useAlert()
  const [fills, setTrades] = useState<Fill[]>([])
  const [pagination, setPagination] = useState<Pagination | null>(null)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)

  // initial fetch
  useEffect(() => {
    const fetchFills = async () => {
      try {
        setLoading(true)
        const res = await getBotFills(id, page, TRADES_LIMIT)
        setTrades(res.data)
        setPagination(res.pagination)
      } catch (err) {
        console.error(err)
        showAlert({ title: 'Failed to load bot fills', type: 'error' })
      } finally {
        setLoading(false)
      }
    }
    fetchFills()
  }, [showAlert, id, page])

  useEffect(() => {
    if (page !== 1) return

    const source = new EventSource(
      `/api/bots/${id}/fills/stream?api_key=${encodeURIComponent(import.meta.env.VITE_SERVER_API_KEY)}`
    )

    source.onmessage = (e) => {
      const trade: Fill = JSON.parse(e.data)
      setTrades((prev) => {
        const next = [trade, ...prev]
        return next.slice(0, TRADES_LIMIT) // or whatever your page size is
      })
      setPagination((prev) => {
        if (!prev) return prev
        const total = prev.total + 1
        const total_pages = Math.ceil(total / TRADES_LIMIT)
        return { ...prev, total, total_pages }
      })
      onNewTrade()
    }

    source.onerror = () => {
      const errorMsg = 'Trade stream error'
      console.error(errorMsg)
      showAlert({ type: 'error', title: errorMsg })
      source.close()
    }

    return () => source.close()
  }, [id, page, showAlert, onNewTrade])

  if (loading) return <BarLoader />
  if (!fills.length)
    return <div className="text-content-secondary text-sm font-mono">No fills yet.</div>

  return (
    <div>
      <div className="overflow-x-auto [&::-webkit-scrollbar]:h-1 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-300 [&::-webkit-scrollbar-thumb]:rounded-full hover:[&::-webkit-scrollbar-thumb]:bg-gray-400">
        <table className="border-collapse w-full select-none">
          <thead>
            <tr className="bg-table-header text-accent font-semibold text-sm uppercase tracking-wider text-nowrap">
              {COLUMNS.map((col) => (
                <th key={col.key} className="border-b border-table-border px-4 py-3 text-left">
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

      {/* Pagination */}
      {pagination && pagination.total_pages > 1 && (
        <div className="flex items-center justify-between mt-4 text-xs font-mono text-content-secondary">
          <span className="text-content-secondary/50">{pagination.total} fills</span>

          <div className="flex items-center gap-1">
            <button
              onClick={() => setPage((p) => p - 1)}
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
                    onClick={() => setPage(p as number)}
                    className={`px-3 py-1 rounded-md border transition-all cursor-pointer
                ${
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
              onClick={() => setPage((p) => p + 1)}
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

export default BotTrades
