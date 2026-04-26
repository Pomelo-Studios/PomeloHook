import { useState, useEffect, useCallback } from 'react'
import { EventList } from './components/EventList'
import { EventDetail } from './components/EventDetail'
import { api } from './api/client'
import type { WebhookEvent } from './types'

export default function App() {
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [selected, setSelected] = useState<WebhookEvent | null>(null)
  const [tunnelID, setTunnelID] = useState<string>('')

  useEffect(() => {
    api.getTunnels().then(tunnels => {
      const active = tunnels.find(t => t.Status === 'active')
      if (active) setTunnelID(active.ID)
    }).catch(() => {})
  }, [])

  useEffect(() => {
    if (!tunnelID) return
    api.getEvents(tunnelID, 100).then(setEvents).catch(() => {})

    const ws = new WebSocket(`ws://${location.host}/api/events/stream?tunnel_id=${tunnelID}`)
    ws.onmessage = e => {
      const event: WebhookEvent = JSON.parse(e.data)
      setEvents(prev => [event, ...prev])
    }
    return () => ws.close()
  }, [tunnelID])

  const handleReplay = useCallback(async (eventID: string, targetURL: string) => {
    await api.replay(eventID, targetURL)
  }, [])

  return (
    <div className="flex h-screen bg-white font-sans text-sm">
      <div className="w-1/2 border-r flex flex-col">
        <div className="p-3 border-b text-xs text-gray-500 font-mono">
          {events.length} events
        </div>
        <EventList events={events} selectedID={selected?.ID ?? null} onSelect={setSelected} />
      </div>
      <div className="w-1/2">
        {selected
          ? <EventDetail event={selected} onReplay={handleReplay} />
          : <div className="flex items-center justify-center h-full text-gray-400">Select an event</div>
        }
      </div>
    </div>
  )
}
