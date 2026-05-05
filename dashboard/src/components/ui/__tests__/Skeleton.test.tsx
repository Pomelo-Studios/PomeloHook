import { render } from '@testing-library/react';
import { Skeleton, EventRowSkeleton, TunnelRowSkeleton } from '../Skeleton';

describe('Skeleton', () => {
  it('renders a div', () => {
    const { container } = render(<Skeleton />);
    expect(container.firstChild).toBeInTheDocument();
  });
  it('applies custom width and height', () => {
    const { container } = render(<Skeleton width={60} height={20} />);
    const el = container.firstChild as HTMLElement;
    expect(el.style.width).toBe('60px');
    expect(el.style.height).toBe('20px');
  });
});

describe('EventRowSkeleton', () => {
  it('renders without crashing', () => {
    const { container } = render(<EventRowSkeleton />);
    expect(container.firstChild).toBeInTheDocument();
  });
});

describe('TunnelRowSkeleton', () => {
  it('renders without crashing', () => {
    const { container } = render(<TunnelRowSkeleton />);
    expect(container.firstChild).toBeInTheDocument();
  });
});
