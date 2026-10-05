import type { ReactNode } from 'react';
import { Icon, type IconName } from './Icon';
import { Text } from './Text';

export function EmptyState({ icon = 'info', title, body, action }: { icon?: IconName; title: string; body?: string; action?: ReactNode }) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-md border border-dashed border-line bg-surface-muted/60 px-6 py-8 text-center">
      <span className="flex size-10 items-center justify-center rounded-sm border border-line bg-raised text-muted">
        <Icon name={icon} size={20} />
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
