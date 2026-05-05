import { createContext, useContext, useState, useCallback } from 'react';
import type { ReactNode } from 'react';

interface ToastItem {
  id: number;
  message: string;
  variant: 'success' | 'error';
}

interface ToastAPI {
  success(message: string): void;
  error(message: string): void;
}

const ToastContext = createContext<ToastAPI | null>(null);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  const add = useCallback((message: string, variant: 'success' | 'error') => {
    const id = Date.now();
    setToasts(prev => [...prev, { id, message, variant }]);
    setTimeout(() => setToasts(prev => prev.filter(t => t.id !== id)), 3000);
  }, []);

  const api: ToastAPI = {
    success: (msg) => add(msg, 'success'),
    error:   (msg) => add(msg, 'error'),
  };

  return (
    <ToastContext.Provider value={api}>
      {children}
      <div style={{
        position: 'fixed', bottom: '20px', right: '20px',
        display: 'flex', flexDirection: 'column', gap: '8px',
        zIndex: 1000, pointerEvents: 'none',
      }}>
        {toasts.map(t => (
          <div key={t.id} style={{
            display: 'flex', alignItems: 'center', gap: '10px',
            padding: '10px 16px', borderRadius: '10px',
            fontFamily: 'var(--font-sans)', fontSize: '12.5px', fontWeight: 500,
            boxShadow: '0 8px 24px rgba(0,0,0,0.4)',
            animation: 'toast-slide-in 0.25s ease',
            ...(t.variant === 'success'
              ? { background: 'var(--ok-bg)',  border: '1px solid rgba(76,212,161,0.28)',  color: 'var(--ok-text)' }
              : { background: 'var(--err-bg)', border: '1px solid rgba(255,107,107,0.28)', color: 'var(--err-text)' }),
          }}>
            <div style={{ width: '7px', height: '7px', borderRadius: '50%', background: 'currentColor', flexShrink: 0 }} />
            {t.message}
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast(): ToastAPI {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error('useToast must be used inside <ToastProvider>');
  return ctx;
}
