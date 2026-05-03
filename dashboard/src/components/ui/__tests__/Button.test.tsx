import { render, screen } from '@testing-library/react';
import { Button } from '../Button';

describe('Button', () => {
  it('renders children', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByRole('button', { name: 'Click me' })).toBeInTheDocument();
  });

  it('applies primary background token', () => {
    render(<Button variant="primary">Save</Button>);
    const btn = screen.getByRole('button');
    expect(btn.style.background).toBe('var(--coral)');
  });

  it('applies ghost border token', () => {
    render(<Button variant="ghost">Cancel</Button>);
    const btn = screen.getByRole('button');
    expect(btn.style.border).toBe('1px solid var(--border)');
  });

  it('applies danger colour tokens', () => {
    render(<Button variant="danger">Delete</Button>);
    const btn = screen.getByRole('button');
    expect(btn.style.color).toBe('var(--err-text)');
  });

  it('applies sm size padding', () => {
    render(<Button size="sm">Small</Button>);
    const btn = screen.getByRole('button');
    expect(btn.style.fontSize).toBe('11.5px');
  });

  it('forwards additional props', () => {
    const onClick = vi.fn();
    render(<Button onClick={onClick}>Click</Button>);
    screen.getByRole('button').click();
    expect(onClick).toHaveBeenCalledOnce();
  });

  it('can be disabled', () => {
    render(<Button disabled>Disabled</Button>);
    expect(screen.getByRole('button')).toBeDisabled();
  });
});
