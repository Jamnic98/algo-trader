import { useEffect, useState, useCallback } from 'react'

import { getAccount } from 'api'
import { BarLoader, Heading } from 'components'
import { useAlert } from 'hooks'
import type { AccountData } from 'types'

const POLL_INTERVAL = 5000

const Accounts = () => {
  const { showAlert } = useAlert()

  const [account, setAccount] = useState<AccountData | null>(null)
  const [loading, setLoading] = useState(true)

  const fetchAccount = useCallback(async () => {
    try {
      const data = await getAccount()
      setAccount(data)
    } catch {
      showAlert({
        title: 'Failed to load account',
        type: 'error',
      })
    } finally {
      setLoading(false)
    }
  }, [showAlert])

  useEffect(() => {
    fetchAccount()

    const interval = setInterval(fetchAccount, POLL_INTERVAL)
    return () => clearInterval(interval)
  }, [fetchAccount])

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
