import { useState, useEffect } from 'react'
import { Header } from './components/Header'
import { LoginForm } from './components/admin/LoginForm'
import { UsersPanel } from './components/admin/UsersPanel'
import { OrgsPanel } from './components/admin/OrgsPanel'
import { TunnelsPanel } from './components/admin/TunnelsPanel'
import { DatabasePanel } from './components/admin/DatabasePanel'
import { useAuth } from './hooks/useAuth'
import { api } from './api/client'

type Section = 'users' | 'orgs' | 'tunnels' | 'database'

export function AdminApp() {
  const { apiKey, isServerMode, loading, login, logout } = useAuth()
  const [section, setSection] = useState<Section>('users')
  const [subdomain, setSubdomain] = useState('')

  useEffect(() => {
    if (loading || (isServerMode && !apiKey)) return
    api.getTunnels().then(tunnels => {
      const active = tunnels.find(t => t.Status === 'active')
      if (active) setSubdomain(active.Subdomain)
    }).catch(() => {})
  }, [loading, isServerMode, apiKey])

  if (loading) {
    return <div className="bg-zinc-950 h-screen flex items-center justify-center text-zinc-600 text-xs font-mono">Loading…</div>
  }

  if (isServerMode && !apiKey) {
    return <LoginForm onLogin={login} />
  }

  const navItem = (id: Section, label: string, icon: string) => (
    <button
      onClick={() => setSection(id)}
      className={`flex items-center gap-2 px-2.5 py-1.5 rounded text-[11px] w-full text-left transition-colors ${
        section === id
          ? 'bg-zinc-800 text-emerald-400 border-l-2 border-emerald-500 pl-2'
          : 'text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800/50'
      }`}
    >
      <span>{icon}</span>{label}
    </button>
  )

  return (
    <div className="flex flex-col h-screen bg-zinc-950 font-mono text-sm">
      <Header subdomain={subdomain} connected={false} isAdmin />
      <div className="flex flex-1 overflow-hidden">
        <aside className="w-44 bg-zinc-900/50 border-r border-zinc-800 flex flex-col gap-1 p-2 flex-shrink-0">
          <p className="text-[9px] text-zinc-600 uppercase tracking-widest px-2 pt-2 pb-1">Manage</p>
          {navItem('users', 'Users', '👤')}
          {navItem('orgs', 'Organizations', '🏢')}
          {navItem('tunnels', 'Tunnels', '⚡')}
          <div className="border-t border-zinc-800 my-1" />
          <p className="text-[9px] text-zinc-600 uppercase tracking-widest px-2 pt-1 pb-1">Developer</p>
          {navItem('database', 'Database', '🗄️')}
          {isServerMode && (
            <button
              onClick={logout}
              className="mt-auto text-[10px] px-2.5 py-1.5 text-zinc-600 hover:text-zinc-400 w-full text-left"
            >
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
