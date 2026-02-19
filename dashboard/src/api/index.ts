export * from './bots'
export * from './trades'

class ApiError extends Error {
  status?: number

  constructor(message: string, status?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export const apiFetch = async <T>(url: string, options?: RequestInit): Promise<T> => {
  let response: Response

  try {
    response = await fetch(url, options)
  } catch {
    throw new ApiError('Network error')
  }

  if (!response.ok) {
    throw new ApiError('Request failed', response.status)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json() as Promise<T>
}
