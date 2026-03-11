import { api } from 'api'

import type { Trade, Pagination, TradeFilters } from 'types'

const tradesEndpoint = '/trades'

type TradesParams = TradeFilters & { page?: number; limit?: number }

type RawTradesResponse = {
  trades: Trade[]
  total: number
  total_pages: number
  page: number
  limit: number
}

type TradesResponse = {
  trades: Trade[]
  pagination: Pagination
}

export const getAllTrades = async (params: TradesParams): Promise<TradesResponse> => {
  const query = new URLSearchParams()
  const { dateFrom, dateTo, ...rest } = params

  Object.entries(rest).forEach(([k, v]) => {
    if (v !== undefined && v !== '') query.set(k, String(v))
  })
  if (dateFrom) query.set('dateFrom', new Date(dateFrom).toISOString())
  if (dateTo) query.set('dateTo', new Date(dateTo + 'T23:59:59').toISOString())

  const raw = await api.fetchJson<RawTradesResponse>(`${tradesEndpoint}?${query.toString()}`)

  return {
    trades: raw.trades,
    pagination: {
      total: raw.total,
      total_pages: raw.total_pages,
      page: raw.page,
      limit: raw.limit,
    },
  }
}
