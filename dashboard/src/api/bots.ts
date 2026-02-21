import { api } from 'api'

type BotStatus = 'created' | 'attached' | 'running'

type BotCreateData = {
  id?: string
  interval: string
  symbol: string
  lookback: string
}

export interface BotData extends BotCreateData {
  id: string
  status: BotStatus
  started?: string | null
}

const botsEndpoint = '/bots'

export const getAllBots = async (): Promise<BotData[]> =>
  (await api.fetch<{ bots: BotData[] }>(`${botsEndpoint}/`)).bots

export const getBot = async (id: string): Promise<BotData> =>
  (await api.fetch<{ bot: BotData }>(`${botsEndpoint}/${id}/`)).bot

export const createBot = async (botData: BotCreateData): Promise<BotData> =>
  (
    await api.fetch<{ bot: BotData }>(botsEndpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(botData),
    })
  ).bot

export const startBot = async (id: string): Promise<BotData> =>
  (await api.fetch<{ bot: BotData }>(`${botsEndpoint}/${id}/start/`, { method: 'POST' })).bot

export const stopBot = async (id: string): Promise<BotData> =>
  (await api.fetch<{ bot: BotData }>(`${botsEndpoint}/${id}/stop/`, { method: 'POST' })).bot

export const attachBot = async (id: string): Promise<BotData> =>
  (await api.fetch<{ bot: BotData }>(`${botsEndpoint}/${id}/attach/`, { method: 'POST' })).bot

export const detachBot = async (id: string): Promise<BotData> =>
  (await api.fetch<{ bot: BotData }>(`${botsEndpoint}/${id}/detach/`, { method: 'POST' })).bot

export const deleteBot = async (id: string) =>
  await api.fetch(`${botsEndpoint}/${id}/`, { method: 'DELETE' })
