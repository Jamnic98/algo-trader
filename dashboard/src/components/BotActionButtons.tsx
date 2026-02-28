import { ArrowDownFromLine, ArrowUpFromLine, Play, Square, X } from 'lucide-react'

type BotActionButtonsProps = {
  botId: string
  botStatus: string
  startBot: (id: string) => void
  stopBot: (id: string) => void
  attachBot: (id: string) => void
  detachBot: (id: string) => void
  deleteBot: (id: string) => void
}

const BotActionButtons = ({
  botId,
  botStatus,
  startBot,
  stopBot,
  attachBot,
  detachBot,
  deleteBot,
}: BotActionButtonsProps) => (
  <>
    {botStatus === 'created' && (
      <button
        onClick={(e) => {
          e.stopPropagation()
          attachBot(botId)
        }}
        className="bg-blue-500 hover:bg-blue-600 w-8 h-8 rounded-full flex justify-center items-center cursor-pointer transition-colors"
        title="Attach"
      >
        <ArrowUpFromLine size={16} />
      </button>
    )}

    {botStatus === 'attached' && (
      <>
        <button
          onClick={(e) => {
            e.stopPropagation()
            startBot(botId)
          }}
          className="bg-green-500 hover:bg-green-600 w-8 h-8 rounded-full flex justify-center items-center cursor-pointer transition-colors"
          title="Start"
        >
          <Play size={16} />
        </button>

        <button
          onClick={(e) => {
            e.stopPropagation()
            detachBot(botId)
          }}
          className="bg-yellow-500 hover:bg-yellow-600 w-8 h-8 rounded-full flex justify-center items-center cursor-pointer transition-colors"
          title="Detach"
        >
          <ArrowDownFromLine size={16} />
        </button>
      </>
    )}

    {botStatus === 'running' && (
      <button
        onClick={(e) => {
          e.stopPropagation()
          stopBot(botId)
        }}
        className="bg-orange-500 hover:bg-orange-600 w-8 h-8 rounded-full flex justify-center items-center cursor-pointer transition-colors"
        title="Stop"
      >
        <Square size={16} />
      </button>
    )}

    {botStatus === 'created' && (
      <button
        onClick={(e) => {
          e.stopPropagation()
          deleteBot(botId)
        }}
        className="bg-red-500 hover:bg-red-600 w-8 h-8 rounded-full flex justify-center items-center cursor-pointer transition-colors"
        title="Delete"
      >
        <X size={16} />
      </button>
    )}
  </>
)

export default BotActionButtons
