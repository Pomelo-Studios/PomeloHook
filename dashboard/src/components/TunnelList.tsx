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
      style={{ color: 'var(--text-dim)' }}
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
        <span className="text-[10px] font-bold tracking-[1.5px] uppercase" style={{ color: 'var(--text-dim)' }}>
          Tunnels
        </span>
        <span
          className="text-[10px] font-medium px-2 py-[1px] rounded-full"
          style={{ background: 'var(--method-dim-bg)', color: 'var(--text-dim)' }}
        >
          {tunnels.length}
        </span>
      </div>
      <div className="flex-1 overflow-y-auto">
        {tunnels.length === 0 && (
          <div className="px-4 py-6 text-[11px] text-center" style={{ color: 'var(--text-dim)' }}>
            No tunnels
          </div>
        )}
        {tunnels.map(tunnel => {
          const selected = tunnel.ID === selectedID
          const isActive = tunnel.Status === 'active'
          const webhookURL = `${origin}/webhook/${tunnel.Subdomain}`
          return (
            <button
              key={tunnel.ID}
              onClick={() => onSelect(tunnel)}
              className="w-full text-left px-4 py-[10px] flex flex-col gap-1 border-b border-l-[3px] transition-colors"
              style={{
                borderBottomColor: 'var(--border-subtle)',
                borderLeftColor: selected ? '#FF6B6B' : 'transparent',
                background: selected ? 'var(--selected-bg)' : 'transparent',
              }}
            >
              <div className="flex items-center gap-[6px]">
                <span
                  className="w-[6px] h-[6px] rounded-full flex-shrink-0"
                  style={{ background: isActive ? '#50cc80' : 'var(--text-dim)' }}
                />
                <div className="flex-1 min-w-0">
                  <div
                    className="text-[11px] font-mono truncate"
                    style={{ color: selected ? 'var(--text-primary)' : 'var(--text-secondary)' }}
                  >
                    {tunnel.DisplayName || tunnel.Subdomain}
                  </div>
                  {tunnel.DisplayName && (
                    <div className="text-[9px] font-mono truncate" style={{ color: 'var(--text-dim)' }}>
                      {tunnel.Subdomain}
                    </div>
                  )}
                </div>
                <CopyButton text={webhookURL} />
              </div>
              {isActive && tunnel.ActiveDevice && (
                <div className="font-mono text-[9px] pl-[12px]" style={{ color: 'var(--text-dim)' }}>
                  {tunnel.ActiveDevice}
                </div>
              )}
              {selected && (
                <div
                  className="font-mono text-[9px] pl-[12px] truncate"
                  style={{ color: 'var(--text-dim)' }}
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
