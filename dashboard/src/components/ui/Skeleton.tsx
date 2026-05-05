import React from 'react';

interface SkeletonProps {
  width?: string | number;
  height?: number;
  style?: React.CSSProperties;
}

export function Skeleton({ width = '100%', height = 12, style }: SkeletonProps) {
  return (
    <div style={{
      background: 'var(--surface2)',
      borderRadius: '4px',
      height: typeof height === 'number' ? `${height}px` : height,
      width:  typeof width  === 'number' ? `${width}px`  : width,
      animation: 'skeleton-pulse 1.6s ease-in-out infinite',
      ...style,
    }} />
  );
}

export function EventRowSkeleton() {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '9px', padding: '9px 14px', borderBottom: '1px solid var(--border)', opacity: 0.5 }}>
      <Skeleton width={38} height={18} style={{ borderRadius: '5px', flexShrink: 0 }} />
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '5px' }}>
        <Skeleton height={12} width="75%" />
        <Skeleton height={10} width="45%" />
      </div>
    </div>
  );
}

export function TunnelRowSkeleton() {
  return (
    <div style={{ padding: '10px 14px', borderBottom: '1px solid var(--border)', opacity: 0.5, display: 'flex', flexDirection: 'column', gap: '6px' }}>
      <Skeleton height={12} width="60%" />
      <Skeleton height={10} width="80%" />
      <Skeleton height={10} width="40%" />
    </div>
  );
}

export function TableRowSkeleton({ cols = 4 }: { cols?: number }) {
  return (
    <tr>
      {Array.from({ length: cols }).map((_, i) => (
        <td key={i} style={{ padding: '10px 12px' }}>
          <Skeleton height={12} width={i === 0 ? '80%' : '60%'} />
        </td>
      ))}
    </tr>
  );
}
