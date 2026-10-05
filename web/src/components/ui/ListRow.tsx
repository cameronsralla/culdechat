import type { HTMLAttributes, ReactNode } from 'react';
import { Link } from 'react-router';
import { cn } from '@/lib/cn';
import { Icon } from './Icon';
import { Text } from './Text';

/** A bordered group of rows; rows get dividers automatically. */
export function ListGroup({ className, ...rest }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('overflow-hidden rounded-md border border-line bg-surface divide-y divide-line-soft', className)} {...rest} />;
}

type RowProps = {
  leading?: ReactNode;
  title: ReactNode;
  subtitle?: ReactNode;
  trailing?: ReactNode;
  to?: string;
  onClick?: () => void;
  className?: string;
};

export function ListRow({ leading, title, subtitle, trailing, to, onClick, className }: RowProps) {
  const interactive = !!(to || onClick);
  const body = (
    <>
      {leading && <div className="shrink-0">{leading}</div>}
      <div className="min-w-0 flex-1 flex flex-col gap-0.5">
        <Text variant="subtitle" className="truncate">
          {title}
        </Text>
        {subtitle && (
          <Text variant="caption" as="p" tone="muted" className="truncate">
            {subtitle}
          </Text>
        )}
      </div>
      {trailing ?? (interactive && <Icon name="chevronRight" className="text-muted" />)}
    </>
  );
  const cls = cn(
    'flex w-full items-center gap-3 px-3 py-3 text-left md:px-4 md:py-2.5 min-h-row md:min-h-row-dense',
    interactive && 'hover:bg-surface-muted active:bg-overlay transition-colors',
    className,
  );
  if (to) return <Link to={to} className={cls}>{body}</Link>;
  if (onClick) return <button type="button" onClick={onClick} className={cls}>{body}</button>;
  return <div className={cls}>{body}</div>;
}
