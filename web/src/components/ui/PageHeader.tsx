import type { ReactNode } from 'react';
import { Badge } from './Badge';
import { Text } from './Text';

/** Page title block with optional eyebrow chip, description, and actions. */
export function PageHeader({ eyebrow, title, description, actions }: { eyebrow?: string; title: string; description?: string; actions?: ReactNode }) {
  return (
    <div className="flex flex-wrap items-end justify-between gap-3">
      <div className="flex flex-col gap-1.5">
        {eyebrow && <Badge className="self-start">{eyebrow}</Badge>}
        <Text variant="title">{title}</Text>
        {description && (
          <Text variant="body" tone="muted">
            {description}
          </Text>
        )}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  );
}
