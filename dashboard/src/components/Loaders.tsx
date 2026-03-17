type BarWidth = 'wide' | 'normal' | 'narrow' | 'uptime'

type BarLoaderProps = { fullscreen?: boolean; width?: BarWidth }

const barWidth = (width: BarWidth): string => {
  switch (width) {
    case 'wide':
      return 'w-36'
    case 'narrow':
      return 'w-15'
    case 'uptime':
      return 'w-[8ch]'
    default:
      return 'w-28'
  }
}

const BarLoader = ({ fullscreen = false, width = 'normal' }: BarLoaderProps) => {
  if (fullscreen)
    return (
      <div className="fixed top-0 left-0 right-0 z-50 h-0.5 bg-surface-secondary">
        <div className="animate-bar-slide absolute inset-y-0 left-0 right-0 bg-accent">
          <div className="animate-shimmer absolute top-0 bottom-0 w-1/3 bg-linear-to-r from-transparent via-white/20 to-transparent" />
        </div>
      </div>
    )
  return (
    <span
      className={`${barWidth(width)} h-0.75 bg-surface-secondary rounded-full overflow-hidden relative inline-flex align-middle`}
    >
      <span className="animate-bar-slide absolute inset-y-0 left-0 right-0 bg-accent rounded-full">
        <span className="animate-shimmer absolute top-0 bottom-0 w-1/3 bg-linear-to-r from-transparent via-white/20 to-transparent" />
      </span>
    </span>
  )
}

const SpinnerLoader = ({ size = 'md' }: { size?: 'sm' | 'md' | 'lg' }) => {
  const sizes = { sm: 'w-4 h-4', md: 'w-10 h-10', lg: 'w-16 h-16' }
  return (
    <div
      className={`animate-spin-slow ${sizes[size]} rounded-full border-2 border-surface-secondary border-t-accent`}
    />
  )
}

export { BarLoader, SpinnerLoader }
