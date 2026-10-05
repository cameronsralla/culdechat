import { NavLink } from 'react-router';
import { useAuth } from '@/auth/AuthContext';
import { Icon } from '@/components/ui';
import { cn } from '@/lib/cn';
import { navFor } from './nav';

/** Compact-screen primary navigation, pinned to the bottom with safe-area padding. */
export function TabBar() {
  const { user } = useAuth();
  const items = navFor(!!user?.is_admin);

  return (
    <nav
      aria-label="Primary"
      className="fixed inset-x-0 bottom-0 z-30 flex h-[calc(var(--spacing-tabbar)+env(safe-area-inset-bottom))] border-t border-line bg-raised pb-[env(safe-area-inset-bottom)]"
    >
      {items.map((n) => (
        <NavLink
          key={n.id}
          to={n.to}
          end={n.end}
          className={({ isActive }) =>
            cn(
              'flex flex-1 flex-col items-center justify-center gap-0.5 text-caption font-medium',
              isActive ? 'text-brand' : 'text-muted',
            )
          }
        >
          <Icon name={n.icon} />
          {n.label}
        </NavLink>
      ))}
    </nav>
  );
}
