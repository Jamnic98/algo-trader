import { api } from 'api'
import type {
  AccountData,
  AccountInfo,
  AccountsResponse,
  CreatePaperAccountRequest,
  LiveAccountResponse,
  ResetPaperAccountRequest,
  SyncLiveResponse,
  TopUpAccountRequest,
} from 'types'

const accountEndpoint = '/account'

export const getAccount = async (): Promise<AccountData> =>
  api.fetchJson<AccountData>(accountEndpoint)

export const listAccounts = async (): Promise<AccountInfo[]> =>
  (await api.fetchJson<AccountsResponse>(`${accountEndpoint}/all`)).accounts

export const getAccountByName = async (name: string): Promise<AccountInfo> =>
  api.fetchJson<AccountInfo>(`${accountEndpoint}/${encodeURIComponent(name)}`)

export const createPaperAccount = async (
  data: CreatePaperAccountRequest
): Promise<AccountInfo> =>
  api.fetchJson<AccountInfo>(`${accountEndpoint}/paper`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })

export const resetPaperAccount = async (
  name: string,
  data: ResetPaperAccountRequest
): Promise<AccountInfo> =>
  api.fetchJson<AccountInfo>(`${accountEndpoint}/paper/${encodeURIComponent(name)}/reset`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })

export const topUpPaperAccount = async (
  name: string,
  data: TopUpAccountRequest
): Promise<AccountInfo> =>
  api.fetchJson<AccountInfo>(`${accountEndpoint}/paper/${encodeURIComponent(name)}/topup`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })

export const getLiveAccount = async (exchange?: string): Promise<LiveAccountResponse> =>
  api.fetchJson<LiveAccountResponse>(
    `${accountEndpoint}/live${exchange ? `?exchange=${encodeURIComponent(exchange)}` : ''}`
  )

export const syncLiveAccount = async (exchange?: string): Promise<SyncLiveResponse> =>
  api.fetchJson<SyncLiveResponse>(
    `${accountEndpoint}/live/sync${exchange ? `?exchange=${encodeURIComponent(exchange)}` : ''}`,
    { method: 'POST' }
  )
