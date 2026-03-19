import type { TradeSignal } from 'types'

type SignalLabelProps = { signal: TradeSignal }

const SignalLabel = ({ signal }: SignalLabelProps) => {
  const lableColour = (signal: TradeSignal): string => {
    switch (signal) {
      case 'BUY':
        return 'text-green-400'
      case 'SELL':
        return 'text-red-400'
      case 'HOLD':
      default:
        return 'text-gray-400'
    }
  }

  return <span className={lableColour(signal)}>{signal}</span>
}

export default SignalLabel
