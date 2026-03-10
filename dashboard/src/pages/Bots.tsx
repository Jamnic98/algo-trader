import { useEffect, useState } from 'react'

import { BarLoader, BotForm, BotTable, Heading } from 'components'
import { getAllBots, startBot, stopBot, attachBot, detachBot, createBot, deleteBot } from 'api'
import { useAlert } from 'hooks'
import type { Bot, LoadingAction } from 'types'

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

const validateCreateBotForm = (formData: CreateBotFormData): boolean => {
  if (Number.parseFloat(formData.quantity) <= 0) {
    alert('Quantity must be greater than 0')
    return false
  }

  // Check that all symbol data filled
  if (![formData.baseAsset, formData.quoteAsset, formData.symbol].every((d) => d.trim() !== '')) {
    return false
  }

  return true
}

const Bots = () => {
  const { showAlert } = useAlert()
  const [bots, setBots] = useState<Bot[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [form, setForm] = useState(defaultFormData)
  const [botLoadingActions, setBotLoadingActions] = useState<Record<string, LoadingAction>>({})

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
        const errorMsg = 'Failed to load bots'
        setError(errorMsg)
        showAlert({
          title: errorMsg,
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
      next.symbol = `${next.baseAsset}${next.quoteAsset}`.toUpperCase()

      return next
    })
  }

  const handleCreateBot = async (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault()

    if (!validateCreateBotForm(form)) {
      return
    }

    try {
      const bot = await createBot(form)
      setBots((prev) => prev && [...prev, bot])
      // reset form
      setForm(defaultFormData)
    } catch (err) {
      console.error(err)
      const errorMsg = 'Failed to create bot'
      showAlert({
        title: errorMsg,
        type: 'error',
      })
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
      setError(errorMsg)
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
      setError(errorMsg)
      showAlert({ title: errorMsg, type: 'error' })
    } finally {
      setBotLoading(id, null)
    }
  }

  const handleAttachBot = async (id: string) => {
    try {
      setBotLoading(id, 'attach')
      const updatedBot = await attachBot(id)
      setBots((prev) => prev.map((b) => (b.id === updatedBot.id ? updatedBot : b)))
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to attach bot with id: ${id}`
      setError(errorMsg)
      showAlert({ title: errorMsg, type: 'error' })
    } finally {
      setBotLoading(id, null)
    }
  }

  const handleDetachBot = async (id: string) => {
    try {
      setBotLoading(id, 'detach')
      const updatedBot = await detachBot(id)
      setBots((prev) => prev.map((b) => (b.id === updatedBot.id ? updatedBot : b)))
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to detach bot with id: ${id}`
      setError(errorMsg)
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
      setError(errorMsg)
      showAlert({ title: errorMsg, type: 'error' })
    } finally {
      setBotLoading(id, null)
    }
  }

  if (loading) return <BarLoader fullscreen />
  if (error) return <div>{error}</div>

  return (
    <div className="space-y-8">
      <Heading title="Bots" />
      <BotForm form={form} onChange={handleChange} onSubmit={handleCreateBot} />

      {bots.length > 0 ? (
        <div className="overflow-x-auto [&::-webkit-scrollbar]:h-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-300 [&::-webkit-scrollbar-thumb]:rounded-full hover:[&::-webkit-scrollbar-thumb]:bg-gray-400">
          <BotTable
            bots={bots}
            botActions={{
              startBot: handleStartBot,
              stopBot: handleStopBot,
              attachBot: handleAttachBot,
              detachBot: handleDetachBot,
              deleteBot: handleDeleteBot,
            }}
            botLoadingActions={botLoadingActions}
          />
        </div>
      ) : (
        <div className="text-gray-500 font">No bots</div>
      )}
    </div>
  )
}

export default Bots
