import { useEffect, useState } from 'react'

import { type Trade, getAllTrades } from 'api'

const Trades = () => {
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
      } finally {
        setLoading(false)
      }
    }

    fetchTrades()
  }, [])

  if (loading) return <div>Loading trades...</div>
  if (error) return <div>{error}</div>

  return (
    <div>
      <h1>Trades Page</h1>
      {trades.length > 0 ? (
        <ul>
          {trades.map((trade, index) => (
            <li key={index} className="mb-8">
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
