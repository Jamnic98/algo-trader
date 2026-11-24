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

export type CreatePaperAccountRequest = { name: string; balance?: string; fee?: string }
export type TopUpAccountRequest = { amount: string }
export type ResetPaperAccountRequest = { balance?: string; fee?: string }
