import { Copy, Check } from 'lucide-react'
import { useState } from 'react'
import type { Tunnel } from '../types'

interface Props {
  tunnels: Tunnel[]
  selectedID: string | null
  onSelect: (tunnel: Tunnel) => void
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  function handleCopy(e: React.MouseEvent) {
    e.stopPropagation()
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    })
  }
  return (
    <button
      onClick={handleCopy}
      className="p-[2px] rounded opacity-60 hover:opacity-100 transition-opacity flex-shrink-0"
      style={{ color: 'var(--text-3)' }}
      title="Copy webhook URL"
    >
      {copied ? <Check size={10} strokeWidth={2.5} /> : <Copy size={10} strokeWidth={2} />}
    </button>
  )
}

export function TunnelList({ tunnels, selectedID, onSelect }: Props) {
  const origin = window.location.origin

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div
        className="px-4 py-[10px] flex items-center justify-between flex-shrink-0 border-b"
        style={{ borderColor: 'var(--border)' }}
      >
        <span className="text-[10px] font-bold tracking-[1.5px] uppercase" style={{ color: 'var(--text-3)' }}>
          Tunnels
        </span>
        <span
          className="text-[10px] font-medium px-2 py-[1px] rounded-full"
          style={{ background: 'var(--surface2)', color: 'var(--text-3)' }}
        >
          {tunnels.length}
        </span>
      </div>
      <div className="flex-1 overflow-y-auto">
        {tunnels.length === 0 && (
          <div className="px-4 py-6 text-[11px] text-center" style={{ color: 'var(--text-3)' }}>
            No tunnels
          </div>
        )}
        {tunnels.map(tunnel => {
          const selected = tunnel.id === selectedID
          const isActive = tunnel.status === 'active'
          const webhookURL = `${origin}/webhook/${tunnel.subdomain}`
          return (
            <button
              key={tunnel.id}
              onClick={() => onSelect(tunnel)}
              className="w-full text-left px-4 py-[10px] flex flex-col gap-1 border-b border-l-[3px] transition-colors"
              style={{
                borderBottomColor: 'var(--border)',
                borderLeftColor: selected ? '#FF6B6B' : 'transparent',
                background: selected ? 'var(--selected-bg)' : 'transparent',
              }}
            >
              <div className="flex items-center gap-[6px]">
                <span
                  className="w-[6px] h-[6px] rounded-full flex-shrink-0"
                  style={{ background: isActive ? '#50cc80' : 'var(--text-3)' }}
                />
                <div className="flex-1 min-w-0">
                  <div
                    className="text-[11px] font-mono truncate"
                    style={{ color: selected ? 'var(--text)' : 'var(--text-2)' }}
                  >
                    {tunnel.display_name || tunnel.subdomain}
                  </div>
                  {tunnel.display_name && (
                    <div className="text-[9px] font-mono truncate" style={{ color: 'var(--text-3)' }}>
                      {tunnel.subdomain}
                    </div>
                  )}
                </div>
                <CopyButton text={webhookURL} />
              </div>
              {isActive && tunnel.active_device && (
                <div className="font-mono text-[9px] pl-[12px]" style={{ color: 'var(--text-3)' }}>
                  {tunnel.active_device}
                </div>
              )}
              {selected && (
                <div
                  className="font-mono text-[9px] pl-[12px] truncate"
                  style={{ color: 'var(--text-3)' }}
                  title={webhookURL}
                >
                  {webhookURL}
                </div>
              )}
            </button>
          )
        })}
      </div>
    </div>
  )
}
