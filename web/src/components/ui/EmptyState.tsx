import type { ReactNode } from 'react';
import { Icon, type IconName } from './Icon';
import { Text } from './Text';

export function EmptyState({ icon = 'info', title, body, action }: { icon?: IconName; title: string; body?: string; action?: ReactNode }) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-md border border-dashed border-line bg-surface-muted px-6 py-10 text-center">
      <span className="flex size-12 items-center justify-center rounded-pill bg-brand-soft text-brand">
        <Icon name={icon} />
      </span>
      <Text variant="subtitle">{title}</Text>
      {body && (
        <Text variant="caption" tone="muted" className="max-w-sm">
          {body}
        </Text>
      )}
      {action}
    </div>
  );
}
