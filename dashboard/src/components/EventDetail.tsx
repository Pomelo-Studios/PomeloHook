import { useState } from 'react'
import { RefreshCw } from 'lucide-react'
import type { WebhookEvent } from '../types'
import { JsonView } from './JsonView'
import { formatTime } from '../utils/formatTime'

interface Props {
  event: WebhookEvent
  onReplay: (eventID: string, targetURL: string) => void
}

function ResponsePill({ event }: { event: WebhookEvent }) {
  const style: React.CSSProperties =
    !event.Forwarded
      ? { background: 'var(--method-dim-bg)', color: 'var(--text-dim)' }
      : event.ResponseStatus >= 400
        ? { background: 'var(--err-bg)', color: 'var(--err-text)' }
        : { background: 'var(--ok-bg)', color: 'var(--ok-text)' }
  const label = !event.Forwarded ? 'not forwarded' : event.ResponseStatus >= 400 ? String(event.ResponseStatus) : `${event.ResponseStatus} OK`
  return <span className="text-[10px] font-semibold px-2 py-[2px] rounded-full flex-shrink-0" style={style}>{label}</span>
}

function responseCodeStyle(event: WebhookEvent): React.CSSProperties {
  if (!event.Forwarded) return { background: 'var(--code-bg)', border: '1px solid var(--code-border)' }
  if (event.ResponseStatus >= 400) return { background: 'var(--err-bg)', border: '1px solid var(--selected-border)' }
  return { background: 'var(--ok-bg)', border: '1px solid var(--ok-bg)' }
}

export function EventDetail({ event, onReplay }: Props) {
  const [targetURL, setTargetURL] = useState('http://localhost:3000')

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div
        className="px-5 py-[14px] flex items-center gap-2 flex-shrink-0 border-b"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <span className="font-mono text-[9px] font-bold px-[7px] py-[2px] rounded-[4px] bg-coral text-white flex-shrink-0">
          {event.Method}
        </span>
        <span className="font-mono text-[13px] font-semibold flex-1 truncate" style={{ color: 'var(--text-primary)' }}>
          {event.Path}
        </span>
        <span className="font-mono text-[10px]" style={{ color: 'var(--text-dim)' }}>
          {formatTime(event.ReceivedAt)}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto p-5 flex flex-col gap-4">
        <div>
          <div className="text-[10px] font-bold tracking-[1.5px] uppercase mb-2" style={{ color: 'var(--text-dim)' }}>
            Request Body
          </div>
          <div className="rounded-[10px] p-[14px]" style={{ background: 'var(--code-bg)', border: '1px solid var(--code-border)' }}>
            <JsonView value={event.RequestBody} />
          </div>
        </div>

        <div>
          <div className="flex items-center gap-2 mb-2">
            <span className="text-[10px] font-bold tracking-[1.5px] uppercase" style={{ color: 'var(--text-dim)' }}>
              Response
            </span>
            <ResponsePill event={event} />
            {event.ResponseMS > 0 && (
              <span className="font-mono text-[10px]" style={{ color: 'var(--text-dim)' }}>{event.ResponseMS}ms</span>
            )}
          </div>
          <div className="rounded-[10px] p-[14px]" style={responseCodeStyle(event)}>
            {event.ResponseBody
              ? <JsonView value={event.ResponseBody} />
              : <span className="font-mono text-[10px]" style={{ color: 'var(--text-dim)' }}>—</span>
            }
          </div>
        </div>
      </div>

      <div
        className="px-5 py-3 flex gap-2 items-center flex-shrink-0 border-t"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <input
          type="text"
          value={targetURL}
          onChange={e => setTargetURL(e.target.value)}
          className="flex-1 rounded-lg px-3 py-2 font-mono text-[11px] outline-none"
          style={{ background: 'var(--bg)', border: '1px solid var(--border)', color: 'var(--text-secondary)' }}
          placeholder="http://localhost:3000"
        />
        <button
          onClick={() => onReplay(event.ID, targetURL)}
          className="flex items-center gap-[6px] bg-coral hover:opacity-90 text-white rounded-lg px-4 py-2 text-[11px] font-bold transition-opacity flex-shrink-0"
        >
          <RefreshCw size={12} strokeWidth={2.5} />
          Replay
        </button>
      </div>
    </div>
  )
}
