import { useEffect, useState, type ReactNode } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import {
  Activity,
  Bot,
  ChartCandlestick,
  Coins,
  Home,
  PanelLeftClose,
  PanelLeftOpen,
} from 'lucide-react'

import { StatusIndicator } from 'components'
import { getHealthStatus } from 'api'

const sidebarLinks: { label: string; url: string; icon: ReactNode }[] = [
  { label: 'dashboard', url: '/', icon: <Home size={18} /> },
  { label: 'bots', url: '/bots', icon: <Bot size={18} /> },
  { label: 'trades', url: '/trades', icon: <ChartCandlestick size={18} /> },
  { label: 'accounts', url: '/accounts', icon: <Coins size={18} /> },
  { label: 'monitoring', url: '/monitoring', icon: <Activity size={18} /> },
]

type SidebarProps = { isOpen: boolean; onToggle: () => void }

const Sidebar = ({ isOpen, onToggle }: SidebarProps) => {
  const navigate = useNavigate()
  const location = useLocation()

  const [healthStatus, setHealthStatus] = useState('')

  useEffect(() => {
    const fetchHealthStatus = async () => {
      try {
        const status = await getHealthStatus()
        setHealthStatus(status)
      } catch (err) {
        console.error(err)
        setHealthStatus('error')
      }
    }

    fetchHealthStatus()
    const intervalId = setInterval(fetchHealthStatus, 60 * 1000)
    return () => clearInterval(intervalId)
  }, [])

  return (
    <div
      className={`flex flex-col justify-between h-full bg-zinc-900 border-r border-zinc-800 transition-all duration-300 ${isOpen ? 'w-48' : 'w-16'}`}
    >
      {/* Logo mark */}
      <div>
        <div className="flex items-center justify-between h-14 border-b border-zinc-800 px-4 space-x-2">
          {isOpen ? (
            <span className="text-emerald-400 font-bold text-lg tracking-tight">tradr</span>
          ) : (
            <span className="text-emerald-400 font-bold text-lg tracking-tight">t</span>
          )}
          <button
            onClick={onToggle}
            className="text-zinc-400 hover:text-zinc-100 transition-colors cursor-pointer"
          >
            {isOpen ? <PanelLeftClose size={18} /> : <PanelLeftOpen size={18} />}
          </button>
        </div>

        {/* Nav links */}
        <ul className="flex flex-col gap-1 p-2 mt-2">
          {sidebarLinks.map((link) => {
            const isActive = location.pathname === link.url
            return (
              <li
                key={link.url}
                onClick={() => navigate(link.url)}
                className={`flex flex-row items-center gap-3 px-3 py-2 rounded-md cursor-pointer transition-colors
                  ${
                    isActive
                      ? 'bg-emerald-500/10 text-emerald-400'
                      : 'text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100'
                  }
                  ${!isOpen ? 'justify-center' : ''}
                `}
              >
                {isOpen && <span className="capitalize text-sm font-medium">{link.label}</span>}
                {link.icon}
              </li>
            )
          })}
        </ul>
      </div>

      {/* Footer */}
      <div className={`p-3 border-t border-zinc-800 ${!isOpen ? 'flex justify-center' : ''}`}>
        <StatusIndicator label={isOpen ? 'api health' : ''} status={healthStatus} />
      </div>
    </div>
  )
}

export default Sidebar
