import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { EventDetail } from './EventDetail'
import { ToastProvider } from './ui'
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
  ResponseBody: '{"success":true}',
  ResponseMS: 42,
  Forwarded: true,
  ReplayedAt: null,
}

describe('EventDetail', () => {
  it('shows method and path in header', () => {
    render(
      <ToastProvider>
        <EventDetail event={mockEvent} onReplay={vi.fn()} />
      </ToastProvider>
    )
    expect(screen.getByText('POST')).toBeInTheDocument()
    expect(screen.getByText('/webhook/stripe')).toBeInTheDocument()
  })

  it('renders request body JSON key', () => {
    render(
      <ToastProvider>
        <EventDetail event={mockEvent} onReplay={vi.fn()} />
      </ToastProvider>
    )
    expect(screen.getByText(/"amount"/)).toBeInTheDocument()
  })

  it('shows response status badge for forwarded event', () => {
    render(
      <ToastProvider>
        <EventDetail event={mockEvent} onReplay={vi.fn()} />
      </ToastProvider>
    )
    expect(screen.getByText(/200 OK/)).toBeInTheDocument()
  })

  it('shows not-forwarded state', () => {
    const unforwarded = { ...mockEvent, Forwarded: false, ResponseStatus: 0, ResponseBody: '' }
    render(
      <ToastProvider>
        <EventDetail event={unforwarded} onReplay={vi.fn()} />
      </ToastProvider>
    )
    expect(screen.getByText('not forwarded')).toBeInTheDocument()
  })

  it('calls onReplay with event ID and target URL when replay clicked', () => {
    const onReplay = vi.fn()
    render(
      <ToastProvider>
        <EventDetail event={mockEvent} onReplay={onReplay} />
      </ToastProvider>
    )
    fireEvent.click(screen.getByRole('button', { name: /replay/i }))
    expect(onReplay).toHaveBeenCalledWith('evt-001', expect.any(String))
  })
})
