import type { WebhookEvent } from '../types'
import { formatTime } from '../utils/formatTime'

interface Props {
  events: WebhookEvent[]
  selectedID: string | null
  onSelect: (event: WebhookEvent) => void
  tunnelSubdomain?: string
}

function StatusPill({ event }: { event: WebhookEvent }) {
  const style: React.CSSProperties =
    !event.Forwarded
      ? { background: 'var(--surface2)', color: 'var(--text-3)' }
      : event.ResponseStatus >= 400
        ? { background: 'var(--err-bg)', color: 'var(--err-text)' }
        : { background: 'var(--ok-bg)', color: 'var(--ok-text)' }

  return (
    <span className="text-[10px] font-semibold px-2 py-[2px] rounded-full flex-shrink-0" style={style}>
      {!event.Forwarded ? '—' : event.ResponseStatus}
    </span>
  )
}

function MethodBadge({ method, selected }: { method: string; selected: boolean }) {
  return (
    <span
      className="font-mono text-[9px] font-bold px-[7px] py-[2px] rounded-[4px] flex-shrink-0"
      style={
        selected
          ? { background: '#FF6B6B', color: 'white' }
          : { background: 'var(--surface2)', color: 'var(--text-3)' }
      }
    >
      {method}
    </span>
  )
}

export function EventList({ events, selectedID, onSelect, tunnelSubdomain }: Props) {
  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div
        className="px-4 py-[10px] flex items-center justify-between flex-shrink-0 border-b"
        style={{ borderColor: 'var(--border)' }}
      >
        <span className="text-[10px] font-bold tracking-[1.5px] uppercase" style={{ color: 'var(--text-3)' }}>
          Events
        </span>
        <span
          className="text-[10px] font-medium px-2 py-[1px] rounded-full"
          style={{ background: 'var(--surface2)', color: 'var(--text-3)' }}
        >
          {events.length}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto">
        {events.map(event => {
          const selected = event.ID === selectedID
          return (
            <button
              key={event.ID}
              onClick={() => onSelect(event)}
              className="w-full text-left px-4 py-[10px] flex flex-col gap-1 border-b border-l-[3px] transition-colors"
              style={{
                borderBottomColor: 'var(--border)',
                borderLeftColor: selected ? '#FF6B6B' : 'transparent',
                background: selected ? 'var(--selected-bg)' : 'transparent',
              }}
            >
              <div className="flex items-center gap-[6px]">
                <MethodBadge method={event.Method} selected={selected} />
                <span
                  className="text-[11px] font-mono flex-1 truncate"
                  style={{ color: selected ? 'var(--text)' : 'var(--text-2)' }}
                >
                  {event.Path}
                </span>
                <StatusPill event={event} />
              </div>
              <div className="font-mono text-[9px]" style={{ color: 'var(--text-3)' }}>
                {event.ResponseMS ? `${event.ResponseMS}ms` : '—'} · {formatTime(event.ReceivedAt)}
              </div>
            </button>
          )
        })}
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
