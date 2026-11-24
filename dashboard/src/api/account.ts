import { api } from 'api'
import type { AccountData } from 'types'

const accountEndpoint = '/account'

export const getAccount = async (): Promise<AccountData> =>
  api.fetchJson<AccountData>(accountEndpoint)
