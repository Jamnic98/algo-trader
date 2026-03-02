import { useEffect, useState } from 'react'

import { BotForm, BotTable, PageTitle } from 'components'
import { getAllBots, startBot, stopBot, attachBot, detachBot, createBot, deleteBot } from 'api'
import type { BotData } from 'types'
import { useAlert } from 'hooks'

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
  const [bots, setBots] = useState<BotData[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [form, setForm] = useState(defaultFormData)

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
      const updatedBot = await startBot(id)
      setBots((prev) => prev && prev.map((b) => (b.id === updatedBot.id ? updatedBot : b)))
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to start bot with id: ${id}`
      setError(errorMsg)
      showAlert({
        title: errorMsg,
        type: 'error',
      })
    }
  }

  const handleStopBot = async (id: string) => {
    try {
      const updatedBot = await stopBot(id)
      setBots((prev) => prev && prev.map((b) => (b.id === updatedBot.id ? updatedBot : b)))
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to stop bot with id: ${id}`
      setError(errorMsg)
      showAlert({
        title: errorMsg,
        type: 'error',
      })
    }
  }

  const handleAttachBot = async (id: string) => {
    try {
      const updatedBot = await attachBot(id)
      setBots((prev) => prev && prev.map((b) => (b.id === updatedBot.id ? updatedBot : b)))
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to attach bot with id: ${id}`
      setError(errorMsg)
      showAlert({
        title: errorMsg,
        type: 'error',
      })
    }
  }

  const handleDetachBot = async (id: string) => {
    try {
      const updatedBot = await detachBot(id)
      setBots((prev) => prev && prev.map((b) => (b.id === updatedBot.id ? updatedBot : b)))
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to detach bot with id: ${id}`
      setError(errorMsg)
      showAlert({
        title: errorMsg,
        type: 'error',
      })
    }
  }

  const handleDeleteBot = async (id: string) => {
    try {
      if (confirm(`Are you sure you want to delete bot: ${id}`) === true) {
        await deleteBot(id)
        setBots((prev) => prev && prev.filter((b) => b.id !== id))
      }
    } catch (err) {
      console.error(err)
      const errorMsg = `Failed to delete bot with id: ${id}`
      setError(errorMsg)
      showAlert({
        title: errorMsg,
        type: 'error',
      })
    }
  }

  if (loading) return <div>Loading bots...</div>

  if (error) return <div>{error}</div>

  return (
    <>
      <PageTitle title="Bots" />
      <BotForm form={form} onChange={handleChange} onSubmit={handleCreateBot} />

      {/* Bots table */}
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
          />
        </div>
      ) : (
        <div>No bots yet</div>
      )}
    </>
  )
}

export default Bots
