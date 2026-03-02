const StatusIndicator = ({ status, label }: { status: string; label: string }) => {
  return (
    <div className="flex flex-row justify-center items-center space-x-2">
      {label && <span className="text-zinc-400 font-medium text-xs text-nowrap">{label}</span>}
      <span
        className={`${status === 'ok' ? 'bg-green-500' : 'bg-red-500'} rounded-full w-2 h-2 flex justify-center items-center`}
      />
    </div>
  )
}

export default StatusIndicator
