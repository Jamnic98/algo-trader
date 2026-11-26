import { api } from 'api'
import type { Exchange } from 'types'

const priceEndpoint = '/price'

export const getPrice = async (symbol: string, exchange: Exchange = 'binance'): Promise<number> => {
  const res = await api.fetchJson<{ price: string }>(
    `${priceEndpoint}?symbol=${encodeURIComponent(symbol)}&exchange=${exchange}`
  )
  return parseFloat(res.price)
}
