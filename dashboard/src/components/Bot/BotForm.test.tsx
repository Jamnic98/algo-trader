import { fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import BotForm from './BotForm'
import type { CreateBot } from 'types'

vi.mock('hooks', () => ({
  useStrategies: () => ({ data: [{ slug: 'simple', display_name: 'Simple' }], loading: false }),
  useAccounts: () => ({
    data: [{ name: 'default', type: 'paper', balance: '10000', fee: '0.001', positions: {} }],
    loading: false,
  }),
}))

const baseForm: CreateBot = {
  mode: 'paper',
  account: 'default',
  exchange: 'binance',
  assetType: 'crypto',
  base: 'BTC',
  quote: 'USDT',
  interval: '1h',
  maxCandles: 200,
  quantity: '0.001',
  strategy_slug: 'simple',
  risk: { max_position_pct: '100', max_daily_loss_pct: '0' },
}

type FormProps = Parameters<typeof BotForm>[0]

const renderForm = (form: CreateBot = baseForm, overrides: Partial<FormProps> = {}) => {
  const props: FormProps = {
    form,
    maxCandles: 200,
    onLookbackChange: vi.fn(),
    onChange: vi.fn(),
    onModeToggle: vi.fn(),
    onStrategyChange: vi.fn(),
    onSubmit: vi.fn(),
    ...overrides,
  }
  const utils = render(
    <MemoryRouter>
      <BotForm {...props} />
    </MemoryRouter>
  )
  return { ...utils, props }
}

const exchangeSelect = (c: HTMLElement) =>
  c.querySelector('select[name="exchange"]') as HTMLSelectElement
const riskInput = (c: HTMLElement, name: string) =>
  c.querySelector(`input[name="${name}"]`) as HTMLInputElement

describe('BotForm exchange picker', () => {
  it('offers both binance and alpaca', () => {
    const { container } = renderForm()
    const options = Array.from(exchangeSelect(container).options).map((o) => o.value)
    expect(options).toEqual(['binance', 'alpaca'])
    expect(exchangeSelect(container).value).toBe('binance')
  })

  it('forwards exchange changes through onChange', () => {
    let captured: { name: string; value: string } | null = null
    const { container } = renderForm(baseForm, {
      onChange: vi.fn((e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        captured = { name: e.target.name, value: e.target.value } // read at call time
      }),
    })
    fireEvent.change(exchangeSelect(container), { target: { value: 'alpaca' } })
    expect(captured).toEqual({ name: 'exchange', value: 'alpaca' })
  })

  it('shows quote input for crypto and hides it for stocks', () => {
    const { container, rerender } = renderForm()
    expect(screen.getByPlaceholderText('USDT')).toBeInTheDocument()
    rerender(
      <MemoryRouter>
        <BotForm
          form={{ ...baseForm, exchange: 'alpaca', assetType: 'stocks', quote: undefined }}
          maxCandles={200}
          onLookbackChange={vi.fn()}
          onChange={vi.fn()}
          onModeToggle={vi.fn()}
          onStrategyChange={vi.fn()}
          onSubmit={vi.fn()}
        />
      </MemoryRouter>
    )
    expect(screen.queryByPlaceholderText('USDT')).not.toBeInTheDocument()
    expect(screen.getByPlaceholderText('AAPL')).toBeInTheDocument()
    expect(container.querySelector('input[name="quote"]')).toBeNull()
  })
})

describe('BotForm risk section', () => {
  it('renders the risk limit fields with defaults', () => {
    const { container } = renderForm({
      ...baseForm,
      risk: {
        max_position_pct: '100',
        max_daily_loss_pct: '0',
        max_drawdown_pct: '0',
        cooldown_losses: 0,
        cooldown_minutes: 0,
      },
    })
    expect(riskInput(container, 'risk.max_position_pct')).toHaveValue(100)
    expect(riskInput(container, 'risk.max_daily_loss_pct')).toHaveValue(0)
    expect(riskInput(container, 'risk.max_drawdown_pct')).toHaveValue(0)
    expect(riskInput(container, 'risk.cooldown_losses')).toHaveValue(0)
    expect(riskInput(container, 'risk.cooldown_minutes')).toHaveValue(0)
  })

  it('routes risk field changes through onChange with the nested name', () => {
    let captured: { name: string; value: string } | null = null
    const { container } = renderForm(baseForm, {
      onChange: vi.fn((e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        captured = { name: e.target.name, value: e.target.value }
      }),
    })
    fireEvent.change(riskInput(container, 'risk.max_daily_loss_pct'), {
      target: { value: '5' },
    })
    expect(captured).toEqual({ name: 'risk.max_daily_loss_pct', value: '5' })
  })
})
