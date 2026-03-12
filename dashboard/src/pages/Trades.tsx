import { useEffect, useState, useCallback } from 'react'

import { Heading, SpinnerLoader, TradesTable, TradesFilters } from 'components'
import { getAllTrades } from 'api'
import { useAlert } from 'hooks'
import type { Trade, Pagination, TradeFilters } from 'types'

const DEFAULT_FILTERS: TradeFilters = {
  limit: 15,
  symbol: '',
  side: '',
  botId: '',
  dateFrom: '',
  dateTo: '',
}

const Trades = () => {
  const { showAlert } = useAlert()
  const [trades, setTrades] = useState<Trade[]>([])
  const [pagination, setPagination] = useState<Pagination | null>(null)
  const [page, setPage] = useState(1)
  const [filters, setFilters] = useState<TradeFilters>(DEFAULT_FILTERS)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchTrades = useCallback(async () => {
    try {
      setLoading(true)
      const res = await getAllTrades({ page, ...filters })
      setTrades(res.trades)
      setPagination(res.pagination)
    } catch {
      const errorMsg = 'Failed to load trades'
      setError(errorMsg)
      showAlert({ title: errorMsg, type: 'error' })
    } finally {
      setLoading(false)
    }
  }, [page, filters, showAlert])

  useEffect(() => {
    fetchTrades()
  }, [fetchTrades])

  const handleFilterChange = (newFilters: TradeFilters) => {
    setFilters(newFilters)
    setPage(1) // reset to page 1 on filter change
  }

  const handleClear = () => {
    setFilters(DEFAULT_FILTERS)
    setPage(1)
  }

  if (error) return <div>{error}</div>

  return (
    <div className="space-y-6">
      <Heading title="Trades" />
      <TradesFilters filters={filters} onChange={handleFilterChange} onClear={handleClear} />
      {loading ? (
        <SpinnerLoader />
      ) : (
        <TradesTable trades={trades} pagination={pagination} page={page} onPageChange={setPage} />
      )}
    </div>
  )
}

export default Trades
