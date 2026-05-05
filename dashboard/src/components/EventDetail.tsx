import { useState } from 'react'
import { RefreshCw } from 'lucide-react'
import type { WebhookEvent } from '../types'
import { JsonView } from './JsonView'
import { formatTime } from '../utils/formatTime'
import { Badge, methodVariant, statusVariant, Button, Input, useToast } from './ui'

interface Props {
  event: WebhookEvent
  onReplay: (eventID: string, targetURL: string) => Promise<void>
}

function responseCodeStyle(event: WebhookEvent): React.CSSProperties {
  if (!event.Forwarded) return { background: 'var(--code-bg)', border: '1px solid var(--code-border)' }
  if (event.ResponseStatus >= 400) return { background: 'var(--err-bg)', border: '1px solid var(--selected-border)' }
  return { background: 'var(--ok-bg)', border: '1px solid var(--ok-bg)' }
}

export function EventDetail({ event, onReplay }: Props) {
  const [targetURL, setTargetURL] = useState('http://localhost:3000')
  const toast = useToast()

  async function handleReplay() {
    try {
      await onReplay(event.ID, targetURL)
      toast.success('Event replayed')
    } catch {
      toast.error('Replay failed')
    }
  }

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div
        className="px-5 py-[14px] flex items-center gap-2 flex-shrink-0 border-b"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <Badge variant={methodVariant(event.Method)} style={{ fontSize: '11px', padding: '3px 8px' }}>
          {event.Method}
        </Badge>
        <span className="font-mono text-[13px] font-semibold flex-1 truncate" style={{ color: 'var(--text)' }}>
          {event.Path}
        </span>
        <span className="font-mono text-[10px]" style={{ color: 'var(--text-3)' }}>
          {formatTime(event.ReceivedAt)}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto p-5 flex flex-col gap-4">
        <div>
          <div className="text-[10px] font-bold tracking-[1.5px] uppercase mb-2" style={{ color: 'var(--text-3)' }}>
            Request Body
          </div>
          <div className="rounded-[10px] p-[14px]" style={{ background: 'var(--code-bg)', border: '1px solid var(--code-border)' }}>
            <JsonView value={event.RequestBody} />
          </div>
        </div>

        <div>
          <div className="flex items-center gap-2 mb-2">
            <span className="text-[10px] font-bold tracking-[1.5px] uppercase" style={{ color: 'var(--text-3)' }}>
              Response
            </span>
            {event.Forwarded
              ? <Badge variant={statusVariant(event.ResponseStatus)} style={{ fontSize: '9px', padding: '1px 5px' }}>
                  {event.ResponseStatus >= 400 ? String(event.ResponseStatus) : `${event.ResponseStatus} OK`}
                </Badge>
              : <Badge variant="selected" style={{ fontSize: '9px', padding: '1px 5px' }}>not forwarded</Badge>
            }
            {event.ResponseMS > 0 && (
              <span className="font-mono text-[10px]" style={{ color: 'var(--text-3)' }}>{event.ResponseMS}ms</span>
            )}
          </div>
          <div className="rounded-[10px] p-[14px]" style={responseCodeStyle(event)}>
            {event.ResponseBody
              ? <JsonView value={event.ResponseBody} />
              : <span className="font-mono text-[10px]" style={{ color: 'var(--text-3)' }}>—</span>
            }
          </div>
        </div>
      </div>

      <div
        className="px-5 py-3 flex gap-2 items-center flex-shrink-0 border-t"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <Input
          type="text"
          value={targetURL}
          onChange={e => setTargetURL(e.target.value)}
          placeholder="http://localhost:3000"
          style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', background: 'var(--bg)' }}
        />
        <Button
          variant="primary"
          size="sm"
          onClick={handleReplay}
          style={{ display: 'flex', alignItems: 'center', gap: '6px', flexShrink: 0 }}
        >
          <RefreshCw size={12} strokeWidth={2.5} />
          Replay
        </Button>
      </div>
    </div>
  )
}
