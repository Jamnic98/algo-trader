import { api } from 'api'
import type { CreateBot, Bot, Pagination, Fill, Position } from 'types'

const botsEndpoint = '/bots'

export const getAllBots = async ({ deleted } = { deleted: false }): Promise<Bot[]> =>
  (await api.fetchJson<{ bots: Bot[] }>(`${botsEndpoint}${deleted ? '?deleted=true' : ''}`)).bots

export const getBot = async (id: string): Promise<Bot> =>
  (await api.fetchJson<{ bot: Bot }>(`${botsEndpoint}/${id}`)).bot

export const createBot = async (botData: CreateBot): Promise<Bot> =>
  (
    await api.fetchJson<{ bot: Bot }>(botsEndpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(botData),
    })
  ).bot

export const startBot = async (id: string): Promise<Bot> =>
  (await api.fetchJson<{ bot: Bot }>(`${botsEndpoint}/${id}/start`, { method: 'POST' })).bot

export const stopBot = async (id: string): Promise<Bot> =>
  (await api.fetchJson<{ bot: Bot }>(`${botsEndpoint}/${id}/stop`, { method: 'POST' })).bot

export const attachBot = async (id: string): Promise<Bot> =>
  (await api.fetchJson<{ bot: Bot }>(`${botsEndpoint}/${id}/attach`, { method: 'POST' })).bot

export const detachBot = async (id: string): Promise<Bot> =>
  (await api.fetchJson<{ bot: Bot }>(`${botsEndpoint}/${id}/detach`, { method: 'POST' })).bot

export const deleteBot = async (id: string) =>
  await api.fetchVoid(`${botsEndpoint}/${id}`, { method: 'DELETE' })

export const getBotFills = async (id: string, page = 1, limit = 15) =>
  await api.fetchJson<{ data: Fill[]; pagination: Pagination }>(
    `${botsEndpoint}/${id}/fills?page=${page}&limit=${limit}`,
    { method: 'GET' }
  )

export const getBotPositions = async (id: string): Promise<Position[]> => {
  const res = await api.fetchJson<{ positions: Position[] }>(`${botsEndpoint}/${id}/positions`)
  return res.positions
}
