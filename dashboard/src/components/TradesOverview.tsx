import { useEffect, useState } from 'react'

import { type Trade, getAllTrades } from 'api'

const TradesOverview = () => {
  const [trades, setTrades] = useState<Trade[] | null>(null)

  useEffect(() => {
    const fetchTrades = async () => {
      try {
        const trades = await getAllTrades()
        setTrades(trades)
      } catch (err) {
        console.error(err)
      }
    }

    fetchTrades()
  }, [])

  return (
    <div>
      <h2>TradesOverview</h2>
      {trades && (
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
      )}
    </div>
  )
}

export default TradesOverview
