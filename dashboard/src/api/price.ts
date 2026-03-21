import { api } from 'api'

const priceEndpoint = '/price'

export const getPrice = async (symbol: string): Promise<number> => {
  const res = await api.fetchJson<{ price: string }>(`${priceEndpoint}?symbol=${symbol}`)
  return parseFloat(res.price)
}
