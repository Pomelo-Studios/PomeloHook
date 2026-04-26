import type { WebhookEvent } from '../types'

interface Props {
  events: WebhookEvent[]
  selectedID: string | null
  onSelect: (event: WebhookEvent) => void
}

export function EventList({ events, selectedID, onSelect }: Props) {
  return (
    <div className="flex flex-col divide-y divide-gray-200 overflow-y-auto h-full">
      {events.map(event => (
        <button
          key={event.ID}
          onClick={() => onSelect(event)}
          className={`flex items-center gap-3 p-3 text-left hover:bg-gray-50 ${selectedID === event.ID ? 'bg-blue-50' : ''}`}
        >
          <span
            data-testid="status-indicator"
            className={`h-2 w-2 rounded-full flex-shrink-0 ${event.Forwarded ? 'bg-green-500' : 'bg-red-500'}`}
          />
          <span className="font-mono text-xs font-bold text-gray-600 w-12">{event.Method}</span>
          <span className="font-mono text-xs text-gray-800 truncate flex-1">{event.Path}</span>
          <span className="text-xs text-gray-400">{event.ResponseStatus || '—'}</span>
          <span className="text-xs text-gray-400">{event.ResponseMS ? `${event.ResponseMS}ms` : '—'}</span>
        </button>
      ))}
    </div>
  )
}
