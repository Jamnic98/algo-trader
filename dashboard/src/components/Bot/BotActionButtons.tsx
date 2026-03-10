import { ArrowDownFromLine, ArrowUpFromLine, Play, Square, X } from 'lucide-react'

import type { LoadingAction } from 'types'

type BotActionButtonsProps = {
  botId: string
  botStatus: string
  loadingAction: LoadingAction
  startBot: (id: string) => void
  stopBot: (id: string) => void
  attachBot: (id: string) => void
  detachBot: (id: string) => void
  deleteBot: (id: string) => void
}

type ActionButtonProps = {
  onClick: (e: React.MouseEvent) => void
  className: string
  title: string
  loading: boolean
  children: React.ReactNode
}

const ActionButton = ({ onClick, className, title, loading, children }: ActionButtonProps) => (
  <button
    onClick={onClick}
    disabled={loading}
    className={`relative w-8 h-8 rounded-full flex justify-center items-center transition-colors ${
      loading ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'
    } ${className}`}
    title={title}
  >
    {loading && (
      <span className="absolute inset-0 rounded-full border-2 border-transparent border-t-white/60 animate-spin" />
    )}
    {children}
  </button>
)

const BotActionButtons = ({
  botId,
  botStatus,
  loadingAction,
  startBot,
  stopBot,
  attachBot,
  detachBot,
  deleteBot,
}: BotActionButtonsProps) => (
  <>
    {botStatus === 'created' && (
      <ActionButton
        onClick={(e) => {
          e.stopPropagation()
          attachBot(botId)
        }}
        className="bg-action-attach hover:bg-action-attach-hover"
        title="Attach"
        loading={loadingAction === 'attach'}
      >
        <ArrowUpFromLine size={16} />
      </ActionButton>
    )}

    {botStatus === 'attached' && (
      <>
        <ActionButton
          onClick={(e) => {
            e.stopPropagation()
            startBot(botId)
          }}
          className="bg-action-start hover:bg-action-start-hover"
          title="Start"
          loading={loadingAction === 'start'}
        >
          <Play size={16} />
        </ActionButton>

        <ActionButton
          onClick={(e) => {
            e.stopPropagation()
            detachBot(botId)
          }}
          className="bg-action-detach hover:bg-action-detach-hover"
          title="Detach"
          loading={loadingAction === 'detach'}
        >
          <ArrowDownFromLine size={16} />
        </ActionButton>
      </>
    )}

    {botStatus === 'running' && (
      <ActionButton
        onClick={(e) => {
          e.stopPropagation()
          stopBot(botId)
        }}
        className="bg-action-stop hover:bg-stop-hover"
        title="Stop"
        loading={loadingAction === 'stop'}
      >
        <Square size={16} />
      </ActionButton>
    )}

    {botStatus === 'created' && (
      <ActionButton
        onClick={(e) => {
          e.stopPropagation()
          deleteBot(botId)
        }}
        className="bg-action-delete hover:bg-action-delete-hover"
        title="Delete"
        loading={loadingAction === 'delete'}
      >
        <X size={16} />
      </ActionButton>
    )}
  </>
)

export default BotActionButtons
