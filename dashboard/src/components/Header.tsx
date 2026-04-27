import { Link, useLocation } from 'react-router-dom'
import { Sun, Moon } from 'lucide-react'
import { HookIcon } from './HookIcon'
import { useTheme } from '../hooks/useTheme'

interface Props {
  subdomain: string
  connected: boolean
  isAdmin?: boolean
}

export function Header({ subdomain, connected, isAdmin }: Props) {
  const location = useLocation()
  const onAdmin = location.pathname.startsWith('/admin')
  const { theme, toggle } = useTheme()

  return (
    <header
      className="h-[52px] flex-shrink-0 flex items-center gap-3 px-5 border-b"
      style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
    >
      <HookIcon size={28} />
      <span className="font-extrabold text-[15px] tracking-tight" style={{ color: 'var(--text-primary)' }}>
        PomeloHook
      </span>

      <div className="w-px h-[18px]" style={{ background: 'var(--border)' }} />

      {subdomain ? (
        <div className="flex items-center gap-2">
          <div
            className="w-[7px] h-[7px] rounded-full flex-shrink-0"
            style={{ background: connected ? '#4CD4A1' : 'var(--text-dim)' }}
          />
          <span className="font-mono text-[10px]" style={{ color: 'var(--text-secondary)' }}>
            {subdomain}
          </span>
        </div>
      ) : (
        <span className="text-[10px]" style={{ color: 'var(--text-dim)' }}>no active tunnel</span>
      )}

      <div className="ml-auto flex items-center gap-2">
        {isAdmin && (
          <div className="flex gap-1">
            {[
              { to: '/', label: 'Dashboard', active: !onAdmin },
              { to: '/admin', label: 'Admin', active: onAdmin },
            ].map(({ to, label, active }) => (
              <Link
                key={to}
                to={to}
                className="text-[11px] font-medium px-[10px] py-1 rounded-md border transition-colors"
                style={
                  active
                    ? { color: '#FF6B6B', background: 'var(--selected-bg)', borderColor: 'var(--selected-border)' }
                    : { color: 'var(--text-dim)', background: 'transparent', borderColor: 'transparent' }
                }
              >
                {label}
              </Link>
            ))}
          </div>
        )}
        {connected && (
          <span
            className="text-[10px] font-semibold px-[10px] py-[3px] rounded-full border"
            style={{ color: 'var(--ok-text)', background: 'var(--ok-bg)', borderColor: 'var(--ok-bg)' }}
          >
            ● connected
          </span>
        )}
        <button
          onClick={toggle}
          className="w-7 h-7 flex items-center justify-center rounded-md transition-colors"
          style={{ color: 'var(--text-dim)' }}
          aria-label="Toggle theme"
        >
          {theme === 'dark' ? <Sun size={14} strokeWidth={2} /> : <Moon size={14} strokeWidth={2} />}
        </button>
      </div>
    </header>
  )
}
