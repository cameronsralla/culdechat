import type { ReactNode } from 'react';
import { cn } from '@/lib/cn';
import { Header } from './Header';
import { useViewport } from './useViewport';

/**
 * A page inside AppShell. Handles the compact header, content width, and
 * bottom padding so the tab bar never covers content. Pages only supply content.
 */
export function Screen({ title, back, wide, children }: { title: string; back?: boolean; wide?: boolean; children: ReactNode }) {
  const { compact } = useViewport();
  return (
    <>
      {compact && <Header title={title} back={back} />}
      <main className={cn('mx-auto w-full px-4 py-5 md:px-8 md:py-10', wide ? 'max-w-5xl' : 'max-w-content', compact && 'pb-[calc(var(--spacing-tabbar)+1.5rem)]')}>
        <div className="flex flex-col gap-6">{children}</div>
      </main>
    </>
  );
}
