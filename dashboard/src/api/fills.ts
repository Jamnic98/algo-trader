import { api } from 'api'

import type { Fill, Pagination, FillFilters } from 'types'

const fillsEndpoint = '/fills'

type FillsParams = FillFilters & { page?: number; limit?: number }

type RawFillsResponse = {
  fills: Fill[]
  total: number
  total_pages: number
  page: number
  limit: number
}

type FillsResponse = {
  fills: Fill[]
  pagination: Pagination
}

export const getAllFills = async (params: FillsParams): Promise<FillsResponse> => {
  const query = new URLSearchParams()
  const { dateFrom, dateTo, ...rest } = params

  Object.entries(rest).forEach(([k, v]) => {
    if (v !== undefined && v !== '') query.set(k, String(v))
  })
  if (dateFrom) query.set('dateFrom', new Date(dateFrom).toISOString())
  if (dateTo) query.set('dateTo', new Date(dateTo + 'T23:59:59').toISOString())

  const raw = await api.fetchJson<RawFillsResponse>(`${fillsEndpoint}?${query.toString()}`)

  return {
    fills: raw.fills,
    pagination: {
      total: raw.total,
      total_pages: raw.total_pages,
      page: raw.page,
      limit: raw.limit,
    },
  }
}
