import { useEffect, useState } from 'react'
import { Outlet } from 'react-router-dom'

import { Breadcrumbs, Sidebar } from 'components'
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
      <main className="flex flex-col justify-center-center grow overflow-y-auto w-full mx-auto py-8 px-8 md:px-16 max-w-7xl overflow-x-auto [&::-webkit-scrollbar]:h-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-300 [&::-webkit-scrollbar-thumb]:rounded-full hover:[&::-webkit-scrollbar-thumb]:bg-gray-400">
        <Breadcrumbs />
        <Outlet />
      </main>
    </div>
  )
}

export default AppLayout
