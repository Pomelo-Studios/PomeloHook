import { useState, useEffect } from 'react'
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
      message: `Delete tunnel ${t.Subdomain}?`,
      detail: 'This also deletes all associated events.',
      onConfirm: async () => {
        setConfirm(null)
        await api.admin.deleteTunnel(apiKey, t.ID).catch(() => setError('Delete failed'))
        load()
      },
    })
  }

  function confirmDisconnect(t: Tunnel) {
    setConfirm({
      message: `Disconnect tunnel ${t.Subdomain}?`,
      detail: 'The active WebSocket connection will be closed.',
      onConfirm: async () => {
        setConfirm(null)
        await api.admin.disconnectTunnel(apiKey, t.ID).catch(() => setError('Disconnect failed'))
        load()
      },
    })
  }

  if (loading) return <div className="p-4 text-zinc-600 text-xs">Loading…</div>

  return (
    <div className="flex flex-col h-full">
      {confirm && <ConfirmDialog message={confirm.message} detail={confirm.detail} onConfirm={confirm.onConfirm} onCancel={() => setConfirm(null)} />}
      <div className="h-11 border-b border-zinc-800 flex items-center px-4 flex-shrink-0 bg-zinc-900/30">
        <span className="text-zinc-300 text-xs font-semibold">Tunnels</span>
      </div>
      {error && <div className="bg-red-950 text-red-400 text-xs px-4 py-2 border-b border-red-900">{error}</div>}
      <div className="flex-1 overflow-auto">
        <table className="w-full border-collapse">
          <thead className="sticky top-0">
            <tr className="bg-zinc-900/80">
              {['Subdomain', 'Type', 'Status', 'Actions'].map(h => (
                <th key={h} className="text-left text-[9px] uppercase tracking-widest text-zinc-500 px-3 py-2 border-b border-zinc-800 font-semibold">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {tunnels.map(t => (
              <tr key={t.ID} className="hover:bg-zinc-900/40 group">
                <td className="px-3 py-2 text-xs font-mono text-zinc-300">{t.Subdomain}</td>
                <td className="px-3 py-2 text-xs text-zinc-500">{t.Type}</td>
                <td className="px-3 py-2">
                  <span className={`text-[9px] px-1.5 py-0.5 rounded font-semibold uppercase ${t.Status === 'active' ? 'bg-emerald-950 text-emerald-400' : 'bg-zinc-800 text-zinc-600'}`}>{t.Status}</span>
                </td>
                <td className="px-3 py-2">
                  <div className="flex gap-1 opacity-0 group-hover:opacity-100">
                    {t.Status === 'active' && (
                      <button onClick={() => confirmDisconnect(t)} className="text-[10px] px-2 py-0.5 border border-zinc-700 text-zinc-400 rounded hover:text-zinc-200">Disconnect</button>
                    )}
                    <button onClick={() => confirmDelete(t)} className="text-[10px] px-2 py-0.5 border border-red-900 text-red-500 rounded hover:text-red-300">Delete</button>
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
