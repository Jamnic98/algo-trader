const BarLoader = ({ fullscreen = false }: { fullscreen?: boolean }) => {
  if (fullscreen)
    return (
      <div className="fixed top-0 left-0 right-0 z-50 h-0.5 bg-surface-secondary">
        <div className="animate-bar-slide absolute inset-y-0 left-0 right-0 bg-accent">
          <div className="animate-shimmer absolute top-0 bottom-0 w-1/3 bg-linear-to-r from-transparent via-white/20 to-transparent" />
        </div>
      </div>
    )
  return (
    <div className="w-36 h-0.75 bg-surface-secondary rounded-full overflow-hidden relative">
      <div className="animate-bar-slide absolute inset-y-0 left-0 right-0 bg-accent rounded-full">
        <div className="animate-shimmer absolute top-0 bottom-0 w-1/3 bg-linear-to-r from-transparent via-white/20 to-transparent" />
      </div>
    </div>
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
