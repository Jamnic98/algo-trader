import { useEffect, useState } from 'react'

import { getAllTrades } from 'api'
import { BarLoader, Heading } from 'components'
import { useAlert } from 'hooks'
import type { Trade } from 'types'

const Trades = () => {
  const { showAlert } = useAlert()
  const [trades, setTrades] = useState<Trade[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchTrades = async () => {
      try {
        const trades = await getAllTrades()
        setTrades(trades)
      } catch {
        setError('Failed to load trades')
        showAlert({
          title: 'Failed to load trades',
          type: 'error',
        })
      } finally {
        setLoading(false)
      }
    }
    fetchTrades()
  }, [showAlert])

  if (loading) return <BarLoader fullscreen />
  if (error) return <div>{error}</div>

  return (
    <div className="space-y-8">
      <Heading title="Trades" />
      <p className="text-gray-500 underline">Last 5 Trades</p>
      {trades.length > 0 ? (
        <ul className="space-y-8 text-gray-500">
          {trades
            .slice(-5, trades.length - 1)
            .reverse()
            .map((trade, index) => (
              <li key={index} className="space-y-1">
                <div>Bot ID: {trade.botID}</div>
                <div>Symbol: {trade.symbol}</div>
                <div>Side: {trade.side}</div>
                <div>Quantity: {trade.quantity}</div>
                <div>Price: {trade.price}</div>
                <div>Fee: {trade.fee}</div>
                <div>Fee Asset: {trade.feeAsset}</div>
                <div>Timestamp: {trade.timestamp.toLocaleString()}</div>
              </li>
            ))}
        </ul>
      ) : (
        <div>No trades yet</div>
      )}
    </div>
  )
}

export default Trades
