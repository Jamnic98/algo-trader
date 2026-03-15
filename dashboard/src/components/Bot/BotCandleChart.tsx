import { useEffect, useRef, useState } from 'react'

import { CandlestickChart } from 'components'
import type { OHLCVCandle } from 'types'
import { useAlert } from 'hooks'

type BotCandleChartProps = { id: string; symbol: string; status: string }

const BotCandleChart = ({ id, symbol, status }: BotCandleChartProps) => {
  const { showAlert } = useAlert()
  const [candles, setCandles] = useState<OHLCVCandle[]>([])
  const newCandleRef = useRef<OHLCVCandle | null>(null)

  useEffect(() => {
    if (status === 'created') return

    const es = new EventSource(
      `/api/bots/${id}/candles/stream?api_key=${import.meta.env.VITE_SERVER_API_KEY}`
    )

    es.onmessage = (e) => {
      const candle = JSON.parse(e.data)
      const mapped: OHLCVCandle = {
        open: candle.open,
        high: candle.high,
        low: candle.low,
        close: candle.close,
        volume: candle.volume,
        openTime: candle.openTime,
        closeTime: candle.closeTime,
      }

      setCandles((prev) => {
        // snapshot — initial bulk load
        if (prev.length === 0 || mapped.openTime > prev[prev.length - 1].openTime) {
          return [...prev, mapped]
        }
        return prev
      })

      newCandleRef.current = mapped
    }

    es.onerror = () => {
      const errorMsg = 'Bot candles stream error'
      console.error(errorMsg)
      showAlert({ type: 'error', title: errorMsg })
      es.close()
    }

    return () => {
      es.close()
      setCandles([])
    }
  }, [id, status, showAlert])

  if (status === 'created') return <div className="text-gray-500">Bot not attached.</div>
  if (!candles.length) return <div className="text-gray-500">No candles yet.</div>

  return <CandlestickChart data={candles} symbol={symbol} height={400} />
}

export default BotCandleChart
