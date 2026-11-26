import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import BacktestPanel from './BacktestPanel'

const runBacktestMock = vi.fn()
const runSweepMock = vi.fn()
vi.mock('api', () => ({
  runBacktest: (...args: unknown[]) => runBacktestMock(...args),
  runSweep: (...args: unknown[]) => runSweepMock(...args),
}))

describe('BacktestPanel exchange picker', () => {
  beforeEach(() => {
    runBacktestMock.mockReset()
    runSweepMock.mockReset()
    runSweepMock.mockResolvedValue({
      strategy_slug: 'simple',
      exchange: 'binance',
      symbols: ['BTCUSDT'],
      interval: '1h',
      start: 0,
      end: 1,
      count: 2,
      sort_by: 'total_return_pct',
      results: [
        {
          params: { fast: 3 },
          stats: {
            initial_balance: '1000',
            final_balance: '1100',
            total_return_pct: '10',
            max_drawdown_pct: '2',
            num_trades: 4,
            win_rate: '50',
            profit_factor: '2',
            sharpe: '1.5',
          },
          equity_curve: [],
          trades: [],
        },
        {
          params: { fast: 5 },
          stats: {
            initial_balance: '1000',
            final_balance: '1050',
            total_return_pct: '5',
            max_drawdown_pct: '3',
            num_trades: 2,
            win_rate: '50',
            profit_factor: '1.5',
            sharpe: '1.1',
          },
          equity_curve: [],
          trades: [],
        },
      ],
    })
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
    expect(container.querySelector('input[name="symbol"]') as HTMLInputElement).toHaveAttribute('placeholder', 'BTCUSDT, ETHUSDT')
    fireEvent.change(container.querySelector('select[name="exchange"]')!, { target: { value: 'alpaca' } })
    expect(container.querySelector('input[name="symbol"]') as HTMLInputElement).toHaveAttribute('placeholder', 'BTC/USD, ETHUSDT')
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

describe('BacktestPanel multi-symbol / slippage / sweep', () => {
  beforeEach(() => {
    runBacktestMock.mockReset()
    runSweepMock.mockReset()
    runBacktestMock.mockResolvedValue({
      strategy_slug: 'simple',
      exchange: 'binance',
      symbol: 'BTCUSDT',
      symbols: ['BTCUSDT', 'ETHUSDT'],
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
    runSweepMock.mockResolvedValue({
      strategy_slug: 'simple',
      symbols: ['BTCUSDT'],
      interval: '1h',
      start: 0,
      end: 1,
      count: 2,
      sort_by: 'total_return_pct',
      results: [
        {
          params: { fast: 3 },
          stats: {
            initial_balance: '1000',
            final_balance: '1100',
            total_return_pct: '10',
            max_drawdown_pct: '2',
            num_trades: 4,
            win_rate: '50',
            profit_factor: '2',
            sharpe: '1.5',
          },
          equity_curve: [],
          trades: [],
        },
        {
          params: { fast: 5 },
          stats: {
            initial_balance: '1000',
            final_balance: '1050',
            total_return_pct: '5',
            max_drawdown_pct: '3',
            num_trades: 2,
            win_rate: '50',
            profit_factor: '1.5',
            sharpe: '1.1',
          },
          equity_curve: [],
          trades: [],
        },
      ],
    })
  })

  it('sends a symbols array for comma-separated input and slippage_bps', async () => {
    const { container } = render(<BacktestPanel strategySlug="simple" />)
    fireEvent.change(container.querySelector('input[name="symbol"]') as HTMLInputElement, {
      target: { value: 'BTCUSDT, ethusdt' },
    })
    fireEvent.change(container.querySelector('input[name="slippage_bps"]') as HTMLInputElement, {
      target: { value: '25' },
    })
    fireEvent.click(screen.getByRole('button', { name: /run backtest/i }))

    await waitFor(() => expect(runBacktestMock).toHaveBeenCalled())
    const req = runBacktestMock.mock.calls[0][0] as Record<string, unknown>
    expect(req.symbols).toEqual(['BTCUSDT', 'ETHUSDT'])
    expect(req.symbol).toBeUndefined()
    expect(req.slippage_bps).toBe('25')
  })

  it('sends a single symbol (not an array) for one-symbol input', async () => {
    const { container } = render(<BacktestPanel strategySlug="simple" />)
    fireEvent.change(container.querySelector('input[name="symbol"]') as HTMLInputElement, {
      target: { value: 'BTCUSDT' },
    })
    fireEvent.click(screen.getByRole('button', { name: /run backtest/i }))

    await waitFor(() => expect(runBacktestMock).toHaveBeenCalled())
    const req = runBacktestMock.mock.calls[0][0] as Record<string, unknown>
    expect(req.symbol).toBe('BTCUSDT')
    expect(req.symbols).toBeUndefined()
    expect(req.slippage_bps).toBeUndefined() // default 0 is not sent
  })

  it('runs a sweep with the parsed param grid and renders the results table', async () => {
    const { container } = render(<BacktestPanel strategySlug="simple" />)
    fireEvent.change(container.querySelector('textarea[name="param_grid"]') as HTMLTextAreaElement, {
      target: { value: '{"fast": [3, 5]}' },
    })
    fireEvent.click(screen.getByRole('button', { name: /run sweep/i }))

    await waitFor(() => expect(runSweepMock).toHaveBeenCalled())
    const req = runSweepMock.mock.calls[0][0] as Record<string, unknown>
    expect(req.param_grid).toEqual({ fast: [3, 5] })
    expect(req.sort_by).toBe('total_return_pct')
    expect(req.strategy_slug).toBe('simple')
    expect(req.symbol).toBe('BTCUSDT')

    await waitFor(() => expect(screen.getByText(/Sweep results — 2 combos/)).toBeInTheDocument())
    expect(screen.getByText('{"fast":3}')).toBeInTheDocument()
    expect(screen.getByText('10.00%')).toBeInTheDocument()
  })

  it('shows an error for an invalid param grid instead of calling the API', async () => {
    const { container } = render(<BacktestPanel strategySlug="simple" />)
    fireEvent.change(container.querySelector('textarea[name="param_grid"]') as HTMLTextAreaElement, {
      target: { value: '{"fast": 3}' },
    })
    fireEvent.click(screen.getByRole('button', { name: /run sweep/i }))

    await waitFor(() =>
      expect(screen.getByText(/must be a non-empty array/)).toBeInTheDocument()
    )
    expect(runSweepMock).not.toHaveBeenCalled()
  })
})
