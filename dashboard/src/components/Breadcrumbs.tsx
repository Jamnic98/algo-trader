import { Link, useLocation } from 'react-router-dom'
import { House } from 'lucide-react'

const LABEL_MAP: Record<string, string> = {
  accounts: 'Accounts',
  bots: 'Bots',
  diagnostics: 'Diagnostics',
  trades: 'Trades',
}

// Segments that look like UUIDs or IDs — shorten to first hyphen chunk
const looksLikeId = (segment: string) => /^[0-9a-f]{8}-/i.test(segment) || segment.length > 20

const formatSegment = (segment: string): string => {
  if (looksLikeId(segment)) return segment.split('-')[0]
  if (LABEL_MAP[segment]) return LABEL_MAP[segment]
  return segment.charAt(0).toUpperCase() + segment.slice(1)
}

const Breadcrumbs = () => {
  const { pathname } = useLocation()

  // Do not render on dashboard
  if (pathname === '/') return <div className="mb-6 h-5" />

  const segments = pathname.split('/').filter(Boolean)

  const crumbs = segments.map((segment, i) => ({
    label: formatSegment(segment),
    path: '/' + segments.slice(0, i + 1).join('/'),
  }))

  return (
    <nav className="flex items-center gap-1.5 text-sm font-mono mb-6 select-none">
      {/* Home icon */}
      <Link
        to="/"
        className="text-content-secondary hover:text-content-primary transition-colors"
        aria-label="Dashboard"
      >
        <House size={14} />
      </Link>

      {crumbs.map((crumb, i) => {
        const isLast = i === crumbs.length - 1
        return (
          <span key={crumb.path} className="flex items-center gap-1.5">
            <span className="text-content-secondary select-none">/</span>
            {isLast ? (
              <span className="text-content-primary">{crumb.label}</span>
            ) : (
              <Link
                to={crumb.path}
                className="text-content-secondary hover:text-content-primary transition-colors"
              >
                {crumb.label}
              </Link>
            )}
          </span>
        )
      })}
    </nav>
  )
}

export default Breadcrumbs
