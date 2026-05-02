import { useState } from 'react'
import type { Me } from '../../types'
import { api } from '../../api/client'

interface Props {
  apiKey: string
  me: Me | null
}

export function OrgSection({ apiKey, me }: Props) {
  const [name, setName] = useState(me?.org_name ?? '')
  const [saved, setSaved] = useState(false)

  async function handleSave() {
    await api.org.updateSettings(apiKey, name)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <div style={{ maxWidth: 380 }}>
      <h3 style={{ fontSize: 13, color: '#ccc', marginBottom: 16, fontWeight: 600 }}>Organization</h3>
      <label style={labelStyle}>Display Name</label>
      <input value={name} onChange={e => setName(e.target.value)} style={inputStyle} />
      <button onClick={handleSave} disabled={!name} style={{ ...btnStyle, opacity: name ? 1 : 0.5 }}>
        {saved ? 'Saved ✓' : 'Save'}
      </button>
    </div>
  )
}

const labelStyle: React.CSSProperties = { display: 'block', fontSize: 11, color: '#555', marginBottom: 3 }
const inputStyle: React.CSSProperties = { display: 'block', width: '100%', marginBottom: 10, padding: '5px 8px', background: '#222', color: '#ccc', border: '1px solid #2a2a2a', borderRadius: 4, fontSize: 12, boxSizing: 'border-box' }
const btnStyle: React.CSSProperties = { fontSize: 11, padding: '4px 12px', background: 'rgba(255,107,107,0.13)', color: '#FF6B6B', border: '1px solid rgba(255,107,107,0.3)', borderRadius: 6, cursor: 'pointer' }
