import { useState, useEffect, useCallback, useRef } from 'react'
import { EventList } from './components/EventList'
import { EventDetail } from './components/EventDetail'
import { Header } from './components/Header'
import { api } from './api/client'
import type { WebhookEvent } from './types'

function useWSEvents(tunnelID: string, onEvent: (e: WebhookEvent) => void) {
  const onEventRef = useRef(onEvent)
  onEventRef.current = onEvent

  useEffect(() => {
    if (!tunnelID) return
    let ws: WebSocket
    let closed = false

    function connect() {
      ws = new WebSocket(`ws://${location.host}/api/events/stream?tunnel_id=${tunnelID}`)
      ws.onmessage = e => {
        try { onEventRef.current(JSON.parse(e.data) as WebhookEvent) } catch {}
      }
      ws.onclose = () => { if (!closed) setTimeout(connect, 2000) }
      ws.onerror = () => ws.close()
    }

    connect()
    return () => { closed = true; ws?.close() }
  }, [tunnelID])
}

export default function App() {
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [selected, setSelected] = useState<WebhookEvent | null>(null)
  const [tunnelID, setTunnelID] = useState('')
  const [tunnelSubdomain, setTunnelSubdomain] = useState('')
  const [replayError, setReplayError] = useState<string | null>(null)

  useEffect(() => {
    api.getTunnels().then(tunnels => {
      const active = tunnels.find(t => t.Status === 'active')
      if (active) { setTunnelID(active.ID); setTunnelSubdomain(active.Subdomain) }
    }).catch(() => {})
  }, [])

  useEffect(() => {
    if (!tunnelID) return
    api.getEvents(tunnelID, 100).then(setEvents).catch(() => {})
  }, [tunnelID])

  useWSEvents(tunnelID, event => setEvents(prev => [event, ...prev].slice(0, 500)))

  const handleReplay = useCallback(async (eventID: string, targetURL: string) => {
    setReplayError(null)
    try {
      await api.replay(eventID, targetURL)
    } catch (err) {
      setReplayError(err instanceof Error ? err.message : 'Replay failed')
    }
  }, [])

  return (
    <div className="flex flex-col h-screen bg-zinc-950 font-mono text-sm">
      <Header subdomain={tunnelSubdomain} connected={!!tunnelID} />
      <div className="flex flex-1 overflow-hidden">
        <div className="w-[38%] border-r border-zinc-800 flex flex-col overflow-hidden">
          <EventList
            events={events}
            selectedID={selected?.ID ?? null}
            onSelect={setSelected}
            tunnelSubdomain={tunnelSubdomain}
          />
        </div>
        <div className="flex-1 flex flex-col overflow-hidden">
          {replayError && (
            <div className="bg-red-950 text-red-400 text-xs px-4 py-2 border-b border-red-900 flex-shrink-0">
              {replayError}
            </div>
          )}
          {selected
            ? <EventDetail event={selected} onReplay={handleReplay} />
            : <div className="flex items-center justify-center h-full text-zinc-600 text-sm">Select an event</div>
          }
        </div>
      </div>
    </div>
  )
}
