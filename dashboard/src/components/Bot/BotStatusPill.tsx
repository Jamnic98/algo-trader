import type { BotStatus } from 'types'

const BotStatusPill = ({ status }: { status: BotStatus }) => (
  <span
    className={`text-xs px-1.5 py-0.5 rounded-full font-mono tracking-wide ${
      status === 'running'
        ? 'bg-emerald-500/10 text-emerald-400'
        : status === 'created'
          ? 'bg-yellow-500/10 text-yellow-400'
          : 'bg-surface-secondary text-content-tertiary'
    }`}
  >
    {status}
  </span>
)

export default BotStatusPill
