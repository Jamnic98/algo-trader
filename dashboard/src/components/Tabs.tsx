import type { Tab } from 'types'

type TabsProps = {
  tabs: Tab[]
  activeTab: number
  onTabChange: (index: number) => void
}

const Tabs = ({ tabs, activeTab, onTabChange }: TabsProps) => {
  return (
    <div>
      {/* Tab Bar */}
      <div className="flex gap-0.5 bg-surface-secondary p-1 rounded-lg w-fit border border-border">
        {tabs.map((tab, index) => {
          const isActive = activeTab === index
          return (
            <button
              key={tab.label}
              onClick={() => onTabChange(index)}
              className={`relative px-4 py-1.5 text-xs font-mono tracking-widest uppercase rounded-md cursor-pointer transition-all duration-200 border
                ${
                  isActive
                    ? 'bg-accent-muted text-accent border-accent/25 font-semibold'
                    : 'bg-transparent text-content-secondary border-transparent hover:text-content-primary hover:bg-white/4'
                }`}
            >
              {isActive && (
                <span className="absolute top-1/2 left-2 -translate-y-1/2 w-1 h-1 rounded-full bg-accent shadow-[0_0_6px_var(--color-accent)]" />
              )}
              <span className={isActive ? 'ml-2' : ''}>{tab.label}</span>
            </button>
          )
        })}
      </div>

      {/* Divider */}
      <div className="h-px my-4 opacity-50 bg-linear-to-r from-accent via-border to-transparent" />

      {/* Content */}
      <div className="text-content-secondary animate-fade-in">{tabs[activeTab].content}</div>
    </div>
  )
}

export default Tabs
