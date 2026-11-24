import { Link } from 'react-router-dom'
import { Plus } from 'lucide-react'

import { BarLoader } from 'components'
import { useAccounts, useStrategies } from 'hooks'
import { lookbackToString } from 'utils'
import {
  CANDLE_INTERVALS,
  EXCHANGE_CONFIG,
  MAX_CANDLES,
  STRATEGY_CONFIG,
  getExchangeLabel,
} from 'utils/constants'
import type { CreateBot, Strategy } from 'types'

type BotFormProps = {
  form: CreateBot
  maxCandles: number
  onLookbackChange: (n: number) => void
  onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void
  onModeToggle: () => void
  onStrategyChange: (strategy: Strategy) => void
  onSubmit: (e: React.SubmitEvent<HTMLFormElement>) => void
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
  const { data: strategies, loading: loadingStrategies } = useStrategies()
  const { data: accounts } = useAccounts()
  const exchangeConfig = EXCHANGE_CONFIG[form.exchange]
  const availableAssetTypes = exchangeConfig?.assetTypes ?? ['crypto']
  const availableIntervals = exchangeConfig?.intervals ?? CANDLE_INTERVALS
  const strategyConfig = STRATEGY_CONFIG[form.strategy_slug] ?? {}

  const renderStrategyField = () => {
    if (loadingStrategies) return <BarLoader />

    if (!strategies || strategies.length === 0)
      return (
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">
            Strategy
          </label>
          <Link
            to="/strategies/create"
            className="text-accent text-sm py-1.5 hover:opacity-80 transition-opacity"
          >
            + Create a strategy
          </Link>
        </div>
      )

    return (
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">Strategy</label>
        <select
          name="strategy_slug"
          value={form.strategy_slug}
          onChange={(e) => {
            const selected = strategies?.find((s) => s.slug === e.target.value)
            if (selected) onStrategyChange(selected)
          }}
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
          required
        >
          {strategies?.map(({ slug, display_name }, i) => (
            <option key={i} value={slug}>
              {display_name}
            </option>
          ))}
        </select>
      </div>
    )
  }

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

        {/* Account */}
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">Account</label>
          <select
            name="account"
            value={form.account ?? (form.mode === 'live' ? 'live' : 'default')}
            onChange={onChange}
            className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
          >
            {form.mode === 'live' && (
              <option value="live">live ({getExchangeLabel(form.exchange)})</option>
            )}
            {(accounts ?? [])
              .filter((a) => a.type === 'paper')
              .map((a) => (
                <option key={a.name} value={a.name}>
                  {a.name} ({a.balance})
                </option>
              ))}
            {form.mode !== 'live' && (
              <option value="live" disabled>
                live (switch mode to live)
              </option>
            )}
          </select>
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
                {getExchangeLabel(exchange)}
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

        {/* Base/Quote (quote only for pairs, e.g. crypto; stocks are just the ticker) */}
        <div className="flex items-center border border-border rounded bg-surface-secondary focus-within:border-accent transition-colors">
          <input
            autoFocus
            name="base"
            value={form.base}
            onChange={onChange}
            placeholder={form.assetType === 'stocks' ? 'AAPL' : 'BTC'}
            className="bg-transparent px-3 py-1.5 w-16 text-center uppercase text-content-primary focus:outline-none placeholder:text-content-secondary/20 text-sm"
            required
          />
          {form.assetType === 'crypto' && (
            <>
              <span className="text-content-secondary px-1 text-sm select-none">/</span>
              <input
                name="quote"
                value={form.quote ?? ''}
                onChange={onChange}
                placeholder={exchangeConfig?.defaultQuote ?? 'USDT'}
                className="bg-transparent px-3 py-1.5 w-16 text-center text-content-primary uppercase text-sm focus:outline-none placeholder:text-content-secondary/20"
                required
              />
            </>
          )}
        </div>

        {/* Strategy */}
        {renderStrategyField()}

        {/* Strategy params — only shown when strategy is loaded */}
        {strategies &&
          strategies.length > 0 &&
          (!strategyConfig.managesInterval ||
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
                        max={MAX_CANDLES[form.interval] ?? 100}
                        value={maxCandles}
                        onChange={(e) => {
                          const max = MAX_CANDLES[form.interval] ?? 100
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
                        {availableIntervals.map((i) => (
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
