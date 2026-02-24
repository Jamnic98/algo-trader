import { useEffect, useState } from 'react'
import { ArrowDownFromLine, ArrowUpFromLine, Play, Square, X } from 'lucide-react'

import { getAllBots, startBot, stopBot, attachBot, detachBot, createBot, deleteBot } from 'api'
import type { BotData } from 'types'

const lookbackTimeframes = ['1m', '5m', '15m', '1h', '4h', '1d']
const defaultFormData = { symbol: 'BTCUSDT', interval: '1m', lookback: '24h' }

const Bots = () => {
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
      } catch {
        setError('Failed to load bots')
      } finally {
        setLoading(false)
      }
    }

    fetchBots()
  }, [])

  // Handle form input change
  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target
    setForm((prev) => ({ ...prev, [name]: value }))
  }

  const handleCreateBot = async (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault()

    try {
      const bot = await createBot(form)
      setBots((prev) => prev && [...prev, bot])
      // reset form
      setForm(defaultFormData)
    } catch (err) {
      console.error(err)
    }
  }

  const handleStartBot = async (id: string) => {
    try {
      const updatedBot = await startBot(id)
      setBots((prev) => prev && prev.map((b) => (b.id === updatedBot.id ? updatedBot : b)))
    } catch (err) {
      console.error(err)
    }
  }

  const handleStopBot = async (id: string) => {
    try {
      const updatedBot = await stopBot(id)
      setBots((prev) => prev && prev.map((b) => (b.id === updatedBot.id ? updatedBot : b)))
    } catch (err) {
      console.error(err)
    }
  }

  const handleAttachBot = async (id: string) => {
    try {
      const updatedBot = await attachBot(id)
      setBots((prev) => prev && prev.map((b) => (b.id === updatedBot.id ? updatedBot : b)))
    } catch (err) {
      console.error(err)
    }
  }

  const handleDetachBot = async (id: string) => {
    try {
      const updatedBot = await detachBot(id)
      setBots((prev) => prev && prev.map((b) => (b.id === updatedBot.id ? updatedBot : b)))
    } catch (err) {
      console.error(err)
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
    }
  }

  if (loading) return <div>Loading bots...</div>
  if (error) return <div>{error}</div>

  return (
    <div>
      <h1>Bots Page</h1>
      {/* Create Bot Form */}
      <form onSubmit={handleCreateBot} className="mb-4 gap-2 flex flex-wrap items-center max-w-fit">
        <div>
          <label>Symbol:</label>
          <input
            name="symbol"
            value={form.symbol}
            onChange={handleChange}
            placeholder="BTCUSDT"
            className="border p-1 rounded"
            required
          />
        </div>
        <div>
          <label>Interval:</label>
          <select
            name="interval"
            value={form.interval}
            onChange={handleChange}
            className="border p-1 rounded"
          >
            {lookbackTimeframes.map((timeframe) => (
              <option key={timeframe} value={timeframe}>
                {timeframe}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label>Lookback:</label>
          <input
            name="lookback"
            value={form.lookback}
            onChange={handleChange}
            placeholder="24h"
            className="border p-1 rounded"
            required
          />
        </div>
        <button type="submit" className="bg-green-400 p-2 rounded cursor-pointer">
          CREATE
        </button>
      </form>

      {/* Bots list */}
      {bots.length > 0 ? (
        <div>
          {bots.map((bot) => (
            <div key={bot.id} className="my-8 border p-4 rounded">
              <div>Id: {bot.id}</div>
              <div>Symbol: {bot.symbol}</div>
              <div>Started At: {bot.started ? new Date(bot.started).toLocaleString() : '-'}</div>
              <div>Status: {bot.status}</div>
              <div>Interval: {bot.interval}</div>
              <div>Lookback period: {bot.lookback}</div>

              <div className="mt-2 flex gap-2">
                {bot.status === 'created' && (
                  <button
                    onClick={() => handleAttachBot(bot.id)}
                    className="bg-blue-500 w-8 h-8 rounded-full text-sm cursor-pointer flex justify-center items-center"
                  >
                    <ArrowUpFromLine size={20} />
                  </button>
                )}

                {bot.status === 'attached' && (
                  <>
                    <button
                      onClick={() => handleStartBot(bot.id)}
                      className="bg-green-500 w-8 h-8 rounded-full text-sm cursor-pointer flex justify-center items-center"
                    >
                      <Play size={20} />
                    </button>

                    <button
                      onClick={() => handleDetachBot(bot.id)}
                      className="bg-yellow-500 w-8 h-8 rounded-full text-sm cursor-pointer flex justify-center items-center"
                    >
                      <ArrowDownFromLine size={20} />
                    </button>
                  </>
                )}

                {bot.status === 'running' && (
                  <button
                    onClick={() => handleStopBot(bot.id)}
                    className="bg-orange-500 w-8 h-8 rounded-full text-sm cursor-pointer flex justify-center items-center"
                  >
                    <Square size={20} />
                  </button>
                )}

                <button
                  onClick={() => handleDeleteBot(bot.id)}
                  className="bg-red-500 w-8 h-8 rounded-full text-sm cursor-pointer flex justify-center items-center"
                >
                  <X size={20} />
                </button>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div>No bots yet</div>
      )}
    </div>
  )
}

export default Bots
