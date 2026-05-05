import { useMemo } from 'react'
import type { ReactNode } from 'react'

function renderValue(v: unknown): ReactNode {
  if (v === null) return <span style={{ color: 'var(--text-3)' }}>null</span>
  if (typeof v === 'boolean') return <span style={{ color: 'var(--ok-text)' }}>{String(v)}</span>
  if (typeof v === 'number') return <span style={{ color: 'var(--text)' }}>{v}</span>
  if (typeof v === 'string') return <span style={{ color: 'var(--text-2)' }}>"{v}"</span>
  if (Array.isArray(v)) {
    if (v.length === 0) return <span style={{ color: 'var(--text-3)' }}>[]</span>
    return (
      <>{`[\n`}{v.map((item, i) => (
        <span key={i}>{'  '}{renderValue(item)}{i < v.length - 1 ? ',' : ''}{'\n'}</span>
      ))}{`]`}</>
    )
  }
  if (typeof v === 'object') {
    const entries = Object.entries(v as Record<string, unknown>)
    if (entries.length === 0) return <span style={{ color: 'var(--text-3)' }}>{'{}'}</span>
    return (
      <>{`{\n`}{entries.map(([k, val], i) => (
        <span key={k}>
          {'  '}<span style={{ color: 'var(--text)' }}>"{k}"</span>{': '}{renderValue(val)}{i < entries.length - 1 ? ',' : ''}{'\n'}
        </span>
      ))}{`}`}</>
    )
  }
  return <span style={{ color: 'var(--text-3)' }}>{String(v)}</span>
}

interface Props {
  value: string
}

export function JsonView({ value }: Props) {
  const content = useMemo(() => {
    try {
      return { ok: true, node: renderValue(JSON.parse(value)) }
    } catch {
      return { ok: false, node: null }
    }
  }, [value])

  return (
    <pre className="text-[11px] leading-relaxed font-mono whitespace-pre-wrap break-all" style={{ color: 'var(--code-text)' }}>
      {content.ok ? content.node : value}
    </pre>
  )
}
