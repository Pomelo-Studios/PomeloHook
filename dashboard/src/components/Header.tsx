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
    <header style={{
      height: '48px',
      background: 'var(--surface)',
      borderBottom: '1px solid var(--border)',
      display: 'flex',
      alignItems: 'center',
      padding: '0 16px',
      gap: '12px',
      flexShrink: 0,
    }}>
      <HookIcon size={28} />
      <span style={{ fontSize: '14px', fontWeight: 700, color: 'var(--text)', fontFamily: 'var(--font-sans)' }}>
        PomeloHook
      </span>

      {subdomain && (
        <>
          <div style={{ width: '1px', height: '16px', background: 'var(--border)', flexShrink: 0 }} />
          <div style={{ width: '7px', height: '7px', borderRadius: '50%', background: connected ? 'var(--mint)' : 'var(--text-3)', animation: connected ? 'blink 2s infinite' : 'none', flexShrink: 0 }} />
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: '11.5px', color: 'var(--ok-text)', background: 'var(--ok-bg)', border: '1px solid rgba(76,212,161,0.2)', padding: '3px 10px', borderRadius: '6px' }}>
            {subdomain}
          </span>
        </>
      )}

      {!subdomain && (
        <>
          <div style={{ width: '1px', height: '16px', background: 'var(--border)', flexShrink: 0 }} />
          <span style={{ fontSize: '11px', color: 'var(--text-3)', fontFamily: 'var(--font-sans)' }}>no active tunnel</span>
        </>
      )}

      <div style={{ flex: 1 }} />

      {isAdmin && (
        <nav style={{ display: 'flex', gap: '2px' }}>
          {[
            { to: '/', label: 'Dashboard', active: !onAdmin },
            { to: '/admin', label: 'Admin', active: onAdmin },
          ].map(({ to, label, active }) => (
            <Link
              key={to}
              to={to}
              style={{
                fontFamily: 'var(--font-sans)',
                fontSize: '12.5px',
                fontWeight: 500,
                color: active ? 'var(--text)' : 'var(--text-2)',
                padding: '5px 10px',
                borderRadius: '7px',
                textDecoration: 'none',
                background: active ? 'var(--surface2)' : 'transparent',
                transition: 'color 0.15s',
              }}
            >
              {label}
            </Link>
          ))}
        </nav>
      )}

      <button
        onClick={toggle}
        style={{
          width: '32px', height: '32px',
          background: 'var(--surface2)',
          border: '1px solid var(--border)',
          borderRadius: '8px',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          cursor: 'pointer',
          color: 'var(--text-2)',
          transition: 'border-color 0.2s',
          flexShrink: 0,
        }}
        onMouseEnter={e => { (e.currentTarget as HTMLElement).style.borderColor = 'var(--coral)'; }}
        onMouseLeave={e => { (e.currentTarget as HTMLElement).style.borderColor = 'var(--border)'; }}
        aria-label="Toggle theme"
      >
        {theme === 'dark' ? <Sun size={14} /> : <Moon size={14} />}
      </button>
    </header>
  )
}
