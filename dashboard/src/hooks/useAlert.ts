import { createContext, useContext } from 'react'

import type { AlertData } from 'types'

export type AlertContextType = {
  alerts: AlertData[]
  showAlert: (alert: Omit<AlertData, 'id'>) => void
  removeAlert: (id: string) => void
}

export const AlertContext = createContext<AlertContextType | undefined>(undefined)

export const useAlert = () => {
  const ctx = useContext(AlertContext)
  if (!ctx) throw new Error('useAlert must be used within an AlertProvider')
  return ctx
}
