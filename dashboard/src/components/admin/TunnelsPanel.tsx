import { useState, useEffect } from 'react'
import { Trash2, Unplug } from 'lucide-react'
import { api } from '../../api/client'
import { ConfirmDialog } from './ConfirmDialog'
import type { Tunnel, ConfirmState } from '../../types'

interface Props { apiKey: string }

export function TunnelsPanel({ apiKey }: Props) {
  const [tunnels, setTunnels] = useState<Tunnel[]>([])
  const [loading, setLoading] = useState(true)
  const [confirm, setConfirm] = useState<ConfirmState | null>(null)
  const [error, setError] = useState('')

  function load() {
    api.admin.listTunnels(apiKey).then(setTunnels).catch(() => setError('Failed to load tunnels')).finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [apiKey])

  function confirmDelete(t: Tunnel) {
    setConfirm({
      message: `Delete tunnel ${t.subdomain}?`,
      detail: 'This also deletes all associated events.',
      onConfirm: async () => {
        setConfirm(null)
        await api.admin.deleteTunnel(apiKey, t.id).catch(() => setError('Delete failed'))
        load()
      },
    })
  }

  function confirmDisconnect(t: Tunnel) {
    setConfirm({
      message: `Disconnect tunnel ${t.subdomain}?`,
      detail: 'The active WebSocket connection will be closed.',
      onConfirm: async () => {
        setConfirm(null)
        await api.admin.disconnectTunnel(apiKey, t.id).catch(() => setError('Disconnect failed'))
        load()
      },
    })
  }

  if (loading) return <div className="p-4 text-xs font-mono" style={{ color: 'var(--text-dim)' }}>Loading…</div>

  return (
    <div className="flex flex-col h-full">
      {confirm && <ConfirmDialog message={confirm.message} detail={confirm.detail} onConfirm={confirm.onConfirm} onCancel={() => setConfirm(null)} />}

      <div
        className="h-[52px] flex items-center px-5 flex-shrink-0 border-b"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <div>
          <div className="text-[14px] font-bold" style={{ color: 'var(--text-primary)' }}>Tunnels</div>
          <div className="text-[11px]" style={{ color: 'var(--text-dim)' }}>{tunnels.length} tunnels</div>
        </div>
      </div>

      {error && (
        <div className="text-xs px-5 py-2 border-b" style={{ background: 'var(--err-bg)', color: 'var(--err-text)', borderColor: 'var(--selected-border)' }}>
          {error}
        </div>
      )}

      <div className="flex-1 overflow-auto">
        <table className="w-full border-collapse">
          <thead className="sticky top-0">
            <tr style={{ background: 'var(--surface)' }}>
              {['Subdomain', 'Type', 'Status', 'Actions'].map(h => (
                <th key={h} className="text-left text-[9px] font-bold tracking-[1.5px] uppercase px-4 py-2 border-b" style={{ color: 'var(--text-dim)', borderColor: 'var(--border)' }}>
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {tunnels.map(t => (
              <tr key={t.id} className="group transition-colors" style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                <td className="px-4 py-3 text-xs font-mono font-semibold" style={{ color: 'var(--text-primary)' }}>{t.subdomain}</td>
                <td className="px-4 py-3 text-xs" style={{ color: 'var(--text-secondary)' }}>{t.type}</td>
                <td className="px-4 py-3">
                  <span
                    className="text-[10px] font-semibold px-2 py-[2px] rounded-full uppercase"
                    style={
                      t.status === 'active'
                        ? { background: 'var(--ok-bg)', color: 'var(--ok-text)' }
                        : { background: 'var(--method-dim-bg)', color: 'var(--text-dim)' }
                    }
                  >
                    {t.status}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    {t.status === 'active' && (
                      <button
                        onClick={() => confirmDisconnect(t)}
                        className="flex items-center gap-1 text-[10px] px-2 py-[3px] rounded-md"
                        style={{ border: '1px solid var(--border)', color: 'var(--text-secondary)' }}
                      >
                        <Unplug size={10} /> Disconnect
                      </button>
                    )}
                    <button
                      onClick={() => confirmDelete(t)}
                      className="flex items-center gap-1 text-[10px] px-2 py-[3px] rounded-md"
                      style={{ border: '1px solid var(--selected-border)', color: 'var(--err-text)' }}
                    >
                      <Trash2 size={10} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
