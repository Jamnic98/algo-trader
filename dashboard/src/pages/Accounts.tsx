import { useCallback, useEffect, useState } from 'react'

import { BarLoader, Heading, SectionLabel } from 'components'
import {
  createPaperAccount,
  getLiveAccount,
  listAccounts,
  resetPaperAccount,
  syncLiveAccount,
  topUpPaperAccount,
} from 'api'
import { useAlert } from 'hooks'
import type { AccountInfo, LiveAccountResponse } from 'types'

const inputCls =
  'border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors'

const btnCls =
  'px-3 py-1.5 rounded bg-accent text-surface-primary font-semibold text-sm uppercase tracking-wider cursor-pointer hover:opacity-90 active:scale-95 transition-all disabled:opacity-50'

const Accounts = () => {
  const { showAlert } = useAlert()

  const [accounts, setAccounts] = useState<AccountInfo[] | null>(null)
  const [live, setLive] = useState<LiveAccountResponse | null>(null)
  const [loading, setLoading] = useState(true)

  // new paper account form
  const [newName, setNewName] = useState('')
  const [newBalance, setNewBalance] = useState('10000')
  const [newFee, setNewFee] = useState('0.001')

  // per-account inline action amounts
  const [topUpAmounts, setTopUpAmounts] = useState<Record<string, string>>({})
  const [resetBalances, setResetBalances] = useState<Record<string, string>>({})

  const reload = useCallback(async () => {
    try {
      const [accs, liveAcc] = await Promise.all([listAccounts(), getLiveAccount()])
      setAccounts(accs)
      setLive(liveAcc)
    } catch (err) {
      console.error(err)
      showAlert({ type: 'error', title: 'Failed to load accounts' })
    } finally {
      setLoading(false)
    }
  }, [showAlert])

  useEffect(() => {
    reload()
    const t = setInterval(reload, 10_000)
    return () => clearInterval(t)
  }, [reload])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await createPaperAccount({ name: newName, balance: newBalance, fee: newFee })
      showAlert({ type: 'success', title: `Created paper account ${newName}` })
      setNewName('')
      await reload()
    } catch (err) {
      showAlert({ type: 'error', title: `Create failed: ${(err as Error).message}` })
    }
  }

  const handleTopUp = async (name: string) => {
    const amount = topUpAmounts[name]
    if (!amount) return
    try {
      await topUpPaperAccount(name, { amount })
      setTopUpAmounts((p) => ({ ...p, [name]: '' }))
      await reload()
    } catch (err) {
      showAlert({ type: 'error', title: `Top up failed: ${(err as Error).message}` })
    }
  }

  const handleReset = async (name: string) => {
    try {
      await resetPaperAccount(name, { balance: resetBalances[name] || undefined })
      showAlert({ type: 'success', title: `Reset ${name}` })
      await reload()
    } catch (err) {
      showAlert({ type: 'error', title: `Reset failed: ${(err as Error).message}` })
    }
  }

  const handleSyncLive = async () => {
    try {
      await syncLiveAccount()
      showAlert({ type: 'success', title: 'Live account synced from the exchange' })
      await reload()
    } catch (err) {
      showAlert({ type: 'error', title: `Sync failed: ${(err as Error).message}` })
    }
  }

  if (loading) return <BarLoader fullscreen />

  const paper = accounts?.filter((a) => a.type === 'paper') ?? []
  const liveAccount = accounts?.find((a) => a.type === 'live')

  return (
    <div className="space-y-8">
      <Heading title="Accounts" />

      {/* Live account */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <SectionLabel>Live account</SectionLabel>
          <button onClick={handleSyncLive} className={btnCls}>
            Sync live
          </button>
        </div>
        <div className="bg-surface-secondary border border-table-border rounded-lg px-5 py-4 flex flex-col gap-2">
          <div className="flex gap-8">
            <div>
              <span className="text-content-secondary text-[11px] font-mono uppercase tracking-widest">
                Ledger balance
              </span>
              <p className="text-lg font-semibold text-accent font-mono">
                {liveAccount?.balance ?? '—'} USDT
              </p>
            </div>
            <div>
              <span className="text-content-secondary text-[11px] font-mono uppercase tracking-widest">
                Ledger positions
              </span>
              <p className="text-sm font-mono text-content-primary">
                {Object.keys(liveAccount?.positions ?? {}).length
                  ? Object.entries(liveAccount!.positions)
                      .map(([s, q]) => `${s} ${q}`)
                      .join(', ')
                  : 'none'}
              </p>
            </div>
          </div>
          {live?.exchange_balances && (
            <div className="text-xs text-content-secondary font-mono">
              exchange: {Object.entries(live.exchange_balances).map(([a, v]) => `${a} ${v}`).join(', ')}
            </div>
          )}
          {live?.exchange_error && (
            <div className="text-xs text-red-400 font-mono">
              exchange fetch failed: {live.exchange_error} (set BINANCE_API_KEY/BINANCE_SECRET_KEY)
            </div>
          )}
        </div>
      </div>

      {/* Paper accounts */}
      <div className="space-y-3">
        <SectionLabel>Paper accounts</SectionLabel>
        <div className="overflow-x-auto border border-table-border rounded-lg">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-content-secondary text-xs uppercase tracking-wider text-left border-b border-table-border">
                <th className="px-3 py-2">Name</th>
                <th className="px-3 py-2">Balance</th>
                <th className="px-3 py-2">Fee</th>
                <th className="px-3 py-2">Positions</th>
                <th className="px-3 py-2">Top up</th>
                <th className="px-3 py-2">Reset</th>
              </tr>
            </thead>
            <tbody>
              {paper.map((a) => (
                <tr key={a.name} className="border-b border-table-border last:border-0">
                  <td className="px-3 py-2 font-mono">{a.name}</td>
                  <td className="px-3 py-2 font-mono text-accent">{a.balance}</td>
                  <td className="px-3 py-2 font-mono">{a.fee}</td>
                  <td className="px-3 py-2 font-mono text-xs">
                    {Object.keys(a.positions).length
                      ? Object.entries(a.positions).map(([s, q]) => `${s} ${q}`).join(', ')
                      : '—'}
                  </td>
                  <td className="px-3 py-2">
                    <div className="flex gap-1">
                      <input
                        className={`${inputCls} w-24`}
                        placeholder="amount"
                        value={topUpAmounts[a.name] ?? ''}
                        onChange={(e) =>
                          setTopUpAmounts((p) => ({ ...p, [a.name]: e.target.value }))
                        }
                      />
                      <button className={btnCls} onClick={() => handleTopUp(a.name)}>
                        +
                      </button>
                    </div>
                  </td>
                  <td className="px-3 py-2">
                    <div className="flex gap-1">
                      <input
                        className={`${inputCls} w-24`}
                        placeholder="new balance"
                        value={resetBalances[a.name] ?? ''}
                        onChange={(e) =>
                          setResetBalances((p) => ({ ...p, [a.name]: e.target.value }))
                        }
                      />
                      <button
                        className="px-3 py-1.5 rounded bg-surface-secondary border border-border text-content-primary text-sm cursor-pointer hover:opacity-80"
                        onClick={() => handleReset(a.name)}
                      >
                        Reset
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Create paper account */}
        <form onSubmit={handleCreate} className="flex flex-wrap items-end gap-3">
          <div className="flex flex-col gap-1">
            <label className="text-content-secondary text-xs uppercase tracking-wider">Name</label>
            <input
              className={inputCls}
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="acc-2"
              required
            />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-content-secondary text-xs uppercase tracking-wider">Balance</label>
            <input
              className={`${inputCls} w-28`}
              value={newBalance}
              onChange={(e) => setNewBalance(e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-content-secondary text-xs uppercase tracking-wider">Fee</label>
            <input
              className={`${inputCls} w-24`}
              value={newFee}
              onChange={(e) => setNewFee(e.target.value)}
            />
          </div>
          <button type="submit" className={btnCls}>
            New paper account
          </button>
        </form>
      </div>
    </div>
  )
}

export default Accounts
