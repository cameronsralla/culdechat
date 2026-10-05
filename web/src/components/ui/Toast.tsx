import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react';
import * as RadixToast from '@radix-ui/react-toast';
import { cn } from '@/lib/cn';
import { Icon } from './Icon';
import { Text } from './Text';

type Tone = 'info' | 'success' | 'danger';
type ToastItem = { id: number; title: string; body?: string; tone: Tone };

const ToastContext = createContext<((t: Omit<ToastItem, 'id'>) => void) | null>(null);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<ToastItem[]>([]);
  const push = useCallback((t: Omit<ToastItem, 'id'>) => {
    setItems((cur) => [...cur, { ...t, id: Date.now() + Math.random() }]);
  }, []);
  const value = useMemo(() => push, [push]);

  return (
    <ToastContext.Provider value={value}>
      <RadixToast.Provider swipeDirection="down" duration={4000}>
        {children}
        {items.map((t) => (
          <RadixToast.Root
            key={t.id}
            onOpenChange={(open) => !open && setItems((cur) => cur.filter((x) => x.id !== t.id))}
            className={cn(
              'flex items-start gap-3 rounded-md border bg-surface px-4 py-3 shadow-card',
              'data-[state=open]:animate-in data-[state=closed]:animate-out',
              t.tone === 'danger' ? 'border-danger/30' : t.tone === 'success' ? 'border-brand-line' : 'border-line-soft',
            )}
          >
            <Icon
              name={t.tone === 'danger' ? 'alert' : t.tone === 'success' ? 'check' : 'info'}
              size={20}
              className={cn('mt-0.5 shrink-0', t.tone === 'danger' ? 'text-danger' : 'text-brand')}
            />
            <div className="min-w-0 flex-1">
              <RadixToast.Title asChild>
                <Text variant="subtitle">{t.title}</Text>
              </RadixToast.Title>
              {t.body && (
                <RadixToast.Description asChild>
                  <Text variant="caption" tone="muted">
                    {t.body}
                  </Text>
                </RadixToast.Description>
              )}
            </div>
            <RadixToast.Close aria-label="Dismiss" className="text-muted hover:text-ink">
              <Icon name="close" size={18} />
            </RadixToast.Close>
          </RadixToast.Root>
        ))}
        <RadixToast.Viewport className="fixed bottom-[calc(var(--spacing-tabbar)+0.5rem+env(safe-area-inset-bottom))] left-1/2 z-50 flex w-[min(92vw,420px)] -translate-x-1/2 flex-col gap-2 md:bottom-6" />
      </RadixToast.Provider>
    </ToastContext.Provider>
  );
}

export function useToast() {
  const push = useContext(ToastContext);
  if (!push) throw new Error('useToast must be used within ToastProvider');
  return push;
}
