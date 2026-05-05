import React from 'react';

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {}

export function Input({ style, onFocus, onBlur, ...props }: InputProps) {
  return (
    <input
      style={{
        width: '100%',
        background: 'var(--surface2)',
        border: '1px solid var(--border)',
        borderRadius: '9px',
        padding: '9px 12px',
        fontFamily: 'var(--font-sans)',
        fontSize: '13px',
        color: 'var(--text)',
        outline: 'none',
        transition: 'border-color 0.2s',
        ...style,
      }}
      onFocus={e => {
        e.currentTarget.style.borderColor = 'var(--coral)';
        onFocus?.(e);
      }}
      onBlur={e => {
        e.currentTarget.style.borderColor = 'var(--border)';
        onBlur?.(e);
      }}
      {...props}
    />
  );
}
