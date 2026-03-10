import { useEffect, useState } from 'react'

import { BarLoader, Heading } from 'components'
import { useAlert } from 'hooks'
import type { AccountData } from 'types'

const Accounts = () => {
  const { showAlert } = useAlert()

  const [account, setAccount] = useState<AccountData | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const source = new EventSource(
      `/api/account/stream?api_key=${import.meta.env.VITE_SERVER_API_KEY}`
    )

    source.onmessage = (e) => {
      setAccount(JSON.parse(e.data))
      setLoading(false)
    }

    source.onerror = () => {
      showAlert({ title: 'Failed to load account', type: 'error' })
      setLoading(false)
      source.close()
    }

    return () => source.close()
  }, [showAlert])

  if (loading) return <BarLoader fullscreen />
  if (!account) return null

  return (
    <div className="space-y-8">
      <Heading title="Accounts" />

      <div className="space-y-3">
        <Heading title="Balance" as={6} size={3} />
        <p className="text-lg font-semibold text-accent font-mono">{account.balance}</p>
      </div>

      <div className="space-y-3">
        <Heading title="Positions" as={6} size={3} />
        {Object.keys(account.positions).length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {Object.entries(account.positions).map(([symbol, qty]) => (
              <div
                key={symbol}
                className="flex items-center gap-2 bg-surface-secondary border border-border rounded-lg px-3 py-2 text-sm font-mono"
              >
                <span className="text-accent font-semibold">{symbol}</span>
                <span className="text-border">·</span>
                <span className="text-content-primary">{qty}</span>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-content-secondary font-semibold text-center">
            - No open positions -
          </div>
        )}
      </div>
    </div>
  )
}

export default Accounts
