import { useEffect, useState } from 'react'
import { Outlet } from 'react-router-dom'

import { Sidebar } from 'components'
import { useIsMobile } from 'hooks'

const AppLayout = () => {
  const isMobile = useIsMobile()
  const [isOpen, setIsOpen] = useState(!isMobile)

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setIsOpen(!isMobile)
  }, [isMobile])

  return (
    <div className="flex flex-row h-screen overflow-hidden bg-surface-page">
      <Sidebar isOpen={isOpen} onToggle={() => setIsOpen((open) => !open)} />
      <main className="flex flex-col justify-center-center grow overflow-y-auto w-full mx-auto mt-16 px-8 md:px-20 max-w-7xl">
        <Outlet />
      </main>
    </div>
  )
}

export default AppLayout
