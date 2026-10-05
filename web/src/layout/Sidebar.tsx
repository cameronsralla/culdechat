import { NavLink } from 'react-router';
import { useAuth } from '@/auth/AuthContext';
import { Avatar, Icon, Logo, Text } from '@/components/ui';
import { cn } from '@/lib/cn';
import { navFor } from './nav';

/** Wide-screen primary navigation. Owns its own width token and styling. */
export function Sidebar({ communityName }: { communityName: string }) {
  const { user, signOut } = useAuth();
  const items = navFor(!!user?.is_admin);

  return (
    <aside className="sticky top-0 flex h-dvh w-sidebar shrink-0 flex-col border-r border-line-soft bg-surface px-4 py-6">
      <div className="flex items-center gap-3 px-2">
        <Logo size={42} />
        <div className="min-w-0">
          <Text variant="subtitle" className="truncate">
            {communityName}
          </Text>
          <Text variant="caption" as="p" tone="muted">
            Cul-de-Chat
          </Text>
        </div>
      </div>

      <nav className="mt-8 flex flex-col gap-1" aria-label="Primary">
        {items.map((n) => (
          <NavLink
            key={n.id}
            to={n.to}
            end={n.end}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 rounded-sm px-3 py-2.5 text-body font-semibold transition-colors',
                isActive ? 'bg-brand-wash text-brand-dark' : 'text-muted hover:bg-surface-muted hover:text-ink',
              )
            }
          >
            <Icon name={n.icon} />
            {n.label}
          </NavLink>
        ))}
      </nav>

      <div className="mt-auto flex items-center gap-3 rounded-md border border-line-soft bg-surface-muted p-3">
        <Avatar name={user?.display_name || user?.email || '?'} size="sm" />
        <div className="min-w-0 flex-1">
          <Text variant="caption" as="p" className="truncate font-semibold">
            {user?.display_name || user?.email}
          </Text>
          <Text variant="caption" as="p" tone="muted" className="truncate">
            Unit {user?.unit_number}
          </Text>
        </div>
        <button type="button" onClick={() => void signOut()} aria-label="Sign out" className="text-muted hover:text-ink">
          <Icon name="logout" size={20} />
        </button>
      </div>
    </aside>
  );
}
