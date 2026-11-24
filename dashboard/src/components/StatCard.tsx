type StatCardProps = { label: string; value: string | number; sub?: string; accent?: boolean }

const StatCard = ({ label, value, sub, accent }: StatCardProps) => (
  <div className="bg-surface-secondary border border-table-border rounded-lg px-5 py-4 flex flex-col gap-1">
    <span className="text-content-secondary text-[11px] font-mono uppercase tracking-widest">
      {label}
    </span>
    <span
      className={`text-[22px] font-semibold tracking-tight ${accent ? 'text-accent' : 'text-content-primary'}`}
    >
      {value}
    </span>
    {sub && <span className="text-content-secondary text-xs">{sub}</span>}
  </div>
)

export default StatCard
