import { useEffect, useState, useCallback } from 'react'

import { Heading, TradesTable, TradesFilters, BarLoader } from 'components'
import { getAllTrades, getAllBots } from 'api'
import { useAlert, useDebounce } from 'hooks'
import type { Trade, Pagination, TradeFilters, Bot } from 'types'

const DEFAULT_FILTERS: TradeFilters = {
  limit: 15,
  symbol: '',
  side: '',
  botId: '',
  dateFrom: '',
  dateTo: '',
}

const Trades = () => {
  const [includeDead, setIncludeDead] = useState(false)
  const [bots, setBots] = useState<Bot[]>([])
  const [trades, setTrades] = useState<Trade[]>([])
  const [pagination, setPagination] = useState<Pagination | null>(null)
  const [page, setPage] = useState(1)
  const [filters, setFilters] = useState<TradeFilters>(DEFAULT_FILTERS)
  const [loading, setLoading] = useState(true)

  const { showAlert } = useAlert()
  const debouncedFilters = useDebounce(filters, 400)

  useEffect(() => {
    getAllBots({ deleted: includeDead }).then(setBots)
  }, [includeDead])

  const fetchTrades = useCallback(async () => {
    try {
      setLoading(true)
      const res = await getAllTrades({ page, ...debouncedFilters })
      setTrades(res.trades)
      setPagination(res.pagination)
    } catch (err) {
      console.log(err)
      showAlert({ type: 'error', title: 'Failed to load trades' })
    } finally {
      setLoading(false)
    }
  }, [page, debouncedFilters, showAlert])

  useEffect(() => {
    fetchTrades()
  }, [fetchTrades])

  useEffect(() => {
    const es = new EventSource(`/api/trades/stream?api_key=${import.meta.env.VITE_SERVER_API_KEY}`)
    es.onmessage = (e) => {
      const trade: Trade = JSON.parse(e.data)
      const { symbol, side, botId, limit } = debouncedFilters

      if (symbol && !trade.symbol.includes(symbol.toUpperCase())) return
      if (side && trade.side !== side) return
      if (botId && trade.botID !== botId) return

      // Only inject into the live view if the user is on page 1.
      // On other pages the new trade is out of scope — let a
      // manual refresh or page navigation pick it up.
      if (page !== 1) return

      setTrades((prev) => {
        const next = [trade, ...prev]
        // keep trimmed to page size
        return next.slice(0, limit)
      })

      // Bump the total count so pagination recalculates correctly
      setPagination((prev) => {
        if (!prev) return prev
        const total = prev.total + 1
        const total_pages = Math.ceil(total / filters.limit)
        return { ...prev, total, total_pages }
      })
    }

    es.onerror = () => {
      const errorMsg = 'Trades stream error'
      console.error(errorMsg)
      showAlert({ type: 'error', title: errorMsg })
      es.close()
    }

    return () => es.close()
  }, [debouncedFilters, filters.limit, page, showAlert])

  const handleFilterChange = (newFilters: TradeFilters) => {
    setFilters(newFilters)
    setPage(1) // reset to page 1 on filter change
  }

  const handleClear = () => {
    setFilters(DEFAULT_FILTERS)
    setIncludeDead(false)
    setPage(1)
  }

  const handleIncludeDeadChange = (val: boolean) => {
    setIncludeDead(val)
    setFilters((f) => ({ ...f, botId: '' }))
  }

  return (
    <div className="space-y-6">
      <Heading title="Trades" />
      <TradesFilters
        filters={filters}
        onChange={handleFilterChange}
        onClear={handleClear}
        bots={bots}
        includeDead={includeDead}
        onIncludeDeadChange={handleIncludeDeadChange}
      />
      {loading ? (
        <BarLoader />
      ) : (
        <TradesTable trades={trades} pagination={pagination} page={page} onPageChange={setPage} />
      )}
    </div>
  )
}

export default Trades
