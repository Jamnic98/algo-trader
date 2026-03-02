type CreateBotFormData = {
  baseAsset: string
  quoteAsset: string
  symbol: string
  interval: string
  lookback: string
  quantity: string
}

const candleIntervals = ['1m', '5m', '15m', '1h', '4h', '1d']
const defaultFormData = {
  baseAsset: '',
  quoteAsset: 'USDT',
  symbol: '',
  interval: candleIntervals[0],
  lookback: '24h',
  quantity: '0',
}

type BotFormProps = {
  form: CreateBotFormData
  onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void
  onSubmit: (e: React.SubmitEvent<HTMLFormElement>) => void
}

const BotForm = ({ form, onChange, onSubmit }: BotFormProps) => {
  return (
    <form onSubmit={onSubmit} className="gap-2 flex flex-wrap items-center">
      <div className="flex items-center gap-1">
        <label className="text-content-secondary text-sm">Symbol:</label>
        <div className="space-x-0.5">
          <input
            tabIndex={0}
            autoFocus={true}
            name="baseAsset"
            value={form.baseAsset}
            onChange={onChange}
            className="border-b border-border bg-transparent p-1 w-14 text-center uppercase text-content-primary focus:outline-none focus:border-accent"
            required
          />
          <span className="text-content-secondary">/</span>
          <input
            name="quoteAsset"
            value={form.quoteAsset}
            className="bg-transparent p-1 w-14 text-center text-content-secondary uppercase"
            disabled
          />
        </div>
      </div>

      <div className="flex items-center gap-1">
        <label className="text-content-secondary text-sm">Quantity:</label>
        <input
          name="quantity"
          type="number"
          value={form.quantity}
          onChange={onChange}
          placeholder={defaultFormData.quantity}
          className="border border-border bg-surface-secondary text-content-primary p-1 rounded w-28 focus:outline-none focus:border-accent"
          required
        />
      </div>

      <div className="flex items-center gap-1">
        <label className="text-content-secondary text-sm">Interval:</label>
        <select
          name="interval"
          value={form.interval}
          onChange={onChange}
          className="border border-border bg-surface-secondary text-content-primary p-1 rounded focus:outline-none focus:border-accent"
          required
        >
          {candleIntervals.map((interval) => (
            <option key={interval} value={interval}>
              {interval}
            </option>
          ))}
        </select>
      </div>

      <div className="flex items-center gap-1">
        <label className="text-content-secondary text-sm">Lookback:</label>
        <input
          name="lookback"
          value={form.lookback}
          onChange={onChange}
          placeholder="24h"
          className="border border-border bg-surface-secondary text-content-primary p-1 rounded w-14 focus:outline-none focus:border-accent"
          required
        />
      </div>

      <button
        type="submit"
        className="bg-accent text-surface-primary font-semibold px-3 py-1.5 rounded cursor-pointer hover:opacity-90 transition-opacity text-sm"
      >
        CREATE
      </button>
    </form>
  )
}

export default BotForm
