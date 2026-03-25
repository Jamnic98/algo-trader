export type ParamType = 'number' | 'string' | 'boolean' | 'select'

export type ParamSchema = {
  key: string
  label: string
  type: ParamType
  default: unknown
  min?: number
  max?: number
  options?: string[]
}

export type StrategySchema = {
  name: string
  display_name: string
  params: ParamSchema[]
}

export type StrategyNode = {
  strategy_id: number
  weight: number
  params: Record<string, unknown>
}

export type CreateStrategy = {
  name: string
  display_name: string
  threshold: number
  nodes: StrategyNode[]
}

export type CompositeConfig = {
  threshold: number
  nodes: StrategyNode[]
}

export type Strategy = {
  id: string
  name: string
  display_name: string
  is_composite: boolean
  created_at: string
  config?: CompositeConfig
}
