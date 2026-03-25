import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'

import { createStrategy } from 'api'
import { useStrategies, useStrategySchemas } from 'hooks'
import { BarLoader } from 'components'
import type { CreateStrategy, StrategyNode, ParamSchema } from 'types'

const defaultForm: CreateStrategy = {
  name: '',
  display_name: '',
  threshold: 0.6,
  nodes: [],
}

const ParamField = ({
  schema,
  value,
  onChange,
}: {
  schema: ParamSchema
  value: unknown
  onChange: (key: string, value: unknown) => void
}) => {
  const base =
    'border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors'

  switch (schema.type) {
    case 'number':
      return (
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">
            {schema.label}
          </label>
          <input
            type="number"
            min={schema.min}
            max={schema.max}
            value={(value as number) ?? (schema.default as number)}
            onChange={(e) => onChange(schema.key, Number(e.target.value))}
            className={`${base} w-28`}
          />
        </div>
      )
    case 'boolean':
      return (
        <div className="flex items-center gap-2">
          <label className="text-content-secondary text-xs uppercase tracking-wider">
            {schema.label}
          </label>
          <input
            type="checkbox"
            checked={(value as boolean) ?? (schema.default as boolean)}
            onChange={(e) => onChange(schema.key, e.target.checked)}
            className="accent-accent"
          />
        </div>
      )
    case 'select':
      return (
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">
            {schema.label}
          </label>
          <select
            value={(value as string) ?? (schema.default as string)}
            onChange={(e) => onChange(schema.key, e.target.value)}
            className={base}
          >
            {schema.options?.map((opt) => (
              <option key={opt} value={opt}>
                {opt}
              </option>
            ))}
          </select>
        </div>
      )
    default: // string
      return (
        <div className="flex flex-col gap-1">
          <label className="text-content-secondary text-xs uppercase tracking-wider">
            {schema.label}
          </label>
          <input
            type="text"
            value={(value as string) ?? (schema.default as string)}
            onChange={(e) => onChange(schema.key, e.target.value)}
            className={`${base} w-36`}
          />
        </div>
      )
  }
}

const StrategyCreationWizard = () => {
  const navigate = useNavigate()
  const { data: strategies, loading: loadingStrategies } = useStrategies()
  const { data: schemas, loading: loadingSchemas } = useStrategySchemas()
  const [form, setForm] = useState<CreateStrategy>(defaultForm)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const loading = loadingStrategies || loadingSchemas

  const schemaFor = (strategyId: number) => {
    const strategy = strategies?.find((s) => Number(s.id) === strategyId)
    return strategy ? schemas?.find((sc) => sc.name === strategy.name) : null
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setForm((prev) => ({
      ...prev,
      [name]: name === 'threshold' ? Number(value) : value,
    }))
  }

  const addNode = () => {
    if (!strategies || strategies.length === 0) return
    const first = strategies[0]
    const schema = schemas?.find((s) => s.name === first.name)
    const params = Object.fromEntries(schema?.params.map((p) => [p.key, p.default]) ?? [])
    setForm((prev) => ({
      ...prev,
      nodes: [...prev.nodes, { strategy_id: Number(first.id), weight: 1, params }],
    }))
  }

  const updateNode = (index: number, patch: Partial<StrategyNode>) => {
    setForm((prev) => ({
      ...prev,
      nodes: prev.nodes.map((n, i) => {
        if (i !== index) return n
        // if strategy changed, reset params to new strategy's defaults
        if (patch.strategy_id !== undefined && patch.strategy_id !== n.strategy_id) {
          const schema = schemaFor(patch.strategy_id)
          const params = Object.fromEntries(schema?.params.map((p) => [p.key, p.default]) ?? [])
          return { ...n, ...patch, params }
        }
        return { ...n, ...patch }
      }),
    }))
  }

  const updateNodeParam = (index: number, key: string, value: unknown) => {
    setForm((prev) => ({
      ...prev,
      nodes: prev.nodes.map((n, i) =>
        i === index ? { ...n, params: { ...n.params, [key]: value } } : n
      ),
    }))
  }

  const removeNode = (index: number) => {
    setForm((prev) => ({ ...prev, nodes: prev.nodes.filter((_, i) => i !== index) }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (form.nodes.length === 0) {
      setError('Add at least one sub-strategy')
      return
    }
    setSubmitting(true)
    setError(null)
    try {
      await createStrategy(form)
      navigate('/strategies')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create strategy')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <BarLoader />

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-6 max-w-lg">
      {/* Name */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">Name</label>
        <input
          name="name"
          value={form.name}
          onChange={handleChange}
          placeholder="my_strategy"
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
          required
        />
      </div>

      {/* Display Name */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">
          Display Name
        </label>
        <input
          name="display_name"
          value={form.display_name}
          onChange={handleChange}
          placeholder="My Strategy"
          className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors"
          required
        />
      </div>

      {/* Threshold */}
      <div className="flex flex-col gap-1">
        <label className="text-content-secondary text-xs uppercase tracking-wider">
          Threshold{' '}
          <span className="text-accent font-mono">{Number(form.threshold).toFixed(2)}</span>
        </label>
        <input
          type="range"
          name="threshold"
          min={0}
          max={1}
          step={0.05}
          value={form.threshold}
          onChange={handleChange}
          className="accent-accent"
        />
        <span className="text-content-secondary text-xs">
          Minimum weighted agreement to trigger a trade
        </span>
      </div>

      {/* Sub-strategies */}
      <div className="flex flex-col gap-3">
        <label className="text-content-secondary text-xs uppercase tracking-wider">
          Sub-strategies
        </label>

        {form.nodes.length === 0 && (
          <p className="text-content-secondary text-sm">No sub-strategies added yet.</p>
        )}

        {form.nodes.map((node, i) => {
          const schema = schemaFor(node.strategy_id)
          return (
            <div key={i} className="flex flex-col gap-2 border border-border rounded p-3">
              <div className="flex items-center gap-2">
                {/* Strategy picker */}
                <select
                  value={node.strategy_id}
                  onChange={(e) => updateNode(i, { strategy_id: Number(e.target.value) })}
                  className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded text-sm focus:outline-none focus:border-accent transition-colors flex-1"
                >
                  {strategies?.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.display_name}
                    </option>
                  ))}
                </select>

                {/* Weight */}
                <input
                  type="number"
                  min={0}
                  max={1}
                  step={0.1}
                  value={node.weight}
                  onChange={(e) => updateNode(i, { weight: Number(e.target.value) })}
                  className="border border-border bg-surface-secondary text-content-primary px-3 py-1.5 rounded w-20 text-sm focus:outline-none focus:border-accent transition-colors"
                />

                <button
                  type="button"
                  onClick={() => removeNode(i)}
                  className="text-content-secondary hover:text-red-400 transition-colors"
                >
                  <Trash2 size={16} />
                </button>
              </div>

              {/* Dynamic param fields */}
              {schema && schema.params.length > 0 && (
                <div className="flex flex-wrap gap-3 pt-1">
                  {schema.params.map((p) => (
                    <ParamField
                      key={p.key}
                      schema={p}
                      value={node.params[p.key]}
                      onChange={(key, val) => updateNodeParam(i, key, val)}
                    />
                  ))}
                </div>
              )}
            </div>
          )
        })}

        <button
          type="button"
          onClick={addNode}
          className="flex items-center gap-1 text-accent text-sm hover:opacity-80 transition-opacity w-fit"
        >
          <Plus size={14} /> Add sub-strategy
        </button>
      </div>

      {error && <p className="text-red-400 text-sm">{error}</p>}

      <button
        type="submit"
        disabled={submitting}
        className="flex items-center gap-2 px-3 py-1.5 rounded bg-accent text-surface-primary font-semibold text-sm uppercase tracking-wider cursor-pointer hover:opacity-90 active:scale-95 transition-all disabled:opacity-50 w-fit"
      >
        {submitting ? 'Creating...' : 'Create'} <Plus size={16} />
      </button>
    </form>
  )
}

export default StrategyCreationWizard
