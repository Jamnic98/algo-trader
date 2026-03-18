import { Plus } from 'lucide-react'

import type { CreateBot, CreateBotStrategy } from 'types'
import { AVAILABLE_STRATEGIES, EXCHANGE_CONFIG, STRATEGY_CONFIG } from 'utils'

type BotFormProps = {
  form: CreateBot
  lookbackCandles: number
  onLookbackChange: (n: number) => void
  onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void
  onModeToggle: () => void
  onStrategyChange: (strategy: CreateBotStrategy) => void
  onSubmit: (e: React.SubmitEvent<HTMLFormElement>) => void
}

const BotForm = ({
  form,
  lookbackCandles,
  onLookbackChange,
  onChange,
  onModeToggle,
  onStrategyChange,
  onSubmit,
}: BotFormProps) => {
  const availableAssetTypes = EXCHANGE_CONFIG[form.exchange]?.assetTypes ?? []
  const strategyConfig = STRATEGY_CONFIG[form.strategy.name] ?? {}

  return (
    <form onSubmit={onSubmit} className="flex flex-wrap items-end gap-4">
      {/* Bot config */}
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

      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">Exchange</label>
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

      {form.assetType === 'crypto' && (
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">Symbol</label>
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
            <span className="text-content-secondary/40 px-0.5">/</span>
            <input
              name="quote"
              value={form.quote}
              onChange={onChange}
              placeholder="USDT"
              className="bg-transparent px-3 py-1.5 w-16 text-center text-content-primary uppercase text-sm focus:outline-none placeholder:text-content-secondary/20"
            />
          </div>
        </div>
      )}

      {/* Strategy */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">Strategy</label>
        <select
          name="strategy"
          value={form.strategy.name}
          onChange={(e) => {
            const selected = AVAILABLE_STRATEGIES.find((s) => s.value.name === e.target.value)
            if (selected) onStrategyChange(selected.value)
          }}
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
          required
        >
          {AVAILABLE_STRATEGIES.map(({ label, value }) => (
            <option key={value.name} value={value.name}>
              {label}
            </option>
          ))}
        </select>
      </div>

      {/* Strategy params — divider + fields only if any are bot-controlled */}
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
                // placeholder="0.001"
                className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded w-24 text-sm focus:outline-none focus:border-accent transition-colors placeholder:text-content-secondary/20"
                required
              />
            </div>
          )}

          {!strategyConfig.managesLookback && (
            <div className="flex flex-col gap-1">
              <label className="text-content-secondary text-xs uppercase tracking-wider">
                Lookback
              </label>
              <input
                type="number"
                min={1}
                value={lookbackCandles}
                onChange={(e) => onLookbackChange(Number(e.target.value))}
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
                {['1m', '5m', '15m', '1h', '4h', '1d'].map((i) => (
                  <option key={i} value={i}>
                    {i}
                  </option>
                ))}
              </select>
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
  )
}

export default BotForm
