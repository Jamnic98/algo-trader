import { Alert } from 'components'

import { useAlert } from 'hooks'

type AlertBannerPositions = 'top-center' | 'bottom-right' | 'top-right' | 'top-left' | 'bottom-left'

type AlertBannerProps = {
  position?: AlertBannerPositions
}

const positionStyles: Record<AlertBannerPositions, string> = {
  'top-center': 'top-8 left-1/2 transform -translate-x-1/2',
  'bottom-right': 'bottom-8 right-4',
  'top-right': 'top-8 right-4',
  'top-left': 'top-8 left-4',
  'bottom-left': 'bottom-8 left-4',
}

const AlertBanner: React.FC<AlertBannerProps> = ({ position = 'bottom-right' }) => {
  const { alerts, removeAlert } = useAlert()

  return (
    <div
      className={`fixed z-100 space-y-2 ${positionStyles[position]}`}
      style={{ pointerEvents: 'none', maxWidth: 320 }}
    >
      {alerts.map((alert) => (
        <div key={alert.id} style={{ pointerEvents: 'auto' }}>
          <Alert
            type={alert.type}
            title={alert.title}
            message={alert?.message}
            onClose={() => removeAlert(alert.id)}
          />
        </div>
      ))}
    </div>
  )
}

export default AlertBanner
