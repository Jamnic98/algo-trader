import { useState, useEffect } from 'react'

import { getBotPositions } from 'api'
import type { Position } from 'types'

const BotStats = ({ id, tick }: { id: string; tick: number }) => {
  const [positions, setPositions] = useState<Position[]>([])

  useEffect(() => {
    getBotPositions(id).then(setPositions)
  }, [id, tick])

  if (!positions.length) return <div className="text-gray-500">No positions yet.</div>

  return (
    <table className="w-full text-sm font-mono">
      <thead>
        <tr className="text-content-tertiary text-left">
          <th>Symbol</th>
          <th>Qty</th>
          <th>Avg Entry</th>
          <th>Total Spent</th>
          <th>Realised PnL</th>
          <th>Fees</th>
        </tr>
      </thead>
      <tbody>
        {positions.map((p) => (
          <tr key={p.symbol} className="text-content-secondary">
            <td>{p.symbol}</td>
            <td>{p.qty}</td>
            <td>{p.avg_entry}</td>
            <td>{p.total_spent}</td>
            <td className={parseFloat(p.realised_pnl) >= 0 ? 'text-green-400' : 'text-red-400'}>
              {p.realised_pnl}
            </td>
            <td>{p.total_fees}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

export default BotStats
