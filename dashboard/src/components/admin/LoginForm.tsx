import { useState } from 'react'
import { api } from '../../api/client'
import { HookIcon } from '../HookIcon'

interface Props {
  onLogin: (apiKey: string) => void
}

export function LoginForm({ onLogin }: Props) {
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const data = await api.login(email)
      onLogin(data.api_key)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex items-center justify-center h-screen" style={{ background: 'var(--bg)' }}>
      <div className="rounded-xl p-8 w-80" style={{ background: 'var(--surface)', border: '1px solid var(--border)' }}>
        <div className="flex items-center gap-2 mb-6">
          <HookIcon size={28} />
          <span className="font-extrabold text-[15px] tracking-tight" style={{ color: 'var(--text-primary)' }}>
            PomeloHook Admin
          </span>
        </div>
        <form onSubmit={handleSubmit} className="flex flex-col gap-3">
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            required
            className="rounded-lg px-3 py-2 text-xs font-mono outline-none"
            style={{ background: 'var(--bg)', border: '1px solid var(--border)', color: 'var(--text-primary)' }}
          />
          {error && <p className="text-xs" style={{ color: 'var(--err-text)' }}>{error}</p>}
          <button
            type="submit"
            disabled={loading}
            className="bg-coral hover:opacity-90 text-white text-xs py-2 rounded-lg font-semibold disabled:opacity-50 transition-opacity"
          >
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  )
}
