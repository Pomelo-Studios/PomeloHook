import { useState, useEffect, useCallback } from 'react'
import { EventList } from './components/EventList'
import { EventDetail } from './components/EventDetail'
import { Header } from './components/Header'
import { api } from './api/client'
import type { WebhookEvent } from './types'

export default function App() {
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [selected, setSelected] = useState<WebhookEvent | null>(null)
  const [tunnelID, setTunnelID] = useState('')
  const [tunnelSubdomain, setTunnelSubdomain] = useState('')
  const [replayError, setReplayError] = useState<string | null>(null)
  const [isAdmin, setIsAdmin] = useState(false)

  useEffect(() => {
    Promise.all([api.getMe(''), api.getTunnels()])
      .then(([me, tunnels]) => {
        if (me.role === 'admin') setIsAdmin(true)
        const active = tunnels.find(t => t.Status === 'active')
        if (active) { setTunnelID(active.ID); setTunnelSubdomain(active.Subdomain) }
      })
      .catch(() => {})
  }, [])

  useEffect(() => {
    if (!tunnelID) return
    const fn = () => api.getEvents(tunnelID, 100).then(setEvents).catch(() => {})
    fn()
    const id = setInterval(fn, 3000)
    return () => clearInterval(id)
  }, [tunnelID])

  const handleReplay = useCallback(async (eventID: string, targetURL: string) => {
    setReplayError(null)
    try {
      await api.replay(eventID, targetURL)
    } catch (err) {
      setReplayError(err instanceof Error ? err.message : 'Replay failed')
    }
  }, [])

  return (
    <div className="flex flex-col h-screen font-sans text-sm" style={{ background: 'var(--bg)' }}>
      <Header subdomain={tunnelSubdomain} connected={!!tunnelID} isAdmin={isAdmin} />
      <div className="flex flex-1 overflow-hidden">
        <div className="w-[240px] flex flex-col overflow-hidden" style={{ borderRight: '1px solid var(--border)' }}>
          <EventList
            events={events}
            selectedID={selected?.ID ?? null}
            onSelect={setSelected}
            tunnelSubdomain={tunnelSubdomain}
          />
        </div>
        <div className="flex-1 flex flex-col overflow-hidden">
          {replayError && (
            <div className="text-xs px-4 py-2 flex-shrink-0" style={{ background: 'var(--err-bg)', color: 'var(--err-text)', borderBottom: '1px solid var(--selected-border)' }}>
              {replayError}
            </div>
          )}
          {selected
            ? <EventDetail event={selected} onReplay={handleReplay} />
            : <div className="flex items-center justify-center h-full text-sm" style={{ color: 'var(--text-dim)' }}>Select an event</div>
          }
        </div>
      </div>
    </div>
  )
}
