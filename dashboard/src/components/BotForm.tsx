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
      <div className="flex items-center">
        <label>Symbol:</label>
        <div className="space-x-0.5">
          <input
            tabIndex={0}
            autoFocus={true}
            name="baseAsset"
            value={form.baseAsset}
            onChange={onChange}
            className="border-b p-1 w-14 text-center uppercase"
            required
          />

          <span>/</span>

          <input
            name="quoteAsset"
            value={form.quoteAsset}
            className="border-none p-1 rounded w-14 text-center text-gray-500 uppercase"
            disabled
          />
        </div>
      </div>

      <div>
        <label>Quantity:</label>
        <input
          name="quantity"
          type="number"
          value={form.quantity}
          onChange={onChange}
          placeholder={defaultFormData.quantity}
          className="border p-1 rounded w-28"
          required
        />
      </div>

      <div>
        <label>Interval:</label>
        <select
          name="interval"
          value={form.interval}
          onChange={onChange}
          className="border p-1 rounded"
          required
        >
          {candleIntervals.map((interval) => (
            <option key={interval} value={interval}>
              {interval}
            </option>
          ))}
        </select>
      </div>

      <div>
        <label>Lookback:</label>
        <input
          name="lookback"
          value={form.lookback}
          onChange={onChange}
          placeholder="24h"
          className="border p-1 rounded w-14"
          required
        />
      </div>

      <button type="submit" className="bg-green-400 p-2 rounded cursor-pointer">
        CREATE
      </button>
    </form>
  )
}

export default BotForm
