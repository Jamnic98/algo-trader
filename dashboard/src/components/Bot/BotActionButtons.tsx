import { Play, Square, X } from 'lucide-react'
import type { LoadingAction } from 'types'

type ActionButtonProps = {
  onClick: (e: React.MouseEvent) => void
  variant: 'start' | 'stop' | 'delete'
  title: string
  loading: boolean
  children: React.ReactNode
}

const variantStyles = {
  start: {
    base: 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/25 hover:bg-emerald-500/25 hover:border-emerald-500/50 hover:text-emerald-300',
    ring: 'focus-visible:ring-emerald-500/40',
    spinner: 'border-t-emerald-400',
  },
  stop: {
    base: 'bg-amber-500/15 text-amber-400 border border-amber-500/25 hover:bg-amber-500/25 hover:border-amber-500/50 hover:text-amber-300',
    ring: 'focus-visible:ring-amber-500/40',
    spinner: 'border-t-amber-400',
  },
  delete: {
    base: 'bg-red-500/15 text-red-400 border border-red-500/25 hover:bg-red-500/25 hover:border-red-500/50 hover:text-red-300',
    ring: 'focus-visible:ring-red-500/40',
    spinner: 'border-t-red-400',
  },
}

const ActionButton = ({ onClick, variant, title, loading, children }: ActionButtonProps) => {
  const styles = variantStyles[variant]
  return (
    <button
      onClick={onClick}
      disabled={loading}
      title={title}
      className={`
        relative w-7 h-7 rounded-full flex items-center justify-center
        transition-all duration-150 ease-out
        focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-1 focus-visible:ring-offset-transparent
        ${styles.base} ${styles.ring}
        ${loading ? 'cursor-not-allowed opacity-50 pointer-events-none' : 'cursor-pointer active:scale-95'}
      `}
    >
      <span className="flex items-center justify-center">{children}</span>
      {loading && (
        <span
          className={`absolute -inset-px rounded-full border-2 border-transparent animate-spin ${styles.spinner}`}
        />
      )}
    </button>
  )
}

type BotActionButtonsProps = {
  botId: string
  botStatus: string
  loadingAction: LoadingAction
  startBot: (id: string) => void
  stopBot: (id: string) => void
  deleteBot: (id: string) => void
}

const BotActionButtons = ({
  botId,
  botStatus,
  loadingAction,
  startBot,
  stopBot,
  deleteBot,
}: BotActionButtonsProps) => (
  <div className="flex items-center gap-2">
    {botStatus === 'created' && (
      <>
        <ActionButton
          onClick={(e) => {
            e.stopPropagation()
            startBot(botId)
          }}
          variant="start"
          title="Start bot"
          loading={loadingAction === 'start'}
        >
          <Play size={12} strokeWidth={2} fill="currentColor" />
        </ActionButton>

        <ActionButton
          onClick={(e) => {
            e.stopPropagation()
            deleteBot(botId)
          }}
          variant="delete"
          title="Delete bot"
          loading={loadingAction === 'delete'}
        >
          <X size={12} strokeWidth={6} />
        </ActionButton>
      </>
    )}

    {botStatus === 'trading' && (
      <ActionButton
        onClick={(e) => {
          e.stopPropagation()
          stopBot(botId)
        }}
        variant="stop"
        title="Stop bot"
        loading={loadingAction === 'stop'}
      >
        <Square size={10} strokeWidth={0} fill="currentColor" />
      </ActionButton>
    )}
  </div>
)

export default BotActionButtons
