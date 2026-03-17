import type { BotStatus } from 'types'

const BotStatusPill = ({ status }: { status: BotStatus }) => (
  <span
    className={`text-xs px-1.5 py-0.5 rounded-full font-mono ${
      status === 'running'
        ? 'bg-green-500/10 text-green-400'
        : status === 'created'
          ? 'bg-yellow-500/10 text-yellow-400'
          : 'bg-surface-secondary text-content-tertiary'
    }`}
  >
    {status}
  </span>
)

export default BotStatusPill
