import { useState } from 'react'
import type { WebhookEvent } from '../types'
import { JsonView } from './JsonView'

interface Props {
  event: WebhookEvent
  onReplay: (eventID: string, targetURL: string) => void
}

function formatTime(iso: string): string {
  const d = new Date(iso)
  return `${String(d.getUTCHours()).padStart(2, '0')}:${String(d.getUTCMinutes()).padStart(2, '0')}:${String(d.getUTCSeconds()).padStart(2, '0')}`
}

function ResponseBadge({ event }: { event: WebhookEvent }) {
  if (!event.Forwarded)
    return <span className="text-[9px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-500">not forwarded</span>
  if (event.ResponseStatus >= 400)
    return <span className="text-[9px] px-1.5 py-0.5 rounded bg-red-950 text-red-400">{event.ResponseStatus}</span>
  return <span className="text-[9px] px-1.5 py-0.5 rounded bg-emerald-950 text-emerald-400">{event.ResponseStatus} OK</span>
}

function responseBoxClass(event: WebhookEvent): string {
  if (!event.Forwarded) return 'bg-zinc-900 border border-zinc-800'
  if (event.ResponseStatus >= 400) return 'bg-red-950/20 border border-red-900/40'
  return 'bg-emerald-950/20 border border-emerald-900/30'
}

export function EventDetail({ event, onReplay }: Props) {
  const [targetURL, setTargetURL] = useState('http://localhost:3000')

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="px-4 py-2.5 border-b border-zinc-800 flex items-center gap-2 flex-shrink-0">
        <span className="text-[9px] font-bold px-1.5 py-0.5 rounded bg-emerald-500 text-zinc-950">
          {event.Method}
        </span>
        <span className="text-zinc-50 text-[11px] font-medium flex-1 truncate">{event.Path}</span>
        <span className="text-zinc-600 text-[9px]">{formatTime(event.ReceivedAt)}</span>
      </div>

      <div className="flex-1 overflow-y-auto p-4 flex flex-col gap-4">
        <div>
          <div className="text-[9px] text-zinc-500 uppercase tracking-widest mb-2">Request Body</div>
          <div className="bg-zinc-900 border border-zinc-800 rounded-md p-3">
            <JsonView value={event.RequestBody} />
          </div>
        </div>

        <div>
          <div className="flex items-center gap-2 mb-2">
            <span className="text-[9px] text-zinc-500 uppercase tracking-widest">Response</span>
            <ResponseBadge event={event} />
            {event.ResponseMS > 0 && (
              <span className="text-[9px] text-zinc-600">{event.ResponseMS}ms</span>
            )}
          </div>
          <div className={`rounded-md p-3 ${responseBoxClass(event)}`}>
            {event.ResponseBody
              ? <JsonView value={event.ResponseBody} />
              : <span className="text-[10px] text-zinc-600 font-mono">—</span>
            }
          </div>
        </div>
      </div>

      <div className="px-4 py-3 border-t border-zinc-800 bg-zinc-950 flex gap-2 flex-shrink-0">
        <input
          type="text"
          value={targetURL}
          onChange={e => setTargetURL(e.target.value)}
          className="flex-1 bg-zinc-900 border border-zinc-800 rounded px-3 py-1.5 text-[10px] text-zinc-400 font-mono focus:outline-none focus:border-zinc-600"
          placeholder="http://localhost:3000"
        />
        <button
          onClick={() => onReplay(event.ID, targetURL)}
          className="bg-emerald-600 hover:bg-emerald-500 text-white rounded px-4 py-1.5 text-[10px] font-bold transition-colors"
        >
          ↺ Replay
        </button>
      </div>
    </div>
  )
}
