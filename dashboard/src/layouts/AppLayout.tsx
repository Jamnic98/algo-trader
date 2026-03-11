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
    <div className="flex flex-row h-screen bg-surface-page overflow-x-auto [&::-webkit-scrollbar]:h-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-300 [&::-webkit-scrollbar-thumb]:rounded-full hover:[&::-webkit-scrollbar-thumb]:bg-gray-400">
      <Sidebar isOpen={isOpen} onToggle={() => setIsOpen((open) => !open)} />
      <main className="flex flex-col grow w-full mx-auto py-8 px-8 overflow-y-auto">
        <Breadcrumbs />
        <Outlet />
      </main>
    </div>
  )
}

export default AppLayout
