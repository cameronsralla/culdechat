import type { ReactNode } from 'react';
import { Logo, Text } from '@/components/ui';

/** Centered card frame for login / registration. */
export function AuthShell({ title, subtitle, children }: { title: string; subtitle?: string; children: ReactNode }) {
  return (
    <div className="flex min-h-dvh flex-col items-center justify-center bg-paper px-4 py-10">
      <div className="flex w-full max-w-auth flex-col gap-6">
        <div className="flex flex-col items-center gap-3 text-center">
          <Logo size={104} className="rounded-lg shadow-card" />
          <Text variant="title">{title}</Text>
          {subtitle && (
            <Text variant="body" tone="muted">
              {subtitle}
            </Text>
          )}
        </div>
        <div className="rounded-lg border border-line-soft bg-surface p-5 shadow-card md:p-6">{children}</div>
      </div>
    </div>
  );
}
