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
    <aside className="sticky top-0 flex h-dvh w-sidebar shrink-0 flex-col border-r border-line bg-surface px-3 py-5">
      <div className="flex items-center gap-2.5 px-2">
        <Logo size={36} className="rounded-sm border border-line" />
        <div className="min-w-0">
          <Text variant="subtitle" className="truncate">
            {communityName}
          </Text>
          <Text variant="caption" as="p" tone="muted">
            Cul-de-Chat
          </Text>
        </div>
      </div>

      <nav className="mt-6 flex flex-col gap-0.5" aria-label="Primary">
        {items.map((n) => (
          <NavLink
            key={n.id}
            to={n.to}
            end={n.end}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-2.5 rounded-sm px-2.5 py-2 text-body font-medium transition-colors',
                isActive
                  ? 'bg-surface-muted text-ink shadow-[inset_2px_0_0_0_var(--color-brand)]'
                  : 'text-muted hover:bg-surface-muted/70 hover:text-ink',
              )
            }
          >
            <Icon name={n.icon} size={20} />
            {n.label}
          </NavLink>
        ))}
      </nav>

      <div className="mt-auto flex items-center gap-2.5 rounded-md border border-line bg-surface-muted p-2.5">
        <Avatar name={user?.display_name || user?.email || '?'} size="sm" />
        <div className="min-w-0 flex-1">
          <Text variant="caption" as="p" className="truncate font-semibold">
            {user?.display_name || user?.email}
          </Text>
          <Text variant="mono" as="p" tone="muted" className="truncate">
            Unit {user?.unit_number}
          </Text>
        </div>
        <button type="button" onClick={() => void signOut()} aria-label="Sign out" className="text-muted hover:text-ink">
          <Icon name="logout" size={18} />
        </button>
      </div>
    </aside>
  );
}
