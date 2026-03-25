import { createContext, useState, useCallback } from 'react'

type BreadcrumbOverrides = Record<string, string>

const BreadcrumbContext = createContext<{
  overrides: BreadcrumbOverrides
  setOverride: (segment: string, label: string) => void
}>({
  overrides: {},
  setOverride: () => {},
})

export const BreadcrumbProvider = ({ children }: { children: React.ReactNode }) => {
  const [overrides, setOverrides] = useState<BreadcrumbOverrides>({})

  const setOverride = useCallback((segment: string, label: string) => {
    setOverrides((prev) => (prev[segment] === label ? prev : { ...prev, [segment]: label }))
  }, [])

  return (
    <BreadcrumbContext.Provider value={{ overrides, setOverride }}>
      {children}
    </BreadcrumbContext.Provider>
  )
}

export const BreadcrumbContext_ = BreadcrumbContext
