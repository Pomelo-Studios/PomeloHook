import React from 'react';

export interface CardProps {
  children: React.ReactNode;
  selected?: boolean;
  style?: React.CSSProperties;
  onClick?: () => void;
}

export function Card({ children, selected = false, style, onClick }: CardProps) {
  return (
    <div
      onClick={onClick}
      style={{
        background: selected ? 'var(--selected-bg)' : 'var(--surface)',
        border: '1px solid var(--border)',
        borderLeft: selected ? '2px solid var(--coral)' : '1px solid var(--border)',
        borderRadius: '12px',
        transition: 'border-color 0.2s',
        cursor: onClick ? 'pointer' : undefined,
        ...style,
      }}
    >
      {children}
    </div>
  );
}
