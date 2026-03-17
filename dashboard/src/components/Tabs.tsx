import type { Tab } from 'types'

type TabsProps = {
  tabs: Tab[]
  activeTab: number
  onTabChange: (index: number) => void
}

const Tabs = ({ tabs, activeTab, onTabChange }: TabsProps) => {
  return (
    <div className="space-y-4">
      {/* Tab Bar */}
      <div className="flex gap-0.5 bg-surface-secondary p-1 rounded-lg w-fit overflow-x-auto border border-border select-none [&::-webkit-scrollbar]:h-1 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-gray-300 [&::-webkit-scrollbar-thumb]:rounded-full hover:[&::-webkit-scrollbar-thumb]:bg-gray-400">
        {tabs.map((tab, index) => {
          const isActive = activeTab === index
          return (
            <button
              key={tab.label}
              onClick={() => onTabChange(index)}
              className={`relative px-2 py-1.5 text-xs font-mono tracking-widest uppercase rounded-md cursor-pointer transition-all duration-200 border
                ${
                  isActive
                    ? 'bg-accent-muted text-accent border-accent/25 font-semibold'
                    : 'bg-transparent text-content-secondary border-transparent hover:text-content-primary hover:bg-white/4'
                }`}
            >
              <span className={`flex items-center gap-1.5`}>
                {tab.icon && (
                  <span
                    className={
                      isActive
                        ? 'text-accent drop-shadow-[0_0_6px_var(--color-accent)]'
                        : 'opacity-70'
                    }
                  >
                    {tab.icon}
                  </span>
                )}
                {tab.label}
              </span>
            </button>
          )
        })}
      </div>

      {/* Divider */}
      <div className="h-px opacity-50 bg-linear-to-r from-accent via-border to-transparent" />

      {/* Content */}
      <div className="text-content-secondary animate-fade-in">{tabs[activeTab].content}</div>
    </div>
  )
}

export default Tabs
