import React from 'react';

export type ButtonVariant = 'primary' | 'ghost' | 'danger';
export type ButtonSize = 'md' | 'sm';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
}

const VARIANT_STYLES: Record<ButtonVariant, React.CSSProperties> = {
  primary: { background: 'var(--coral)', color: '#fff', border: 'none' },
  ghost:   { background: 'transparent', border: '1px solid var(--border)', color: 'var(--text)' },
  danger:  { background: 'var(--err-bg)', color: 'var(--err-text)', border: '1px solid var(--selected-border)' },
};

const SIZE_STYLES: Record<ButtonSize, React.CSSProperties> = {
  md: { fontSize: '12.5px', padding: '7px 16px', borderRadius: '8px' },
  sm: { fontSize: '11.5px', padding: '4px 10px', borderRadius: '7px' },
};

export function Button({ variant = 'ghost', size = 'md', style, ...props }: ButtonProps) {
  return (
    <button
      style={{
        fontFamily: 'var(--font-sans)',
        fontWeight: 600,
        cursor: 'pointer',
        transition: 'opacity 0.2s, transform 0.15s',
        lineHeight: 1,
        ...VARIANT_STYLES[variant],
        ...SIZE_STYLES[size],
        ...style,
      }}
      {...props}
    />
  );
}
