import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TunnelList } from './TunnelList'
import type { Tunnel } from '../types'

const tunnels: Tunnel[] = [
  { ID: 't1', Type: 'personal', Subdomain: 'abc123', Status: 'active', ActiveUserID: 'u1', ActiveDevice: 'MONSTER-2352' },
  { ID: 't2', Type: 'personal', Subdomain: 'def456', Status: 'inactive', ActiveUserID: '', ActiveDevice: '' },
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
