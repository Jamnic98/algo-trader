import { useEffect, useState, useCallback } from 'react'

import { getAccount } from 'api'
import { Heading } from 'components'
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

  if (loading) return <div>Loading account...</div>
  if (!account) return null

  return (
    <div className="space-y-8">
      <Heading title="Accounts" />

      <div className="space-y-3">
        <Heading title="Balance" as={6} size={3} />
        <p className="text-lg font-semibold">{account.balance}</p>
      </div>

      <div className="space-y-3">
        <Heading title="Positions" as={6} size={3} />
        {Object.keys(account.positions).length > 0 ? (
          <div className="overflow-x-auto [&::-webkit-scrollbar]:h-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-300 [&::-webkit-scrollbar-thumb]:rounded-full hover:[&::-webkit-scrollbar-thumb]:bg-gray-400">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200 text-left text-gray-500">
                  <th className="pb-2 pr-8 font-medium">Symbol</th>
                  <th className="pb-2 font-medium">Quantity</th>
                </tr>
              </thead>
              <tbody>
                {Object.entries(account.positions).map(([symbol, qty]) => (
                  <tr key={symbol} className="border-b border-gray-100 last:border-0">
                    <td className="py-2 pr-8 font-medium">{symbol}</td>
                    <td className="py-2 text-gray-600">{qty}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="text-gray-500 font-semibold text-center">- No open positions -</div>
        )}
      </div>
    </div>
  )
}

export default Accounts
