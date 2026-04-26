import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { EventDetail } from './EventDetail'
import type { WebhookEvent } from '../types'

const mockEvent: WebhookEvent = {
  ID: 'evt-001',
  TunnelID: 't1',
  ReceivedAt: '2026-04-26T10:00:00Z',
  Method: 'POST',
  Path: '/webhook/stripe',
  Headers: '{"Content-Type":["application/json"]}',
  RequestBody: '{"amount":100}',
  ResponseStatus: 200,
  ResponseBody: 'ok',
  ResponseMS: 42,
  Forwarded: true,
  ReplayedAt: null,
}

describe('EventDetail', () => {
  it('shows request body', () => {
    render(<EventDetail event={mockEvent} onReplay={vi.fn()} />)
    expect(screen.getByText(/amount/)).toBeInTheDocument()
  })

  it('calls onReplay with target URL when replay clicked', () => {
    const onReplay = vi.fn()
    render(<EventDetail event={mockEvent} onReplay={onReplay} />)
    fireEvent.click(screen.getByRole('button', { name: /replay/i }))
    expect(onReplay).toHaveBeenCalledWith('evt-001', expect.any(String))
  })
})
