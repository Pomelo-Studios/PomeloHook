import { useState } from 'react'
import type { WebhookEvent } from '../types'

interface Props {
  event: WebhookEvent
  onReplay: (eventID: string, targetURL: string) => void
}

export function EventDetail({ event, onReplay }: Props) {
  const [targetURL, setTargetURL] = useState('http://localhost:3000')

  return (
    <div className="flex flex-col gap-4 p-4 overflow-y-auto h-full font-mono text-xs">
      <div>
        <div className="text-gray-500 mb-1">Request</div>
        <div className="bg-gray-100 rounded p-2">
          <div><span className="font-bold">{event.Method}</span> {event.Path}</div>
          <pre className="mt-2 whitespace-pre-wrap break-all">{event.RequestBody}</pre>
        </div>
      </div>

      <div>
        <div className="text-gray-500 mb-1">Response</div>
        <div className={`rounded p-2 ${event.ResponseStatus >= 400 ? 'bg-red-50' : 'bg-green-50'}`}>
          <div className="font-bold">{event.ResponseStatus} · {event.ResponseMS}ms</div>
          <pre className="mt-2 whitespace-pre-wrap break-all">{event.ResponseBody}</pre>
        </div>
      </div>

      <div className="flex gap-2 items-center">
        <input
          type="text"
          value={targetURL}
          onChange={e => setTargetURL(e.target.value)}
          className="flex-1 border rounded px-2 py-1 text-xs"
          placeholder="http://localhost:3000"
        />
        <button
          onClick={() => onReplay(event.ID, targetURL)}
          className="bg-blue-600 text-white rounded px-3 py-1 text-xs hover:bg-blue-700"
        >
          Replay
        </button>
      </div>
    </div>
  )
}
