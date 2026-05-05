import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TunnelList } from './TunnelList'
import type { Tunnel } from '../types'

const tunnels: Tunnel[] = [
  { id: 't1', type: 'personal', subdomain: 'abc123', display_name: '', status: 'active', user_id: 'u1', org_id: '', active_user_id: 'u1', active_device: 'MONSTER-2352' },
  { id: 't2', type: 'personal', subdomain: 'def456', display_name: '', status: 'inactive', user_id: 'u1', org_id: '', active_user_id: '', active_device: '' },
]

test('renders tunnel subdomains', () => {
  render(<TunnelList tunnels={tunnels} selectedID={null} onSelect={() => {}} />)
  expect(screen.getByText('abc123')).toBeInTheDocument()
  expect(screen.getByText('def456')).toBeInTheDocument()
})

test('shows device name for active tunnel', () => {
  render(<TunnelList tunnels={tunnels} selectedID={null} onSelect={() => {}} />)
  expect(screen.getByText('MONSTER-2352')).toBeInTheDocument()
})

test('calls onSelect with correct tunnel', async () => {
  const onSelect = vi.fn()
  render(<TunnelList tunnels={tunnels} selectedID={null} onSelect={onSelect} />)
  await userEvent.click(screen.getByText('abc123'))
  expect(onSelect).toHaveBeenCalledWith(tunnels[0])
})
