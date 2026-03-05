import { api } from 'api'
import type { Trade } from 'types'

const tradesEndpoint = '/trades'

export const getAllTrades = async () =>
  (await api.fetchJson<{ trades: Trade[] }>(tradesEndpoint)).trades
