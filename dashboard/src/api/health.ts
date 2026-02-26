import { api } from 'api'

const healthEndpoint = '/health'

export const getHealthStatus = async () =>
  (await api.fetchJson<{ status: string }>(healthEndpoint)).status
