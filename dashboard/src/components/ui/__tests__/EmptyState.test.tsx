import { render, screen } from '@testing-library/react';
import { EmptyState } from '../EmptyState';

describe('EmptyState', () => {
  it('renders icon and title', () => {
    render(<EmptyState icon="⚡" title="No events yet" />);
    expect(screen.getByText('⚡')).toBeInTheDocument();
    expect(screen.getByText('No events yet')).toBeInTheDocument();
  });
  it('renders optional subtitle', () => {
    render(<EmptyState icon="⚡" title="Empty" subtitle="Connect a tunnel" />);
    expect(screen.getByText('Connect a tunnel')).toBeInTheDocument();
  });
  it('renders optional command', () => {
    render(<EmptyState icon="⚡" title="Empty" command="$ pomelo-hook connect" />);
    expect(screen.getByText('$ pomelo-hook connect')).toBeInTheDocument();
  });
});
