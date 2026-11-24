export type AlertType = 'info' | 'success' | 'warning' | 'error'

export type AlertData = {
  id: string
  type: AlertType
  title?: string
  message?: string
}

export type AlertContextType = {
  alerts: AlertData[]
  showAlert: (alert: Omit<AlertData, 'id'>) => void
  removeAlert: (id: string) => void
}
