import { render, screen, act } from '@testing-library/react';
import { ToastProvider, useToast } from '../Toast';
import { vi } from 'vitest';

function TestComponent({ variant }: { variant: 'success' | 'error' }) {
  const toast = useToast();
  return <button onClick={() => toast[variant]('Test message')}>show</button>;
}

describe('Toast', () => {
  it('shows success toast on call', async () => {
    render(<ToastProvider><TestComponent variant="success" /></ToastProvider>);
    act(() => { screen.getByRole('button').click(); });
    expect(await screen.findByText('Test message')).toBeInTheDocument();
  });

  it('shows error toast on call', async () => {
    render(<ToastProvider><TestComponent variant="error" /></ToastProvider>);
    act(() => { screen.getByRole('button').click(); });
    expect(await screen.findByText('Test message')).toBeInTheDocument();
  });

  it('throws when useToast used outside provider', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    expect(() => render(<TestComponent variant="success" />)).toThrow();
    spy.mockRestore();
  });
});
