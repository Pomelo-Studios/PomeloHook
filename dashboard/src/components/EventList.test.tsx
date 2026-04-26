import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { EventList } from './EventList'
import type { WebhookEvent } from '../types'

const mockEvent: WebhookEvent = {
  ID: 'evt-001',
  TunnelID: 't1',
  ReceivedAt: '2026-04-26T10:00:00Z',
  Method: 'POST',
  Path: '/webhook/stripe',
  Headers: '{}',
  RequestBody: '{"amount":100}',
  ResponseStatus: 200,
  ResponseBody: 'ok',
  ResponseMS: 42,
  Forwarded: true,
  ReplayedAt: null,
}

describe('EventList', () => {
  it('renders event method and path', () => {
    render(<EventList events={[mockEvent]} onSelect={() => {}} selectedID={null} />)
    expect(screen.getByText('POST')).toBeInTheDocument()
    expect(screen.getByText('/webhook/stripe')).toBeInTheDocument()
  })

  it('shows status code badge for forwarded event', () => {
    render(<EventList events={[mockEvent]} onSelect={() => {}} selectedID={null} />)
    expect(screen.getByText('200')).toBeInTheDocument()
  })

  it('shows err badge for non-forwarded event', () => {
    const failed = { ...mockEvent, Forwarded: false, ResponseStatus: 0 }
    render(<EventList events={[failed]} onSelect={() => {}} selectedID={null} />)
    expect(screen.getByText('err')).toBeInTheDocument()
  })

  it('shows latency in milliseconds', () => {
    render(<EventList events={[mockEvent]} onSelect={() => {}} selectedID={null} />)
    expect(screen.getByText(/42ms/)).toBeInTheDocument()
  })
})
