import { Link } from 'react-router';
import { Icon, type IconName } from './Icon';

/**
 * Floating action button. On compact screens it sits above the tab bar;
 * on wide screens it tucks into the bottom-right of the content column.
 */
export function Fab({ to, label, icon = 'plus' }: { to: string; label: string; icon?: IconName }) {
  return (
    <Link
      to={to}
      aria-label={label}
      className="fixed bottom-[calc(var(--spacing-tabbar)+1rem+env(safe-area-inset-bottom))] right-4 z-20 flex size-14 items-center justify-center rounded-pill bg-brand text-white shadow-raised hover:bg-brand-dark md:bottom-8 md:right-8"
    >
      <Icon name={icon} size={26} />
    </Link>
  );
}
