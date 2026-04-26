import { useState, useEffect, useCallback, useRef } from 'react'
import { EventList } from './components/EventList'
import { EventDetail } from './components/EventDetail'
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
        try {
          onEventRef.current(JSON.parse(e.data) as WebhookEvent)
        } catch {}
      }
      ws.onclose = () => {
        if (!closed) setTimeout(connect, 2000)
      }
      ws.onerror = () => ws.close()
    }

    connect()
    return () => {
      closed = true
      ws?.close()
    }
  }, [tunnelID])
}

export default function App() {
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [selected, setSelected] = useState<WebhookEvent | null>(null)
  const [tunnelID, setTunnelID] = useState<string>('')
  const [replayError, setReplayError] = useState<string | null>(null)

  useEffect(() => {
    api.getTunnels().then(tunnels => {
      const active = tunnels.find(t => t.Status === 'active')
      if (active) setTunnelID(active.ID)
    }).catch(() => {})
  }, [])

  useEffect(() => {
    if (!tunnelID) return
    api.getEvents(tunnelID, 100).then(setEvents).catch(() => {})
  }, [tunnelID])

  useWSEvents(tunnelID, event => setEvents(prev => [event, ...prev]))

  const handleReplay = useCallback(async (eventID: string, targetURL: string) => {
    setReplayError(null)
    try {
      await api.replay(eventID, targetURL)
    } catch (err) {
      setReplayError(err instanceof Error ? err.message : 'Replay failed')
    }
  }, [])

  return (
    <div className="flex h-screen bg-white font-sans text-sm">
      <div className="w-1/2 border-r flex flex-col">
        <div className="p-3 border-b text-xs text-gray-500 font-mono">
          {events.length} events
        </div>
        <EventList events={events} selectedID={selected?.ID ?? null} onSelect={setSelected} />
      </div>
      <div className="w-1/2 flex flex-col">
        {replayError && (
          <div className="bg-red-50 text-red-700 text-xs px-4 py-2 border-b">{replayError}</div>
        )}
        {selected
          ? <EventDetail event={selected} onReplay={handleReplay} />
          : <div className="flex items-center justify-center h-full text-gray-400">Select an event</div>
        }
      </div>
    </div>
  )
}
