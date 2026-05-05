import { useEffect } from 'react'
import type { WebhookEvent } from '../types'
import { formatTime } from '../utils/formatTime'
import { Badge, methodVariant, statusVariant, EventRowSkeleton, EmptyState } from './ui'

interface Props {
  events: WebhookEvent[]
  selectedID: string | null
  onSelect: (event: WebhookEvent) => void
  tunnelSubdomain?: string
  loading?: boolean
}

export function EventList({ events, selectedID, onSelect, tunnelSubdomain, loading = false }: Props) {
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (!events.length) return
      const idx = events.findIndex(ev => ev.ID === selectedID)
      if (e.key === 'ArrowDown' || e.key === 'j') {
        e.preventDefault()
        const next = events[Math.min(idx + 1, events.length - 1)]
        if (next) onSelect(next)
      }
      if (e.key === 'ArrowUp' || e.key === 'k') {
        e.preventDefault()
        const prev = events[Math.max(idx - 1, 0)]
        if (prev) onSelect(prev)
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [events, selectedID, onSelect])

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div
        className="px-4 py-[10px] flex items-center justify-between flex-shrink-0 border-b"
        style={{ borderColor: 'var(--border)' }}
      >
        <span className="text-[10px] font-bold tracking-[1.5px] uppercase" style={{ color: 'var(--text-3)' }}>
          Events
        </span>
        <div className="flex items-center gap-2">
          <span style={{ fontSize: '10px', color: 'var(--text-3)', fontFamily: 'var(--font-mono)' }}>↑↓ navigate</span>
          <span
            className="text-[10px] font-medium px-2 py-[1px] rounded-full"
            style={{ background: 'var(--surface2)', color: 'var(--text-3)' }}
          >
            {events.length}
          </span>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {loading ? (
          Array.from({ length: 5 }).map((_, i) => <EventRowSkeleton key={i} />)
        ) : events.length === 0 ? (
          <EmptyState
            icon="📭"
            title="Henüz event yok"
            subtitle="Bu tunnel'a ilk istek geldiğinde burada görünecek."
          />
        ) : (
          events.map(event => {
            const selected = event.ID === selectedID
            return (
              <button
                key={event.ID}
                onClick={() => onSelect(event)}
                className="w-full text-left px-4 py-[10px] flex flex-col gap-1 border-b border-l-[3px] transition-colors"
                style={{
                  borderBottomColor: 'var(--border)',
                  borderLeftColor: selected ? 'var(--coral)' : 'transparent',
                  background: selected ? 'var(--selected-bg)' : 'transparent',
                }}
              >
                <div className="flex items-center gap-[6px]">
                  <Badge variant={methodVariant(event.Method)}>{event.Method}</Badge>
                  <span
                    className="text-[11px] font-mono flex-1 truncate"
                    style={{ color: selected ? 'var(--text)' : 'var(--text-2)' }}
                  >
                    {event.Path}
                  </span>
                  {!event.Forwarded ? (
                    <Badge variant="selected" style={{ fontSize: '9px', padding: '1px 5px' }}>—</Badge>
                  ) : (
                    <Badge variant={statusVariant(event.ResponseStatus)} style={{ fontSize: '9px', padding: '1px 5px' }}>{event.ResponseStatus}</Badge>
                  )}
                </div>
                <div className="font-mono text-[9px]" style={{ color: 'var(--text-3)' }}>
                  {event.ResponseMS ? `${event.ResponseMS}ms` : '—'} · {formatTime(event.ReceivedAt)}
                </div>
              </button>
            )
          })
        )}
      </div>

      {tunnelSubdomain && (
        <div className="px-4 py-[10px] flex-shrink-0 border-t" style={{ borderColor: 'var(--border)' }}>
          <div className="text-[9px] font-medium mb-[3px]" style={{ color: 'var(--text-3)' }}>Webhook URL</div>
          <div className="font-mono text-[10px] truncate" style={{ color: 'var(--text-2)' }}>
            /webhook/{tunnelSubdomain}
          </div>
        </div>
      )}
    </div>
  )
}
