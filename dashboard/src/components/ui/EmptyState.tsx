
interface EmptyStateProps {
  icon: string;
  title: string;
  subtitle?: string;
  command?: string;
}

export function EmptyState({ icon, title, subtitle, command }: EmptyStateProps) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', gap: '10px', padding: '40px', textAlign: 'center' }}>
      <div style={{ fontSize: '32px', opacity: 0.7 }}>{icon}</div>
      <div style={{ fontSize: '14px', fontWeight: 700, color: 'var(--text)' }}>{title}</div>
      {subtitle && (
        <div style={{ fontSize: '12.5px', color: 'var(--text-2)', lineHeight: 1.5 }}>{subtitle}</div>
      )}
      {command && (
        <div style={{ marginTop: '4px', background: 'var(--code-bg)', border: '1px solid var(--code-border)', borderRadius: '9px', padding: '8px 14px', fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--ok-text)' }}>
          {command}
        </div>
      )}
    </div>
  );
}
