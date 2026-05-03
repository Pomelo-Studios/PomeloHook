import React from 'react';

export type BadgeVariant =
  | 'method-get' | 'method-post' | 'method-put' | 'method-delete' | 'method-patch'
  | 'status-2xx' | 'status-4xx' | 'status-5xx'
  | 'role-admin' | 'role-member'
  | 'online' | 'offline'
  | 'selected';

const VARIANT_STYLES: Record<BadgeVariant, React.CSSProperties> = {
  'method-get':    { background: 'rgba(76,212,161,0.12)',  color: 'var(--mint)',     border: '1px solid rgba(76,212,161,0.22)' },
  'method-post':   { background: 'rgba(167,184,250,0.12)', color: 'var(--purple)',   border: '1px solid rgba(167,184,250,0.22)' },
  'method-put':    { background: 'rgba(255,163,73,0.12)',  color: 'var(--orange)',   border: '1px solid rgba(255,163,73,0.22)' },
  'method-patch':  { background: 'rgba(255,163,73,0.08)',  color: 'var(--orange)',   border: '1px solid rgba(255,163,73,0.18)' },
  'method-delete': { background: 'rgba(255,107,107,0.12)', color: 'var(--coral)',    border: '1px solid rgba(255,107,107,0.22)' },
  'status-2xx':    { background: 'var(--ok-bg)',           color: 'var(--ok-text)',  border: '1px solid rgba(76,212,161,0.2)' },
  'status-4xx':    { background: 'rgba(255,163,73,0.10)',  color: 'var(--orange)',   border: '1px solid rgba(255,163,73,0.2)' },
  'status-5xx':    { background: 'var(--err-bg)',          color: 'var(--err-text)', border: '1px solid rgba(255,107,107,0.2)' },
  'role-admin':    { background: 'rgba(167,184,250,0.12)', color: 'var(--purple)',   border: '1px solid rgba(167,184,250,0.22)' },
  'role-member':   { background: 'rgba(255,255,255,0.05)', color: 'var(--text-2)',   border: '1px solid var(--border)' },
  'online':        { background: 'var(--ok-bg)',           color: 'var(--ok-text)',  border: '1px solid rgba(76,212,161,0.2)' },
  'offline':       { background: 'rgba(55,65,81,0.4)',     color: 'var(--text-3)',   border: '1px solid var(--border)' },
  'selected':      { background: 'rgba(255,107,107,0.12)', color: 'var(--coral)',    border: '1px solid rgba(255,107,107,0.25)' },
};

export interface BadgeProps {
  variant: BadgeVariant;
  children: React.ReactNode;
  style?: React.CSSProperties;
}

export function Badge({ variant, children, style }: BadgeProps) {
  return (
    <span style={{
      fontFamily: 'var(--font-mono)',
      fontSize: '9.5px',
      fontWeight: 700,
      padding: '2px 6px',
      borderRadius: '5px',
      textTransform: 'uppercase' as const,
      display: 'inline-block',
      flexShrink: 0,
      ...VARIANT_STYLES[variant],
      ...style,
    }}>
      {children}
    </span>
  );
}

export function methodVariant(method: string): BadgeVariant {
  switch (method.toUpperCase()) {
    case 'GET':    return 'method-get';
    case 'POST':   return 'method-post';
    case 'PUT':    return 'method-put';
    case 'PATCH':  return 'method-patch';
    case 'DELETE': return 'method-delete';
    default:       return 'method-post';
  }
}

export function statusVariant(status: number): BadgeVariant {
  if (status >= 500) return 'status-5xx';
  if (status >= 400) return 'status-4xx';
  return 'status-2xx';
}
