import { useNavigate } from 'react-router';
import { Icon, Logo, Text } from '@/components/ui';

/** Compact-screen top bar. Shows a back button on sub-pages. */
export function Header({ title, back }: { title: string; back?: boolean }) {
  const nav = useNavigate();
  return (
    <header className="sticky top-0 z-20 flex h-12 items-center gap-2.5 border-b border-line bg-raised/95 px-3 backdrop-blur pt-[env(safe-area-inset-top)] md:h-14">
      {back ? (
        <button type="button" onClick={() => nav(-1)} aria-label="Back" className="flex size-10 items-center justify-center rounded-sm text-brand hover:bg-brand-wash">
          <Icon name="chevronLeft" />
        </button>
      ) : (
        <Logo size={28} className="rounded-sm border border-line" />
      )}
      <Text variant="subtitle" className="truncate">
        {title}
      </Text>
    </header>
  );
}
