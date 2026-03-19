import { useEffect, useState } from 'react'

import { BarLoader, BotForm, BotTable, Heading } from 'components'
import { getAllBots, startBot, stopBot, createBot, deleteBot } from 'api'
import { useAlert } from 'hooks'
import { MAX_CANDLES, STRATEGY_CONFIG } from 'utils'
import type { Bot, CreateBot, CreateBotStrategy, LoadingAction } from 'types'

// Strategy only has name now
const botStrategyValidators: ((s: CreateBotStrategy) => ValidationResult)[] = [
  (s) => (s.name ? { valid: true } : { valid: false, error: 'Strategy name is required' }),
]

// interval, quantity, maxCandles live on the form now
const createBotValidators: ((f: CreateBot) => ValidationResult)[] = [
  (f) => (f.base ? { valid: true } : { valid: false, error: 'Base is required' }),
  (f) => (f.exchange ? { valid: true } : { valid: false, error: 'Exchange is required' }),
  (f) => (f.mode ? { valid: true } : { valid: false, error: 'Mode is required' }),
  (f) => (f.interval ? { valid: true } : { valid: false, error: 'Interval is required' }),
  (f) => (f.quantity ? { valid: true } : { valid: false, error: 'Quantity is required' }),
  (f) => {
    for (const validator of botStrategyValidators) {
      const result = validator(f.strategy)
      if (!result.valid) return result
    }
    return { valid: true }
  },
]
const validateCreateBotForm = (formData: CreateBot): ValidationResult => {
  for (const validator of createBotValidators) {
    const result = validator(formData)
    if (!result.valid) return result // short-circuits on first failure
  }
  return { valid: true }
}

const defaultStrategy: CreateBotStrategy = {
  name: 'simple',
}
const defaultFormData: CreateBot = {
  mode: 'paper',
  exchange: 'binance',
  assetType: 'crypto',
  base: '',
  quote: 'USDT',
  interval: '1h', //1h
  maxCandles: 200,
  quantity: '0.001',
  strategy: defaultStrategy,
}

type ValidationResult = { valid: true } | { valid: false; error: string }

const Bots = () => {
  const { showAlert } = useAlert()
  const [bots, setBots] = useState<Bot[]>([])
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState(defaultFormData)
  const [botLoadingActions, setBotLoadingActions] = useState<Record<string, LoadingAction>>({})
  const [maxCandles, setMaxCandles] = useState(200)

  const setBotLoading = (id: string, action: LoadingAction) =>
    setBotLoadingActions((prev) => ({ ...prev, [id]: action }))

  // Fetch all bots on mount
  useEffect(() => {
    const fetchBots = async () => {
      try {
        const bots = await getAllBots()
        setBots(bots)
      } catch (err) {
        console.error(err)
        showAlert({
          title: 'Failed to load bots',
          type: 'error',
        })
      } finally {
        setLoading(false)
      }
    }

    fetchBots()
  }, [showAlert])

  // Handle form input change
  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target

    setForm((prev) => {
      const next = { ...prev, [name]: value }

      if (name === 'base') next.base = value.toUpperCase()
      if (name === 'quote') next.quote = value.toUpperCase()
      if (name === 'assetType' && value !== 'crypto') next.quote = undefined

      return next
    })

    if (name === 'interval') {
      const max = MAX_CANDLES[value] ?? 500
      if (maxCandles > max) setMaxCandles(max)
    }
  }

  const handleCreateBot = async (e: React.SyntheticEvent<HTMLFormElement>) => {
    e.preventDefault()

    if (maxCandles < 1) {
      showAlert({
        title: 'Invalid form',
        message: 'Candles must be at least 1',
        type: 'error',
      })
      return
    }

    const payload: CreateBot = {
      ...form,
    }
    const result = validateCreateBotForm(payload)
    if (!result.valid) {
      showAlert({ title: 'Invalid form', message: result.error, type: 'error' })
      return
    }

    try {
      const bot = await createBot(payload)
      setBots((prev) => [...prev, bot])
      setForm(defaultFormData)
      setMaxCandles(200)
    } catch (err) {
      console.error(err)
      showAlert({ title: 'Failed to create bot', type: 'error' })
    }
  }

  const handleStartBot = async (id: string) => {
    try {
      setBotLoading(id, 'start')
      const updatedBot = await startBot(id)
      setBots((prev) => prev.map((b) => (b.id === updatedBot.id ? updatedBot : b)))
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to start bot with id: ${id}`
      showAlert({ title: errorMsg, type: 'error' })
    } finally {
      setBotLoading(id, null)
    }
  }

  const handleStopBot = async (id: string) => {
    try {
      if (confirm(`Stop bot ${id}?`) === true) {
        setBotLoading(id, 'stop')
        const updatedBot = await stopBot(id)
        setBots((prev) => prev.map((b) => (b.id === updatedBot.id ? updatedBot : b)))
      }
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to stop bot with id: ${id}`
      showAlert({ title: errorMsg, type: 'error' })
    } finally {
      setBotLoading(id, null)
    }
  }

  const handleDeleteBot = async (id: string) => {
    try {
      if (confirm(`Delete bot ${id}?`) === true) {
        setBotLoading(id, 'delete')
        await deleteBot(id)
        setBots((prev) => prev.filter((b) => b.id !== id))
      }
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to delete bot with id: ${id}`
      showAlert({ title: errorMsg, type: 'error' })
    } finally {
      setBotLoading(id, null)
    }
  }

  const handleStrategyChange = (strategy: CreateBotStrategy) => {
    setForm((prev) => ({ ...prev, strategy }))
    const config = STRATEGY_CONFIG[strategy.name]
    if (config?.defaultLookback) setMaxCandles(config.defaultLookback)
  }

  if (loading) return <BarLoader fullscreen />

  return (
    <div className="space-y-8">
      <Heading title="Bots" />
      <div className="space-y-4">
        <BotForm
          form={form}
          maxCandles={maxCandles}
          onLookbackChange={setMaxCandles}
          onChange={handleChange}
          onStrategyChange={handleStrategyChange}
          onModeToggle={() =>
            setForm((f) => ({ ...f, mode: f.mode === 'live' ? 'paper' : 'live' }))
          }
          onSubmit={handleCreateBot}
        />

        {bots.length > 0 ? (
          <div className="overflow-x-auto [&::-webkit-scrollbar]:h-0.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-300 [&::-webkit-scrollbar-thumb]:rounded-full hover:[&::-webkit-scrollbar-thumb]:bg-gray-400">
            <BotTable
              bots={bots}
              botActions={{
                startBot: handleStartBot,
                stopBot: handleStopBot,
                deleteBot: handleDeleteBot,
              }}
              botLoadingActions={botLoadingActions}
            />
          </div>
        ) : (
          <div className="text-gray-500 font">No bots.</div>
        )}
      </div>
    </div>
  )
}

export default Bots
