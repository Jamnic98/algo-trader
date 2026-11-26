import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import BacktestPanel from './BacktestPanel'

const runBacktestMock = vi.fn()
vi.mock('api', () => ({
  runBacktest: (...args: unknown[]) => runBacktestMock(...args),
}))

describe('BacktestPanel exchange picker', () => {
  beforeEach(() => {
    runBacktestMock.mockReset()
    runBacktestMock.mockResolvedValue({
      strategy_slug: 'simple',
      exchange: 'alpaca',
      symbol: 'BTC/USD',
      interval: '1h',
      start: 0,
      end: 1,
      stats: {
        initial_balance: '1000',
        final_balance: '1000',
        total_return_pct: '0',
        max_drawdown_pct: '0',
        num_trades: 0,
        win_rate: '0',
        profit_factor: null,
        sharpe: null,
      },
      equity_curve: [],
      trades: [],
    })
  })

  it('offers binance and alpaca, defaulting to binance', () => {
    const { container } = render(<BacktestPanel strategySlug="simple" />)
    const select = container.querySelector('select[name="exchange"]') as HTMLSelectElement
    expect(Array.from(select.options).map((o) => o.value)).toEqual(['binance', 'alpaca'])
    expect(select.value).toBe('binance')
  })

  it('switches symbol placeholder to the alpaca default when alpaca is picked', () => {
    const { container } = render(<BacktestPanel strategySlug="simple" />)
    expect(container.querySelector('input[name="symbol"]') as HTMLInputElement).toHaveAttribute('placeholder', 'BTCUSDT')
    fireEvent.change(container.querySelector('select[name="exchange"]')!, { target: { value: 'alpaca' } })
    expect(container.querySelector('input[name="symbol"]') as HTMLInputElement).toHaveAttribute('placeholder', 'BTC/USD')
    // alpaca interval list must not offer 3m
    const intervalOptions = Array.from(
      (container.querySelector('select[name="interval"]') as HTMLSelectElement).options
    ).map((o) => o.value)
    expect(intervalOptions).toContain('1h')
    expect(intervalOptions).not.toContain('3m')
  })

  it('sends the selected exchange in the backtest request', async () => {
    const { container } = render(<BacktestPanel strategySlug="simple" />)
    fireEvent.change(container.querySelector('select[name="exchange"]')!, { target: { value: 'alpaca' } })
    fireEvent.change(container.querySelector('input[name="symbol"]') as HTMLInputElement, { target: { value: 'BTC/USD' } })
    fireEvent.click(screen.getByRole('button', { name: /run backtest/i }))

    await waitFor(() => expect(runBacktestMock).toHaveBeenCalled())
    const req = runBacktestMock.mock.calls[0][0] as Record<string, unknown>
    expect(req.exchange).toBe('alpaca')
    expect(req.symbol).toBe('BTC/USD')
    expect(req.strategy_slug).toBe('simple')
  })
})
