import { useEffect, useRef, useState } from 'react'
import {
  createChart,
  CandlestickSeries,
  HistogramSeries,
  ColorType,
  CrosshairMode,
  type IChartApi,
  type ISeriesApi,
  type CandlestickSeriesOptions,
  type CandlestickData,
  type HistogramData,
  type DeepPartial,
  type Time,
} from 'lightweight-charts'

import type { OHLCVCandle } from 'types'
import { minutesToDurationString } from 'utils'

type CandlestickChartProps = {
  data: OHLCVCandle[]
  symbol?: string
  height?: number
  newCandle?: OHLCVCandle | null
  initialPrice?: number
  intervalMinutes: number
}

type TooltipData = {
  x: number
  y: number
  volume: number
  pctChange: number
  amountChange: number
  price: number
  time: number
}

const fmtVol = (n: number) => n.toLocaleString(undefined, { maximumFractionDigits: 3 })
const fmtTime = (t: number) =>
  new Date(t * 1000).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
const fmtChange = (n: number) => {
  if (Math.abs(n) < 0.01) return n.toFixed(4)
  return n.toFixed(2)
}
const fmtPrice = (n: number) => {
  if (n < 0.01) return n.toFixed(6)
  if (n < 1) return n.toFixed(4)
  if (n < 10) return n.toFixed(3)
  if (n < 100) return n.toFixed(2)
  return n.toFixed(0)
}
const getPricePrecision = (n: number): { precision: number; minMove: number } => {
  if (n < 0.01) return { precision: 7, minMove: 0.0000001 }
  if (n < 1) return { precision: 5, minMove: 0.00001 }
  if (n < 10) return { precision: 4, minMove: 0.0001 }
  if (n < 100) return { precision: 3, minMove: 0.001 }
  return { precision: 1, minMove: 0.1 }
}

const CandlestickChart = ({
  data,
  newCandle,
  symbol = 'CHART',
  height = 400,
  initialPrice,
  intervalMinutes,
}: CandlestickChartProps) => {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const hasSeeded = useRef(false)
  const candleSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null)
  const [tooltip, setTooltip] = useState<TooltipData | null>(null)
  const [visibleCandles, setVisibleCandles] = useState<number | null>(null)
  const [containerWidth, setContainerWidth] = useState(0)

  // Chart init — once on mount
  useEffect(() => {
    if (!chartContainerRef.current) return

    const style = getComputedStyle(document.documentElement)
    const bg = style.getPropertyValue('--color-surface-secondary').trim() || '#111'
    const bgPrimary = style.getPropertyValue('--color-surface-primary').trim() || '#0a0a0a'
    const border = style.getPropertyValue('--color-table-border').trim() || '#2a2a2a'
    const contentSecondary = style.getPropertyValue('--color-content-secondary').trim() || '#666'

    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: bg },
        textColor: contentSecondary,
        fontFamily: 'monospace',
        fontSize: 10,
      },
      grid: { vertLines: { color: 'transparent' }, horzLines: { color: border } },
      crosshair: {
        mode: CrosshairMode.Normal,
        vertLine: { color: border, labelBackgroundColor: bgPrimary },
        horzLine: { color: border, labelBackgroundColor: bgPrimary },
      },
      rightPriceScale: {
        borderColor: border,
        textColor: contentSecondary,
        scaleMargins: { top: 0.08, bottom: 0.28 },
      },
      timeScale: {
        borderColor: border,
        timeVisible: true,
        secondsVisible: false,
        fixLeftEdge: true,
        fixRightEdge: true,
      },
      handleScroll: true,
      handleScale: true,
      width: chartContainerRef.current.clientWidth,
      height,
    })

    chartRef.current = chart

    chart.timeScale().subscribeVisibleLogicalRangeChange((range) => {
      if (range) {
        setVisibleCandles(Math.round(range.to - range.from))
      }
    })

    const { precision, minMove } = getPricePrecision(initialPrice ?? 100)
    const candleSeries = chart.addSeries(CandlestickSeries, {
      upColor: '#22c55e',
      downColor: '#ef4444',
      borderUpColor: '#22c55e',
      borderDownColor: '#ef4444',
      wickUpColor: '#22c55e',
      wickDownColor: '#ef4444',
      priceFormat: { type: 'price', precision, minMove },
    } as DeepPartial<CandlestickSeriesOptions>)
    candleSeriesRef.current = candleSeries

    const volumeSeries = chart.addSeries(HistogramSeries, {
      priceFormat: { type: 'volume' },
      priceScaleId: 'volume',
    })
    chart.priceScale('volume').applyOptions({ scaleMargins: { top: 0.78, bottom: 0 } })
    volumeSeriesRef.current = volumeSeries

    chart.subscribeCrosshairMove((param) => {
      if (!param.time || !param.point || !param.seriesData) {
        setTooltip(null)
        return
      }
      const candle = param.seriesData.get(candleSeries) as CandlestickData<Time> | undefined
      const vol = param.seriesData.get(volumeSeries) as HistogramData<Time> | undefined
      if (!candle) {
        setTooltip(null)
        return
      }

      const amountChange = candle.close - candle.open
      const pctChange = ((candle.close - candle.open) / candle.open) * 100

      setTooltip({
        x: param.point.x,
        y: param.point.y,
        price: candle.close,
        volume: vol?.value ?? 0,
        amountChange,
        pctChange,
        time: param.time as number,
      })
    })

    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        chart.applyOptions({ width: entry.contentRect.width })
        setContainerWidth(entry.contentRect.width)
      }
    })
    ro.observe(chartContainerRef.current)
    setContainerWidth(chartContainerRef.current.clientWidth)

    return () => {
      ro.disconnect()
      chart.remove()
      hasSeeded.current = false
    }
  }, [height, initialPrice])

  // Seed/reseed data when it changes
  useEffect(() => {
    if (!candleSeriesRef.current || !volumeSeriesRef.current || !data.length) return

    const sorted = [...data].sort((a, b) => a.openTime - b.openTime)

    const latestPrice = sorted[sorted.length - 1].close
    const { precision, minMove } = getPricePrecision(latestPrice)

    candleSeriesRef.current.applyOptions({
      priceFormat: { type: 'price', precision, minMove },
    })

    candleSeriesRef.current.setData(
      sorted.map((c) => ({
        time: Math.floor(c.openTime / 1000) as unknown as Time,
        open: c.open,
        high: c.high,
        low: c.low,
        close: c.close,
      }))
    )

    volumeSeriesRef.current.setData(
      sorted.map((c) => ({
        time: Math.floor(c.openTime / 1000) as unknown as Time,
        value: c.volume,
        color: c.close >= c.open ? 'rgba(34,197,94,0.25)' : 'rgba(239,68,68,0.25)',
      }))
    )

    if (!hasSeeded.current) {
      chartRef.current?.timeScale().fitContent()
      hasSeeded.current = true
    }
  }, [data])

  // Live tick update
  useEffect(() => {
    if (!newCandle || !candleSeriesRef.current || !volumeSeriesRef.current) return

    candleSeriesRef.current.update({
      time: Math.floor(newCandle.openTime / 1000) as unknown as Time,
      open: newCandle.open,
      high: newCandle.high,
      low: newCandle.low,
      close: newCandle.close,
    })

    volumeSeriesRef.current.update({
      time: Math.floor(newCandle.openTime / 1000) as unknown as Time,
      value: newCandle.volume,
      color: newCandle.close >= newCandle.open ? 'rgba(34,197,94,0.25)' : 'rgba(239,68,68,0.25)',
    })
  }, [newCandle])

  const first = data.length > 0 ? data[0] : null
  const pctChangeNum =
    first && newCandle ? ((newCandle.close - first.close) / first.close) * 100 : 0
  const pctChange = Math.abs(pctChangeNum).toFixed(2)

  // price color — based on current candle (green if close >= open)
  const isUp = newCandle ? newCandle.close >= newCandle.open : true
  const priceColor = isUp ? '#22c55e' : '#ef4444'

  // pct change — based on first candle in dataset
  const isUpFromFirst = pctChangeNum >= 0
  const pctColor = isUpFromFirst ? '#22c55e' : '#ef4444'

  const tooltipIsUp = tooltip ? tooltip.amountChange >= 0 : true
  const tooltipColor = tooltipIsUp ? '#22c55e' : '#ef4444'

  return (
    <div className="bg-surface-secondary border border-table-border rounded-lg overflow-hidden">
      <div className="flex items-center justify-between px-5 py-3 border-b border-table-border">
        <div className="flex items-center gap-3">
          <span className="text-content-secondary text-[11px] font-mono uppercase tracking-widest">
            {symbol}
          </span>
          {/* OHLC numerical data */}
          {newCandle && (
            <span className="text-[11px] font-mono text-content-secondary">
              {' '}
              <span className="text-nowrap">
                O <span className="text-content-primary">${fmtPrice(newCandle.open)}</span>
              </span>
              {'  '}
              <span className="text-nowrap">
                H <span className="text-content-primary">${fmtPrice(newCandle.high)}</span>
              </span>
              {'  '}
              <span className="text-nowrap">
                L <span className="text-content-primary">${fmtPrice(newCandle.low)}</span>
              </span>
              {'  '}
              <span className="text-nowrap">
                C <span className="text-content-primary">${fmtPrice(newCandle.close)}</span>
              </span>
            </span>
          )}
        </div>
        {newCandle && (
          <div className="flex items-center gap-2">
            <span
              className="text-[20px] font-semibold tracking-tight"
              style={{ color: priceColor }}
            >
              ${fmtPrice(newCandle?.close)}
            </span>
            <span
              className="text-[11px] font-mono px-1.5 py-0.5 rounded"
              style={{
                color: pctColor,
                backgroundColor: isUpFromFirst ? 'rgba(34,197,94,0.1)' : 'rgba(239,68,68,0.1)',
              }}
            >
              {isUpFromFirst ? '+' : '-'}
              {pctChange}%
            </span>
          </div>
        )}
      </div>

      <div className="relative">
        <div
          ref={chartContainerRef}
          className="w-full [&_.tv-lightweight-charts_table_tr:last-child_td:last-child]:hidden cursor-crosshair"
          style={{ height }}
        />
        {tooltip && (
          <div
            className="absolute pointer-events-none z-10 border border-table-border rounded px-2.5 py-2 text-[11px] font-mono flex flex-col gap-1"
            style={{
              left: tooltip.x + 16,
              top: Math.max(8, tooltip.y - 60),
              backgroundColor: 'var(--color-surface-primary)',
              ...(tooltip.x > containerWidth - 160 && {
                left: 'auto',
                right: containerWidth - tooltip.x + 16,
              }),
            }}
          >
            <span className="text-content-secondary">{fmtTime(tooltip.time)}</span>
            <span style={{ color: tooltipColor }}>
              {tooltipIsUp ? '+' : ''}
              {fmtChange(tooltip.amountChange)}{' '}
              <span className="opacity-70">
                ({tooltipIsUp ? '+' : ''}
                {tooltip.pctChange.toFixed(2)}%)
              </span>
            </span>
            <span className="text-content-secondary">
              Vol <span className="text-content-primary">{fmtVol(tooltip.volume)}</span>
            </span>
          </div>
        )}
      </div>
      <div className="flex items-center gap-4 px-5 py-2 border-t border-table-border">
        {newCandle && (
          <>
            <span className="text-[10px] font-mono text-content-secondary uppercase tracking-widest">
              Vol
            </span>
            <span className="text-[11px] font-mono text-content-primary">
              {newCandle.volume.toLocaleString(undefined, { maximumFractionDigits: 3 })}
            </span>
          </>
        )}
        {visibleCandles !== null && (
          <>
            <span className="text-[10px] font-mono text-content-secondary uppercase tracking-widest">
              Lookback
            </span>
            <span className="text-[11px] font-mono text-content-primary">
              {minutesToDurationString(visibleCandles * intervalMinutes)}
            </span>
          </>
        )}
        <div className="ml-auto flex items-center gap-4">
          {visibleCandles !== null && (
            <>
              <span className="text-[10px] font-mono text-content-secondary uppercase tracking-widest">
                Zoom
              </span>
              <span className="text-[11px] font-mono text-content-primary">
                {visibleCandles} / {data.length} candles (
                {Math.round((visibleCandles / data.length) * 100)}%)
              </span>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

export default CandlestickChart
