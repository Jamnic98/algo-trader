import { useEffect, useRef } from 'react'
import {
  createChart,
  CandlestickSeries,
  HistogramSeries,
  ColorType,
  CrosshairMode,
  type IChartApi,
  type ISeriesApi,
  type CandlestickSeriesOptions,
  type DeepPartial,
} from 'lightweight-charts'

import type { OHLCVCandle } from 'types'

type CandlestickChartProps = {
  data: OHLCVCandle[]
  symbol?: string
  height?: number
}

const CandlestickChart = ({ data, symbol = 'CHART', height = 400 }: CandlestickChartProps) => {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candleSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null)

  useEffect(() => {
    if (!chartContainerRef.current) return

    // Read CSS vars for theming
    const style = getComputedStyle(document.documentElement)
    const bg = style.getPropertyValue('--color-surface-secondary').trim() || '#111'
    const bgPrimary = style.getPropertyValue('--color-surface-primary').trim() || '#0a0a0a'
    const border = style.getPropertyValue('--color-table-border').trim() || '#2a2a2a'
    const contentSecondary = style.getPropertyValue('--color-content-secondary').trim() || '#666'
    // const contentPrimary = style.getPropertyValue('--color-content-primary').trim() || '#e5e5e5'

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
        scaleMargins: { top: 0.08, bottom: 0.28 }, // leave room for volume
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

    // Candlestick series
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

    // Volume histogram series — overlaid on a separate price scale
    const volumeSeries = chart.addSeries(HistogramSeries, {
      priceFormat: { type: 'volume' },
      priceScaleId: 'volume',
    })
    chart.priceScale('volume').applyOptions({
      scaleMargins: { top: 0.78, bottom: 0 }, // sits in the bottom 22%
    })
    volumeSeriesRef.current = volumeSeries

    // Transform & sort data
    const sorted = [...data].sort((a, b) => a.openTime - b.openTime)

    const candleData = sorted.map((c) => ({
      time: Math.floor(c.openTime / 1000) as unknown as import('lightweight-charts').Time,
      open: c.open,
      high: c.high,
      low: c.low,
      close: c.close,
    }))

    const volumeData = sorted.map((c) => ({
      time: Math.floor(c.openTime / 1000) as unknown as import('lightweight-charts').Time,
      value: c.volume,
      color: c.close >= c.open ? 'rgba(34,197,94,0.25)' : 'rgba(239,68,68,0.25)',
    }))

    candleSeries.setData(candleData)
    volumeSeries.setData(volumeData)
    chart.timeScale().fitContent()

    // Responsive resize
    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        chart.applyOptions({ width: entry.contentRect.width })
      }
    })
    ro.observe(chartContainerRef.current)

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

      {/* Chart */}
      <div
        ref={chartContainerRef}
        className="w-full [&_.tv-lightweight-charts_table_tr:last-child_td:last-child]:hidden"
        style={{ height }}
      />

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
