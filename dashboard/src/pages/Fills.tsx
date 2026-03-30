import { useEffect, useState, useCallback } from 'react'

import { Heading, FillsTable, FillsFilters, BarLoader } from 'components'
import { getAllFills, getAllBots } from 'api'
import { useAlert, useDebounce } from 'hooks'
import type { Fill, Pagination, FillFilters, Bot } from 'types'

const DEFAULT_FILTERS: FillFilters = {
  limit: 15,
  symbol: '',
  side: '',
  botId: '',
  dateFrom: '',
  dateTo: '',
}

const Fills = () => {
  const [includeDead, setIncludeDead] = useState(false)
  const [bots, setBots] = useState<Bot[]>([])
  const [fills, setTrades] = useState<Fill[]>([])
  const [pagination, setPagination] = useState<Pagination | null>(null)
  const [page, setPage] = useState(1)
  const [filters, setFilters] = useState<FillFilters>(DEFAULT_FILTERS)
  const [loading, setLoading] = useState(true)

  const { showAlert } = useAlert()
  const debouncedFilters = useDebounce(filters, 400)

  useEffect(() => {
    getAllBots({ deleted: includeDead }).then(setBots)
  }, [includeDead])

  const fetchFills = useCallback(async () => {
    try {
      setLoading(true)
      const res = await getAllFills({ page, ...debouncedFilters })
      setTrades(res.fills)
      setPagination(res.pagination)
    } catch (err) {
      console.error(err)
      showAlert({ type: 'error', title: 'Failed to load fills' })
    } finally {
      setLoading(false)
    }
  }, [page, debouncedFilters, showAlert])

  useEffect(() => {
    fetchFills()
  }, [fetchFills])

  useEffect(() => {
    const es = new EventSource(`/api/fills/stream?api_key=${import.meta.env.VITE_SERVER_API_KEY}`)
    es.onmessage = (e) => {
      const fill: Fill = JSON.parse(e.data)
      const { symbol, side, botId, limit } = debouncedFilters

      if (symbol && !fill.symbol.includes(symbol.toUpperCase())) return
      if (side && fill.side !== side) return
      if (botId && fill.botID !== botId) return

      // Only inject into the live view if the user is on page 1.
      // On other pages the new fill is out of scope — let a
      // manual refresh or page navigation pick it up.
      if (page !== 1) return

      setTrades((prev) => {
        const next = [fill, ...prev]
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
      const errorMsg = 'Fills stream error'
      console.error(errorMsg)
      showAlert({ type: 'error', title: errorMsg })
      es.close()
    }

    return () => es.close()
  }, [debouncedFilters, filters.limit, page, showAlert])

  const handleFilterChange = (newFilters: FillFilters) => {
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
    <div className="space-y-8">
      <Heading title="Fills" />
      <div className="space-y-4">
        <FillsFilters
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
          <FillsTable fills={fills} pagination={pagination} page={page} onPageChange={setPage} />
        )}
      </div>
    </div>
  )
}

export default Fills
