import { useState, useEffect, useCallback } from 'react'
import { LogOut } from 'lucide-react'
import { EventList } from './components/EventList'
import { EventDetail } from './components/EventDetail'
import { TunnelList } from './components/TunnelList'
import { LoginForm } from './components/admin/LoginForm'
import { useAuth } from './hooks/useAuth'
import { api } from './api/client'
import type { WebhookEvent, Tunnel } from './types'

type Tab = 'personal' | 'org'

export function OrgApp() {
  const { apiKey, isServerMode, loading, login, logout } = useAuth()
  const [tab, setTab] = useState<Tab>('personal')
  const [tunnels, setTunnels] = useState<Tunnel[]>([])
  const [selectedTunnelID, setSelectedTunnelID] = useState<string | null>(null)
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [selectedEvent, setSelectedEvent] = useState<WebhookEvent | null>(null)
  const [replayError, setReplayError] = useState<string | null>(null)

  const selectedTunnel = tunnels.find(t => t.ID === selectedTunnelID) ?? null

  useEffect(() => {
    if (loading || (isServerMode && !apiKey)) return

    function fetchTunnels() {
      const req = tab === 'personal'
        ? api.org.getUserTunnels(apiKey)
        : api.org.getTunnels(apiKey)
      req.then(data => {
        const filtered = tab === 'personal' ? data.filter(t => t.Type === 'personal') : data
        setTunnels(filtered)
        setSelectedTunnelID(prev =>
          filtered.some(t => t.ID === prev) ? prev : (filtered[0]?.ID ?? null)
        )
      }).catch(() => {})
    }

    fetchTunnels()
    const id = setInterval(fetchTunnels, 5000)
    return () => clearInterval(id)
  }, [loading, isServerMode, apiKey, tab])

  useEffect(() => {
    if (!selectedTunnelID) { setEvents([]); return }

    function fetchEvents() {
      api.getEvents(selectedTunnelID, 100).then(setEvents).catch(() => {})
    }

    fetchEvents()
    const id = setInterval(fetchEvents, 5000)
    return () => clearInterval(id)
  }, [selectedTunnelID])

  const handleReplay = useCallback(async (eventID: string, targetURL: string) => {
    setReplayError(null)
    try {
      await api.replay(eventID, targetURL)
    } catch (err) {
      setReplayError(err instanceof Error ? err.message : 'Replay failed')
    }
  }, [])

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

  return (
    <div className="flex flex-col h-screen font-sans text-sm" style={{ background: 'var(--bg)' }}>
      <div
        className="flex items-center px-4 flex-shrink-0"
        style={{ height: '42px', borderBottom: '1px solid var(--border)', background: 'var(--surface)' }}
      >
        <span className="font-mono text-[13px] font-bold mr-4" style={{ color: '#FF6B6B' }}>
          PomeloHook
        </span>
        <div className="flex gap-1">
          {(['personal', 'org'] as Tab[]).map(t => (
            <button
              key={t}
              onClick={() => { setTab(t); setSelectedTunnelID(null); setSelectedEvent(null) }}
              className="px-3 py-1 rounded text-[11px] font-semibold capitalize transition-colors"
              style={
                tab === t
                  ? { background: 'rgba(255,107,107,0.13)', color: '#FF6B6B' }
                  : { color: 'var(--text-dim)' }
              }
            >
              {t}
            </button>
          ))}
        </div>
        <div className="flex-1" />
        {isServerMode && (
          <button
            onClick={logout}
            className="p-1"
            style={{ color: 'var(--text-dim)' }}
            title="Sign out"
          >
            <LogOut size={14} strokeWidth={2} />
          </button>
        )}
      </div>

      <div className="flex flex-1 overflow-hidden">
        <div
          className="w-[180px] flex flex-col overflow-hidden flex-shrink-0"
          style={{ borderRight: '1px solid var(--border)' }}
        >
          <TunnelList
            tunnels={tunnels}
            selectedID={selectedTunnelID}
            onSelect={t => { setSelectedTunnelID(t.ID); setSelectedEvent(null) }}
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
            tunnelSubdomain={selectedTunnel?.Subdomain}
          />
        </div>

        <div className="flex-1 flex flex-col overflow-hidden">
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
              <div className="flex items-center justify-center h-full text-sm" style={{ color: 'var(--text-dim)' }}>
                Select an event
              </div>
            )
          }
        </div>
      </div>
    </div>
  )
}
