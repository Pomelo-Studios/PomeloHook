import { render, screen } from '@testing-library/react';
import { Badge, methodVariant, statusVariant } from '../Badge';

describe('Badge', () => {
  it('renders children', () => {
    render(<Badge variant="method-post">POST</Badge>);
    expect(screen.getByText('POST')).toBeInTheDocument();
  });

  it('applies purple colour for method-post', () => {
    render(<Badge variant="method-post">POST</Badge>);
    expect(screen.getByText('POST').style.color).toBe('var(--purple)');
  });

  it('applies mint colour for method-get', () => {
    render(<Badge variant="method-get">GET</Badge>);
    expect(screen.getByText('GET').style.color).toBe('var(--mint)');
  });

  it('applies coral colour for method-delete', () => {
    render(<Badge variant="method-delete">DELETE</Badge>);
    expect(screen.getByText('DELETE').style.color).toBe('var(--coral)');
  });

  it('applies ok-text colour for status-2xx', () => {
    render(<Badge variant="status-2xx">200</Badge>);
    expect(screen.getByText('200').style.color).toBe('var(--ok-text)');
  });

  it('applies err-text colour for status-5xx', () => {
    render(<Badge variant="status-5xx">500</Badge>);
    expect(screen.getByText('500').style.color).toBe('var(--err-text)');
  });
});

describe('methodVariant', () => {
  it.each([
    ['GET', 'method-get'],
    ['POST', 'method-post'],
    ['PUT', 'method-put'],
    ['PATCH', 'method-patch'],
    ['DELETE', 'method-delete'],
    ['get', 'method-get'],
  ])('%s → %s', (method, expected) => {
    expect(methodVariant(method)).toBe(expected);
  });
});

describe('statusVariant', () => {
  it.each([
    [200, 'status-2xx'],
    [201, 'status-2xx'],
    [404, 'status-4xx'],
    [422, 'status-4xx'],
    [500, 'status-5xx'],
    [503, 'status-5xx'],
  ])('%d → %s', (status, expected) => {
    expect(statusVariant(status)).toBe(expected);
  });
});
