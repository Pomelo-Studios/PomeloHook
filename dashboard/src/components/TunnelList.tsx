import type { Tunnel } from '../types'

interface Props {
  tunnels: Tunnel[]
  selectedID: string | null
  onSelect: (tunnel: Tunnel) => void
}

export function TunnelList({ tunnels, selectedID, onSelect }: Props) {
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
        {tunnels.map(tunnel => {
          const selected = tunnel.ID === selectedID
          const isActive = tunnel.Status === 'active'
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
                <span
                  className="text-[11px] font-mono flex-1 truncate"
                  style={{ color: selected ? 'var(--text-primary)' : 'var(--text-secondary)' }}
                >
                  {tunnel.Subdomain}
                </span>
              </div>
              {isActive && tunnel.ActiveDevice && (
                <div className="font-mono text-[9px] pl-[12px]" style={{ color: 'var(--text-dim)' }}>
                  {tunnel.ActiveDevice}
                </div>
              )}
            </button>
          )
        })}
      </div>
    </div>
  )
}
