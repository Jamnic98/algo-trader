import { useEffect, useRef, useState } from 'react'

import { BarLoader, CandlestickChart } from 'components'
import type { Bot, OHLCVCandle } from 'types'
import { useAlert } from 'hooks'
import { INTERVAL_TO_HOURS } from 'utils'

const candleCache = new Map<string, OHLCVCandle[]>()

type BotCandleChartProps = { bot: Bot }

const BotCandleChart = ({ bot }: BotCandleChartProps) => {
  const { id, status, base, quote } = bot

  const { showAlert } = useAlert()
  const showAlertRef = useRef(showAlert)
  const snapshotDoneRef = useRef(false)
  const candlesRef = useRef<OHLCVCandle[]>([])
  const [candles, setCandles] = useState<OHLCVCandle[]>(() => candleCache.get(id) ?? [])
  const [latestTick, setLatestTick] = useState<OHLCVCandle | null>(
    () => candleCache.get(id)?.at(-1) ?? null
  )
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
        // deduplicate by openTime
        const exists = prev.some((c) => c.openTime === mapped.openTime)
        if (exists) return prev
        const next = [...prev, mapped]
        candlesRef.current = next
        candleCache.set(id, next)
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
      showAlertRef.current({ type: 'error', title: 'Bot candles stream error' })
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
  }, [id, status, candleCache])

  useEffect(() => {
    if (status === 'created') return

    const es = new EventSource(
      `/api/bots/${id}/ticks/stream?api_key=${import.meta.env.VITE_SERVER_API_KEY}`
    )

    es.onmessage = (e) => {
      console.log('tick raw:', e.data)
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

  if (bot.started && !snapshotDone) return <BarLoader />

  return (
    <CandlestickChart
      data={candles}
      newCandle={latestTick}
      symbol={`${base}/${quote}`}
      height={400}
      initialPrice={candles ? candles[0]?.close : 0}
      intervalMinutes={Math.round((INTERVAL_TO_HOURS[bot.interval] ?? 1) * 60)}
    />
  )
}

export default BotCandleChart
