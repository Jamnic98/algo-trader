import { useEffect, useRef, useState } from 'react'

import { BarLoader, CandlestickChart } from 'components'
import type { OHLCVCandle } from 'types'
import { useAlert } from 'hooks'

type BotCandleChartProps = { id: string; symbol: string; status: string }

const BotCandleChart = ({ id, symbol, status }: BotCandleChartProps) => {
  const { showAlert } = useAlert()
  const snapshotDoneRef = useRef(false)
  const candlesRef = useRef<OHLCVCandle[]>([])
  const [candles, setCandles] = useState<OHLCVCandle[]>([])
  const [latestTick, setLatestTick] = useState<OHLCVCandle | null>(null)
  const [snapshotDone, setSnapshotDone] = useState(false)

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
        const next = [...prev, mapped]
        candlesRef.current = next
        return next
      })
    }

    const timeout = setTimeout(() => {
      snapshotDoneRef.current = true
      setSnapshotDone(true)
    }, 10000)

    es.addEventListener('ready', () => {
      clearTimeout(timeout)
      snapshotDoneRef.current = true
      setSnapshotDone(true)
    })

    es.onerror = () => {
      showAlert({ type: 'error', title: 'Bot candles stream error' })
      es.close()
    }

    return () => {
      clearTimeout(timeout)
      es.close()
      setCandles([])
      candlesRef.current = []
      snapshotDoneRef.current = false
      setSnapshotDone(false)
    }
  }, [id, status, showAlert])

  useEffect(() => {
    if (status === 'created') return

    const es = new EventSource(
      `/api/bots/${id}/ticks/stream?api_key=${import.meta.env.VITE_SERVER_API_KEY}`
    )

    es.onmessage = (e) => {
      const candle = JSON.parse(e.data)
      if (!snapshotDoneRef.current) {
        snapshotDoneRef.current = true
        setSnapshotDone(true)
      }

      setLatestTick({
        open: candle.open,
        high: candle.high,
        low: candle.low,
        close: candle.close,
        volume: candle.volume,
        openTime: candle.openTime,
        closeTime: candle.closeTime,
      })
    }

    es.onerror = () => es.close()
    return () => {
      es.close()
      setLatestTick(null)
    }
  }, [id, status])

  if (status === 'created') return <div className="text-gray-500">Bot not attached.</div>
  if (!snapshotDone) return <BarLoader />

  return (
    <CandlestickChart
      data={candles}
      newCandle={latestTick}
      symbol={symbol}
      height={400}
      initialPrice={candles[0]?.close}
    />
  )
}

export default BotCandleChart
