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

class ApiClient {
  private apiKey: string
  private baseUrl: string

  constructor(baseUrl: string, apiKey: string) {
    this.baseUrl = baseUrl
    this.apiKey = apiKey
  }

  async fetch<T>(path: string, options: RequestInit = {}): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers: {
        ...(options.headers ?? {}),
        Authorization: `ApiKey ${this.apiKey}`,
      },
    })

    if (!response.ok) {
      throw new ApiError(`HTTP ${response.status}`, response.status)
    }

    return response.json() as Promise<T>
  }
}

export const api = new ApiClient(import.meta.env.VITE_API_URL, import.meta.env.VITE_SERVER_API_KEY)
