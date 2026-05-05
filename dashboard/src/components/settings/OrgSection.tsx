import { useState } from 'react'
import type { Me } from '../../types'
import { api } from '../../api/client'

interface Props {
  apiKey: string
  me: Me | null
  onOrgNameSaved?: (name: string) => void
}

export function OrgSection({ apiKey, me, onOrgNameSaved }: Props) {
  const [name, setName] = useState(me?.org_name ?? '')
  const [saved, setSaved] = useState(false)

  async function handleSave() {
    await api.org.updateSettings(apiKey, name)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
    onOrgNameSaved?.(name)
  }

  return (
    <div style={{ maxWidth: 380 }}>
      <h3 style={{ fontSize: 13, color: 'var(--text-2)', marginBottom: 16, fontWeight: 600 }}>Organization</h3>
      <label style={labelStyle}>Display Name</label>
      <input value={name} onChange={e => setName(e.target.value)} style={inputStyle} />
      <button onClick={handleSave} disabled={!name} style={{ ...btnStyle, opacity: name ? 1 : 0.5 }}>
        {saved ? 'Saved ✓' : 'Save'}
      </button>
    </div>
  )
}

const labelStyle: React.CSSProperties = { display: 'block', fontSize: 11, color: 'var(--text-3)', marginBottom: 3 }
const inputStyle: React.CSSProperties = { display: 'block', width: '100%', marginBottom: 10, padding: '5px 8px', background: 'var(--surface2)', color: 'var(--text)', border: '1px solid var(--border)', borderRadius: 4, fontSize: 12, boxSizing: 'border-box' }
const btnStyle: React.CSSProperties = { fontSize: 11, padding: '4px 12px', background: 'var(--selected-bg)', color: 'var(--coral)', border: '1px solid var(--selected-border)', borderRadius: 6, cursor: 'pointer' }
