import { Plus } from 'lucide-react'

import type { BotMode } from 'types'

type CreateBotFormData = {
  mode: BotMode
  base: string
  quote: string
  interval: string
  lookback: string
  quantity: string
}

const candleIntervals = ['1m', '5m', '15m', '1h', '4h', '1d']

type BotFormProps = {
  form: CreateBotFormData
  onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void
  onModeToggle: () => void
  onSubmit: (e: React.SubmitEvent<HTMLFormElement>) => void
}

const BotForm = ({ form, onChange, onModeToggle, onSubmit }: BotFormProps) => {
  return (
    <form onSubmit={onSubmit} className="flex flex-wrap items-end gap-4">
      {/* Mode */}
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

      {/* Symbol */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">Symbol</label>
        <div className="flex items-center border border-border rounded bg-surface-secondary focus-within:border-accent transition-colors">
          <input
            tabIndex={0}
            autoFocus
            name="base"
            value={form.base}
            onChange={onChange}
            placeholder="BTC"
            className="bg-transparent px-3 py-1.5 w-16 text-center uppercase text-content-primary focus:outline-none placeholder:text-content-secondary/10 text-sm"
            required
          />
          <span className="text-gray-500 px-0.5">| </span>
          <input
            name="quote"
            value={form.quote}
            className="bg-transparent px-3 py-1.5 w-16 text-center text-content-secondary uppercase text-sm"
            disabled
          />
        </div>
      </div>

      {/* Quantity */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">Quantity</label>
        <input
          name="quantity"
          type="number"
          value={form.quantity}
          onChange={onChange}
          placeholder="0"
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded w-28 text-sm focus:outline-none focus:border-accent transition-colors placeholder:text-content-secondary/10"
          required
        />
      </div>

      {/* Interval */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">Interval</label>
        <select
          name="interval"
          value={form.interval}
          onChange={onChange}
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
          required
        >
          {candleIntervals.map((interval) => (
            <option key={interval} value={interval}>
              {interval}
            </option>
          ))}
        </select>
      </div>

      {/* Lookback */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">Lookback</label>
        <input
          name="lookback"
          value={form.lookback}
          onChange={onChange}
          placeholder="24h"
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded w-20 text-sm focus:outline-none focus:border-accent transition-colors placeholder:text-content-secondary/10"
          required
        />
      </div>

      <button
        type="submit"
        className="flex flex-row flex-nowra gap-2 p px-2 py-1.5 rounded bg-accent text-surface-primary font-semibold text-sm uppercase tracking-wider cursor-pointer hover:opacity-90 active:scale-95 transition-all"
      >
        New <Plus size={18} />
      </button>
    </form>
  )
}
export default BotForm
