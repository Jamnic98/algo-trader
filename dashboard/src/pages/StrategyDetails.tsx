import { useEffect } from 'react'
import { /* useNavigate, */ useParams } from 'react-router-dom'

import { BarLoader, Heading } from 'components'
import { useStrategy, useStrategies, useBreadcrumbs } from 'hooks'

const StrategyDetails = () => {
  const { slug } = useParams<{ slug: string }>()
  const { setOverride } = useBreadcrumbs()
  const { data: strategy, loading, error } = useStrategy(slug!)

  useEffect(() => {
    if (strategy && slug) setOverride(slug, strategy.display_name)
  }, [strategy, slug, setOverride])

  const { data: allStrategies } = useStrategies()

  const strategyName = (strategySlug: string) =>
    allStrategies?.find((s) => s.slug === strategySlug)?.display_name ?? strategySlug

  if (loading) return <BarLoader fullscreen />
  if (error || !strategy)
    return <p className="text-red-400 text-sm">{error ?? 'Strategy not found'}</p>

  return (
    <div className="space-y-8">
      <Heading title={strategy.display_name} />

      <div className="flex flex-col gap-6 max-w-lg">
        <div className="flex flex-col gap-3">
          <Row label="Type" value={strategy.is_composite ? 'Composite' : 'Primitive'} />
          <Row label="Created" value={new Date(strategy.created_at).toLocaleString()} />
        </div>

        {strategy.is_composite && strategy.config && (
          <>
            <div className="w-full h-px bg-border" />
            <div className="flex flex-col gap-4">
              <Row
                label="Threshold"
                value={`${(strategy.config.threshold * 100).toFixed(0)}%`}
                hint="Minimum weighted agreement to trigger a trade"
              />

              <div className="flex flex-col gap-2">
                <span className="text-content-secondary text-xs uppercase tracking-wider">
                  Sub-strategies
                </span>
                {strategy.config.nodes.map((node, i) => (
                  <div
                    key={i}
                    className="flex items-center justify-between border border-border rounded px-3 py-2 text-sm"
                  >
                    <span className="text-content-primary">{strategyName(node.strategy_slug)}</span>
                    <div className="flex items-center gap-4">
                      {Object.entries(node.params ?? {}).map(([k, v]) => (
                        <span key={k} className="text-content-secondary font-mono text-xs">
                          {k}: {String(v)}
                        </span>
                      ))}
                      <span className="text-accent font-mono text-xs">
                        {(node.weight * 100).toFixed(0)}%
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  )
}

const Row = ({ label, value, hint }: { label: string; value: string; hint?: string }) => (
  <div className="flex flex-col gap-0.5">
    <span className="text-content-secondary text-xs uppercase tracking-wider">{label}</span>
    <span className="text-content-primary text-sm">{value}</span>
    {hint && <span className="text-content-secondary text-xs">{hint}</span>}
  </div>
)

export default StrategyDetails
