type SectionLabelProps = { children: React.ReactNode }

const SectionLabel = ({ children }: SectionLabelProps) => (
  <p className="text-content-secondary text-[11px] font-mono uppercase tracking-widest mb-2.5 mt-0">
    {children}
  </p>
)

export default SectionLabel
