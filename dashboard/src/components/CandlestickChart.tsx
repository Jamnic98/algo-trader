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

type CandlestickChartProps = {
  data: OHLCVCandle[]
  symbol?: string
  height?: number
  newCandle?: OHLCVCandle | null
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

const CandlestickChart = ({
  data,
  newCandle,
  symbol = 'CHART',
  height = 400,
}: CandlestickChartProps) => {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candleSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null)

  const [tooltip, setTooltip] = useState<TooltipData | null>(null)
  const [containerWidth, setContainerWidth] = useState(0)

  // Candle stream update
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

  useEffect(() => {
    if (!chartContainerRef.current) return

    const style = getComputedStyle(document.documentElement)
    const bg = style.getPropertyValue('--color-surface-secondary').trim() || '#111'
    const bgPrimary = style.getPropertyValue('--color-surface-primary').trim() || '#0a0a0a'
    const border = style.getPropertyValue('--color-table-border').trim() || '#2a2a2a'
    const contentSecondary = style.getPropertyValue('--color-content-secondary').trim() || '#666'

    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: bg || '#111114' },
        textColor: contentSecondary || '#666',
        fontFamily: 'monospace',
        fontSize: 10,
      },
      grid: {
        vertLines: { color: 'transparent' },
        horzLines: { color: border || '#1e1e22' },
      },
      crosshair: {
        mode: CrosshairMode.Normal,
        vertLine: {
          color: border || '#333',
          labelBackgroundColor: bgPrimary || '#0a0a0a',
        },
        horzLine: {
          color: border || '#333',
          labelBackgroundColor: bgPrimary || '#0a0a0a',
        },
      },
      rightPriceScale: {
        borderColor: border || '#1e1e22',
        textColor: contentSecondary || '#666',
        scaleMargins: { top: 0.08, bottom: 0.28 },
      },
      timeScale: {
        borderColor: border || '#1e1e22',
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

    const candleOptions: DeepPartial<CandlestickSeriesOptions> = {
      upColor: '#22c55e',
      downColor: '#ef4444',
      borderUpColor: '#22c55e',
      borderDownColor: '#ef4444',
      wickUpColor: '#22c55e',
      wickDownColor: '#ef4444',
    }
    const candleSeries = chart.addSeries(CandlestickSeries, candleOptions)
    candleSeriesRef.current = candleSeries

    const volumeSeries = chart.addSeries(HistogramSeries, {
      priceFormat: { type: 'volume' },
      priceScaleId: 'volume',
    })
    chart.priceScale('volume').applyOptions({
      scaleMargins: { top: 0.78, bottom: 0 },
    })
    volumeSeriesRef.current = volumeSeries

    const sorted = [...data].sort((a, b) => a.openTime - b.openTime)
    const latestClose = sorted.length > 0 ? sorted[sorted.length - 1].close : 0

    const candleData = sorted.map((c) => ({
      time: Math.floor(c.openTime / 1000) as unknown as Time,
      open: c.open,
      high: c.high,
      low: c.low,
      close: c.close,
    }))

    const volumeData = sorted.map((c) => ({
      time: Math.floor(c.openTime / 1000) as unknown as Time,
      value: c.volume,
      color: c.close >= c.open ? 'rgba(34,197,94,0.25)' : 'rgba(239,68,68,0.25)',
    }))

    candleSeries.setData(candleData)
    volumeSeries.setData(volumeData)
    chart.timeScale().fitContent()

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

      const amountChange = latestClose - candle.close
      const pctChange = ((latestClose - candle.close) / candle.close) * 100

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
    }
  }, [data, height])

  const latest = data.length > 0 ? data[data.length - 1] : null
  const isUp = latest ? latest.close >= latest.open : true
  const priceColor = isUp ? '#22c55e' : '#ef4444'
  const pctChange = latest
    ? (((latest.close - latest.open) / latest.open) * 100).toFixed(2)
    : '0.00'

  const tooltipIsUp = tooltip ? tooltip.amountChange >= 0 : true
  const tooltipColor = tooltipIsUp ? '#22c55e' : '#ef4444'

  return (
    <div className="bg-surface-secondary border border-table-border rounded-lg overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-5 py-3 border-b border-table-border">
        <div className="flex items-center gap-3">
          <span className="text-content-secondary text-[11px] font-mono uppercase tracking-widest">
            {symbol}
          </span>
          {latest && (
            <span className="text-[11px] font-mono text-content-secondary">
              O <span className="text-content-primary">${latest.open.toFixed(2)}</span>
              {'  '}H <span className="text-content-primary">${latest.high.toFixed(2)}</span>
              {'  '}L <span className="text-content-primary">${latest.low.toFixed(2)}</span>
              {'  '}C <span className="text-content-primary">${latest.close.toFixed(2)}</span>
            </span>
          )}
        </div>
        {latest && (
          <div className="flex items-center gap-2">
            <span
              className="text-[20px] font-semibold tracking-tight"
              style={{ color: priceColor }}
            >
              ${latest.close.toFixed(2)}
            </span>
            <span
              className="text-[11px] font-mono px-1.5 py-0.5 rounded"
              style={{
                color: priceColor,
                backgroundColor: isUp ? 'rgba(34,197,94,0.1)' : 'rgba(239,68,68,0.1)',
              }}
            >
              {isUp ? '+' : ''}
              {pctChange}%
            </span>
          </div>
        )}
      </div>

      {/* Chart + Tooltip */}
      <div className="relative">
        <div
          ref={chartContainerRef}
          className="w-full [&_.tv-lightweight-charts_table_tr:last-child_td:last-child]:hidden cursor-crosshair"
          style={{ height }}
        />

        {/* Tooltip */}
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
              {tooltip.amountChange.toFixed(2)}{' '}
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

      {/* Footer */}
      {latest && (
        <div className="flex items-center gap-4 px-5 py-2 border-t border-table-border">
          <span className="text-[10px] font-mono text-content-secondary uppercase tracking-widest">
            Vol
          </span>
          <span className="text-[11px] font-mono text-content-primary">
            {latest.volume.toLocaleString(undefined, { maximumFractionDigits: 3 })}
          </span>
        </div>
      )}
    </div>
  )
}

export default CandlestickChart
