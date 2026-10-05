import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { Button } from './Button';
import { initials } from './Avatar';
import { ListGroup, ListRow } from './ListRow';
import { TextField } from './TextField';

describe('ui kit', () => {
  it('Button shows a spinner and disables while loading', () => {
    render(<Button loading>Save</Button>);
    const btn = screen.getByRole('button');
    expect(btn).toBeDisabled();
    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.queryByText('Save')).not.toBeInTheDocument();
  });

  it('TextField wires label and error to the input', () => {
    render(<TextField label="Email" error="Required" />);
    const input = screen.getByLabelText('Email');
    expect(input).toHaveAttribute('aria-invalid', 'true');
    expect(screen.getByRole('alert')).toHaveTextContent('Required');
  });

  it('ListRow renders a link when given `to`', () => {
    render(
      <MemoryRouter>
        <ListGroup>
          <ListRow title="Unit 12" subtitle="Res Ident" to="/directory/1" />
        </ListGroup>
      </MemoryRouter>,
    );
    expect(screen.getByRole('link', { name: /Unit 12/ })).toHaveAttribute('href', '/directory/1');
  });

  it('initials', () => {
    expect(initials('Res Ident')).toBe('RI');
    expect(initials('cam')).toBe('CA');
    expect(initials('')).toBe('?');
  });
});
