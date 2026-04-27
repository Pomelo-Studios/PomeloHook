import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { DatabasePanel } from './DatabasePanel'
import { api } from '../../api/client'

vi.mock('../../api/client', () => ({
  api: {
    admin: {
      listTables: vi.fn(),
      getTableRows: vi.fn(),
    },
  },
}))

beforeEach(() => {
  vi.mocked(api.admin.listTables).mockResolvedValue([{ name: 'users', row_count: 3 }])
  vi.mocked(api.admin.getTableRows).mockResolvedValue({ columns: ['id', 'email'], rows: [['u1', 'a@b.com']] })
})

describe('DatabasePanel', () => {
  it('renders table list after loading', async () => {
    render(<DatabasePanel apiKey="" />)
    await waitFor(() => expect(screen.getByText('users')).toBeInTheDocument())
    expect(screen.getByText('3')).toBeInTheDocument()
  })
})
