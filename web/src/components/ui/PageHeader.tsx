import type { ReactNode } from 'react';
import { Text } from './Text';

/** Page title block with optional description and actions. */
export function PageHeader({ title, description, actions }: { title: string; description?: ReactNode; actions?: ReactNode }) {
  return (
    <div className="flex flex-wrap items-end justify-between gap-3">
      <div className="flex flex-col gap-1">
        <Text variant="title">{title}</Text>
        {description && (
          <Text variant="body" tone="muted" as="div">
            {description}
          </Text>
        )}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  );
}
