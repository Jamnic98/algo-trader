import { Plus } from 'lucide-react'

import {
  candleIntervals,
  lookbackToString,
  AVAILABLE_STRATEGIES,
  EXCHANGE_CONFIG,
  MAX_CANDLES,
  STRATEGY_CONFIG,
} from 'utils'
import type { CreateBot, CreateBotStrategy } from 'types'

type BotFormProps = {
  form: CreateBot
  maxCandles: number
  onLookbackChange: (n: number) => void
  onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void
  onModeToggle: () => void
  onStrategyChange: (strategy: CreateBotStrategy) => void
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
}

const BotForm = ({
  form,
  maxCandles,
  onLookbackChange,
  onChange,
  onModeToggle,
  onStrategyChange,
  onSubmit,
}: BotFormProps) => {
  const availableAssetTypes = EXCHANGE_CONFIG[form.exchange]?.assetTypes ?? []
  const strategyConfig = STRATEGY_CONFIG[form.strategy.name] ?? {}

  return (
    <div className="flex flex-col gap-1">
      <form onSubmit={onSubmit} className="flex flex-wrap items-end gap-4">
        {/* Live toggle */}
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">Live</label>
          <div className="flex items-center py-1.5">
            <div
              onClick={onModeToggle}
              className={`relative w-11 h-6 rounded-full cursor-pointer transition-colors duration-200 ${
                form.mode === 'live' ? 'bg-accent' : 'bg-surface-secondary border border-border'
              }`}
            >
              <div
                className={`absolute top-0.5 w-5 h-5 rounded-full bg-surface-primary shadow transition-all duration-200 ${
                  form.mode === 'live' ? 'left-5.5' : 'left-0.5'
                }`}
              />
            </div>
          </div>
        </div>

        {/* Exchange */}
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">
            Exchange
          </label>
          <select
            name="exchange"
            value={form.exchange}
            onChange={onChange}
            className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
            required
          >
            {Object.keys(EXCHANGE_CONFIG).map((exchange) => (
              <option key={exchange} value={exchange}>
                {exchange}
              </option>
            ))}
          </select>
        </div>

        {/* Asset Type */}
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">
            Asset Type
          </label>
          <select
            name="assetType"
            value={form.assetType}
            onChange={onChange}
            className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
            required
          >
            {availableAssetTypes.map((assetType) => (
              <option key={assetType} value={assetType}>
                {assetType}
              </option>
            ))}
          </select>
        </div>

        {/* Base/Quote */}
        {form.assetType === 'crypto' && (
          <div className="flex items-center border border-border rounded bg-surface-secondary focus-within:border-accent transition-colors">
            <input
              autoFocus
              name="base"
              value={form.base}
              onChange={onChange}
              placeholder="BTC"
              className="bg-transparent px-3 py-1.5 w-16 text-center uppercase text-content-primary focus:outline-none placeholder:text-content-secondary/20 text-sm"
              required
            />
            <span className="text-content-secondary px-1 text-sm select-none">/</span>
            <input
              name="quote"
              value={form.quote ?? ''}
              onChange={onChange}
              placeholder="USDT"
              className="bg-transparent px-3 py-1.5 w-16 text-center text-content-primary uppercase text-sm focus:outline-none placeholder:text-content-secondary/20"
              required
            />
          </div>
        )}

        {/* Strategy */}
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">
            Strategy
          </label>
          <select
            name="strategy"
            value={form.strategy.name}
            onChange={(e) => {
              const selected = AVAILABLE_STRATEGIES.find((s) => s.name === e.target.value)
              if (selected) onStrategyChange({ name: selected.name })
            }}
            className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
            required
          >
            {AVAILABLE_STRATEGIES.map(({ label, name }) => (
              <option key={name} value={name}>
                {label}
              </option>
            ))}
          </select>
        </div>

        {/* Strategy params */}
        {(!strategyConfig.managesInterval ||
          !strategyConfig.managesLookback ||
          !strategyConfig.managesQuantity) && (
          <>
            <div className="w-px self-stretch bg-border" />

            {!strategyConfig.managesQuantity && (
              <div className="flex flex-col gap-1">
                <label className="text-content-secondary text-xs uppercase tracking-wider">
                  Quantity
                </label>
                <input
                  name="quantity"
                  value={form.quantity}
                  onChange={onChange}
                  placeholder="0.001"
                  className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded w-24 text-sm focus:outline-none focus:border-accent transition-colors placeholder:text-content-secondary/20"
                  required
                />
              </div>
            )}

            {(!strategyConfig.managesLookback || !strategyConfig.managesInterval) && (
              <div className="flex items-end gap-2">
                {!strategyConfig.managesLookback && (
                  <div className="flex flex-col gap-1">
                    <label className="text-content-secondary text-xs uppercase tracking-wider">
                      Candles
                    </label>
                    <input
                      type="number"
                      min={1}
                      max={MAX_CANDLES[form.interval] ?? 500}
                      value={maxCandles}
                      onChange={(e) => {
                        const max = MAX_CANDLES[form.interval] ?? 500
                        onLookbackChange(Math.min(Number(e.target.value), max))
                      }}
                      placeholder="200"
                      className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded w-24 text-sm focus:outline-none focus:border-accent transition-colors placeholder:text-content-secondary/20"
                      required
                    />
                  </div>
                )}

                {!strategyConfig.managesInterval && (
                  <div className="flex flex-col gap-1">
                    <label className="text-content-secondary text-xs uppercase tracking-wider">
                      Interval
                    </label>
                    <select
                      name="interval"
                      value={form.interval}
                      onChange={onChange}
                      className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
                      required
                    >
                      {candleIntervals.map((i) => (
                        <option key={i} value={i}>
                          {i}
                        </option>
                      ))}
                    </select>
                  </div>
                )}

                {!strategyConfig.managesLookback && (
                  <span className="text-accent text-xs font-mono pb-2 whitespace-nowrap">
                    ≈ {lookbackToString(maxCandles, form.interval)}
                  </span>
                )}
              </div>
            )}
          </>
        )}

        <button
          type="submit"
          className="flex flex-row flex-nowrap gap-2 px-2 py-1.5 rounded bg-accent text-surface-primary font-semibold text-sm uppercase tracking-wider cursor-pointer hover:opacity-90 active:scale-95 transition-all"
        >
          New <Plus size={18} />
        </button>
      </form>
    </div>
  )
}

export default BotForm
