import { useState, useEffect } from 'react'
import { Users, Building2, Network, Database, LogOut } from 'lucide-react'
import { Header } from './components/Header'
import { LoginForm } from './components/admin/LoginForm'
import { UsersPanel } from './components/admin/UsersPanel'
import { OrgsPanel } from './components/admin/OrgsPanel'
import { TunnelsPanel } from './components/admin/TunnelsPanel'
import { DatabasePanel } from './components/admin/DatabasePanel'
import { useAuth } from './hooks/useAuth'
import { api } from './api/client'
import type { ReactNode } from 'react'

type Section = 'users' | 'orgs' | 'tunnels' | 'database'

type NavItem = { id: Section; label: string; icon: ReactNode }

export function AdminApp() {
  const { apiKey, isServerMode, loading, login, logout } = useAuth()
  const [section, setSection] = useState<Section>('users')
  const [subdomain, setSubdomain] = useState('')

  useEffect(() => {
    if (loading || (isServerMode && !apiKey)) return
    api.admin.listTunnels(apiKey).then(tunnels => {
      const active = tunnels.find(t => t.Status === 'active')
      if (active) setSubdomain(active.Subdomain)
    }).catch(() => {})
  }, [loading, isServerMode, apiKey])

  if (loading) {
    return (
      <div className="h-screen flex items-center justify-center text-xs font-mono" style={{ background: 'var(--bg)', color: 'var(--text-dim)' }}>
        Loading…
      </div>
    )
  }

  if (isServerMode && !apiKey) {
    return <LoginForm onLogin={login} />
  }

  const manageItems: NavItem[] = [
    { id: 'users', label: 'Users', icon: <Users size={14} strokeWidth={2} /> },
    { id: 'orgs', label: 'Organizations', icon: <Building2 size={14} strokeWidth={2} /> },
    { id: 'tunnels', label: 'Tunnels', icon: <Network size={14} strokeWidth={2} /> },
  ]
  const devItems: NavItem[] = [
    { id: 'database', label: 'Database', icon: <Database size={14} strokeWidth={2} /> },
  ]

  function NavButton({ item }: { item: NavItem }) {
    const active = section === item.id
    return (
      <button
        onClick={() => setSection(item.id)}
        className="flex items-center gap-2 px-[10px] py-[7px] rounded-lg w-full text-left text-[12px] font-medium transition-colors border"
        style={
          active
            ? { color: '#FF6B6B', background: 'var(--selected-bg)', borderColor: 'var(--selected-border)' }
            : { color: 'var(--text-secondary)', background: 'transparent', borderColor: 'transparent' }
        }
      >
        {item.icon}
        {item.label}
      </button>
    )
  }

  return (
    <div className="flex flex-col h-screen font-sans text-sm" style={{ background: 'var(--bg)' }}>
      <Header subdomain={subdomain} connected={false} isAdmin={!isServerMode} />
      <div className="flex flex-1 overflow-hidden">
        <aside
          className="w-[200px] flex flex-col gap-[2px] p-2 flex-shrink-0"
          style={{ background: 'var(--surface)', borderRight: '1px solid var(--border)' }}
        >
          <p className="text-[9px] font-bold tracking-[2px] uppercase px-2 pt-2 pb-1" style={{ color: 'var(--text-dim)' }}>
            Manage
          </p>
          {manageItems.map(item => <NavButton key={item.id} item={item} />)}

          <div className="my-1 mx-1" style={{ height: 1, background: 'var(--border)' }} />

          <p className="text-[9px] font-bold tracking-[2px] uppercase px-2 pt-1 pb-1" style={{ color: 'var(--text-dim)' }}>
            Developer
          </p>
          {devItems.map(item => <NavButton key={item.id} item={item} />)}

          {isServerMode && (
            <button
              onClick={logout}
              className="mt-auto flex items-center gap-2 px-[10px] py-[7px] rounded-lg text-[11px] transition-colors w-full"
              style={{ color: 'var(--text-dim)' }}
            >
              <LogOut size={13} strokeWidth={2} />
              Sign out
            </button>
          )}
        </aside>
        <main className="flex-1 overflow-hidden flex flex-col">
          {section === 'users' && <UsersPanel apiKey={apiKey} />}
          {section === 'orgs' && <OrgsPanel apiKey={apiKey} />}
          {section === 'tunnels' && <TunnelsPanel apiKey={apiKey} />}
          {section === 'database' && <DatabasePanel apiKey={apiKey} />}
        </main>
      </div>
    </div>
  )
}
