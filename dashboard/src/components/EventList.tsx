import type { WebhookEvent } from '../types'

interface Props {
  events: WebhookEvent[]
  selectedID: string | null
  onSelect: (event: WebhookEvent) => void
  tunnelSubdomain?: string
}

function StatusBadge({ event }: { event: WebhookEvent }) {
  if (!event.Forwarded)
    return <span className="text-[9px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-500">err</span>
  if (event.ResponseStatus >= 400)
    return <span className="text-[9px] px-1.5 py-0.5 rounded bg-red-950 text-red-400">{event.ResponseStatus}</span>
  return <span className="text-[9px] px-1.5 py-0.5 rounded bg-emerald-950 text-emerald-400">{event.ResponseStatus}</span>
}

function formatTime(iso: string): string {
  const d = new Date(iso)
  return `${String(d.getUTCHours()).padStart(2, '0')}:${String(d.getUTCMinutes()).padStart(2, '0')}:${String(d.getUTCSeconds()).padStart(2, '0')}`
}

export function EventList({ events, selectedID, onSelect, tunnelSubdomain }: Props) {
  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="px-3 py-2 border-b border-zinc-800 flex items-center justify-between flex-shrink-0">
        <span className="text-[10px] text-zinc-500 uppercase tracking-widest">Events</span>
        <span className="text-[9px] px-2 py-0.5 rounded-full bg-zinc-900 border border-zinc-800 text-zinc-600">
          {events.length}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto divide-y divide-zinc-900">
        {events.map(event => {
          const selected = event.ID === selectedID
          return (
            <button
              key={event.ID}
              onClick={() => onSelect(event)}
              className={`w-full text-left px-3 py-2.5 border-l-2 transition-colors ${
                selected ? 'bg-zinc-900 border-emerald-500' : 'border-transparent hover:bg-zinc-900/50'
              }`}
            >
              <div className="flex items-center gap-2 mb-1">
                <span className={`text-[8px] font-bold px-1.5 py-0.5 rounded flex-shrink-0 ${
                  selected ? 'bg-emerald-500 text-zinc-950' : 'bg-zinc-800 text-zinc-400'
                }`}>
                  {event.Method}
                </span>
                <span className={`text-[10px] truncate flex-1 ${selected ? 'text-zinc-50' : 'text-zinc-400'}`}>
                  {event.Path}
                </span>
                <StatusBadge event={event} />
              </div>
              <div className="text-[9px] text-zinc-600 pl-0.5">
                {event.ResponseMS ? `${event.ResponseMS}ms` : '—'} · {formatTime(event.ReceivedAt)}
              </div>
            </button>
          )
        })}
      </div>

      {tunnelSubdomain && (
        <div className="px-3 py-2 border-t border-zinc-900 bg-zinc-950 flex-shrink-0">
          <div className="text-[9px] text-zinc-600 mb-0.5">Webhook URL</div>
          <div className="text-[9px] text-zinc-500 font-mono truncate">/webhook/{tunnelSubdomain}</div>
        </div>
      )}
    </div>
  )
}
