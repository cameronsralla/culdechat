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
      className="fixed bottom-[calc(var(--spacing-tabbar)+1rem+env(safe-area-inset-bottom))] right-4 z-20 flex size-12 items-center justify-center rounded-pill bg-brand text-white shadow-raised transition-transform active:scale-[0.96] hover:bg-brand-dark md:bottom-8 md:right-8 md:size-11"
    >
      <Icon name={icon} size={26} />
    </Link>
  );
}
