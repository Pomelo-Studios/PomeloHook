import { useState } from 'react'
import type { Me } from '../types'
import { MembersSection } from './settings/MembersSection'
import { RolesSection } from './settings/RolesSection'
import { OrgSection } from './settings/OrgSection'

type SettingsSection = 'members' | 'roles' | 'org'

interface Props {
  apiKey: string
  me: Me | null
  can: (perm: string) => boolean
}

export function SettingsTab({ apiKey, me, can }: Props) {
  const sections: { id: SettingsSection; label: string; show: boolean }[] = [
    { id: 'members', label: 'Members', show: true },
    { id: 'roles', label: 'Roles', show: true },
    { id: 'org', label: 'Organization', show: can('edit_org_settings') },
  ]
  const visible = sections.filter(s => s.show)
  const [active, setActive] = useState<SettingsSection>(visible[0]?.id ?? 'members')

  return (
    <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
      <div style={{ width: 140, borderRight: '1px solid var(--border)', paddingTop: 8, flexShrink: 0 }}>
        {visible.map(s => (
          <button
            key={s.id}
            onClick={() => setActive(s.id)}
            style={{
              display: 'block', width: '100%', textAlign: 'left',
              padding: '6px 14px', fontSize: 11,
              background: active === s.id ? 'rgba(255,107,107,0.08)' : 'transparent',
              color: active === s.id ? '#FF6B6B' : '#555',
              fontWeight: active === s.id ? 600 : 400,
              borderLeft: active === s.id ? '2px solid #FF6B6B' : '2px solid transparent',
              border: 'none', cursor: 'pointer',
            }}
          >
            {s.label}
          </button>
        ))}
      </div>
      <div style={{ flex: 1, overflow: 'auto', padding: 20 }}>
        {active === 'members' && <MembersSection apiKey={apiKey} can={can} />}
        {active === 'roles' && <RolesSection apiKey={apiKey} can={can} />}
        {active === 'org' && can('edit_org_settings') && <OrgSection apiKey={apiKey} me={me} />}
      </div>
    </div>
  )
}
