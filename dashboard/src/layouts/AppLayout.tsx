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
    <div className="flex flex-row h-screen bg-surface-page">
      <Sidebar isOpen={isOpen} onToggle={() => setIsOpen((open) => !open)} />
      <div className="flex flex-col grow overflow-y-auto [&::-webkit-scrollbar]:h-1 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-300 [&::-webkit-scrollbar-thumb]:rounded-full hover:[&::-webkit-scrollbar-thumb]:bg-gray-400">
        <main className="flex flex-col w-full max-w-7xl py-8 px-8 mx-auto">
          <Breadcrumbs />
          <Outlet />
        </main>
      </div>
    </div>
  )
}

export default AppLayout
