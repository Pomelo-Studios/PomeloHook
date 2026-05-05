import { useState, useEffect, useCallback, useRef } from 'react'
import { LogOut } from 'lucide-react'
import { EventList } from './components/EventList'
import { EventDetail } from './components/EventDetail'
import { TunnelList } from './components/TunnelList'
import { LoginForm } from './components/admin/LoginForm'
import { useAuth } from './hooks/useAuth'
import { api } from './api/client'
import type { WebhookEvent, Tunnel, Me } from './types'
import { SettingsTab } from './components/SettingsTab'
import { Button } from './components/ui'

type Tab = 'personal' | 'org' | 'settings' | 'profile'

function useWSEvents(tunnelID: string, apiKey: string, onEvent: (e: WebhookEvent) => void) {
  const onEventRef = useRef(onEvent)
  onEventRef.current = onEvent

  useEffect(() => {
    if (!tunnelID || !apiKey) return
    let ws: WebSocket
    let closed = false
    let delay = 1000

    function connect() {
      const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
      ws = new WebSocket(
        `${proto}//${location.host}/api/events/stream?tunnel_id=${encodeURIComponent(tunnelID)}&api_key=${encodeURIComponent(apiKey)}`
      )
      ws.onmessage = e => {
        try { onEventRef.current(JSON.parse(e.data) as WebhookEvent) } catch {}
      }
      ws.onopen = () => { delay = 1000 }
      ws.onclose = () => {
        if (!closed) setTimeout(() => { delay = Math.min(delay * 2, 30000); connect() }, delay)
      }
      ws.onerror = () => ws.close()
    }

    connect()
    return () => { closed = true; ws?.close() }
  }, [tunnelID, apiKey])
}

export function OrgApp() {
  const { apiKey, isServerMode, loading, login, logout } = useAuth()
  const [tab, setTab] = useState<Tab>('personal')
  const [tunnels, setTunnels] = useState<Tunnel[]>([])
  const [selectedTunnelID, setSelectedTunnelID] = useState<string | null>(null)
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [selectedEvent, setSelectedEvent] = useState<WebhookEvent | null>(null)
  const [replayError, setReplayError] = useState<string | null>(null)
  const [me, setMe] = useState<Me | null>(null)
  const [creating, setCreating] = useState(false)
  const [editingDisplayName, setEditingDisplayName] = useState(false)
  const [displayNameInput, setDisplayNameInput] = useState('')

  const selectedTunnel = tunnels.find(t => t.id === selectedTunnelID) ?? null

  useEffect(() => {
    if (loading || (isServerMode && !apiKey)) return

    function fetchTunnels() {
      const req = tab === 'personal'
        ? api.org.getUserTunnels(apiKey)
        : api.org.getTunnels(apiKey)
      req.then(data => {
        const filtered = tab === 'personal' ? data.filter(t => t.type === 'personal') : data
        setTunnels(filtered)
        setSelectedTunnelID(prev =>
          filtered.some(t => t.id === prev) ? prev : (filtered[0]?.id ?? null)
        )
      }).catch(() => {})
    }

    fetchTunnels()
    const id = setInterval(fetchTunnels, 5000)
    return () => clearInterval(id)
  }, [loading, isServerMode, apiKey, tab])

  useEffect(() => {
    if (isServerMode && !apiKey) return
    api.getMe(apiKey).then(setMe).catch(() => {})
  }, [apiKey, isServerMode])

  useEffect(() => {
    if (!selectedTunnelID) { setEvents([]); return }
    setEditingDisplayName(false)
    setDisplayNameInput(tunnels.find(t => t.id === selectedTunnelID)?.display_name ?? '')

    const tunnelID = selectedTunnelID
    function fetchEvents() {
      api.getEvents(tunnelID, 100, apiKey).then(next =>
        setEvents(prev => {
          if (prev.length === next.length && prev.every((e, i) => e.ID === next[i].ID)) return prev
          return next
        })
      ).catch(() => {})
    }

    fetchEvents()
    const id = setInterval(fetchEvents, 5000)
    return () => clearInterval(id)
  }, [selectedTunnelID, apiKey])

  useWSEvents(
    selectedTunnelID ?? '',
    apiKey,
    event => setEvents(prev => [event, ...prev.filter(e => e.ID !== event.ID)].slice(0, 100))
  )

  const handleReplay = useCallback(async (eventID: string, targetURL: string) => {
    setReplayError(null)
    try {
      await api.replay(eventID, targetURL, apiKey)
    } catch (err) {
      setReplayError(err instanceof Error ? err.message : 'Replay failed')
    }
  }, [apiKey])

  async function handleCreateTunnel() {
    setCreating(true)
    try {
      const tun = tab === 'org'
        ? await api.org.createOrgTunnel(apiKey)
        : await api.org.createPersonalTunnel(apiKey)
      setTunnels(prev => [...prev, tun])
      setSelectedTunnelID(tun.id)
    } catch (err) {
      setReplayError(err instanceof Error ? err.message : 'Failed to create tunnel')
    } finally { setCreating(false) }
  }

  if (loading) {
    return (
      <div className="h-screen flex items-center justify-center text-xs font-mono" style={{ background: 'var(--bg)', color: 'var(--text-3)' }}>
        Loading…
      </div>
    )
  }

  if (isServerMode && !apiKey) {
    return <LoginForm onLogin={login} />
  }

  const can = (perm: string) => me?.role === 'admin' || (me?.permissions ?? []).includes(perm)
  const hasAnySettingsPerm = can('manage_members') || can('change_member_role') || can('manage_roles') || can('edit_org_settings')

  return (
    <div className="flex flex-col h-screen font-sans text-sm" style={{ background: 'var(--bg)' }}>
      <div
        className="flex items-center px-4 flex-shrink-0"
        style={{ height: '42px', borderBottom: '1px solid var(--border)', background: 'var(--surface)' }}
      >
        <span className="font-mono text-[13px] font-bold mr-2" style={{ color: 'var(--coral)' }}>
          PomeloHook
        </span>
        {me?.org_name && (
          <span
            className="text-[10px] font-medium px-2 py-[2px] rounded mr-3"
            style={{ background: 'var(--surface2)', color: 'var(--text-3)', border: '1px solid var(--border)' }}
          >
            {me.org_name}
          </span>
        )}
        <div className="flex gap-1">
          {(['personal', 'org', ...(hasAnySettingsPerm ? ['settings'] : []), 'profile'] as Tab[]).map(t => (
            <button
              key={t}
              onClick={() => { setTab(t as Tab); setSelectedTunnelID(null); setSelectedEvent(null) }}
              className="px-3 text-[11px] font-semibold capitalize transition-colors"
              style={
                tab === t
                  ? { color: 'var(--coral)', borderBottom: '2px solid var(--coral)', paddingBottom: '10px', marginBottom: '-1px', background: 'transparent' }
                  : { color: 'var(--text-3)', borderBottom: '2px solid transparent', paddingBottom: '10px', marginBottom: '-1px', background: 'transparent' }
              }
            >
              {t}
            </button>
          ))}
        </div>
        <div className="flex-1" />
        {(tab === 'personal' || tab === 'org') && (tab === 'personal' || can('create_org_tunnel')) && (
          <Button
            variant="primary"
            size="sm"
            onClick={handleCreateTunnel}
            disabled={creating}
            className="mr-3"
          >
            {creating ? 'Creating…' : '+ New Tunnel'}
          </Button>
        )}
        {me?.role === 'admin' && (
          <a
            href="/admin"
            className="text-[11px] font-medium px-3 py-1 rounded transition-colors mr-2"
            style={{ color: 'var(--text-3)', background: 'var(--surface)' }}
          >
            Admin Panel →
          </a>
        )}
        {isServerMode && (
          <button
            onClick={logout}
            className="p-1"
            style={{ color: 'var(--text-3)' }}
            title="Sign out"
          >
            <LogOut size={14} strokeWidth={2} />
          </button>
        )}
      </div>

      {tab === 'profile' ? (
        <div className="flex-1 overflow-y-auto p-6 max-w-lg">
          <ProfilePanel apiKey={apiKey} me={me} onUpdated={setMe} />
        </div>
      ) : tab === 'settings' ? (
        <SettingsTab apiKey={apiKey} me={me} can={can} />
      ) : (
        <div className="flex flex-1 overflow-hidden">
          <div
            className="w-[180px] flex flex-col overflow-hidden flex-shrink-0"
            style={{ borderRight: '1px solid var(--border)' }}
          >
            <TunnelList
              tunnels={tunnels}
              selectedID={selectedTunnelID}
              onSelect={t => { setSelectedTunnelID(t.id); setSelectedEvent(null) }}
            />
          </div>

          <div
            className="w-[240px] flex flex-col overflow-hidden flex-shrink-0"
            style={{ borderRight: '1px solid var(--border)' }}
          >
            <EventList
              events={events}
              selectedID={selectedEvent?.ID ?? null}
              onSelect={setSelectedEvent}
              tunnelSubdomain={selectedTunnel?.subdomain}
            />
          </div>

          <div className="flex-1 flex flex-col overflow-hidden">
            {selectedTunnel && (
              <div style={{ padding: '6px 12px', borderBottom: '1px solid var(--border)', flexShrink: 0 }}>
                <div style={{ fontSize: 10, color: 'var(--text-3)', marginBottom: 3 }}>Display name</div>
                {editingDisplayName ? (
                  <div style={{ display: 'flex', gap: 6 }}>
                    <input
                      value={displayNameInput}
                      onChange={e => setDisplayNameInput(e.target.value)}
                      style={{ flex: 1, padding: '3px 6px', background: 'var(--surface)', color: 'var(--text)', border: '1px solid var(--border)', borderRadius: 4, fontSize: 11 }}
                      autoFocus
                    />
                    <Button
                      variant="primary"
                      size="sm"
                      onClick={async () => {
                        const updated = await api.org.updateTunnel(apiKey, selectedTunnel.id, displayNameInput)
                        setTunnels(ts => ts.map(t => t.id === updated.id ? updated : t))
                        setEditingDisplayName(false)
                      }}
                    >
                      Save
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => { setEditingDisplayName(false); setDisplayNameInput(selectedTunnel.display_name ?? '') }}
                    >
                      ✕
                    </Button>
                  </div>
                ) : (
                  <div
                    onClick={() => { setDisplayNameInput(selectedTunnel.display_name ?? ''); setEditingDisplayName(true) }}
                    style={{ fontSize: 12, color: selectedTunnel.display_name ? 'var(--text)' : 'var(--text-3)', cursor: 'pointer', padding: '2px 4px', borderRadius: 4 }}
                    title="Click to rename"
                  >
                    {selectedTunnel.display_name || '(click to add display name)'}
                  </div>
                )}
              </div>
            )}
            {replayError && (
              <div
                className="text-xs px-4 py-2 flex-shrink-0"
                style={{ background: 'var(--err-bg)', color: 'var(--err-text)', borderBottom: '1px solid var(--selected-border)' }}
              >
                {replayError}
              </div>
            )}
            {selectedEvent
              ? <EventDetail event={selectedEvent} onReplay={handleReplay} />
              : (
                <div className="flex items-center justify-center h-full text-sm" style={{ color: 'var(--text-3)' }}>
                  Select an event
                </div>
              )
            }
          </div>
        </div>
      )}
    </div>
  )
}

const profileInputStyle: React.CSSProperties = {
  background: 'var(--surface)', border: '1px solid var(--border)',
  borderRadius: 6, padding: '6px 10px', fontSize: 12, color: 'var(--text)', width: '100%',
}
const profileLabelStyle: React.CSSProperties = {
  fontSize: 10, fontWeight: 600, color: 'var(--text-3)', textTransform: 'uppercase', letterSpacing: 1,
}
const profileBtnStyle: React.CSSProperties = {
  background: 'var(--coral)', color: 'white', border: 'none', borderRadius: 6,
  padding: '6px 14px', fontSize: 11, fontWeight: 600, cursor: 'pointer',
}

function ProfilePanel({
  apiKey,
  me,
  onUpdated,
}: {
  apiKey: string
  me: Me | null
  onUpdated: (u: Me) => void
}) {
  const [name, setName] = useState(me?.name ?? '')
  const [email, setEmail] = useState(me?.email ?? '')
  const [profileMsg, setProfileMsg] = useState('')
  const [currentPwd, setCurrentPwd] = useState('')
  const [newPwd, setNewPwd] = useState('')
  const [pwdMsg, setPwdMsg] = useState('')
  const [showKey, setShowKey] = useState(false)

  useEffect(() => { setName(me?.name ?? ''); setEmail(me?.email ?? '') }, [me])

  async function handleProfileSave(e: React.FormEvent) {
    e.preventDefault()
    try {
      const updated = await api.updateMe(apiKey, name, email)
      onUpdated({ ...(me ?? {} as Me), ...updated })
      setProfileMsg('Saved.')
      setTimeout(() => setProfileMsg(''), 2000)
    } catch {
      setProfileMsg('Save failed.')
    }
  }

  async function handlePasswordChange(e: React.FormEvent) {
    e.preventDefault()
    try {
      await api.changePassword(apiKey, currentPwd, newPwd)
      setCurrentPwd(''); setNewPwd('')
      setPwdMsg('Password changed.')
      setTimeout(() => setPwdMsg(''), 2000)
    } catch {
      setPwdMsg('Failed. Check your current password.')
    }
  }

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h3 className="text-[13px] font-semibold mb-4" style={{ color: 'var(--text)' }}>Profile</h3>
        <form onSubmit={handleProfileSave} className="flex flex-col gap-3">
          <div className="flex flex-col gap-1">
            <label style={profileLabelStyle}>Name</label>
            <input style={profileInputStyle} value={name} onChange={e => setName(e.target.value)} required />
          </div>
          <div className="flex flex-col gap-1">
            <label style={profileLabelStyle}>Email</label>
            <input style={profileInputStyle} type="email" value={email} onChange={e => setEmail(e.target.value)} required />
          </div>
          <div className="flex items-center gap-3">
            <button type="submit" style={profileBtnStyle}>Save</button>
            {profileMsg && <span style={{ fontSize: 11, color: 'var(--text-3)' }}>{profileMsg}</span>}
          </div>
        </form>
      </div>

      <div>
        <h3 className="text-[13px] font-semibold mb-4" style={{ color: 'var(--text)' }}>API Key</h3>
        <div className="flex items-center gap-2">
          <span className="font-mono text-[11px]" style={{ color: 'var(--text-2)' }}>
            {showKey ? (me?.api_key ?? '—') : '••••••••••••••••••••'}
          </span>
          <button
            onClick={() => setShowKey(v => !v)}
            style={{ fontSize: 10, color: 'var(--text-3)', background: 'none', border: 'none', cursor: 'pointer' }}
          >
            {showKey ? 'hide' : 'reveal'}
          </button>
        </div>
      </div>

      <div>
        <h3 className="text-[13px] font-semibold mb-4" style={{ color: 'var(--text)' }}>Change Password</h3>
        <form onSubmit={handlePasswordChange} className="flex flex-col gap-3">
          <div className="flex flex-col gap-1">
            <label style={profileLabelStyle}>Current password</label>
            <input style={profileInputStyle} type="password" value={currentPwd} onChange={e => setCurrentPwd(e.target.value)} required />
          </div>
          <div className="flex flex-col gap-1">
            <label style={profileLabelStyle}>New password (min 8 chars)</label>
            <input style={profileInputStyle} type="password" value={newPwd} onChange={e => setNewPwd(e.target.value)} minLength={8} required />
          </div>
          <div className="flex items-center gap-3">
            <button type="submit" style={profileBtnStyle}>Change Password</button>
            {pwdMsg && <span style={{ fontSize: 11, color: 'var(--text-3)' }}>{pwdMsg}</span>}
          </div>
        </form>
      </div>
    </div>
  )
}
