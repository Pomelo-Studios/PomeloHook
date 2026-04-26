interface Props {
  subdomain: string
  connected: boolean
}

export function Header({ subdomain, connected }: Props) {
  return (
    <header className="h-11 bg-zinc-900 border-b border-zinc-800 px-4 flex items-center gap-3 flex-shrink-0">
      <div className="flex items-center gap-2">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2.5">
          <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
        </svg>
        <span className="text-zinc-50 text-[13px] font-bold tracking-tight">PomeloHook</span>
      </div>
      <div className="w-px h-4 bg-zinc-800" />
      {subdomain ? (
        <div className="flex items-center gap-1.5">
          <div className={`w-1.5 h-1.5 rounded-full ${connected ? 'bg-emerald-500' : 'bg-zinc-600'}`} />
          <span className="text-zinc-400 text-[10px] font-mono">{subdomain}</span>
        </div>
      ) : (
        <span className="text-zinc-600 text-[10px]">no active tunnel</span>
      )}
      {connected && (
        <span className="ml-auto text-[9px] bg-emerald-950 text-emerald-400 px-2 py-0.5 rounded font-medium">
          connected
        </span>
      )}
    </header>
  )
}
