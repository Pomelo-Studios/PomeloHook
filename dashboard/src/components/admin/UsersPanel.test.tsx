import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { UsersPanel } from './UsersPanel'
import { api } from '../../api/client'

vi.mock('../../api/client', () => ({
  api: {
    admin: {
      listUsers: vi.fn(),
    },
  },
}))

const mockUsers = [
  { ID: 'u1', OrgID: 'org1', Email: 'alice@a.com', Name: 'Alice', APIKey: 'ph_abc123', Role: 'admin' },
]

beforeEach(() => {
  vi.mocked(api.admin.listUsers).mockResolvedValue(mockUsers)
})

describe('UsersPanel', () => {
  it('renders user rows after loading', async () => {
    render(<UsersPanel apiKey="" />)
    await waitFor(() => expect(screen.getByText('Alice')).toBeInTheDocument())
    expect(screen.getByText('alice@a.com')).toBeInTheDocument()
    expect(screen.getByText('admin')).toBeInTheDocument()
  })
})
