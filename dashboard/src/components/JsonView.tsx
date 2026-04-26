import { useMemo } from 'react'
import type { ReactNode } from 'react'

function renderValue(v: unknown): ReactNode {
  if (v === null) return <span className="text-zinc-500">null</span>
  if (typeof v === 'boolean') return <span className="text-blue-400">{String(v)}</span>
  if (typeof v === 'number') return <span className="text-blue-400">{v}</span>
  if (typeof v === 'string') return <span className="text-amber-300">"{v}"</span>
  if (Array.isArray(v)) {
    if (v.length === 0) return <span className="text-zinc-400">[]</span>
    return (
      <>{`[\n`}{v.map((item, i) => (
        <span key={i}>{'  '}{renderValue(item)}{i < v.length - 1 ? ',' : ''}{'\n'}</span>
      ))}{`]`}</>
    )
  }
  if (typeof v === 'object') {
    const entries = Object.entries(v as Record<string, unknown>)
    if (entries.length === 0) return <span className="text-zinc-400">{'{}'}</span>
    return (
      <>{`{\n`}{entries.map(([k, val], i) => (
        <span key={k}>
          {'  '}<span className="text-emerald-400">"{k}"</span>{': '}{renderValue(val)}{i < entries.length - 1 ? ',' : ''}{'\n'}
        </span>
      ))}{`}`}</>
    )
  }
  return <span className="text-zinc-400">{String(v)}</span>
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
    <pre className="text-xs leading-relaxed font-mono whitespace-pre-wrap break-all text-zinc-400">
      {content.ok ? content.node : value}
    </pre>
  )
}
