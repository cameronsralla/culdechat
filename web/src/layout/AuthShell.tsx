import type { ReactNode } from 'react';
import { Logo, Text } from '@/components/ui';

/** Centered card frame for login / registration. */
export function AuthShell({ title, subtitle, children }: { title: string; subtitle?: string; children: ReactNode }) {
  return (
    <div className="flex min-h-dvh flex-col items-center justify-center bg-paper px-4 py-10">
      <div className="flex w-full max-w-auth flex-col gap-5">
        <div className="flex flex-col items-center gap-2.5 text-center">
          <Logo size={88} className="rounded-md border border-line bg-raised" />
          <Text variant="title">{title}</Text>
          {subtitle && (
            <Text variant="body" tone="muted">
              {subtitle}
            </Text>
          )}
        </div>
        <div className="rounded-md border border-line bg-surface p-4 md:p-5">{children}</div>
      </div>
    </div>
  );
}
