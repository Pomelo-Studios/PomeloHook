import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { JsonView } from './JsonView'

describe('JsonView', () => {
  it('renders string values with quotes', () => {
    render(<JsonView value='{"event":"payment"}' />)
    expect(screen.getByText(/"payment"/)).toBeInTheDocument()
  })

  it('renders number values', () => {
    render(<JsonView value='{"amount":99}' />)
    expect(screen.getByText('99')).toBeInTheDocument()
  })

  it('falls back to plain text for invalid JSON', () => {
    render(<JsonView value='not json' />)
    expect(screen.getByText('not json')).toBeInTheDocument()
  })

  it('renders empty object', () => {
    render(<JsonView value='{}' />)
    expect(screen.getByText('{}')).toBeInTheDocument()
  })
})
