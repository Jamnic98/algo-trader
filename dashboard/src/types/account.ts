export type AccountData = {
  balance: string
  positions: Record<string, string>
}

export type AccountType = 'paper' | 'live'

export type AccountInfo = {
  name: string
  type: AccountType
  balance: string
  fee: string
  positions: Record<string, string>
}

export type AccountsResponse = { accounts: AccountInfo[] }

export type LiveAccountResponse = {
  account: AccountInfo
  exchange_balances: Record<string, string> | null
  exchange_error?: string
}

// POST /account/live/sync?exchange=X returns one account; without a param it
// syncs every registered venue and returns per-exchange results + errors.
export type SyncLiveOneResponse = { account: AccountInfo }
export type SyncLiveAllResponse = {
  accounts: Record<string, AccountInfo>
  errors: Record<string, string>
}
export type SyncLiveResponse = SyncLiveOneResponse | SyncLiveAllResponse

export type CreatePaperAccountRequest = { name: string; balance?: string; fee?: string }
export type TopUpAccountRequest = { amount: string }
export type ResetPaperAccountRequest = { balance?: string; fee?: string }
